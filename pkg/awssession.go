package awssession

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/ini.v1"

	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
)

const (
	awsAccessKey    string = "aws_access_key_id"
	awsSecretKey    string = "aws_secret_access_key"
	awsSessionToken string = "aws_session_token"
)

type awsSession struct{
	Profile string
	Duration string
	Token string
	HomeDir string
}
//New create new awsSession object
func New () *awsSession{
	return new(awsSession)
}
func(sess * awsSession) loadToken() (string, error) {
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



func (sess * awsSession) Save() {

	roleDuration, err := time.ParseDuration(sess.Duration)
	var sessOpts session.Options
	sessOpts.Profile = sess.Profile
	sessOpts.SharedConfigState = session.SharedConfigEnable
	sessOpts.AssumeRoleTokenProvider = sess.loadToken //stscreds.StdinTokenProvider
	sessOpts.AssumeRoleDuration = roleDuration

	newSession, err := session.NewSessionWithOptions(sessOpts)
	checkErrorAndExit(err, "Failed to Created Session")

	updateCredentialsFile(sess, newSession)

}

func updateCredentialsFile(sess *awsSession, currSession *session.Session) {

	filePath := sess.HomeDir+"/.aws/credentials"
	creds, err := currSession.Config.Credentials.Get()
	checkErrorAndExit(err, "Error loading Credentials from Current Session")
	credsFile, err := ini.Load(filePath)
	checkErrorAndExit(err, "Error Loading credentials from "+filePath)
	mfaProfileName := sess.Profile + "_mfa"

	mfaSection, err := credsFile.GetSection(mfaProfileName)

	if err != nil {
		fmt.Printf("Credentils Not Found...Creating Section for %s \n", mfaProfileName)

		mfaSection, err := credsFile.NewSection(mfaProfileName)
		checkErrorAndExit(err, "Failed to Created New Section")
		_, err = mfaSection.NewKey(awsAccessKey, creds.AccessKeyID)
		checkErrorAndExit(err, "Failed to add:"+awsAccessKey)
		_, err = mfaSection.NewKey(awsSecretKey, creds.SecretAccessKey)
		checkErrorAndExit(err, "Failed to add:"+awsAccessKey)
		_, err = mfaSection.NewKey(awsSessionToken, creds.SessionToken)
		checkErrorAndExit(err, "Failed to add:"+awsAccessKey)
		err = credsFile.SaveTo(sess.HomeDir)
		checkErrorAndExit(err, "Failed to Save credentials" )
		fmt.Println("New Credentials added to Credentials File")
	} else {
		fmt.Println("Previous Credentials Found....Updating Them")
		mfaSection.Key(awsAccessKey).SetValue(creds.AccessKeyID)
		mfaSection.Key(awsSecretKey).SetValue(creds.SecretAccessKey)
		mfaSection.Key(awsSessionToken).SetValue(creds.SessionToken)
		credsFile.SaveTo(sess.HomeDir)
		checkErrorAndExit(err, "Failed to Save credentials" )
		fmt.Println("Credentials File updated with New Credentials")
}

}
