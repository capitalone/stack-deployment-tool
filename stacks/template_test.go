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
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/capitalone/stack-deployment-tool/utils"

	log "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestEnvKeys(t *testing.T) {
	fmt.Printf("env keys: %+v\n", EnvKeys())
	assert.True(t, len(EnvKeys()) > 0)
}

type FakeDeploymentFinder struct {
	StackName string
	OutputKey string
}

func (f *FakeDeploymentFinder) FindDeploymentOutput(stackName string, outputKey string) (string, error) {
	f.StackName = stackName
	f.OutputKey = outputKey
	fmt.Printf("stackName: %s\n\n", f.StackName)
	if stackName != "error" {
		return "something", nil
	}
	return "", errors.New("asdf")
}

func TestBasicTemplate(t *testing.T) {
	tmpl := NewTemplate(&FakeDeploymentFinder{}, nil)
	home := tmpl.Render("{{ env.HOME }}")
	assert.NotNil(t, home)
}

func TestOutputMacro(t *testing.T) {
	envKey := "_TEST_TMPL_VAL"
	os.Setenv(envKey, "test")
	fake := &FakeDeploymentFinder{}
	tmpl := NewTemplate(fake, nil)
	out := tmpl.Render("{{output stack=\"bank-nagios-elb-build-{{env._TEST_TMPL_VAL}}\" key=\"ELBDNS\"}}")

	assert.Equal(t, "bank-nagios-elb-build-test", fake.StackName)
	assert.Equal(t, "something", out)
	os.Unsetenv(envKey)
}

func TestFromYamlTemplate(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	c := NewConfig(ResourcePath("from_yaml.yml"), tempOutputFinder())
	tmpl := NewTemplate(&FakeDeploymentFinder{}, c)
	val := tmpl.Render("{{ from_yaml \"test\" }}")
	assert.NotNil(t, val)
}

func TestTemplateParsing(t *testing.T) {
	u := utils.GetenvWithDefault("HOME", "nothing")
	assert.NotEqual(t, u, "nothing")
}

func TestGetEnv(t *testing.T) {
	u := utils.GetenvWithDefault("HOME", "nothing")
	assert.NotEqual(t, u, "nothing")

	u = utils.GetenvWithDefault("SOMETHINGREALLYINVALID", "blah")
	assert.Equal(t, u, "blah")
}

func TestEnvValueMap(t *testing.T) {
	b := EnvValueMap("BLAH")
	assert.NotNil(t, b["BLAH"])
	assert.Equal(t, reflect.TypeOf(b["BLAH"]).Kind(), reflect.Func)
}
