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
	"bufio"
	"bytes"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/prometheus/common/log"
)

const (
	FN_INCLUDE_FILE_LINES = "Fn::Local::IncludeFileLines"
)

var (
	// This is the preferred way to include a file when using the yaml version of a CFT
	// Matches literal string: ${Local::IncludeFileLines file_content.txt}
	literalIncludeRe = regexp.MustCompile(`\${[ ]*Local::IncludeFile(Lines)?[ ]+([[:ascii:]]+)}`)

	// Matches tag: "Fn::Base64": !Local::IncludeFileLines file_content.txt
	valueTagRe = regexp.MustCompile(`!Local::IncludeFile(Lines)?[ ]+([[:ascii:]]+)`)
)

func init() {
	AddDirective(ApplyIncludeFileLinesDirective)
}

func ApplyIncludeFileLinesDirective(reader io.Reader) []byte {
	output := bytes.NewBuffer([]byte{})
	scanner := bufio.NewScanner(reader)
	scanner.Split(CustomScanLines)
	for scanner.Scan() {
		line := scanner.Bytes()
		loc := literalIncludeRe.FindIndex(line)
		tagLoc := valueTagRe.FindIndex(line)
		if len(loc) == 2 {
			fileName := strings.TrimSpace(string(literalIncludeRe.FindSubmatch(line)[2]))
			log.Debugf("loading include file: %s", fileName)
			output.Write(indentedFileLines(loc[0], fileName))
		} else if len(tagLoc) == 2 {
			// case for tag value include
			fileName := strings.TrimSpace(string(valueTagRe.FindSubmatch(line)[2]))
			log.Debugf("loading include file: %s", fileName)
			spaces := indentation(line)
			output.Write(line[0:tagLoc[0]])
			// TODO: AWS CFT specific
			output.WriteString("!Sub |\n") // append the aws sub function tag..
			output.Write(indentedFileLines(spaces+2, fileName))
		} else {
			output.Write(line)
		}
	}

	return output.Bytes()
}

func indentation(line []byte) int {
	spaces := 0
	for _, b := range line {
		if b == ' ' {
			spaces++
		}
	}
	return spaces
}

func indentedFileLines(indentDepth int, filename string) []byte {
	lines := bytes.NewBuffer([]byte{})

	f, err := os.Open(filename)
	if err != nil {
		log.Errorf("Error opening file: %v", err)
		return lines.Bytes()
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		cmd := []byte(scanner.Text() + "\n")
		lines.Write(bytes.Repeat([]byte{' '}, indentDepth))
		lines.Write(cmd)
	}

	return lines.Bytes()
}
