//
// Copyright 2016 Capital One Services, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and limitations under the License.
//
// SPDX-Copyright: Copyright (c) Capital One Services, LLC
// SPDX-License-Identifier: Apache-2.0
//
package cmd

import (
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	projectName    = "stack-deployment-tool"
	dryModeFlag    = "drymode"
	debugFlag      = "debug"
	assumeRoleFlag = "assume-role"
)

var (
	cfgFile string
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   projectName,
	Short: "Stack Deployment Tool",
	Long: `Stack Deployment Tool
	that will help with deploying multiple CloudFormation stacks
	`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// setup debugging
		if IsDebug() {
			log.SetLevel(log.DebugLevel)
		}
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func IsDryMode() bool {
	return viper.GetBool(dryModeFlag)
}

func IsDebug() bool {
	return viper.GetBool(debugFlag)
}

func init() {
	// setup logging
	log.SetLevel(log.InfoLevel)
	cobra.OnInitialize(initConfig)

	// global flags for the application.
	RootCmd.PersistentFlags().BoolP(debugFlag, "d", false, "enable debug")
	viper.BindPFlag(debugFlag, RootCmd.PersistentFlags().Lookup(debugFlag))

	RootCmd.PersistentFlags().BoolP(dryModeFlag, "q", false, "enable dry mode")
	viper.BindPFlag(dryModeFlag, RootCmd.PersistentFlags().Lookup(dryModeFlag))

	RootCmd.PersistentFlags().Bool(assumeRoleFlag, false,
		"force assume role, if AWS_ROLE_ARN is not provided it will be guessed based on AWS_PROFILE and ~/.aws/config")
	viper.BindPFlag(assumeRoleFlag, RootCmd.PersistentFlags().Lookup(assumeRoleFlag))

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/."+projectName+".yaml)")
	// local flags
	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName("." + projectName) // name of config file (without extension)
	viper.AddConfigPath("$HOME")           // adding home directory as first search path
	viper.AutomaticEnv()                   // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Info("Using config file:", viper.ConfigFileUsed())
	}
}

func ValidateArgLen(wantLen int, args []string, msg interface{}) {
	if len(args) != wantLen {
		log.Fatal("Error: ", msg)
	}
}

func ValidateArgMinLen(wantLen int, args []string, msg interface{}) {
	if len(args) < wantLen {
		log.Fatal("Error: ", msg)
	}
}

func ValidateFlagStr(flag string, msg interface{}) {
	if len(flag) == 0 {
		log.Fatal("Error: ", msg)
	}
}
