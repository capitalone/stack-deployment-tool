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
package providers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/capitalone/stack-deployment-tool/sdt"
	"github.com/capitalone/stack-deployment-tool/utils"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/s3"
	ini "github.com/go-ini/ini"
)

const (
	retry_max = 5
)

type AWSApi struct {
	Session *session.Session
	cfsrvc  *cloudformation.CloudFormation
	s3srvc  *s3.S3
	ec2srvc *ec2.EC2
	dryMode bool
}

func NewAWSApi() *AWSApi {
	return &AWSApi{Session: createSession(http.DefaultClient)}
}

func NewAWSApiWithHttpClient(httpClient *http.Client) *AWSApi {
	return &AWSApi{Session: createSession(httpClient)}
}

func sessionName() string {
	return fmt.Sprintf("%s-%d", utils.GetenvWithDefault("USER", ""), time.Now().UTC().Unix())
}

func RoleArn() string {
	role := utils.GetenvWithDefault("AWS_ROLE_ARN", "")
	if len(role) == 0 { // maybe try $HOME/.aws/config
		profile := utils.GetenvWithDefault("AWS_PROFILE", "")
		configFile := filepath.Join(os.Getenv("HOME"), ".aws", "config")
		if len(profile) > 0 && utils.FileExists(configFile) {
			cfg, err := ini.Load(configFile)
			if err != nil {
				log.Debugf("Error loading %s: %+v", configFile, err)
			} else {
				role = cfg.Section("profile " + profile).Key("saml_role").Value()
			}
		}
	}
	if len(role) == 0 { // check if it is still empty..
		log.Infof("AWS_ROLE_ARN is empty")
	} else {
		log.Infof("Using AWS ROLE: %s", role)
	}
	return role
}

func Region() string {
	return utils.GetenvWithDefault("AWS_REGION", utils.GetenvWithDefault("AWS_DEFAULT_REGION", "us-east-1"))
}

// addNameAndVersionToUserAgent will add the name and version of this utility to the
// user-agent in the request from the aws sdk
var addNameAndVersionToUserAgent = request.NamedHandler{
	Name: "stack-deployment-tool.providers.aws.UserAgentHandler",
	Fn:   request.MakeAddToUserAgentHandler("sdt", sdt.Version),
}

func createSession(httpClient *http.Client) *session.Session {
	region := Region()
	config := aws.NewConfig().WithRegion(region).
		WithMaxRetries(retry_max).WithCredentialsChainVerboseErrors(true).
		WithHTTPClient(httpClient)

	sess := session.New(config)

	role := RoleArn()
	if len(role) > 0 {
		sess.Config.Credentials = stscreds.NewCredentials(sess,
			role,
			func(p *stscreds.AssumeRoleProvider) { p.RoleSessionName = sessionName() })
	}

	sess.Handlers.Build.PushFrontNamed(addNameAndVersionToUserAgent)

	return sess
}

func (a *AWSApi) IsDryMode() bool {
	return a.dryMode
}

func (a *AWSApi) DryMode(enable bool) {
	a.dryMode = enable
}

func (a *AWSApi) MustHaveAccess() {
	params := &cloudformation.DescribeAccountLimitsInput{}
	_, err := cloudformation.New(a.Session, a.Session.Config).DescribeAccountLimits(params)
	if err != nil {
		// check for ExpiredToken
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "ExpiredToken" {
			log.Fatalf("%v %v", awsErr.Code(), awsErr.Message())
		}
	}
}

func (a *AWSApi) CFService() *cloudformation.CloudFormation {
	if a.cfsrvc == nil {
		a.MustHaveAccess()
		a.cfsrvc = cloudformation.New(a.Session, a.Session.Config)
	}
	return a.cfsrvc
}

func (a *AWSApi) S3Service() *s3.S3 {
	if a.s3srvc == nil {
		a.MustHaveAccess()
		a.s3srvc = s3.New(a.Session, a.Session.Config)
	}
	return a.s3srvc
}

func (a *AWSApi) EC2Service() *ec2.EC2 {
	if a.ec2srvc == nil {
		a.MustHaveAccess()
		a.ec2srvc = ec2.New(a.Session, a.Session.Config)
	}
	return a.ec2srvc
}
