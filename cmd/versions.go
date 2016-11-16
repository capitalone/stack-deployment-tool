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
package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"github.com/capitalone/stack-deployment-tool/versioning"

	"github.com/spf13/cobra"
)

var versionsCmd = &cobra.Command{
	Use:   "versions",
	Short: "versioning commands",
	Long:  "versioning support for your project",
}

var versionsBumpCmd = &cobra.Command{
	Use:   "bump <minor|major|patch>",
	Short: "bump version",
	Long:  "bump version based on options",
	Run: func(cmd *cobra.Command, args []string) {
		v := versioning.LoadVersion()
		fmt.Printf("Current version: %+v\n", v)
		ValidateArgLen(1, args, "bump <minor|major|patch>")
		switch args[0] {
		case "minor":
			v.BumpMinor()
		case "major":
			v.BumpMajor()
		case "patch":
			v.BumpPatch()
		case "metadata":
			v.Build = []string{versioning.GitHash()}
		}

		fmt.Printf("New version: %+v\n", v)
		if !dryFlag {
			v.WriteVersion()
		}
	},
}

var versionsPrintCmd = &cobra.Command{
	Use:   "print",
	Short: "print version",
	Long:  "print version based on options",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version: %+v\n", versioning.LoadVersion())
	},
}

var versionsSetCmd = &cobra.Command{
	Use:   "set [version] [filename]",
	Short: "set version",
	Long:  "set version based on options",
	Run: func(cmd *cobra.Command, args []string) {
		ValidateArgMinLen(1, args, "set [version]")
		v := versioning.New(args[0])
		fmt.Printf("New version: %+v\n", v)
		if len(args) > 1 {
			v.WriteVersionProps(args[1])
		} else {
			v.WriteVersion()
		}
	},
}

var versionsInitCmd = &cobra.Command{
	Use:   "init [filename]",
	Short: "initialize version",
	Long:  "initialize version based on options",
	Run: func(cmd *cobra.Command, args []string) {
		v := versioning.New("0.0.0+" + versioning.GitHash())
		fmt.Printf("New version: %+v\n", v)
		if len(args) > 0 {
			v.WriteVersionProps(args[0])
		} else {
			v.WriteVersion()
		}
	},
}

// TODO: i dont like this, it's too tied to jenkins - not going to be supported in the future.
// we should just use the version.properties..
var versionsBuildCmd = &cobra.Command{
	Use:   "build [filename]",
	Short: "Generate build.properties with the artifact information",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		v := versioning.LoadVersion()
		v.Build = []string{versioning.GitHash()}
		propFileName := "build.properties"
		if len(args) > 0 {
			propFileName = args[0]
		}

		w := bytes.NewBufferString("")
		fmt.Fprintf(w, "ARTIFACT_VERSION=%s\n", v.String())
		fmt.Fprintf(w, "ARTIFACT_GIT_COMMIT=%s\n", versioning.GitHash())
		fmt.Fprintf(w, "ARTIFACT_GIT_BRANCH=%s\n", versioning.GetBranch())
		ioutil.WriteFile(propFileName, w.Bytes(), 0666)
	},
}

func init() {
	versionsCmd.AddCommand(versionsBumpCmd)
	versionsCmd.AddCommand(versionsPrintCmd)
	versionsCmd.AddCommand(versionsSetCmd)
	versionsCmd.AddCommand(versionsInitCmd)
	versionsCmd.AddCommand(versionsBuildCmd)
	RootCmd.AddCommand(versionsCmd)

	//versionsBumpCmd.PersistentFlags().StringVarP(&imageVer, "version", "v", "", "image version [AmiVersion]")
}
