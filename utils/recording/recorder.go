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
package recording

import (
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/dnaeon/go-vcr/cassette"
	"github.com/dnaeon/go-vcr/recorder"
)

/// vcr helpers ...
type CleanerFunc func()

func FixtureName() string {
	return CallerName(1)
}

func CallerName(up int) string {
	caller, _, _, _ := runtime.Caller(up)
	frames := runtime.CallersFrames([]uintptr{caller})
	frame, _ := frames.Next()
	_, pkgFunc := filepath.Split(frame.Function)
	return strings.Replace(pkgFunc, ".", "/", -1)
}

func RecorderHttpClient(r *recorder.Recorder) *http.Client {
	return &http.Client{
		Transport: r, // Inject our transport!
	}
}

func CreateHttpRecorder(dir string, fixtureName ...string) (*recorder.Recorder, CleanerFunc) {
	f := filepath.Join(fixtureName...)
	if len(fixtureName) == 0 {
		f = CallerName(2)
	}
	name := filepath.Join(dir, f)

	if _, err := os.Stat(filepath.Dir(name)); err != nil {
		log.Fatalf("Error recording fixture: %s", err)
	}

	r, err := recorder.New(name)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	return r, func() {
		r.Stop()
		scrubHeaders(name)
	}
}

func scrubHeaders(cassetteName string) {
	c, err := cassette.Load(cassetteName)
	if err == nil {
		for _, inter := range c.Interactions {
			scrubHeader(inter.Request.Headers, "Authorization")
			scrubHeader(inter.Request.Headers, "X-Amz-Security-Token")
			scrubBody(&inter.Response, "<AccessKeyId>(.*)</AccessKeyId>", "<AccessKeyId>xxxx</AccessKeyId>")
			scrubBody(&inter.Response, "<SecretAccessKey>(.*)</SecretAccessKey>", "<SecretAccessKey>xxxx</SecretAccessKey>")
			scrubBody(&inter.Response, "<SessionToken>(.*)</SessionToken>", "<SessionToken>xxxx</SessionToken>")
		}
		c.Save()
	}
}

func scrubHeader(headers http.Header, headerName string) {
	auth := headers.Get(headerName)
	if len(auth) > 0 {
		headers.Set(headerName, "xxxxx")
	}
}

func scrubBody(resp *cassette.Response, match, replace string) {
	re := regexp.MustCompile(match)
	resp.Body = re.ReplaceAllString(resp.Body, replace)
}
