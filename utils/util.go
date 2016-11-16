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
package utils

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect"

	log "github.com/Sirupsen/logrus"
	"github.com/hjson/hjson-go"
	"gopkg.in/yaml.v2"
)

func FileExists(filename string) bool {
	if _, err := os.Stat(filename); err == nil {
		return true
	}
	return false
}

// dir functions
func TemporaryChdir(dir string, runItInDir func()) {
	curDir, err := os.Getwd()
	os.Chdir(dir)
	defer func() {
		if err == nil {
			os.Chdir(curDir)
		}
	}()
	runItInDir()
}

// map conversion functions

func ToStrMap(in interface{}) map[string]interface{} {
	if in == nil || (reflect.ValueOf(in).Kind() == reflect.Interface && reflect.ValueOf(in).IsNil()) {
		return make(map[string]interface{})
	}
	return in.(map[string]interface{})
}

func KeyExists(key string, input map[string]interface{}) bool {
	if _, ok := input[key]; ok {
		return true
	}
	return false
}

func DeepToStrMap(input map[string]interface{}) map[string]interface{} {
	for k, v := range input {
		input[k] = toStrItem(v)
	}
	return input
}

func shallowToStrMap(input map[interface{}]interface{}) map[string]interface{} {
	strmap := make(map[string]interface{})
	for mk, mv := range input {
		strmap[mk.(string)] = mv
	}
	return strmap
}

func toStrItem(item interface{}) interface{} {
	result := item
	if m, ok := item.(map[interface{}]interface{}); ok {
		strmap := shallowToStrMap(m)
		DeepToStrMap(strmap)
		result = strmap
	} else if a, ok := item.([]interface{}); ok {
		for i, item := range a {
			a[i] = toStrItem(item)
		}
	}
	return result
}

// ENV funcs

func GetenvWithDefault(key string, defaultValue string) string {
	val := os.Getenv(key)
	if len(val) == 0 {
		return defaultValue
	}
	return val
}

func SetenvIfAbsent(key, value string) {
	if _, found := os.LookupEnv(key); !found {
		os.Setenv(key, value)
	}
}

// YAML Utils

func DecodeYAMLFile(yamlFilePath string) (map[string]interface{}, error) {
	yamlInput, err := ioutil.ReadFile(yamlFilePath)
	if err != nil {
		return nil, err
	}
	return DecodeYAML(yamlInput)
}

func DecodeYAML(yamlInput []byte) (map[string]interface{}, error) {
	document := new(map[string]interface{})
	err := yaml.Unmarshal(yamlInput, document)
	if err != nil {
		log.Fatalf("Error parsing yaml: %+v\n", err)
	}

	if document != nil {
		return DeepToStrMap(*document), err
	}
	return nil, err
}

func DecodeHJSON(hjsonInput []byte) (map[string]interface{}, error) {
	document := new(map[string]interface{})
	err := hjson.Unmarshal(hjsonInput, document)
	if err != nil {
		log.Fatalf("Error parsing hjson: %+v\n", err)
	}

	if document != nil {
		return DeepToStrMap(*document), err
	}
	return nil, err
}

func EncodeYAML(data interface{}) []byte {
	d, err := yaml.Marshal(&data)
	if err != nil {
		log.Fatalf("Error marshalling to yaml: %+v", err)
	}
	return d
}

func EncodeJSON(document interface{}) []byte {
	jsonDoc, err := json.Marshal(document)
	if err != nil {
		log.Debugf("document: %#v", document)
		log.Fatalf("Error generating json: %+v", err)
	}
	return jsonDoc
}

func GenerateJSONFromYaml(yamlInput []byte) []byte {
	document, err := DecodeYAML(yamlInput)
	if err != nil {
		log.Fatalf("Error decoding yaml: %+v", err)
		return []byte{}
	}
	return EncodeJSON(document)
}

func EscapeStackVer(version string) string {
	ver := []byte(version)
	for i, c := range ver {
		if (c <= 'Z' && c >= 'A') || (c <= 'z' && c >= 'a') || (c <= '9' && c >= '0') {
			continue
		} else {
			ver[i] = '-'
		}
	}
	return string(ver)
}

func FnFileLinesInclude(input interface{}) interface{} {
	if m, ok := input.(map[string]interface{}); ok {
		if lines, found := fnFileLines(m); found {
			input = lines
		} else {
			for key, val := range m {
				m[key] = FnFileLinesInclude(val)
			}
		}
	} else if a, ok := input.([]interface{}); ok {
		// start from the back since we are modifying the array at & after this index
		for i := len(a) - 1; i >= 0; i-- {
			item := a[i]
			if m, ok := item.(map[string]interface{}); ok {
				if lines, found := fnFileLines(m); found {
					a[i] = "" // make it empty for now, clean after iteration
					if len(lines) > 0 {
						a = InsertAtIndex(a, i, lines...) // expand out the array
					}
					// clean up the empty node that has moved after the lines
					a = DeleteAtIndex(a, len(lines)+i)
				} else {
					a[i] = FnFileLinesInclude(m)
				}
			} else {
				a[i] = FnFileLinesInclude(a[i])
			}
		}
		input = a // assign the updated slice
	}

	return input
}

func toStrSlice(input []interface{}) []string {
	result := make([]string, len(input))
	for i, v := range input {
		result[i] = v.(string)
	}
	return result
}

func toIntfSlice(input []string) []interface{} {
	result := make([]interface{}, len(input))
	for i, v := range input {
		result[i] = v
	}
	return result
}

func fnFileLines(input map[string]interface{}) ([]interface{}, bool) {
	var lines []interface{}
	for key, val := range input {
		if key == FN_INCLUDE_FILE_LINES {
			filename := val.(string)
			lines := FileLines(filename, true)
			return lines, true
		}
	}

	return lines, false
}

func FileLines(filename string, sub bool) []interface{} {
	var lines []interface{}
	f, err := os.Open(filename)

	if err != nil {
		log.Errorf("Error opening file: %v", err)
		return lines
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		cmd := scanner.Text() + "\n"
		if sub {
			line := map[string]string{"Fn::Sub": cmd}
			lines = append(lines, line)
		} else {
			lines = append(lines, cmd)
		}
	}

	return lines
}

// string array insert
func InsertStrAtIndex(into []string, index int, val ...string) []string {
	shift := len(val)
	for i := 0; i < shift; i++ {
		into = append(into, "") // placeholder
	}
	copy(into[index+shift:], into[index:]) // shift everything down to make room
	copy(into[index:], val)
	return into
}

func InsertAtIndex(into []interface{}, index int, val ...interface{}) []interface{} {
	shift := len(val)
	for i := 0; i < shift; i++ {
		into = append(into, "") // placeholder
	}
	copy(into[index+shift:], into[index:]) // shift everything down to make room
	copy(into[index:], val)
	return into
}

func DeleteAtIndex(into []interface{}, index int) []interface{} {
	intoLen := len(into)
	if index+1 < intoLen {
		copy(into[index:], into[index+1:])
	}
	into[intoLen-1] = nil // or the zero value of T
	return into[:intoLen-1]
}

// min/max

func MinInt(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func MaxInt(x, y int) int {
	if x > y {
		return x
	}
	return y
}
