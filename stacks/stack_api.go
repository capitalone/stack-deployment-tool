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
package stacks

import (
	"strings"

	"github.com/capitalone/stack-deployment-tool/providers"
)

type StackApi interface {
	DeploymentOutputFinder
	CreateOrUpdateStacks(envStacks *EnvStacksConfig)
	DeleteStacks(envStacks *EnvStacksConfig)
	StacksStatus(envStacks *EnvStacksConfig)
	PrintChangesToStacks(envStacks *EnvStacksConfig)

	DryMode(enable bool)
}

func DefaultStackApi() StackApi {
	return GetStackApi("aws")
}

func GetStackApi(stackType string) StackApi {
	var api StackApi
	switch strings.ToLower(stackType) {
	case "aws":
		api = NewAWSStackApi(providers.NewAWSApi())
	}
	return &ScriptRunnerStackProxy{api: api}
}

// proxy to run scripts for stack apis

type ScriptRunnerStackProxy struct {
	api StackApi
}

func (p *ScriptRunnerStackProxy) FindDeploymentOutput(stackName string, outputKey string) (string, error) {
	return p.api.FindDeploymentOutput(stackName, outputKey)
}

func (p *ScriptRunnerStackProxy) CreateOrUpdateStacks(envStacks *EnvStacksConfig) {
	p.api.CreateOrUpdateStacks(envStacks)
}

func (p *ScriptRunnerStackProxy) DeleteStacks(envStacks *EnvStacksConfig) {
	p.api.DeleteStacks(envStacks)
}
func (p *ScriptRunnerStackProxy) StacksStatus(envStacks *EnvStacksConfig) {
	p.api.StacksStatus(envStacks)
}
func (p *ScriptRunnerStackProxy) PrintChangesToStacks(envStacks *EnvStacksConfig) {
	p.api.PrintChangesToStacks(envStacks)
}

func (p *ScriptRunnerStackProxy) DryMode(enable bool) {
	p.api.DryMode(enable)
}
