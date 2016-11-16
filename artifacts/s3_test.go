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
package artifacts

import (
	"fmt"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func initDefaults() {
	// set defaults
	viper.SetDefault("PromotionPath", map[string]string{
		// from  : to
		"sandbox":  "snapshot",
		"snapshot": "staging",
		"staging":  "release",
	})
	viper.SetDefault("S3Path", "artifacts/{{Repo}}/{{Key}}/{{Name}}/{{Version}}/{{FileName}}")
	viper.SetDefault("S3BucketUrl", "s3://{{Bucket}}/")
}

func TestS3Path(t *testing.T) {
	initDefaults()
	s3 := S3Artifact{Artifact: Artifact{FileName: "testing.tar.gz", Group: "com.something", Repo: "dev", Name: "test"}, Bucket: "stuff",
		Path: viper.GetString("S3Path"), BucketUrl: viper.GetString("S3BucketUrl")}

	p := s3.s3ArtifactPath()

	assert.NotNil(t, p)

	fmt.Printf("p: %s\n", p)

	assert.Equal(t, "s3://stuff/artifacts/dev/com/something/test/testing.tar.gz", p)
}

func TestS3PathWithVer(t *testing.T) {
	s3 := S3Artifact{Artifact: Artifact{FileName: "testing.tar.gz", Group: "com.something", Repo: "dev", Name: "test", Version: "0.1.2"}, Bucket: "stuff",
		Path: viper.GetString("S3Path"), BucketUrl: viper.GetString("S3BucketUrl")}

	p := s3.s3ArtifactPath()

	assert.NotNil(t, p)

	fmt.Printf("p: %s\n", p)

	assert.Equal(t, "s3://stuff/artifacts/dev/com/something/test/0.1.2/testing.tar.gz", p)
}
