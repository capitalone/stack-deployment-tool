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
package images

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestImageFilterRegex(t *testing.T) {
	viper.SetDefault("ImageFilterRegex", map[string]string{
		"name": `^SDT-[0-9A-Za-z_.\-]+-x64-HVM-Enc-[0-9A-Za-z_.\-]+`,
	})
	regexMap := imageFilterRegexMap()
	assert.NotNil(t, regexMap)
	assert.NotNil(t, regexMap["name"])
}

func TestImageFilters(t *testing.T) {
	viper.SetDefault("ImageFilter", map[string]string{
		"state":     "available",
		"is-public": "false",
	})
	filters := imageFilters()
	assert.Equal(t, 2, len(filters))

	// order not guaranteed
	assert.True(t, "state" == *filters[0].Name || "state" == *filters[1].Name || "state" == *filters[2].Name)
	assert.True(t, "available" == *filters[0].Values[0] || "available" == *filters[1].Values[0] || "available" == *filters[2].Values[0])
	assert.True(t, "is-public" == *filters[0].Name || "is-public" == *filters[1].Name || "is-public" == *filters[2].Name)
}

var redhatImage = ec2.Image{
	VirtualizationType: aws.String("hvm"),
	Name:               aws.String("RHEL-7.3_HVM_GA-20161026-x86_64-1-Hourly2-GP2"),
	Hypervisor:         aws.String("xen"),
	SriovNetSupport:    aws.String("simple"),
	ImageId:            aws.String("ami-b63769a1"),
	State:              aws.String("available"),
	BlockDeviceMappings: []*ec2.BlockDeviceMapping{{
		DeviceName: aws.String("/dev/sda1"),
		Ebs: &ec2.EbsBlockDevice{
			DeleteOnTermination: aws.Bool(true),
			SnapshotId:          aws.String("snap-b9a222c1"),
			VolumeSize:          aws.Int64(10),
			VolumeType:          aws.String("gp2"),
			Encrypted:           aws.Bool(false),
		},
	}},
	Architecture:   aws.String("x86_64"),
	ImageLocation:  aws.String("309956199498/RHEL-7.3_HVM_GA-20161026-x86_64-1-Hourly2-GP2"),
	RootDeviceType: aws.String("ebs"),
	OwnerId:        aws.String("309956199498"),
	RootDeviceName: aws.String("/dev/sda1"),
	CreationDate:   aws.String("2016-10-26T22:32:29.000Z"),
	Public:         aws.Bool(true),
	ImageType:      aws.String("machine"),
	Description:    aws.String("Provided by Red Hat, Inc."),
}

var redhatImage2 = ec2.Image{
	VirtualizationType: aws.String("hvm"),
	Name:               aws.String("RHEL-6.5_GA_HVM-20140929-x86_64-11-Hourly2-GP2"),
	Hypervisor:         aws.String("xen"),
	SriovNetSupport:    aws.String("simple"),
	ImageId:            aws.String("ami-00a11e68"),
	State:              aws.String("available"),
	BlockDeviceMappings: []*ec2.BlockDeviceMapping{{
		DeviceName: aws.String("/dev/sda1"),
		Ebs: &ec2.EbsBlockDevice{
			DeleteOnTermination: aws.Bool(true),
			SnapshotId:          aws.String("snap-bbafa91c"),
			VolumeSize:          aws.Int64(10),
			VolumeType:          aws.String("gp2"),
			Encrypted:           aws.Bool(false),
		},
	}},
	Architecture:   aws.String("x86_64"),
	ImageLocation:  aws.String("309956199498/RHEL-6.5_GA_HVM-20140929-x86_64-11-Hourly2-GP2"),
	RootDeviceType: aws.String("ebs"),
	OwnerId:        aws.String("309956199498"),
	RootDeviceName: aws.String("/dev/sda1"),
	CreationDate:   aws.String("2014-10-10T13:34:58.000Z"),
	Public:         aws.Bool(true),
	ImageType:      aws.String("machine"),
	Description:    aws.String("Provided by Red Hat, Inc."),
}

func SetTestDefaults() {
	// Red Hat Enterprise Linux 7.3 (HVM), SSD Volume Type
	viper.SetDefault("ImageFilterRegex", map[string]string{
		"name": `^[A-Za-z]+-([\d]+(\.[\d]+)*)_`,
	})
	viper.SetDefault("ImageFilter", map[string]string{})

	viper.SetDefault("VersionRegex", `^[A-Za-z]+-([\d]+(\.[\d]+)*)_`)
	viper.SetDefault("PlatformRegex", `^([A-Za-z]+)-`)
}

func TestImageRegexFilters(t *testing.T) {
	SetTestDefaults()

	r := regexFilterAndSortImages([]*ec2.Image{&redhatImage, &redhatImage2}, "RHEL", "7.3")
	assert.Equal(t, "7.3", ImageVersion(&redhatImage))
	assert.Equal(t, "RHEL", ImagePlatform(&redhatImage))
	assert.Equal(t, 1, len(r))
	assert.Equal(t, *redhatImage.ImageId, *r[0].ImageId)

	r = regexFilterAndSortImages([]*ec2.Image{&redhatImage, &redhatImage2}, "RHEL", "")
	assert.Equal(t, 2, len(r))
	assert.Equal(t, *redhatImage.ImageId, *r[0].ImageId) // sorted newest first..
}

func TestNoRegex(t *testing.T) {
	viper.SetDefault("ImageFilterRegex", map[string]string{})
	viper.SetDefault("ImageFilter", map[string]string{})

	viper.SetDefault("VersionRegex", "")
	viper.SetDefault("PlatformRegex", "")

	imgFinder := NewAmiFinder()
	imgFinder.DescribeImages = func(params *ec2.DescribeImagesInput) (*ec2.DescribeImagesOutput, error) {
		//ranImageFunc = true
		return &ec2.DescribeImagesOutput{Images: []*ec2.Image{&redhatImage}}, nil
	}
	img := imgFinder.FindImageId("", "")
	assert.Equal(t, "ami-b63769a1", img)
}

func TestNoMatch(t *testing.T) {
	SetTestDefaults()

	imgFinder := NewAmiFinder()
	imgFinder.DescribeImages = func(params *ec2.DescribeImagesInput) (*ec2.DescribeImagesOutput, error) {
		//ranImageFunc = true
		return &ec2.DescribeImagesOutput{Images: []*ec2.Image{&redhatImage}}, nil
	}
	img := imgFinder.FindImageId("ubuntu", "14.04")
	assert.Equal(t, "", img)
}
