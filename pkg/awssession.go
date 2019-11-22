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

type awsSession struct {
	Profile  string
	Duration string
	Token    string
	HomeDir  string
}

type credentialResult struct {
	accessKey    string `ini:"aws_access_key_id"`
	secretKey    string `ini:"aws_secret_access_key"`
	sessionToken string `ini:"aws_session_token"`
}

//New create new awsSession object
func New() *awsSession {
	return new(awsSession)
}
func (sess *awsSession) loadToken() (string, error) {
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

func (sess *awsSession) AssumeRoleFromConfig() {

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
	values := &credentialResult{
		accessKey:    c.AccessKeyID,
		secretKey:    c.SecretAccessKey,
		sessionToken: c.SessionToken,
	}
	

	updateCredentialsFile(sess, values)
	export(values)
}
func generateMFASerialNumber(sess *session.Session) string {
	svc := sts.New(sess)
	input := &sts.GetCallerIdentityInput{}

	callerInfo, err := svc.GetCallerIdentity(input)
	checkErrorAndExit(err, "unable to retrieve Get Caller Identity")
	arn := callerInfo.Arn
	mfaserial := strings.ReplaceAll(*arn, "user", "mfa")
	fmt.Printf("User's MFA Serial is : %s\n", mfaserial)
	return mfaserial
}
func (sess *awsSession) GetUserSession() {

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
	values := &credentialResult{
		accessKey:    *creds.AccessKeyId,
		secretKey:    *creds.SecretAccessKey,
		sessionToken: *creds.SessionToken,
	}

	updateCredentialsFile(sess, values)
	export(values)
}

func updateCredentialsFile(sess *awsSession, value *credentialResult) {

	filePath := sess.HomeDir + "/.aws/credentials"
	fmt.Printf("File Path is %s\n", filePath)
	credsFile, err := ini.Load(filePath)
	checkErrorAndExit(err, "Error Loading credentials from "+filePath)
	mfaProfileName := sess.Profile + "_mfa"

	section, err := credsFile.GetSection(mfaProfileName)

	if err != nil {
		section, err = credsFile.NewSection(mfaProfileName)
		checkErrorAndExit(err, "Unable to create new section for "+mfaProfileName)
		section.NewKey("aws_access_key_id", value.accessKey)
		section.NewKey("aws_secret_access_key", value.secretKey)
		section.NewKey("aws_session_token", value.sessionToken)
		
	}
	section.Key("aws_access_key_id").SetValue(value.accessKey)
	section.Key("aws_secret_access_key").SetValue(value.secretKey)
	section.Key("aws_session_token").SetValue(value.sessionToken)
	err = credsFile.SaveTo(filePath)
	checkErrorAndExit(err, "Failed to Save credentials")
	fmt.Printf("Temporary Credentials for Profile:%s\n",mfaProfileName)

}

func export (value *credentialResult){
	
	format := "bash"
	if runtime.GOOS == "windows"{
		format = "cmdpwshell"
	}
	
	switch format {
	case "bash":
		
		fmt.Printf("AWS_ACCESS_KEY_ID=%s\n", value.accessKey)
		fmt.Printf("AWS_SECRET_ACCESS_KEY=%s\n", value.secretKey)
		fmt.Printf("AWS_SESSION_TOKEN=%s\n", value.sessionToken)
		fmt.Println("Run below in Terminal")
		fmt.Printf("export AWS_ACCESS_KEY_ID=%s\n", value.accessKey)
		fmt.Printf("export AWS_SECRET_ACCESS_KEY=%s\n", value.secretKey)
		fmt.Printf("export AWS_SESSION_TOKEN=%s\n", value.sessionToken)
		
	case "cmdpwshell":
		fmt.Printf("AWS_ACCESS_KEY_ID=\"%s\";", value.accessKey)
		fmt.Printf("AWS_SECRET_ACCESS_KEY=\"%s\";", value.secretKey)
		fmt.Printf("AWS_SESSION_TOKEN=\"%s\";", value.sessionToken)
		fmt.Println("Run below Windows Command Prompt")
		fmt.Printf("setx AWS_ACCESS_KEY_ID=\"%s\";", value.accessKey)
		fmt.Printf("setx AWS_SECRET_ACCESS_KEY=\"%s\";", value.secretKey)
		fmt.Printf("setx AWS_SESSION_TOKEN=\"%s\";", value.sessionToken)
		fmt.Println("Run below Windows PowerShell")
		fmt.Printf("[Environment]::SetEnvironmentVariable(\"AWS_ACCESS_KEY_ID\", \"%s\", \"User\");", value.accessKey)
		fmt.Printf("[Environment]::SetEnvironmentVariable(\"AWS_SECRET_ACCESS_KEY\", \"%s\", \"User\");", value.secretKey)
		fmt.Printf("[Environment]::SetEnvironmentVariable(\"AWS_SESSION_TOKEN\", \"%s\", \"User\");", value.sessionToken)	
		
	default:
		fmt.Printf("%s is an unrecognized option", format)
	}
	
}