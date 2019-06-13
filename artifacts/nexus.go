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
package artifacts

import (
	"bytes"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path"

	log "github.com/Sirupsen/logrus"
)

type NexusArtifact struct {
	Artifact
	User     string // Repository user
	Password string // Repository Password
	Url      string // Repository URL
}

// Nexus Upload
func (a *NexusArtifact) Upload() string {
	buf := bytes.NewBuffer([]byte{})
	multiWriter := multipart.NewWriter(buf)

	fields := map[string]string{
		"r":      a.Repo,
		"hasPom": "false",
		"e":      "gz",
		"p":      "gz",
		"g":      a.Group,
		"a":      a.Name,
		"v":      a.Version,
	}

	for k, v := range fields {
		err := multiWriter.WriteField(k, v)
		if err != nil {
			log.Errorf("Error writing field: %v", err)
		}
	}

	f, err := multiWriter.CreateFormFile("file", path.Base(a.FileName))
	arcFile, err := os.Open(a.FileName)

	if err != nil {
		log.Fatalf("Error opening archive file: %v", err)
	}

	defer arcFile.Close()
	io.Copy(f, arcFile)
	multiWriter.Close()

	req, err := http.NewRequest("POST", a.Url, buf)
	if err != nil {
		log.Fatalf("Error posting file: %v", err)
	}
	req.Header.Add("Content-Type", multiWriter.FormDataContentType()) //"multipart/form-data")
	if len(a.User) > 0 && len(a.Password) > 0 {
		req.SetBasicAuth(a.User, a.Password)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Error posting archive file: %v %v", err, resp)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	log.Debugf("Response: %v\n", resp)
	log.Debugf("Response: %s\n", string(body))
	if !OkResponse(resp) {
		log.Fatalf("Error uploading archive file: %s %v", a.FileName, resp.Status)
	}
	return a.Url
}

func (a *NexusArtifact) Download() {
	log.Fatalf("Not supported at this time")
}

func (a *NexusArtifact) Promote(fromRepo string) {
	log.Fatalf("Not supported at this time")
}

func OkResponse(resp *http.Response) bool {
	return (resp.StatusCode%100) >= 2 && (resp.StatusCode%100) < 3
}
