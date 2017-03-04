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

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/capitalone/stack-deployment-tool/stacks"

	"github.com/stretchr/testify/assert"
)

func TestAmi(t *testing.T) {
	SetTestDefaults()

	ranImageFunc := false

	imgFinder := NewAmiFinder()
	imgFinder.DescribeImages = func(params *ec2.DescribeImagesInput) (*ec2.DescribeImagesOutput, error) {
		ranImageFunc = true
		return &ec2.DescribeImagesOutput{Images: []*ec2.Image{&redhatImage}}, nil
	}
	tmpl := stacks.NewTemplate(nil, nil)
	tmpl.HelpersCtx["ami"] = imgFinder // override with ours just for this render

	amiId := tmpl.Render("{{ami platform=\"rhel\" version=\"7.3\"}}")
	assert.Equal(t, "ami-b63769a1", amiId)
	assert.True(t, ranImageFunc)
}

func TestLatestAmi(t *testing.T) {
	SetTestDefaults()

	ranImageFunc := false

	imgFinder := NewAmiFinder()
	imgFinder.DescribeImages = func(params *ec2.DescribeImagesInput) (*ec2.DescribeImagesOutput, error) {
		ranImageFunc = true
		return &ec2.DescribeImagesOutput{Images: []*ec2.Image{&redhatImage}}, nil
	}
	tmpl := stacks.NewTemplate(nil, nil)
	tmpl.HelpersCtx[latestAmiTemplCmd] = imgFinder // override with ours just for this render

	amiId := tmpl.Render("{{latest-ami platform=\"rhel\"}}")
	assert.Equal(t, "ami-b63769a1", amiId)
	assert.True(t, ranImageFunc)
}
