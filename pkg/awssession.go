package awssession

import (
	"fmt"
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

const (
	awsAccessKey    string = "aws_access_key_id"
	awsSecretKey    string = "aws_secret_access_key"
	awsSessionToken string = "aws_session_token"
)

type awsSession struct {
	Profile  string
	Duration string
	Token    string
	HomeDir  string
}

type credentialResult struct {
	accessKey    string
	secretKey    string
	sessionToken string
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

}

func updateCredentialsFile(sess *awsSession, value *credentialResult) {

	filePath := sess.HomeDir + "/.aws/credentials"
	fmt.Printf("File Path is %s", filePath)
	credsFile, err := ini.Load(filePath)
	checkErrorAndExit(err, "Error Loading credentials from "+filePath)
	mfaProfileName := sess.Profile + "_mfa"

	mfaSection, err := credsFile.GetSection(mfaProfileName)

	if err != nil {
		fmt.Printf("Credentils Not Found...Creating Section for %s \n", mfaProfileName)

		mfaSection, err := credsFile.NewSection(mfaProfileName)
		checkErrorAndExit(err, "Failed to Created New Section")
		_, err = mfaSection.NewKey(awsAccessKey, value.accessKey)
		checkErrorAndExit(err, "Failed to add:"+awsAccessKey)
		_, err = mfaSection.NewKey(awsSecretKey, value.secretKey)
		checkErrorAndExit(err, "Failed to add:"+awsAccessKey)
		_, err = mfaSection.NewKey(awsSessionToken, value.sessionToken)
		checkErrorAndExit(err, "Failed to add:"+awsAccessKey)
		err = credsFile.SaveTo(filePath)
		checkErrorAndExit(err, "Failed to Save credentials")
		fmt.Println("New Credentials added to Credentials File")
	} else {
		fmt.Println("Previous Credentials Found....Updating Them")
		mfaSection.Key(awsAccessKey).SetValue(value.accessKey)
		mfaSection.Key(awsSecretKey).SetValue(value.secretKey)
		mfaSection.Key(awsSessionToken).SetValue(value.sessionToken)
		credsFile.SaveTo(filePath)
		checkErrorAndExit(err, "Failed to Save credentials")
		fmt.Println("Credentials File updated with New Credentials")
	}

}
