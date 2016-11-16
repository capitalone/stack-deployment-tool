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

	"github.com/capitalone/stack-deployment-tool/utils"

	log "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func ResourcePath(filename string) string {
	return "../resources/" + filename
}

// temporary
type TempDeploymentFinder struct{}

func (f *TempDeploymentFinder) FindDeploymentOutput(stackName string, outputKey string) (string, error) {
	return "", nil
}

func tempOutputFinder() DeploymentOutputFinder {
	return new(TempDeploymentFinder)
}

func TestStacksConfigInclude(t *testing.T) {
	r, err := utils.DecodeYAMLFile(ResourcePath("stacks_include.json"))
	assert.NotNil(t, r)
	assert.Nil(t, err)
	fmt.Printf("err: %#v\n", err)
	fmt.Printf("result: %#v\n", r)
}

func TestStacksConfig(t *testing.T) {
	c := NewConfig(ResourcePath("stacks.yml"), tempOutputFinder())
	fmt.Printf("c: %+v\n", c)

	os.Setenv("APP_VERSION", "1.3.3")

	artifacts := utils.ToStrMap(c.Fetch("artifacts"))
	assert.NotNil(t, artifacts)
	assert.Equal(t, "1.3.3", utils.ToStrMap(artifacts["s3"])["version"])

	os.Unsetenv("APP_VERSION")
}

func TestFetchEnvStacks(t *testing.T) {
	c := NewConfig(ResourcePath("stacks.yml"), tempOutputFinder())
	fake := &FakeDeploymentFinder{}
	c.Templ = NewTemplate(fake, c)

	os.Setenv("STACK_VERSION", "4.3.2")

	assert.NotNil(t, c)
	e := c.FetchEnvStacks("build.nagios-server")
	assert.NotNil(t, e)
	ep := e.Fetch("Endpoint")
	assert.NotNil(t, ep)

	assert.Equal(t, "nagios-elb-build-4.3.2", fake.StackName)
	assert.Equal(t, "ELBDNS", fake.OutputKey)

	assert.Equal(t, ep, "https://something")
	os.Unsetenv("STACK_VERSION")
}

func TestFetchEnvStacksNonExistingItem(t *testing.T) {
	c := NewConfig(ResourcePath("stacks.yml"), tempOutputFinder())
	fake := &FakeDeploymentFinder{}
	c.Templ = NewTemplate(fake, c)

	assert.NotNil(t, c)
	e := c.FetchEnvStacks("build.nagios-server")
	assert.NotNil(t, e)
	ep := e.Fetch("asdfasdf")
	assert.Nil(t, ep)
}

func TestFetchEnvStacksEsc(t *testing.T) {
	c := NewConfig(ResourcePath("stacks.yml"), tempOutputFinder())
	fake := &FakeDeploymentFinder{}
	c.Templ = NewTemplate(fake, c)

	assert.NotNil(t, c)
	e := c.FetchEnvStacks("build.nagios-server")
	assert.NotNil(t, e)
	ep := e.Fetch("E/P")
	assert.Equal(t, ep, "testing")
	ab := e.Fetch("A~B")
	assert.Equal(t, ab, "hello")
}

func TestFetchEnvStacksAll(t *testing.T) {
	c := NewConfig(ResourcePath("stacks.yml"), tempOutputFinder())
	fake := &FakeDeploymentFinder{}
	c.Templ = NewTemplate(fake, c)

	assert.NotNil(t, c)
	e := c.FetchEnvStacks("build")
	assert.NotNil(t, e)
	fmt.Printf("e.Stacks: %+v\n", e.Stacks)
	assert.Equal(t, len(e.Stacks), 2)
}

func TestDependsOn(t *testing.T) {
	log.SetLevel(log.DebugLevel)

	os.Setenv("PIPELINE_VERSION", "1.2.3")
	defer os.Unsetenv("PIPELINE_VERSION")

	c := NewConfig(ResourcePath("stacks_dag.yml"), tempOutputFinder())

	build := c.FetchEnvStacks("build")
	assert.Equal(t, []string{"nagios-elb", "nagios-server"}, build.StackLabels)

	// check label to name mapping
	elbStack := build.Stack("nagios-elb")
	assert.Equal(t, "nagios-elb-build-1-2-3", elbStack.Name())

	// dev will fail
	qa := c.FetchEnvStacks("qa")
	// no order constraints
	assert.Contains(t, qa.StackLabels, "nagios-elb")
	assert.Contains(t, qa.StackLabels, "nagios-server")
	qa2 := c.FetchEnvStacks("qa2")

	dns1 := indexInArray("nagios-internal-dns", qa2.StackLabels)
	dns2 := indexInArray("nagios-r53", qa2.StackLabels)
	app1 := indexInArray("nagios-elb", qa2.StackLabels)
	app2 := indexInArray("nagios-server", qa2.StackLabels)
	assert.True(t, dns1 != -1 && dns2 != -1 && app1 != -1 && app2 != -1)
	assert.True(t, dns1 < dns2, qa2.StackLabels)
	assert.True(t, app1 < app2, qa2.StackLabels)

	prod := c.FetchEnvStacks("prod")
	// no order constraints
	assert.Contains(t, prod.StackLabels, "nagios-elb")
	assert.Contains(t, prod.StackLabels, "nagios-server")
}

func indexInArray(key string, arr []string) int {
	result := -1
	for i, v := range arr {
		if v == key {
			return i
		}
	}
	return result
}
