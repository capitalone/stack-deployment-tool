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
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	stacksIncludeFile = "stacks_include.json"
)

func TestEscapeVer(t *testing.T) {
	s := EscapeStackVer("0.0.1-preview+alsdkfj")
	assert.Equal(t, "0-0-1-preview-alsdkfj", s)
}

func TestMinMax(t *testing.T) {
	assert.Equal(t, MinInt(1, 2), 1)
	assert.Equal(t, MinInt(2, 2), 2)
	assert.Equal(t, MaxInt(2, 2), 2)
	assert.Equal(t, MaxInt(3, 2), 3)
}

func TestToStrMap(t *testing.T) {
	m := make(map[string]interface{})
	m["params"] = nil
	assert.NotNil(t, ToStrMap(m["params"]))
	assert.Equal(t, len(ToStrMap(m["params"])), 0)
	m["params"] = "here"
	assert.NotNil(t, ToStrMap(m))
	assert.Equal(t, len(ToStrMap(m)), 1)
}

func TestFileLines(t *testing.T) {
	lines := FileLines("../resources/"+"file_content.txt", false)
	joined := strings.Join(toStrSlice(lines), "")
	expected := "asdfoasdf\noqoruiqweour\necho 'Blue' > /usr/share/nginx/html/index.html\necho \"Green\" > /usr/share/nginx/html/index.html\n"
	assert.Equal(t, expected, joined)

	lines = FileLines("../resources/"+"file_content.txt", true)
	expectedMap := []interface{}{
		map[string]string{"Fn::Sub": "asdfoasdf\n"},
		map[string]string{"Fn::Sub": "oqoruiqweour\n"},
		map[string]string{"Fn::Sub": "echo 'Blue' > /usr/share/nginx/html/index.html\n"},
		map[string]string{"Fn::Sub": "echo \"Green\" > /usr/share/nginx/html/index.html\n"},
	}
	assert.Equal(t, expectedMap, lines)
}

func TestDeleteAtIndexInSlice(t *testing.T) {
	a := []interface{}{"a", "b", "c", "d"}
	b := DeleteAtIndex(a, 1)
	assert.Equal(t, []interface{}{"a", "c", "d"}, b)
	assert.Equal(t, 3, len(b))

	a = []interface{}{"a", "b", "c", "d"}
	d := DeleteAtIndex(a, 3)
	assert.Equal(t, []interface{}{"a", "b", "c"}, d)
	assert.Equal(t, 3, len(d))
}

func TestHJsonHandling(t *testing.T) {
	fname := "../resources/" + stacksIncludeFile
	input, err := ioutil.ReadFile(fname)
	m, err := DecodeHJSON(input)
	assert.Nil(t, err)
	out := EncodeJSON(m)
	assert.NotNil(t, out)
}

func TestSpecialFnIncludeLines(t *testing.T) {
	fname := "../resources/" + stacksIncludeFile
	y, err := DecodeYAMLFile(fname)
	assert.Nil(t, err)

	var result interface{}
	TemporaryChdir(filepath.Dir(fname), func() {
		result = FnFileLinesInclude(y)
	})

	cftJson := GenerateJSONFromYaml(EncodeYAML(result))

	/*
		valid := `{"File":["asdfoasdf\n","oqoruiqweour\n","echo 'Blue' \u003e /usr/share/nginx/html/index.html\n","echo \"Green\" \u003e /usr/share/nginx/html/index.html\n"],"UserData":{"Fn::Base64":{"Fn::Join":["",["#!/usr/bin/env bash\n","asdfoasdf\n","oqoruiqweour\n","echo 'Blue' \u003e // /usr/share/nginx/html/index.html\n","echo \"Green\" \u003e /usr/share/nginx/html/index.html\n","","apt-get update -y\n"]]}}}`
	*/
	valid := "{\"File\":[{\"Fn::Sub\":\"asdfoasdf\\n\"},{\"Fn::Sub\":\"oqoruiqweour\\n\"},{\"Fn::Sub\":\"echo 'Blue' \\u003e /usr/share/nginx/html/index.html\\n\"},{\"Fn::Sub\":\"echo \\\"Green\\\" \\u003e /usr/share/nginx/html/index.html\\n\"}],\"UserData\":{\"Fn::Base64\":{\"Fn::Join\":[\"\",[\"#!/usr/bin/env bash\\n\",{\"Fn::Sub\":\"asdfoasdf\\n\"},{\"Fn::Sub\":\"oqoruiqweour\\n\"},{\"Fn::Sub\":\"echo 'Blue' \\u003e /usr/share/nginx/html/index.html\\n\"},{\"Fn::Sub\":\"echo \\\"Green\\\" \\u003e /usr/share/nginx/html/index.html\\n\"},\"apt-get update -y\\n\"]]}}}"

	assert.Equal(t, valid, string(cftJson))
}

func TestSpecialFnIncludeLinesMissingFile(t *testing.T) {

	fname := "../resources/" + stacksIncludeFile
	y, err := DecodeYAMLFile(fname)
	assert.Nil(t, err)

	result := FnFileLinesInclude(y)

	cftJson := GenerateJSONFromYaml(EncodeYAML(result))
	//fmt.Printf("%s", cftJson)
	valid := `{"File":[],"UserData":{"Fn::Base64":{"Fn::Join":["",["#!/usr/bin/env bash\n","apt-get update -y\n"]]}}}`

	assert.Equal(t, valid, string(cftJson))

}
