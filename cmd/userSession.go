package cmd

/*
Copyright © 2019 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

import (
	"fmt"
	"os"

	"github.com/IndranilVyas/awsmfa/pkg"
	homedir "github.com/mitchellh/go-homedir"

	"github.com/spf13/cobra"
)

// userSessionCmd represents the userSession command
var userSessionCmd = &cobra.Command{
	Use:   "userSession",
	Short: "aws sts get-session-token implementation",
	Long: `This command calls aws sts get-session token and saves output to default
~/.aws/credentials file`,
	Run: func(cmd *cobra.Command, args []string) {
		session := awssession.New()
		session.Profile, _ = cmd.Flags().GetString("profile")
		session.HomeDir, err = homedir.Dir()
		session.Duration, _ = cmd.Flags().GetString("duration")
		session.Token , _ = cmd.Flags().GetString("token")
		if err != nil {
			fmt.Printf("Unable get Home directory \nError: %v", err.Error())
			os.Exit(1)
		}
		session.GetUserSession()

	},
}

func init() {
	rootCmd.AddCommand(userSessionCmd)
	userSessionCmd.Flags().StringP("token", "t", "", "MFA Device Token")
	userSessionCmd.Flags().StringP("duration", "d", "1h", "Session Duration like 1h, 2h.")
	userSessionCmd.Flags().StringP("profile", "p", "default", "Profile name where IAM USER is defined")

}
