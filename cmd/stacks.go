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
	"fmt"

	"github.com/capitalone/stack-deployment-tool/stacks"
	"github.com/capitalone/stack-deployment-tool/utils"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	stacksRef string
	process   bool
	api       stacks.StackApi
)

// stacksCmd represents the stacks command
var stacksCmd = &cobra.Command{
	Use:   "stacks",
	Short: "Stack manipulation commands",
	Long:  "Stack manipulation commands",
}

var stacksTemplateCmd = &cobra.Command{
	Use:   "template [stack_config.yml]",
	Short: "Interpret the stacks template for cloudformation stacks",
	Long:  "Interpret the stacks template that will be used to launch the cloudformation stacks defined",
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug("stacks template called")
		ValidateArgLen(1, args, "stacks config file required")
		conf := stacks.NewConfig(args[0], StacksApi())
		var result interface{}
		if len(stacksRef) > 0 {
			log.Debugf("stacksRef: %+v", stacksRef)
			item := conf.FetchEnvStacks(stacksRef)
			result = item.FetchAll()
		} else if process {
			result = conf.FetchAll()
		} else {
			result = conf.Yaml
		}

		fmt.Println(utils.EncodeYAML(result))
	},
}

var stacksCreateOrUpdateCmd = &cobra.Command{
	Use:     "deploy [stack_config.yml]",
	Short:   "Deploy a set of cloudformation stacks",
	Long:    "Deploy a set of cloudformation stacks",
	Aliases: []string{"deploy", "create"},
	Run: func(cmd *cobra.Command, args []string) {
		if len(stacksRef) == 0 {
			log.Fatalf("specify stack option: -s <environment>.<stack name> or <environment>[<stack name>, ...]")
		}
		ValidateArgLen(1, args, "stacks config file required")
		conf := stacks.NewConfig(args[0], StacksApi())
		item := conf.FetchEnvStacks(stacksRef)
		StacksApi().CreateOrUpdateStacks(item)
	},
}

var stacksDeleteCmd = &cobra.Command{
	Use:     "delete [stack_config.yml]",
	Short:   "Delete a set of cloudformation stacks",
	Long:    "Delete a set of cloudformation stacks",
	Aliases: []string{"teardown", "destroy"},
	Run: func(cmd *cobra.Command, args []string) {
		if len(stacksRef) == 0 {
			log.Fatalf("specify stack option: -s <environment>.<stack name> or <environment>[<stack name>, ...]")
		}
		ValidateArgLen(1, args, "stacks config file required")
		conf := stacks.NewConfig(args[0], StacksApi())
		item := conf.FetchEnvStacks(stacksRef)
		StacksApi().DeleteStacks(item)
	},
}

var stacksStatusCmd = &cobra.Command{
	Use:   "status [stack_config.yml]",
	Short: "Status a set of cloudformation stacks",
	Long:  "Status a set of cloudformation stacks",
	Run: func(cmd *cobra.Command, args []string) {
		if len(stacksRef) == 0 {
			log.Fatalf("specify stack option: -s <environment>.<stack name> or <environment>[<stack name>, ...]")
		}
		ValidateArgLen(1, args, "stacks config file required")
		conf := stacks.NewConfig(args[0], StacksApi())
		item := conf.FetchEnvStacks(stacksRef)
		StacksApi().StacksStatus(item)
	},
}

var stacksChangesCmd = &cobra.Command{
	Use:   "changes [stack_config.yml]",
	Short: "Show changes a set of cloudformation stacks will make",
	Long:  "Show changes a set of cloudformation stacks will make to existing stacks",
	Run: func(cmd *cobra.Command, args []string) {
		if len(stacksRef) == 0 {
			log.Fatalf("specify stack option: -s <environment>.<stack name> or <environment>[<stack name>, ...]")
		}
		ValidateArgLen(1, args, "stacks config file required")
		conf := stacks.NewConfig(args[0], StacksApi())
		item := conf.FetchEnvStacks(stacksRef)
		StacksApi().PrintChangesToStacks(item)
	},
}

var stacksJsonToYamlCmd = &cobra.Command{
	Use:   "yaml [stack.json]",
	Short: "Convert a CloudFormation stack in json to yaml",
	Long:  "Convert a CloudFormation stack in json to yaml",
	Run: func(cmd *cobra.Command, args []string) {
		ValidateArgLen(1, args, "file required")
		// naive convert, we dont change to the custom tags (yet)
		y, err := utils.DecodeYAMLFile(args[0])
		if err != nil {
			log.Fatalf("Error loading file: %s", args[0])
		}
		fmt.Printf("\n%s\n", utils.EncodeYAML(y))
	},
}

func StacksApi() stacks.StackApi {
	if api != nil {
		return api
	}
	api = stacks.DefaultStackApi()
	api.DryMode(dryFlag)
	if dryFlag {
		log.Infof("-- DRY MODE --")
	}
	return api
}

func init() {
	stacksCmd.AddCommand(stacksTemplateCmd)
	stacksCmd.AddCommand(stacksCreateOrUpdateCmd)
	stacksCmd.AddCommand(stacksDeleteCmd)
	stacksCmd.AddCommand(stacksStatusCmd)
	stacksCmd.AddCommand(stacksChangesCmd)
	stacksCmd.AddCommand(stacksJsonToYamlCmd)
	RootCmd.AddCommand(stacksCmd)

	stacksCmd.PersistentFlags().StringVarP(&stacksRef, "stacks", "s", "", "<environment>.<stack name> or <environment>[<stack name>, ...]")
	stacksTemplateCmd.PersistentFlags().BoolVarP(&process, "process", "p", false, "process the template")
}
