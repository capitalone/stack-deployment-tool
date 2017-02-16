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
package versioning

import (
	"fmt"
	"io/ioutil"
	"strings"

	semver "github.com/blang/semver"
)

const (
	VERSION_PROPS = "version.properties"
	BUILD_PROPS   = "build.properties"
)

type Version struct {
	semver.Version
	raw string
}

func LoadVersion() *Version {
	return LoadVersionProps(VERSION_PROPS)
}

func LoadVersionProps(filename string) *Version {
	b, err := ioutil.ReadFile(filename)
	result := "0.0.0"
	if err == nil {
		result = strings.TrimSpace(strings.Replace(string(b), "version=", "", -1))
	}
	return New(result)
}

func (v *Version) WriteVersion() {
	v.WriteVersionProps(VERSION_PROPS)
}

func (v *Version) WriteVersionProps(filename string) {
	ioutil.WriteFile(filename, []byte(v.ToPropString()), 0666)
}

func New(version string) *Version {
	v := semver.MustParse(version)
	return &Version{raw: version, Version: v}
}

func (v *Version) ToPropString() string {
	return fmt.Sprintf("version=%s", v.String())
}

func (v *Version) String() string {
	return v.Version.String()
}

func (v *Version) BumpMajor() string {
	v.Version.Major += 1
	return v.String()
}

func (v *Version) BumpMinor() string {
	v.Version.Minor += 1
	return v.String()
}

func (v *Version) BumpPatch() string {
	v.Version.Patch += 1
	return v.String()
}
