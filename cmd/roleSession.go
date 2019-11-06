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

// roleSessionCmd represents the roleSession command
var roleSessionCmd = &cobra.Command{
	Use:   "roleSession",
	Short: "Manage your AWS Session Credentials for IAM Roles with MFA enabled",
	Long: `Manage your AWS Session Credentials for aws cli/api access IAM Role that has MFA enabled.
awsmfa will generate Session Credentials and save them in default credentials file`,
	Run: func(cmd *cobra.Command, args []string) {
		session := awssession.New()
		session.Profile, _ = cmd.Flags().GetString("profile")
		session.Duration, _ = cmd.Flags().GetString("duration")
		session.Token, _ = cmd.Flags().GetString("token")
		session.HomeDir, err = homedir.Dir()
		if err != nil {
			fmt.Printf("Unable get Home directory \nError: %v", err.Error())
			os.Exit(1)
		}
		session.AssumeRoleFromConfig()

	},
}

func init() {
	rootCmd.AddCommand(roleSessionCmd)

	roleSessionCmd.Flags().StringP("token", "t", "", "MFA Device Token")
	roleSessionCmd.Flags().StringP("duration", "d", "1h", "Session Duration like 1h, 2h.")
	roleSessionCmd.Flags().StringP("profile", "p", "default", "Profile name where IAM Role is defined")
}
