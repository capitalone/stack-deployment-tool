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
	"reflect"
	"strings"

	"github.com/capitalone/stack-deployment-tool/artifacts"
	"github.com/capitalone/stack-deployment-tool/stacks"
	"github.com/capitalone/stack-deployment-tool/utils"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	bucketFlag, repoFlag, regionFlag, groupFlag, artifactFlag, versionFlag, stkconfFlag, providerFlag string
	urlFlag, userFlag, passwdFlag                                                                     string
)

var artifactsCmd = &cobra.Command{
	Use:   "artifacts",
	Short: "artifacts functions, like finding uploading, downloading, promoting",
	Long:  "upload, download, and promote artifacts",
}

var artifactsUploadCmd = &cobra.Command{
	Use:   "upload [file]",
	Short: "upload artifact",
	Long:  "upload an artifact to a destination",
	Run: func(runCmd *cobra.Command, args []string) {
		ValidateArgLen(1, args, "no file to upload specified")
		fileName := args[0]
		setAwsRegion()

		var artifact artifacts.ArtifactAccessor
		if providerFlag == "s3" {
			artifact = newS3Artifact(fileName)
		} else if providerFlag == "nexus" {
			artifact = newNexusArtifact(fileName)
		} else {
			log.Fatalf("Invalid Provider: %s", providerFlag)
		}

		mergeStacksConfIfAvailable(artifact)

		// artifact.Valid()

		log.Debugf("artifact: %#v", artifact)

		if !dryFlag {
			loc := artifact.Upload()
			log.Infof("Uploaded to: %s", loc)
		}
	},
}

var artifactsDownoadCmd = &cobra.Command{
	Use:   "download [archive name]",
	Short: "download artifact",
	Long:  "download an artifact to a destination",
	Run: func(runCmd *cobra.Command, args []string) {
		ValidateArgLen(1, args, "no file to download specified")
		fileName := args[0]
		setAwsRegion()

		var artifact artifacts.ArtifactAccessor
		if providerFlag == "s3" {
			artifact = newS3Artifact(fileName)
		} else if providerFlag == "nexus" {
			artifact = newNexusArtifact(fileName)
		} else {
			log.Fatalf("Invalid Provider: %s", providerFlag)
		}

		mergeStacksConfIfAvailable(artifact)
		log.Debugf("artifact: %#v", artifact)

		if !dryFlag {
			artifact.Download()
			log.Infof("Downloaded: %s", fileName)
		}
	},
}

// try to find & load the stacks Config
func mergeStacksConfIfAvailable(artifact artifacts.ArtifactAccessor) {
	stkconf := "stacks.yml"
	if len(stkconfFlag) > 0 {
		stkconf = stkconfFlag
	}
	if len(stkconf) > 0 && utils.FileExists(stkconf) {
		conf := stacks.NewConfig(stkconf, StacksApi())
		val := utils.ToStrMap(conf.FetchJsonPtr("/artifacts/" + providerFlag))
		if val != nil {
			mergeConf(artifact, val, false)
		}
	}
}

func newArtifact(fileName string) artifacts.Artifact {
	return artifacts.Artifact{Version: versionFlag, Name: artifactFlag, Group: groupFlag, Repo: repoFlag, FileName: fileName}
}

func newS3Artifact(fileName string) *artifacts.S3Artifact {
	return &artifacts.S3Artifact{Bucket: bucketFlag, Region: regionFlag,
		Path: viper.GetString("S3Path"), BucketUrl: viper.GetString("S3BucketUrl"),
		Artifact: newArtifact(fileName), Encrypt: true}
}

func newNexusArtifact(fileName string) *artifacts.NexusArtifact {
	url := urlFlag
	if len(urlFlag) == 0 {
		url = viper.GetString("NexusUrl")
	}
	return &artifacts.NexusArtifact{Artifact: newArtifact(fileName), Url: url, User: userFlag, Password: passwdFlag}
}

func mergeConf(artifact interface{}, conf map[string]interface{}, overwrite bool) {
	for k, v := range conf {
		field := reflect.ValueOf(artifact).Elem().FieldByName(strings.Title(k))
		if field.IsValid() && field.CanSet() {
			if field.Len() == 0 || overwrite {
				field.SetString(v.(string))
			}
		}
	}
}

var artifactsPromoteCmd = &cobra.Command{
	Use:   "promote [archive name] [from repo] [optional: to repo]",
	Short: "promote artifact",
	Long:  "promote an artifact to a destination",
	Run: func(runCmd *cobra.Command, args []string) {

		ValidateArgLen(2, args, "file & source repo need to be specified")

		fileName := args[0]
		repoFlag = targetRepo(args[1])
		setAwsRegion()

		var artifact artifacts.ArtifactAccessor
		if providerFlag == "s3" {
			artifact = newS3Artifact(fileName)
		} else if providerFlag == "nexus" {
			artifact = newNexusArtifact(fileName)
		} else {
			log.Fatalf("Invalid Provider: %s", providerFlag)
		}

		mergeStacksConfIfAvailable(artifact)

		log.Debugf("artifact: %#v", artifact)

		if !dryFlag {
			artifact.Promote(args[1])
			log.Infof("Promoted: %s", fileName)
		}
	},
}

func targetRepo(fromRepo string) string {
	promotionPath := viper.GetStringMapString("PromotionPath")
	if target, ok := promotionPath[fromRepo]; ok {
		return target
	}
	log.Fatalf("promotion path from: %s not found", fromRepo)
	return ""
}

func setAwsRegion() {
	if len(regionFlag) > 0 {
		utils.SetenvIfAbsent("AWS_REGION", regionFlag)
	}
}

func init() {
	viper.SetDefault("PromotionPath", map[string]string{
		// from  : to
		"sandbox":  "snapshot",
		"snapshot": "staging",
		"staging":  "release",
	})

	viper.SetDefault("S3Path", "artifacts/{{Repo}}/{{Key}}/{{Name}}/{{Version}}/{{FileName}}")
	viper.SetDefault("S3BucketUrl", "s3://{{Bucket}}/")

	artifactsCmd.AddCommand(artifactsUploadCmd)
	artifactsCmd.AddCommand(artifactsDownoadCmd)
	artifactsCmd.AddCommand(artifactsPromoteCmd)
	RootCmd.AddCommand(artifactsCmd)

	artifactsCmd.PersistentFlags().StringVarP(&stkconfFlag, "stacksconf", "s", "stacks.yml", "specify the stack_config.yml")
	artifactsCmd.PersistentFlags().StringVarP(&providerFlag, "provider", "p", "s3", "target provider [default: s3]")

	// nexus
	artifactsUploadCmd.PersistentFlags().StringVarP(&urlFlag, "url", "u", "", "repository URL for posting archive")
	artifactsDownoadCmd.PersistentFlags().StringVarP(&urlFlag, "url", "u", "", "repository URL for pulling archive")
	artifactsPromoteCmd.PersistentFlags().StringVarP(&urlFlag, "url", "u", "", "repository URL for promoting archive")

	artifactsUploadCmd.PersistentFlags().StringVarP(&userFlag, "user", "", "", "user for posting archive")
	artifactsDownoadCmd.PersistentFlags().StringVarP(&userFlag, "user", "", "", "user for pulling archive")
	artifactsPromoteCmd.PersistentFlags().StringVarP(&userFlag, "user", "", "", "user for promoting archive")

	artifactsUploadCmd.PersistentFlags().StringVarP(&passwdFlag, "pass", "", "", "password for posting archive")
	artifactsDownoadCmd.PersistentFlags().StringVarP(&passwdFlag, "pass", "", "", "password for pulling archive")
	artifactsPromoteCmd.PersistentFlags().StringVarP(&passwdFlag, "pass", "", "", "password for promoting archive")

	// s3
	artifactsUploadCmd.PersistentFlags().StringVarP(&repoFlag, "repo", "r", "sandbox", "target repository (default sandbox)")
	artifactsDownoadCmd.PersistentFlags().StringVarP(&repoFlag, "repo", "r", "sandbox", "target repository (default sandbox)")

	artifactsUploadCmd.PersistentFlags().StringVarP(&bucketFlag, "bucket", "b", "", "target bucket")
	artifactsDownoadCmd.PersistentFlags().StringVarP(&bucketFlag, "bucket", "b", "", "target bucket")
	artifactsPromoteCmd.PersistentFlags().StringVarP(&bucketFlag, "bucket", "b", "", "target bucket")

	artifactsUploadCmd.PersistentFlags().StringVarP(&regionFlag, "region", "n", "us-east-1", "target bucket region")
	artifactsDownoadCmd.PersistentFlags().StringVarP(&regionFlag, "region", "n", "us-east-1", "target bucket region")
	artifactsPromoteCmd.PersistentFlags().StringVarP(&regionFlag, "region", "n", "us-east-1", "target bucket region")

	artifactsUploadCmd.PersistentFlags().StringVarP(&groupFlag, "group", "g", "", "repositiory group")
	artifactsDownoadCmd.PersistentFlags().StringVarP(&groupFlag, "group", "g", "", "repositiory group")
	artifactsPromoteCmd.PersistentFlags().StringVarP(&groupFlag, "group", "g", "", "repositiory group")

	artifactsUploadCmd.PersistentFlags().StringVarP(&artifactFlag, "artifact", "a", "", "repositiory artifact name")
	artifactsDownoadCmd.PersistentFlags().StringVarP(&artifactFlag, "artifact", "a", "", "repositiory artifact name")
	artifactsPromoteCmd.PersistentFlags().StringVarP(&artifactFlag, "artifact", "a", "", "repositiory artifact name")

	artifactsUploadCmd.PersistentFlags().StringVarP(&versionFlag, "version", "v", "", "specify the version for the artifact")
	artifactsDownoadCmd.PersistentFlags().StringVarP(&versionFlag, "version", "v", "", "specify the version for the artifact")
	artifactsPromoteCmd.PersistentFlags().StringVarP(&versionFlag, "version", "v", "", "specify the version for the artifact")

}
