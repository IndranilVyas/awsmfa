package awssession

import (
	"fmt"
	"runtime"
	"os"
	"strings"
	"time"

	"gopkg.in/ini.v1"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/defaults"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

//AwsSession is used to store AWS Session Config for generating STS token.
type AwsSession struct {
	Profile  string
	Duration string
	Token    string
	HomeDir  string
	Eval bool
}

//CredentialResult is caching STS session credentials and then writing them to ~/.aws/credentials file
type CredentialResult struct {
	AccessKey    string `ini:"aws_access_key_id"`
	SecretKey    string `ini:"aws_secret_access_key"`
	SessionToken string `ini:"aws_session_token"`
}

func(c *CredentialResult) save(profile ,homeDir string) error{
	fileName := homeDir + "/.aws/credentials"
	credsFile, err := ini.Load(fileName)
	checkErrorAndExit(err, "Error Loading credentials from "+fileName)
	mfaProfileName := profile + "_mfa"

	section, err := credsFile.GetSection(mfaProfileName)

	if  err != nil {

		section, err := credsFile.NewSection(mfaProfileName)

		checkErrorAndExit(err,"unable to create new section")
		section.ReflectFrom(c)
		err = credsFile.SaveTo(fileName)

		return err
	}

	err = section.ReflectFrom(c)
	checkErrorAndExit(err,"unable add to new keys")
	
	err = credsFile.SaveTo(fileName)

	return err


	

}

//New create new AwsSession object
func New() *AwsSession {
	return new(AwsSession)
}
func (sess *AwsSession) loadToken() (string, error) {
	if sess.Token == "" {
		return stscreds.StdinTokenProvider()
	}
	return sess.Token, nil
}

func checkErrorAndExit(err error, message string) {
	if err != nil {
		fmt.Printf("%s \nError: %s \n", message, err.Error())
		os.Exit(1)
	}
}

//AssumeRoleFromConfig allows to assume IAM role with MFA that are defined in ~/.aws/config defined
func (sess *AwsSession) AssumeRoleFromConfig() {

	roleDuration, err := time.ParseDuration(sess.Duration)
	var sessOpts session.Options
	sessOpts.Profile = sess.Profile
	sessOpts.SharedConfigState = session.SharedConfigEnable
	sessOpts.AssumeRoleTokenProvider = sess.loadToken //stsvalue.StdinTokenProvider
	sessOpts.AssumeRoleDuration = roleDuration

	newSession, err := session.NewSessionWithOptions(sessOpts)
	checkErrorAndExit(err, "Failed to Created Session")

	creds := newSession.Config.Credentials

	c, err := creds.Get()
	checkErrorAndExit(err, "Unable to retrieve credentials from the session")
	values := &CredentialResult{
		AccessKey:    c.AccessKeyID,
		SecretKey:    c.SecretAccessKey,
		SessionToken: c.SessionToken,
	}
	


	//Save Credentials to File.
	err = values.save(sess.Profile,sess.HomeDir)
	checkErrorAndExit(err,"Unable to save credentials to the file")
	printOrExport(values, sess.Eval)
}
func generateMFASerialNumber(sess *session.Session) string {
	svc := sts.New(sess)
	input := &sts.GetCallerIdentityInput{}

	callerInfo, err := svc.GetCallerIdentity(input)
	checkErrorAndExit(err, "unable to retrieve Get Caller Identity")
	arn := callerInfo.Arn
	mfaserial := strings.ReplaceAll(*arn, "user", "mfa")
	return mfaserial
}
//GetUserSession implements GetSessionToken with MFA for the AWS IAM User.
func (sess *AwsSession) GetUserSession() {

	var sessOpts session.Options

	sessOpts.Config.Region = aws.String("us-east-1")
	sessOpts.Config.Credentials = credentials.NewSharedCredentials(defaults.SharedCredentialsFilename(), sess.Profile)
	t, err := time.ParseDuration(sess.Duration)
	seconds := t.Seconds()
	newSession, err := session.NewSessionWithOptions(sessOpts)
	mfaSerial := generateMFASerialNumber(newSession)
	token, err := sess.loadToken()
	checkErrorAndExit(err, "Token Not Found")
	params := &sts.GetSessionTokenInput{
		DurationSeconds: aws.Int64(int64(seconds)),
		TokenCode:       aws.String(token),
		SerialNumber:    aws.String(mfaSerial),
	}

	checkErrorAndExit(err, "Failed to Created Session")
	stsSession := sts.New(newSession)
	stsOutput, err := stsSession.GetSessionToken(params)
	checkErrorAndExit(err, "Could Not create user session")

	creds := stsOutput.Credentials
	values := &CredentialResult{
		AccessKey:    *creds.AccessKeyId,
		SecretKey:    *creds.SecretAccessKey,
		SessionToken: *creds.SessionToken,
	}

	//Save Credentials to File.
	err = values.save(sess.Profile,sess.HomeDir)
	checkErrorAndExit(err,"Unable to save credentials to the file")
	printOrExport(values, sess.Eval)
	
}


func printOrExport (value *CredentialResult, eval bool){
	
	format := "bash"
	if runtime.GOOS == "windows"{
		format = "cmdpwshell"
	}
	
	switch format {
	case "bash":
		
		if eval {

			fmt.Printf("export AWS_ACCESS_KEY_ID=%s\n", value.AccessKey)
			fmt.Printf("export AWS_SECRET_ACCESS_KEY=%s\n", value.SecretKey)
			fmt.Printf("export AWS_SESSION_TOKEN=%s\n", value.SessionToken)
			fmt.Printf("echo \"Following Environment Variables are set\"\n")
			fmt.Printf("echo \"AWS_ACCESS_KEY_ID\nAWS_SECRET_ACCESS_KEY\nAWS_SESSION_TOKEN\"")

		}else {

			fmt.Println("Run below in Terminal")
			fmt.Printf("export AWS_ACCESS_KEY_ID=%s\n", value.AccessKey)
			fmt.Printf("export AWS_SECRET_ACCESS_KEY=%s\n", value.SecretKey)
			fmt.Printf("export AWS_SESSION_TOKEN=%s\n", value.SessionToken)
		}

		
	case "cmdpwshell":
		fmt.Printf("AWS_ACCESS_KEY_ID=\"%s\";", value.AccessKey)
		fmt.Printf("AWS_SECRET_ACCESS_KEY=\"%s\";", value.SecretKey)
		fmt.Printf("AWS_SESSION_TOKEN=\"%s\";", value.SessionToken)
		fmt.Println("Run below Windows Command Prompt")
		fmt.Printf("setx AWS_ACCESS_KEY_ID=\"%s\";", value.AccessKey)
		fmt.Printf("setx AWS_SECRET_ACCESS_KEY=\"%s\";", value.SecretKey)
		fmt.Printf("setx AWS_SESSION_TOKEN=\"%s\";", value.SessionToken)
		fmt.Println("Run below Windows PowerShell")
		fmt.Printf("[Environment]::SetEnvironmentVariable(\"AWS_ACCESS_KEY_ID\", \"%s\", \"User\");", value.AccessKey)
		fmt.Printf("[Environment]::SetEnvironmentVariable(\"AWS_SECRET_ACCESS_KEY\", \"%s\", \"User\");", value.SecretKey)
		fmt.Printf("[Environment]::SetEnvironmentVariable(\"AWS_SESSION_TOKEN\", \"%s\", \"User\");", value.SessionToken)	
		
	default:
		fmt.Printf("%s is an unrecognized option", format)
	}
	
}