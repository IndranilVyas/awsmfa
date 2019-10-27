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
	"github.com/IndranilVyas/awsmfa/pkg"
	"github.com/spf13/cobra"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	// "github.com/spf13/viper"
)

var (
	profile  string
	duration string
	token    string
	err      error
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "awsmfa",
	Short: "Manage your AWS Secruity Credentials for MFA enabled AWS API Access",
	Long: `Manage your AWS Secruity Credentials for aws cli/api access that has MFA enabled.
  awsmfa allows user to temporary credentials using named profile`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		 

		session := awssession.New()
		session.Profile = profile
		session.Duration = duration
		session.Token = token
		session.HomeDir, err = homedir.Dir()
		if err != nil {
			fmt.Printf("Unable get Home directory \nError: %v", err.Error())
			os.Exit(1)
		}
		session.Save()

	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	//cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.Flags().StringVarP(&profile, "profile", "p", "", "profile name found config file (default config file is $HOME/.aws/config)")
	rootCmd.Flags().StringVarP(&duration, "duration", "d", "1h", "Session Duration like 1h, 2h.")
	rootCmd.Flags().StringVarP(&token, "token", "t", "", "MFA Token for User")
	rootCmd.MarkFlagRequired("profile")
}
