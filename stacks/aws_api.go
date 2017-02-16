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
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/capitalone/stack-deployment-tool/providers"
	"github.com/capitalone/stack-deployment-tool/utils"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

const (
	max_wait_time = 15 * time.Minute
)

type AWSStackApi struct {
	providers.AWSApi
}

func NewAWSStackApi(api *providers.AWSApi) *AWSStackApi {
	return &AWSStackApi{AWSApi: *api}
}

func (a *AWSStackApi) FindDeploymentOutput(stackName string, outputKey string) (string, error) {
	log.Debugf("FindDeploymentOutput(%s, %s)", stackName, outputKey)
	stack := a.FindStack(stackName)
	if stack == nil {
		return "", fmt.Errorf("stack (%s) not found", stackName)
	}
	for _, o := range stack.Outputs {
		log.Debugf("stack %s output: %#v", stackName, o)
		if *o.OutputKey == outputKey {
			return *o.OutputValue, nil
		}
	}
	return "", fmt.Errorf("stack (%s) output key (%s) not found", stackName, outputKey)
}

func (a *AWSStackApi) FindStack(stackName string) *cloudformation.Stack {
	stackOutput, err := a.CFService().DescribeStacks(&cloudformation.DescribeStacksInput{StackName: &stackName})

	log.Debugf("stackOutput: %s\n", stackOutput)
	if err != nil {
		log.Errorf("Find Stack Error: %+v\n", err)
		return nil
	}

	if len(stackOutput.Stacks) == 0 {
		return nil
	}

	return stackOutput.Stacks[0]
}

func changeSetName(stackName string) string {
	return fmt.Sprintf("%s-%d", stackName, time.Now().Unix())
}

// TODO: wait for stack param
func (a *AWSStackApi) updateStack(stack *cloudformation.Stack, template string,
	parameters map[string]interface{}, tags map[string]interface{}) {

	stackName := *stack.StackName
	log.Debugf("updateStack stack: %s", stackName)

	cftags := cftTags(tags)
	log.Infof("CF Tags: %+v", cftags)

	cfparams := cftParams(parameters)
	log.Infof("CF Params: %+v", cfparams)

	// short-circuit in drymode
	if a.IsDryMode() {
		return
	}

	changeSetName := changeSetName(stackName)

	params := &cloudformation.CreateChangeSetInput{
		Capabilities:     stack.Capabilities,
		ChangeSetName:    aws.String(changeSetName),
		StackName:        aws.String(stackName),
		NotificationARNs: stack.NotificationARNs,
		Parameters:       cfparams,
		Tags:             cftags,
		TemplateBody:     aws.String(template),
	}

	resp, err := a.CFService().CreateChangeSet(params)
	if err != nil {
		log.Errorf("Error determining a changeset: %s", err)
		return
	}
	if resp.Id != nil {
		// wait for changeset to be created...
		a.waitForChangeSet(*resp.Id)

		a.printChangeSet(*resp.Id)
		_, err = a.CFService().ExecuteChangeSet(&cloudformation.ExecuteChangeSetInput{
			ChangeSetName: aws.String(*resp.Id),
		})

		if err != nil {
			log.Errorf("Error applying a changeset: %s", err)
		}

		err = a.waitForStackOperation(stackName)
	}
}

func (a *AWSStackApi) PrintChangesToStacks(envStacks *EnvStacksConfig) {
	log.Debugf("PrintChangesToStacks: %#v", envStacks.StackLabels)
	for _, stackLabel := range envStacks.StackLabels {
		//stackmap := ToStrMap(envStacks.Fetch(stackLabel))
		stack := envStacks.Stack(stackLabel)
		templateName := stackLabel
		stackmap := utils.ToStrMap(stack.FetchAll())
		if n, ok := stackmap["template"]; ok {
			templateName = n.(string)
		}
		p := filepath.Dir(envStacks.Config.FileName)
		template := a.loadTemplateJSON(filepath.Join(p, templateName), filepath.Join(p, stack.Name()))
		params := utils.ToStrMap(stackmap["parameters"])
		tags := utils.ToStrMap(stackmap["tags"])
		a.determineChangeSet(stack.Name(), template, params, tags)
	}
}

func (a *AWSStackApi) printChangeSet(changeSetName string) {
	params := &cloudformation.DescribeChangeSetInput{
		ChangeSetName: aws.String(changeSetName),
	}
	resp, err := a.CFService().DescribeChangeSet(params)
	if err != nil {
		log.Errorf("Error determining a changeset: %s", err)
		return
	}
	// make into a table?
	fmt.Printf("ChangeSet: %s\n", *resp.ChangeSetName)
	for i, chg := range resp.Changes {
		fmt.Printf("Change[%d]: %+v\n", i+1, *chg)
	}
}

func (a *AWSStackApi) determineChangeSet(stackName string, template string,
	parameters map[string]interface{}, tags map[string]interface{}) {

	cftags := cftTags(tags)
	log.Infof("CF Tags: %+v", cftags)

	cfparams := cftParams(parameters)
	log.Infof("CF Params: %+v", cfparams)

	changeSetName := fmt.Sprintf("%s-%d", stackName, time.Now().Unix())

	params := &cloudformation.CreateChangeSetInput{
		ChangeSetName: aws.String(changeSetName),
		StackName:     aws.String(stackName),
		//NotificationARNs: stack.NotificationARNs,
		Parameters:   cfparams,
		Tags:         cftags,
		TemplateBody: aws.String(template),
		// some rakefiles were default to DO_NOTHING, but i think we default to rollback..
		//OnFailure: aws.String(cloudformation.OnFailureRollback),
	}

	resp, err := a.CFService().CreateChangeSet(params)
	if err != nil {
		log.Errorf("Error determining a changeset: %s", err)
		return
	}
	if resp.Id != nil {
		a.printChangeSet(*resp.Id)
		a.deleteChangeSet(*resp.Id)
	}
	fmt.Printf("%#v\n", resp)
}

func (a *AWSStackApi) deleteChangeSet(changeSetName string) {
	params := &cloudformation.DeleteChangeSetInput{
		ChangeSetName: aws.String(changeSetName), // Required
	}
	_, err := a.CFService().DeleteChangeSet(params)
	if err != nil {
		log.Errorf("Error deleting a changeset: %s", err)
	}
}

func (a *AWSStackApi) CreateOrUpdateStacks(envStacks *EnvStacksConfig) {
	log.Debugf("Creating stacks: %#v", envStacks.StackLabels)
	for _, stackLabel := range envStacks.StackLabels {
		//stack := ToStrMap(envStacks.Fetch(stackName))
		stack := envStacks.Stack(stackLabel)
		stackmap := utils.ToStrMap(stack.FetchAll())
		templateName := stackLabel
		if n, ok := stackmap["template"]; ok {
			templateName = n.(string)
		}
		p := filepath.Dir(envStacks.Config.FileName)
		template := a.loadTemplateJSON(filepath.Join(p, templateName), filepath.Join(p, stack.Name()))
		params := utils.ToStrMap(stackmap["parameters"])
		tags := utils.ToStrMap(stackmap["tags"])

		existingStack := a.FindStack(stack.Name())
		if existingStack == nil {
			a.createStack(stack.Name(), template, params, tags)
		} else {
			a.updateStack(existingStack, template, params, tags)
		}

	}
	log.Info("Stacks Create Complete")
}

func (a *AWSStackApi) DeleteStacks(envStacks *EnvStacksConfig) {
	log.Debugf("Deleting stacks: %#v", envStacks.StackLabels)

	// short-circuit in drymode
	if a.IsDryMode() {
		return
	}

	index := len(envStacks.StackLabels)
	for index > 0 {
		label := envStacks.StackLabels[index-1]
		a.deleteStack(envStacks.Stack(label).Name())
		index--
	}
	log.Info("Stacks Delete Complete")
}

func (a *AWSStackApi) StacksStatus(envStacks *EnvStacksConfig) {
	log.Debugf("Stacks stacks: %#v", envStacks.StackLabels)

	tbl := utils.NewTableWriter(os.Stdout, 40, 50)
	tbl.WriteHeader("Stack", "Status")

	for _, stackLabel := range envStacks.StackLabels {
		stackName := envStacks.Stack(stackLabel).Name()
		stack := a.FindStack(stackName)
		if stack != nil {
			tbl.WriteRow(stackName, *stack.StackStatus)
		} else {
			tbl.WriteRow(stackName, "Not Found")
		}
	}
	tbl.Footer()
	fmt.Println()
}

func (a *AWSStackApi) loadTemplateJSON(templateNames ...string) string {
	var template string

	for _, templateName := range templateNames {
		// check for: .yml, .yaml, .json
		possibles := []string{templateName, templateName + ".yml", templateName + ".yaml",
			templateName + ".json", templateName + ".hjson"}

		for _, p := range possibles {
			log.Debugf("looking for template: %s", p)
			if utils.FileExists(p) {
				b, err := ioutil.ReadFile(p)
				if err != nil {
					continue
				}
				log.Debugf("Found template: %s", p)

				if filepath.Ext(p) == ".yml" || filepath.Ext(p) == ".yaml" {
					// we dont change the yaml very much because of the custom / local tags for CloudFormation
					utils.TemporaryChdir(filepath.Dir(p), func() {
						template = string(utils.ApplyIncludeFileLinesDirective(bytes.NewReader(b)))
					})
				} else {
					y, err := utils.DecodeHJSON(b)

					if err != nil {
						log.Errorf("Error loading file: %s %v", p, err)
						continue
					} else {
						utils.TemporaryChdir(filepath.Dir(p), func() {
							template = string(utils.EncodeJSON(utils.FnFileLinesInclude(y)))
						})
					}
				}

				break
			}
		}
		if len(template) > 0 {
			break
		}
	}

	return template
}

func cftTags(tags map[string]interface{}) []*cloudformation.Tag {
	cftags := []*cloudformation.Tag{}
	for k, v := range tags {
		cftags = append(cftags, &cloudformation.Tag{
			Key:   aws.String(k),
			Value: aws.String(v.(string)),
		})
	}
	return cftags
}

func cftParams(parameters map[string]interface{}) []*cloudformation.Parameter {
	cfparams := []*cloudformation.Parameter{}
	for k, v := range parameters {
		cfparams = append(cfparams, &cloudformation.Parameter{
			ParameterKey:   aws.String(k),
			ParameterValue: aws.String(v.(string)),
		})
	}
	return cfparams
}

// TODO: on failure param
// TODO: wait for stack param
func (a *AWSStackApi) createStack(stackName string, template string,
	parameters map[string]interface{}, tags map[string]interface{}) {

	log.Infof("createStack(%s)", stackName)
	log.Debugf("createStack(%s, %s, %#v, %#v)", stackName, template, parameters, tags)

	cftags := cftTags(tags)
	log.Infof("CF Tags: %+v", cftags)

	cfparams := cftParams(parameters)
	log.Infof("CF Params: %+v", cfparams)

	// short-circuit in drymode
	if a.IsDryMode() {
		return
	}

	params := &cloudformation.CreateStackInput{
		StackName: aws.String(stackName),
		//NotificationARNs: stack.NotificationARNs,
		Parameters:   cfparams,
		Tags:         cftags,
		TemplateBody: aws.String(template),
		// some rakefiles were default to DO_NOTHING, but i think we default to rollback..
		OnFailure: aws.String(cloudformation.OnFailureRollback),
	}

	resp, err := a.CFService().CreateStack(params)
	if err != nil {
		log.Errorf("Error creating stack: %+v", err)
		return
	}
	log.Infof("CreateStack Started: %s", resp)
	err = a.waitForStackOperation(stackName)
	if err != nil {
		log.Errorf("Error waiting for stack operation: %+v", err)
	}
}

func (a *AWSStackApi) deleteStack(stackName string) error {
	params := &cloudformation.DeleteStackInput{
		StackName:       aws.String(stackName), // Required
		RetainResources: []*string{
		//aws.String("LogicalResourceId"), // Required
		// More values...
		},
	}
	_, err := a.CFService().DeleteStack(params)
	if err != nil {
		log.Errorf("Error deleting stack: %s", stackName)
		return fmt.Errorf("Error deleting stack: %s %s", stackName, err)
	}
	err = a.waitForStackOperation(stackName)
	if err != nil && !strings.Contains(err.Error(), "does not exist") {
		log.Errorf("Error deleting stack: %s error: %s", stackName, err)
	}
	return nil
}

func (a *AWSStackApi) waitForStackOperation(stackName string) error {
	log.Infof("Waiting for stack operation to complete")

	params := &cloudformation.DescribeStackEventsInput{
		StackName: aws.String(stackName),
	}

	eventIdsSeen := make(map[string]interface{})

	startTime := time.Now()
	waitTime := startTime.Add(max_wait_time)

	tbl := utils.NewTableWriter(os.Stdout, 40, 45, 30)
	tbl.WriteHeader("Status", "Type", "LogicalID")

	var result error
	done := false
	var nextToken *string
	for time.Now().Before(waitTime) && !done {
		stack := a.FindStack(stackName)

		if stack == nil {
			return fmt.Errorf("stack: %s does not exist", stackName)
		}

		if strings.HasSuffix(*stack.StackStatus, "_FAILED") || strings.HasSuffix(*stack.StackStatus, "_COMPLETE") {
			if strings.HasSuffix(*stack.StackStatus, "_FAILED") {
				result = fmt.Errorf("Stack operation failed: %s", *stack.StackStatus)
			}
			if !a.isChangeSetPending(stackName) {
				done = true
			} else {
				log.Debugf("Changes Pending")
			}

		} else {
			time.Sleep(15 * time.Second)
		}

		params.NextToken = nextToken
		resp, err := a.CFService().DescribeStackEvents(params)
		if err != nil {
			return err
		}
		nextToken = resp.NextToken
		for i := len(resp.StackEvents) - 1; i >= 0; i-- {
			event := resp.StackEvents[i]
			if !utils.KeyExists(*event.EventId, eventIdsSeen) {
				eventIdsSeen[*event.EventId] = true
				tbl.WriteRow(*event.ResourceStatus, *event.ResourceType, *event.LogicalResourceId)
			}
		}
	}
	tbl.Footer()
	return result
}

func (a *AWSStackApi) isChangeSetPending(stackName string) bool {
	resp, err := a.CFService().ListChangeSets(&cloudformation.ListChangeSetsInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		log.Errorf("Error listing changeset: %v", err)
		return false
	}
	if resp != nil {
		for _, s := range resp.Summaries {
			log.Debugf("Change Status: %v %v", *s.ChangeSetId, *s.Status)
			if strings.HasSuffix(*s.Status, "_IN_PROGRESS") || strings.HasSuffix(*s.Status, "_PENDING") {
				return true
			}
		}
	}
	return false
}

func (a *AWSStackApi) waitForChangeSet(changeSetName string) {
	startTime := time.Now()
	waitTime := startTime.Add(max_wait_time)
	done := false

	for time.Now().Before(waitTime) && !done {
		params := &cloudformation.DescribeChangeSetInput{
			ChangeSetName: aws.String(changeSetName),
		}
		resp, err := a.CFService().DescribeChangeSet(params)
		if err != nil {
			log.Errorf("Error determining a changeset: %s", err)
			return
		}
		if resp != nil && !strings.HasSuffix(*resp.Status, "_IN_PROGRESS") && !strings.HasSuffix(*resp.Status, "_PENDING") {
			done = true
		} else {
			log.Infof("Waiting for change set: %s to be available: %s", changeSetName, *resp.Status)
			time.Sleep(15 * time.Second)
		}
	}
}
