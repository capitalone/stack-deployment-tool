//
// Copyright 2017 Capital One Services, LLC
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
	"fmt"
	"os"
	"regexp"

	"github.com/capitalone/stack-deployment-tool/images"
	"github.com/capitalone/stack-deployment-tool/providers"
	"github.com/capitalone/stack-deployment-tool/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var imagesCmd = &cobra.Command{
	Use:   "images",
	Short: "image functions, like finding latest ami id",
	Long:  "image related functions, AMI images, docker images, etc.",
}

var imagesAmiCmd = &cobra.Command{
	Use:     "ami [platform] [version]",
	Short:   "ami id",
	Aliases: []string{"amis"},
	Long:    "get ami information",
	Run: func(runCmd *cobra.Command, args []string) {
		ValidateArgMinLen(1, args, "required: [platform] [version]")
		var imageVer string
		if len(args) > 1 {
			imageVer = args[1]
		}
		imgs := images.NewAmiFinder().FindImages(args[0], imageVer)
		fmt.Printf("Matching Images\n")
		tw := utils.NewTableWriter(os.Stdout, 20, 40, 25, 30)
		tw.WriteHeader("Image Id", "Name", "Version", "Creation Date")
		for _, img := range imgs {
			tw.WriteRow(*img.ImageId, *img.Name, images.ImageVersion(img), *img.CreationDate)
		}
		tw.Footer()
	},
}

var imagesLatestAmiCmd = &cobra.Command{
	Use:   "latest-ami [platform]",
	Short: "latest-ami id",
	Long:  "get latest ami information",
	Run: func(runCmd *cobra.Command, args []string) {
		ValidateArgMinLen(1, args, "required: [platform] [version]")
		fmt.Println(images.NewAmiFinder().FindLatestImageId(args[0]))
	},
}

func init() {
	viper.SetDefault("ImageFilter", map[string]string{
		"state":     "available",
		"is-public": "false",
	})

	if matched, err := regexp.MatchString(`.*role\/.*_COF_.*`, providers.RoleArn()); err == nil && matched {
		// i expect these to be customized..
		viper.SetDefault("ImageFilterRegex", map[string]string{
			"name": `^COF-[0-9A-Za-z_.\-]+-x64-HVM-Enc-[0-9A-Za-z_.\-]+`,
		})
	}

	viper.SetDefault("VersionRegex", `-([\d]+(-[\d]+)*)$`)
	viper.SetDefault("PlatformRegex", `^[0-9A-Za-z_.]+-([0-9A-Za-z_.]+)`)

	imagesCmd.AddCommand(imagesAmiCmd)
	imagesCmd.AddCommand(imagesLatestAmiCmd)
	RootCmd.AddCommand(imagesCmd)
}
