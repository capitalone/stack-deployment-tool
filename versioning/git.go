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
package versioning

import (
	"bytes"
	"os/exec"
	"strings"

	log "github.com/Sirupsen/logrus"
)

func GetBranch() string {
	result, _ := Sh("git", "rev-parse", "--abbrev-ref", "HEAD")
	log.Debugf("GetBranch: %s\n", result)
	return result
}

func GitHash() string {
	result, _ := Sh("git", "log", "-n", "1", "--pretty=format:%H")
	log.Debugf("GitHash: %s\n", result)
	return result
}

func GitLatestTag() string {
	result, _ := Sh("git", "describe", "--tags", "--abbrev=0")
	log.Debugf("GitLatestTag: %s\n", result)
	return result
}

func MustSh(name string, arg ...string) string {
	o, err := Sh(name, arg...)
	if err != nil {
		log.Fatal(err)
	}
	return o
}

func Sh(name string, arg ...string) (string, error) {
	cmd := exec.Command(name, arg...)
	cmd.Stdin = strings.NewReader("")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return out.String(), err
}
