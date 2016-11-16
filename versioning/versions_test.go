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
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGit(t *testing.T) {
	assert.NotNil(t, GitLatestTag())
}

func TestVersionLoad(t *testing.T) {
	v := LoadVersionProps("../resources/full_ver.properties")
	assert.NotNil(t, v)
	assert.True(t, len(v.String()) > 0)
	assert.Equal(t, "3.0.1-alpha.1+abc1231467392993", v.String())
	assert.Equal(t, "version=3.0.1-alpha.1+abc1231467392993", v.ToPropString())

	v.BumpMajor()
	assert.Equal(t, "4.0.1-alpha.1+abc1231467392993", v.String())

	v.BumpMinor()
	assert.Equal(t, "4.1.1-alpha.1+abc1231467392993", v.String())

	v.BumpPatch()
	assert.Equal(t, "4.1.2-alpha.1+abc1231467392993", v.String())
}

func TestVersionWriting(t *testing.T) {
	v := LoadVersionProps("../resources/full_ver.properties")
	v.BumpPatch()
	v.BumpPatch()
	v.BumpPatch()
	f, err := ioutil.TempFile("", "ver")
	assert.Nil(t, err)
	defer os.Remove(f.Name()) // clean up
	v.WriteVersionProps(f.Name())
	v2 := LoadVersionProps(f.Name())
	assert.Equal(t, v.String(), v2.String())
}
