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

var expectedUserData = `UserData: 
    "Fn::Base64": !Sub |
        #!/bin/bash
        yum install -y aws-cfn-bootstrap
        asdfoasdf
        oqoruiqweour
        echo 'Blue' > /usr/share/nginx/html/index.html
        echo "Green" > /usr/share/nginx/html/index.html
        /opt/aws/bin/cfn-init -v --region ${AWS::Region} --stack ${AWS::StackName} --resource ECSLaunchConfiguration
        /opt/aws/bin/cfn-signal -e $? --region ${AWS::Region} --stack ${AWS::StackName} --resource ECSAutoScalingGroup
UserData2: 
    "Fn::Base64": !Sub |
        asdfoasdf
        oqoruiqweour
        echo 'Blue' > /usr/share/nginx/html/index.html
        echo "Green" > /usr/share/nginx/html/index.html

UserData3: # unsupported
    !Base64 "Fn::Local::IncludeFileLines": "file_content.txt"
UserData4: { # unsupported
    "Fn::Base64": {"Fn::Local::IncludeFileLines" : "file_content.txt" }
}
`

func TestApplyIncludeFileLinesDirective(t *testing.T) {
	fname := "../resources/" + "user-data.yaml"
	TemporaryChdir("../resources", func() {
		f, err := os.Open(fname)
		assert.Nil(t, err)
		defer f.Close()
		result := ApplyIncludeFileLinesDirective(f)
		//fmt.Printf("result: \n%s\n", result)
		assert.Equal(t, expectedUserData, string(result))
	})
}
