package awssession

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/ini.v1"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/defaults"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
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

func (creds *credentialResult) setAssumeRoleCreds(awsCreds *credentials.Credentials) {

	value, err := awsCreds.Get()
	checkErrorAndExit(err, "Unable to retrieve Credentials")
	creds.accessKey = value.AccessKeyID
	creds.secretKey = value.SecretAccessKey
	creds.sessionToken = value.SessionToken
}
func (creds *credentialResult) setUserSessionCreds(awsCreds *sts.Credentials) {
	creds.accessKey = *awsCreds.AccessKeyId
	creds.secretKey = *awsCreds.SecretAccessKey
	creds.sessionToken = *awsCreds.SessionToken

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

func (sess *awsSession) Save() {

	var credentialsValue = &credentialResult{}
	roleDuration, err := time.ParseDuration(sess.Duration)
	var sessOpts session.Options
	sessOpts.Profile = sess.Profile
	sessOpts.SharedConfigState = session.SharedConfigEnable
	sessOpts.AssumeRoleTokenProvider = sess.loadToken //stsvalue.StdinTokenProvider
	sessOpts.AssumeRoleDuration = roleDuration

	newSession, err := session.NewSessionWithOptions(sessOpts)
	checkErrorAndExit(err, "Failed to Created Session")

	creds := newSession.Config.Credentials
	
	credentialsValue.setAssumeRoleCreds(creds)
	updateCredentialsFile(sess, credentialsValue)
}

func (sess *awsSession) GetUserSession() {

	//  var credentialsValue *credentialResult
	var sessOpts session.Options
	//sessOpts.Profile = sess.Profile
	sessOpts.Config.Region = aws.String("us-east-1")
	sessOpts.Config.Credentials = credentials.NewSharedCredentials(defaults.SharedCredentialsFilename(),sess.Profile)
	t, err := time.ParseDuration(sess.Duration)
	seconds := t.Seconds()
	fmt.Println(sess.Token)
	params := &sts.GetSessionTokenInput{
		DurationSeconds: aws.Int64(int64(seconds)),
		TokenCode: aws.String(sess.Token),
		SerialNumber: aws.String("arn:aws:iam::016737941129:mfa/testMFA"),
	}
	newSession, err := session.NewSessionWithOptions(sessOpts)
	checkErrorAndExit(err, "Failed to Created Session")
	stsSession := sts.New(newSession)
	stsOutput, err := stsSession.GetSessionToken(params)
	checkErrorAndExit(err, "Could Not create user session")

	creds := stsOutput.Credentials
	// fmt.Println(*creds.AccessKeyId)
	// fmt.Println(*creds.SecretAccessKey)
	// fmt.Println(*creds.SessionToken)
	// valueA := *creds.AccessKeyId
	// valueB := *creds.SecretAccessKey
	// valueC := *creds.SessionToken

	// fmt.Println(valueA)
	// fmt.Println(valueB)
	// fmt.Println(valueC)
	values := &credentialResult{
		accessKey: *creds.AccessKeyId,
		secretKey: *creds.SecretAccessKey,
		sessionToken: *creds.SessionToken,
	}
	//  credentialsValue.accessKey = valueA
	//  credentialsValue.secretKey = valueB
	//  credentialsValue.sessionToken = valueC
	 fmt.Println(values)
	// credentialsValue.setUserSessionCreds(creds)
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
