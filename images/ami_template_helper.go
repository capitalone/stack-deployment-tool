//
// Copyright 2017 Capital One Services, LLC
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
package images

import (
	"sync"

	"github.com/capitalone/stack-deployment-tool/stacks"

	"github.com/aymerick/raymond"
)

const (
	amiTemplCmd       = "ami"
	latestAmiTemplCmd = "latest-ami"
)

func init() {
	lazyImageFinder := &LazyImageFinder{}
	stacks.RegisterTemplateHelper(amiTemplCmd, lazyImageFinder, amiHelper)
	stacks.RegisterTemplateHelper(latestAmiTemplCmd, lazyImageFinder, latestAmiHelper)
}

type LazyImageFinder struct {
	finderInit sync.Once
	finderInst *AmiFinder
}

func (l *LazyImageFinder) FindImageId(platform, version string) string {
	return l.finder().FindImageId(platform, version)
}

func (l *LazyImageFinder) FindLatestImageId(platform string) string {
	return l.finder().FindLatestImageId(platform)
}

func (l *LazyImageFinder) finder() *AmiFinder {
	l.finderInit.Do(func() {
		l.finderInst = NewAmiFinder()
	})
	return l.finderInst
}

func amiHelper(options *raymond.Options) raymond.SafeString {
	imageFinder := stacks.HelperCtx(options, amiTemplCmd).(ImageFinder)

	return raymond.SafeString(imageFinder.FindImageId(
		options.HashStr("platform"), options.HashStr("version")))
}

func latestAmiHelper(options *raymond.Options) raymond.SafeString {
	imageFinder := stacks.HelperCtx(options, latestAmiTemplCmd).(ImageFinder)

	return raymond.SafeString(imageFinder.FindLatestImageId(
		options.HashStr("platform")))
}
