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
package stacks

import (
	"fmt"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Exit(fakeAwsEnv(func() int { return m.Run() }))
}

/*
func TestAwsApiCreateStackOnly(t *testing.T) {
	log.SetLevel(log.DebugLevel)

	r, cleaner := recording.CreateHttpRecorder("../fixtures/")
	defer cleaner() // Make sure recorder is stopped once done with it

	client := recording.RecorderHttpClient(r)

	aws := NewAWSStackApi(providers.NewAWSApiWithHttpClient(client))

	conf := NewConfig(ResourcePath("bluegreen/stack_blue.yaml"), aws)

	item := conf.FetchEnvStacks("dev-techops.blue")

	aws.CreateOrUpdateStacks(item)
}

func TestAwsApiFindStack(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	r, cleaner := recording.CreateHttpRecorder("../fixtures/")
	defer cleaner() // Make sure recorder is stopped once done with it

	client := recording.RecorderHttpClient(r)

	aws := NewAWSStackApi(providers.NewAWSApiWithHttpClient(client))

	stack := aws.FindStack("drone-ecs")
	assert.NotNil(t, stack)
	assert.Equal(t, *stack.StackStatus, "DELETE_FAILED")
}

func TestAwsApiCreateStackOutput(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	r, cleaner := recording.CreateHttpRecorder("../fixtures/")
	defer cleaner() // Make sure recorder is stopped once done with it

	client := recording.RecorderHttpClient(r)

	aws := NewAWSStackApi(providers.NewAWSApiWithHttpClient(client))

	out, err := aws.FindDeploymentOutput("drone-rr-build-0-0-2", "TargetHostName")
	assert.Nil(t, err)
	//assert.Equal(t, out, "*.com.")
}
*/

func fakeAwsEnv(toRun func() int) int {
	if os.Getenv("RECORDING") == "1" {
		fmt.Printf("-\\/\\/-  RECORDING -\\/\\/-\n")
		return toRun()
	} else { // unless recording is enabled...
		os.Setenv("AWS_PROFILE", "Developer")
		os.Setenv("AWS_ROLE_ARN", "arn:aws:iam::000000000000:role/Developer")
		os.Setenv("AWS_ACCESS_KEY_ID", "0")
		os.Setenv("AWS_ACCESS_KEY", "0")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "0")
		os.Setenv("AWS_SECRET_KEY", "0")

		defer func() {
			os.Unsetenv("AWS_ACCESS_KEY_ID")
			os.Unsetenv("AWS_ACCESS_KEY")
			os.Unsetenv("AWS_SECRET_ACCESS_KEY")
			os.Unsetenv("AWS_SECRET_KEY")
		}()
		return toRun()
	}
}
