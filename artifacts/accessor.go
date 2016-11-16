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

type Artifact struct {
	Repo     string // Target Repositiory to use for the artifact
	Group    string // Target Repositiory group (default: com.capitalone.bankoao)
	Name     string // Target Repositiory artifact name (default: bankoao_inf_nagios)
	FileName string // File to upload
	Version  string // version of the artifact
}

type ArtifactAccessor interface {
	Upload() string
	Download()
	Promote(fromRepo string)
}
