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
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/capitalone/stack-deployment-tool/utils"
	"github.com/capitalone/stack-deployment-tool/versioning"

	log "github.com/Sirupsen/logrus"
	"github.com/aymerick/raymond"
)

const (
	TEMPLATE_CTX_KEY = "_templateCtx"
)

var (
	helpersCtx map[string]interface{} = make(map[string]interface{}) // helper -> ctx
)

//
// Supported Template Parameters:
// env.BUILD_NUMBER - the build number env variable or the user id + timestamp
// env.LATEST_TAG - the latest git tag
// env.ARTIFACT_VERSION - the version from the build.properties (see: versioning.rb)
// env.STACK_VERSION - stack friendly artifact version string (using - instead of .)
// env.<ANY ENV VARIABLE> - so anything from ENV[] can be referenced
//   optional default value for the env value, for example:
//   CreatedByURL: '{{env.BUILD_URL default="NA"}}'
// output stack=<stack name> key=<output key to pull value from>  - use the output value from one stack
// s3artifact repo=<one of the valid artifact repos, default: sandbox>
// pipeline_version - PIPELINE_VERSION environment variable
//                    shortcut for: env.PIPELINE_VERSION default="NA"
// from_yaml - reference a yaml key
// All of these should be enclosed in a mustache style template within single-quotes for yaml
// i.e.   version: '{{env.ARTIFACT_VERSION}}'
// and parenthesis for subexpression
// i.e.  {{sum 1 (sum 1 1)}}
//

type DeploymentOutputFinder interface {
	FindDeploymentOutput(stackName string, outputKey string) (string, error)
}

func RegisterTemplateHelper(cmd string, ctx interface{}, helper interface{}) {
	helpersCtx[cmd] = ctx
	raymond.RegisterHelper(cmd, helper)
}

type Template struct {
	OutputFinder DeploymentOutputFinder
	YamlFetcher  Fetcher
	HelpersCtx   map[string]interface{} // helper -> ctx
}

// TODO: move these to render?
func NewTemplate(outputFinder DeploymentOutputFinder, yamlFetcher Fetcher) *Template {
	return &Template{
		OutputFinder: outputFinder,
		YamlFetcher:  yamlFetcher,
		HelpersCtx:   helpersCtx,
	}
}

func (t *Template) Render(input string) string {
	s, err := t.renderTemplate(raymond.MustParse(string(input)))
	if err != nil {
		log.Errorf("%v", err)
		return ""
	}
	return s
}

func init() {
	// can only register these once, and they get all info from the ctx since they are reused
	raymond.RegisterHelper("output", outputHelper)
	raymond.RegisterHelper("pipeline_version", func(options *raymond.Options) raymond.SafeString {
		return raymond.SafeString(utils.GetenvWithDefault("PIPELINE_VERSION", "NA"))
	})
	raymond.RegisterHelper("s3artifact", artifactHelper)
	raymond.RegisterHelper("from_yaml", fromYamlHelper)
	raymond.RegisterHelper("env", envHelper)
	raymond.RegisterHelper("timestamp", timestampHelper)
	raymond.RegisterHelper("escstackname", escNameHelper)

	// initialize default env variables
	initDefaultEnvs()
}

func initDefaultEnvs() {
	utils.SetenvIfAbsent("LATEST_TAG", versioning.GitLatestTag())
	utils.SetenvIfAbsent("BUILD_NUMBER", fmt.Sprintf("%s%s", utils.GetenvWithDefault("USER", "unknown"), timeNowInUTC()))
	v := versioning.LoadVersion()
	utils.SetenvIfAbsent("ARTIFACT_VERSION", v.String())
	utils.SetenvIfAbsent("STACK_VERSION", utils.EscapeStackVer(v.String()))
}

func ctxMap(options *raymond.Options) map[string]interface{} {
	return utils.ToStrMap(options.Ctx())
}

func CtxTemplate(options *raymond.Options) *Template {
	return ctxMap(options)[TEMPLATE_CTX_KEY].(*Template)
}

func HelperCtx(options *raymond.Options, helperCmd string) interface{} {
	val, ok := CtxTemplate(options).HelpersCtx[helperCmd]
	if !ok {
		return nil
	}
	return val
}

func escNameHelper(options *raymond.Options) raymond.SafeString {
	return raymond.SafeString(utils.EscapeStackVer(options.Fn()))
}

func fromYamlHelper(jsonPtr string, options *raymond.Options) raymond.SafeString {
	if jsonPtr[0:1] != "/" {
		jsonPtr = "/" + jsonPtr
	}
	log.Debugf("yaml_from %s", jsonPtr)
	result := CtxTemplate(options).YamlFetcher.FetchJsonPtr(jsonPtr)
	if result != nil {
		conv, err := json.Marshal(result)
		if err == nil {
			return raymond.SafeString(conv)
		}
	}
	return raymond.SafeString("")
}

func artifactHelper(options *raymond.Options) raymond.SafeString {
	// TODO
	return raymond.SafeString("")
}

func timeNowInUTC() string {
	return fmt.Sprintf("%d", time.Now().UTC().Unix())
}

func timestampHelper(options *raymond.Options) raymond.SafeString {
	return raymond.SafeString(timeNowInUTC())
}

func envHelper(options *raymond.Options) raymond.SafeString {
	key := options.HashStr("key")
	defaultVal := options.HashStr("default")
	log.Debugf("envHelper(%s) default: %s", key, defaultVal)
	return raymond.SafeString(utils.GetenvWithDefault(key, defaultVal))
}

func outputHelper(options *raymond.Options) raymond.SafeString {
	stackNameTempl := options.HashStr("stack")
	key := options.HashStr("key")
	stackName, err := raymond.MustParse(stackNameTempl).Exec(options.Ctx())
	if err != nil {
		log.Fatalf("Error parsing: %s\n", stackNameTempl)
	}
	log.Debugf("looking for %s %s\n", stackName, key)
	// find the Stack output
	outputFinder := CtxTemplate(options).OutputFinder
	log.Debugf("outputFinder: %#v", outputFinder)
	if outputFinder != nil {
		val, err := outputFinder.FindDeploymentOutput(stackName, key)
		if err != nil {
			log.Fatalf("Error finding stack output: %s\n", key)
		}
		// cache the result.
		// ctx [cache key] = val
		log.Debugf("found %s %s = %s\n", stackName, key, val)

		return raymond.SafeString(val)
	}
	return raymond.SafeString("")
}

func EnvKeys() (result []string) {
	e := os.Environ() // "key=value"
	for _, ev := range e {
		k := strings.Split(ev, "=")[0]
		result = append(result, k)
	}
	return result
}

func (t *Template) renderTemplate(template *raymond.Template) (string, error) {
	ctx := map[string]interface{}{
		"eid":            EnvValueMap("USER"), // convenience mapping
		"user":           EnvValueMap("USER"), // convenience mapping
		"env":            EnvValueMap(EnvKeys()...),
		TEMPLATE_CTX_KEY: t,
	}

	//log.Debugf("template ctx: %#v", ctx)

	str, err := template.Exec(ctx)
	if err != nil {
		log.Errorf("Error rendering template: %+v\n", err)
	}

	return str, err
}

func GetenvFunc(key string) func(*raymond.Options) string {
	return func(options *raymond.Options) string {
		//		fmt.Printf("GetenvFunc(%s): %+v\n", key, options)
		// support for default values
		if hashVal, found := options.Hash()["default"]; found {
			log.Debugf("GetenvFunc(%s) default: %s", key, hashVal)
			return utils.GetenvWithDefault(key, raymond.Str(hashVal))
		}
		log.Debugf("GetenvFunc(%s) not default", key)
		return os.Getenv(key)
	}
}

func EnvValueMap(keys ...string) map[string]interface{} {
	result := make(map[string]interface{})
	for _, k := range keys {
		result[k] = GetenvFunc(k)
	}
	return result
}
