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
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/capitalone/stack-deployment-tool/providers"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aymerick/raymond"
)

const (
	default_sse = s3.ServerSideEncryptionAes256
)

var (
	reFixPath = regexp.MustCompile(`([^:])(//)`)
)

type S3Artifact struct {
	Artifact
	BucketUrl string
	Path      string
	Bucket    string
	Region    string
	Key       string
	Encrypt   bool
}

func (a *S3Artifact) Upload() string {
	path := a.artifactPath()
	log.Infof("Uploading artifact %v to: %s", a.FileName, path)

	// create a file reader
	f, err := os.Open(a.FileName)
	if err != nil {
		log.Fatalf("Error opening file to upload %s", a.FileName)
	}
	defer f.Close()

	params := &s3.PutObjectInput{
		Bucket: aws.String(a.Bucket), // Required
		Key:    aws.String(path),     // Required

		Body: f,
		//ContentEncoding:    aws.String("ContentEncoding"),
		ContentType: aws.String("binary/octet-stream"),
		// Metadata: map[string]*string{
		// 	"Key": aws.String("MetadataValue"), // Required
		// },

		ACL:          aws.String(s3.ObjectCannedACLBucketOwnerFullControl),
		StorageClass: aws.String(s3.StorageClassStandard),
	}

	if a.Encrypt {
		params.ServerSideEncryption = aws.String(default_sse)
	}

	api := providers.NewAWSApi()
	resp, err := api.S3Service().PutObject(params)
	s3path := a.s3ArtifactPath()
	if err != nil {
		log.Fatalf("Error (%v) uploading file %s to %s", err, a.FileName, s3path)
	}
	log.Debugf("PutObject: %+v", resp)
	return s3path
}

func (a *S3Artifact) Download() {
	f, err := os.Create(a.FileName)
	if err != nil {
		log.Fatalf("Error (%v) writing file %s", err, a.FileName)
	}

	path := a.artifactPath()
	s3path := a.s3ArtifactPath()
	log.Infof("Downloading artifact %v from: %s", a.FileName, s3path)

	params := &s3.GetObjectInput{
		Bucket: aws.String(a.Bucket), // Required
		Key:    aws.String(path),     // Required
	}

	api := providers.NewAWSApi()
	resp, err := api.S3Service().GetObject(params)

	if err != nil {
		log.Fatalf("Error (%v) downloading archive %s to %s", err, a.FileName, params)
	}
	log.Debugf("GetObject: %+v", resp)

	defer resp.Body.Close()
	_, err = io.Copy(f, resp.Body)
	if err != nil {
		log.Fatalf("Error (%v) writing to file %s", err, a.FileName)
	}
}

func (a *S3Artifact) Promote(fromRepo string) {
	toPath := a.artifactPath()
	toS3path := a.s3ArtifactPath()

	fromArtifact := *a
	fromArtifact.Repo = fromRepo
	fromPath := fromArtifact.artifactPath()
	fromS3path := fromArtifact.s3ArtifactPath()

	log.Infof("Promoting artifact %#v from: %s to: %s", a.FileName, fromS3path, toS3path)
	params := &s3.CopyObjectInput{
		Bucket:       aws.String(a.Bucket),                  // Required
		CopySource:   aws.String(a.Bucket + "/" + fromPath), // Required
		Key:          aws.String(toPath),                    // Required
		ACL:          aws.String(s3.ObjectCannedACLBucketOwnerFullControl),
		StorageClass: aws.String(s3.StorageClassStandard),
	}

	if a.Encrypt {
		params.ServerSideEncryption = aws.String(default_sse)
	}

	api := providers.NewAWSApi()
	resp, err := api.S3Service().CopyObject(params)
	if err != nil {
		log.Fatalf("Error (%v) promoting archive %s to %s", err, fromPath, params)
	}
	log.Debugf("CopyObject: %+v", resp)
}

func (a *S3Artifact) artifactPath() string {
	a.Key = strings.Replace(a.Group, ".", "/", -1)
	templ := a.Path
	p, err := raymond.MustParse(templ).Exec(a)
	if err != nil {
		log.Errorf("Error processing s3 path template: %s %v", templ, err)
		return ""
	}

	return cleanDoubleSlashes(p)
}

func (a *S3Artifact) s3ArtifactPath() string {
	templ := a.BucketUrl + a.artifactPath()

	p, err := raymond.MustParse(templ).Exec(a)
	if err != nil {
		log.Errorf("Error processing s3 path template: %s %v", templ, err)
		return ""
	}
	return cleanDoubleSlashes(p)
}

func cleanDoubleSlashes(path string) string {
	return reFixPath.ReplaceAllString(path, "$1/")
}
