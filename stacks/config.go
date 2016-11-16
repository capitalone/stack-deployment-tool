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
	"reflect"
	"regexp"
	"strings"

	"github.com/capitalone/stack-deployment-tool/utils"

	log "github.com/Sirupsen/logrus"
	jsonptr "github.com/dustin/go-jsonpointer"
	dag "github.com/hashicorp/terraform/dag"
)

type Fetcher interface {
	Fetch(item string) interface{}
	FetchJsonPtr(jsonPtr string) interface{}
	FetchAll() interface{}
}

type StacksConfig struct {
	Yaml     map[string]interface{}
	Templ    *Template
	FileName string
}

type EnvStacksConfig struct {
	Yaml        map[string]interface{} // scoped yaml to the env
	Env         string
	StackLabels []string
	Stacks      map[string]StackConfig // stack id to stackconfig mapping
	Config      *StacksConfig
}

type StackConfig struct {
	Yaml   map[string]interface{} // scoped yaml to the stack
	label  string
	name   string
	Config *StacksConfig
}

func NewConfig(yamlFilePath string, outputFinder DeploymentOutputFinder) (c *StacksConfig) {
	c = &StacksConfig{FileName: yamlFilePath}
	c.Templ = NewTemplate(outputFinder, c)
	out, err := utils.DecodeYAMLFile(yamlFilePath)
	if err != nil {
		log.Fatalf("Cannot load stack config: %s", yamlFilePath)
	}
	if out == nil {
		log.Fatalf("Template empty")
	}

	c.Yaml = out

	return c
}

func (c *StacksConfig) Fetch(item string) interface{} {
	return c.ProcessValue(fetchValue(c.Yaml, item))
}

func (c *StacksConfig) FetchJsonPtr(jsonPtr string) interface{} {
	return c.ProcessValue(jsonptr.Get(c.Yaml, jsonPtr))
}

func (c *StacksConfig) FetchAll() interface{} {
	return c.ProcessValue(c.Yaml)
}

// "build.nagios-elb" or "build[nagios-elb,nagios-app]"
//                    or just "build" - which implies all stacks
func (c *StacksConfig) FetchEnvStacks(stackRef string) *EnvStacksConfig {
	stackEnvs := regexp.MustCompile("\\.|\\[").Split(strings.Replace(stackRef, "]", "", 1), -1)

	if len(stackEnvs) < 1 {
		return nil
	}

	env := stackEnvs[0]
	st := jsonptr.Get(c.Yaml, "/stacks/"+escJsonPtr(env))
	if st == nil {
		log.Fatalf("Environment: %s not found", env)
	}
	envStackYaml := utils.ToStrMap(st)

	if len(stackEnvs) == 1 { // only env, so add all the stacks for this env.
		for k, v := range envStackYaml {
			if reflect.TypeOf(v).Kind() == reflect.Map {
				stackEnvs = append(stackEnvs, k)
			}
		}
	}

	var stackLabels []string
	stacks := make(map[string]StackConfig)

	for _, s := range stackEnvs[1:] {
		log.Debugf("looking for stack: %+v", s)
		stackLabels = append(stackLabels, s)
		stackYaml := jsonptr.Get(envStackYaml, "/"+escJsonPtr(s))
		if stackYaml != nil {
			log.Debugf("newStackConfig: %+v", s)
			stacks[s] = *newStackConfig(s, utils.ToStrMap(stackYaml), c)
		}
	}

	stackLabels = orderedArray(stackLabels, depsGraph(stacks))
	log.Debugf("stackLabels: %#v", stackLabels)
	return &EnvStacksConfig{
		Yaml: envStackYaml, // scoped yaml
		Env:  env, StackLabels: stackLabels, Config: c, Stacks: stacks}
}

func orderedArray(stackLabels []string, deps *dag.AcyclicGraph) []string {
	var result []string
	root, err := deps.Root()
	// stacks that need to be ordered, go first, then everything else.
	if err == nil {
		rootName := root.(dag.NamedVertex).Name()
		log.Debugf("root: %+v\n", rootName)
		// walk children
		result = childrenNames(root, deps)
	} else {
		log.Errorf("Error walking DAG: %#v", err)
	}
	log.Debugf("orderedArray result: %#v\n", result)
	return result
}

type LabeledVertex interface {
	Label() string
}

type RootNamedVertex struct {
}

func (r *RootNamedVertex) Name() string {
	return "ROOT"
}

func depsGraph(stacks map[string]StackConfig) *dag.AcyclicGraph {
	root := RootNamedVertex{}
	// order stack names based on DAG
	depsGraph := &dag.AcyclicGraph{}
	depsGraph.Add(&root)
	for _, st := range stacks {
		s := st
		depsGraph.Add(&s)
		// depends on maps to the stack label
		for _, d := range s.dependsOn() {
			src := stacks[d]
			if s.Label() != d { // dont add a connection to myself.
				e := dag.BasicEdge(&src, &s)
				depsGraph.Add(&src)
				depsGraph.Connect(e)
			}
		}
		if len(s.dependsOn()) == 0 {
			// add to the root
			depsGraph.Connect(dag.BasicEdge(&root, &s))
		}
	}
	// make sure there are not multiple roots..
	for _, v := range depsGraph.Vertices() {
		if v == &root {
			continue
		}
		if depsGraph.UpEdges(v).Len() == 0 {
			depsGraph.Connect(dag.BasicEdge(&root, v))
		}
	}

	err := depsGraph.Validate()
	if err != nil {
		log.Fatalf("dependencies invalid, %+v", err)
	}
	depsGraph.TransitiveReduction()
	return depsGraph
}

// built in methods dont walk correctly because they use Set's that are built on maps with no order guarantees..
func childrenNames(vertex dag.Vertex, deps *dag.AcyclicGraph) []string {
	var verts []string
	start := dag.AsVertexList(deps.DownEdges(vertex))
	memoFunc := func(v dag.Vertex, d int) error {
		verts = append(verts, v.(LabeledVertex).Label())
		return nil
	}

	if err := deps.DepthFirstWalk(start, memoFunc); err != nil {
		log.Errorf("Error finding children: %+v", err)
	}
	return verts
}

func (s *StacksConfig) ProcessValue(val interface{}) interface{} {
	var result interface{}
	rval := reflect.ValueOf(val)
	switch rval.Kind() {
	case reflect.Map:
		mapval := utils.ToStrMap(val)
		for k, v := range mapval {
			mapval[k] = s.ProcessValue(v)
		}
		result = mapval
	case reflect.Array:
		arrval := val.([]interface{})
		for i, v := range arrval {
			arrval[i] = s.ProcessValue(v)
		}
		result = arrval
	case reflect.String:
		result = s.Templ.Render(rval.String())
	default:
		result = val
	}
	log.Debugf("ProcessValue result: %+v", result)
	return result
}

// EnvStacksConfig

func (e *EnvStacksConfig) Fetch(item string) interface{} {
	return e.Config.ProcessValue(jsonptr.Get(e.Yaml, fmt.Sprintf("/%s", escJsonPtr(item))))
}

func (e *EnvStacksConfig) FetchAll() interface{} {
	return e.Config.ProcessValue(e.Yaml)
}

func (e *EnvStacksConfig) Stack(stackLabel string) *StackConfig {
	if s, ok := e.Stacks[stackLabel]; ok {
		return &s
	}
	return nil
}

// StackConfig

func newStackConfig(label string, yaml map[string]interface{}, c *StacksConfig) *StackConfig {
	name := label
	if utils.KeyExists("stack_name", yaml) {
		val := c.ProcessValue(fetchValue(yaml, "stack_name"))
		if val != nil {
			name = val.(string)
		}
	}
	return &StackConfig{
		Config: c,
		Yaml:   yaml,
		label:  label,
		name:   name,
	}
}

func (s *StackConfig) Fetch(item string) interface{} {
	return s.Config.ProcessValue(fetchValue(s.Yaml, item))
}

func (s *StackConfig) FetchAll() interface{} {
	return s.Config.ProcessValue(s.Yaml)
}

func (s *StackConfig) Label() string {
	return s.label
}

func (s *StackConfig) Name() string {
	return s.name
}

func (s *StackConfig) Hashcode() interface{} {
	return s.Label() // label is unique
}

func (s *StackConfig) dependsOn() []string {
	switch deps := s.Yaml["depends_on"].(type) {
	case string:
		return []string{deps}
	case []string:
		return deps
	default:
		return []string{}
	}
}

// Map & Value Utils

func escJsonPtr(item string) string {
	return strings.Replace(strings.Replace(item, "~", "~0", -1), "/", "~1", -1)
}

func fetchValue(yaml map[string]interface{}, item string) interface{} {
	if v, ok := yaml[item]; ok {
		return v
	}
	return nil
}
