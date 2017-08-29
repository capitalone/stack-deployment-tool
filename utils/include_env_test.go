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
package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var expectedEnvIncl = `UserData: 
    "Fn::Base64": !Sub |
        #!/bin/bash
        yum install -y aws-cfn-bootstrap
        something
        /opt/aws/bin/cfn-init -v --region ${AWS::Region} --stack ${AWS::StackName} --resource ECSLaunchConfiguration
        /opt/aws/bin/cfn-signal -e $? --region ${AWS::Region} --stack ${AWS::StackName} --resource ECSAutoScalingGroup
UserData2-something:
    temp: something
`

func TestApplyIncludeEnvDirective(t *testing.T) {
	fname := "../resources/" + "user-data-env.yaml"
	TemporaryChdir("../resources", func() {
		f, err := os.Open(fname)
		assert.Nil(t, err)
		defer f.Close()
		var result []byte
		TemporaryEnv("TEMP_ENV_VAL", "something", func() {
			result = ApplyIncludeEnvDirective(f)
		})

		assert.Equal(t, expectedEnvIncl, string(result))
	})
}
