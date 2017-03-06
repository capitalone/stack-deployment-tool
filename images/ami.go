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
package images

import (
	"reflect"
	"regexp"
	"sort"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/capitalone/stack-deployment-tool/providers"
	"github.com/spf13/viper"
)

type ImageFinder interface {
	FindImageId(platform, version string) string
	FindLatestImageId(platform string) string
}

type DescribeImagesFunc func(params *ec2.DescribeImagesInput) (*ec2.DescribeImagesOutput, error)

type AmiFinder struct {
	providers.AWSApi
	DescribeImages DescribeImagesFunc
}

func NewAmiFinder() *AmiFinder {
	return NewAmiFinderFromProvider(providers.NewAWSApi())
}

func NewAmiFinderFromProvider(api *providers.AWSApi) *AmiFinder {
	return &AmiFinder{AWSApi: *api}
}

func (a *AmiFinder) FindLatestImageId(platform string) string {
	return a.FindImageId(platform, "")
}

func (a *AmiFinder) FindImageId(platform, version string) string {
	ids := a.FindImageIds(platform, version)
	if len(ids) > 0 {
		return ids[0]
	}
	return ""
}

func (a *AmiFinder) FindImageIds(platform, version string) []string {
	matching := a.FindImages(platform, version)
	results := []string{}
	for _, match := range matching {
		results = append(results, *match.ImageId)
	}
	return results
}

func imageFilters() []*ec2.Filter {
	filterConfig := viper.GetStringMapString("ImageFilter")

	filters := make([]*ec2.Filter, 0)
	for k, v := range filterConfig {
		filters = append(filters, &ec2.Filter{
			Name: aws.String(k),
			Values: []*string{
				aws.String(v),
			},
		})
	}
	return filters
}

func imageFilterRegexMap() map[string]*regexp.Regexp {
	filterRegex := viper.GetStringMapString("ImageFilterRegex")
	regexMap := make(map[string]*regexp.Regexp)
	for k, v := range filterRegex {
		regexMap[k] = regexp.MustCompile(v)
	}
	return regexMap
}

func (a *AmiFinder) describeImages(params *ec2.DescribeImagesInput) (*ec2.DescribeImagesOutput, error) {
	log.Debugf("DescribeImages: %#v", params)
	var resp *ec2.DescribeImagesOutput
	var err error
	if a.DescribeImages == nil {
		resp, err = a.EC2Service().DescribeImages(params)
	} else {
		resp, err = a.DescribeImages(params)
	}
	log.Debugf("DescribeImages Err: %#v Response: %#v", err, resp)
	return resp, err
}

func (a *AmiFinder) FindImages(platform, version string) []*ec2.Image {
	filters := imageFilters()

	dryMode := viper.GetBool("drymode")

	params := &ec2.DescribeImagesInput{
		DryRun:  aws.Bool(dryMode),
		Filters: filters,
		Owners: []*string{
			aws.String("self"), // Required
		},
	}

	resp, err := a.describeImages(params)
	if err == nil {
		// apply regex Filters & sort by version desc
		return regexFilterAndSortImages(resp.Images, platform, version)
	}
	return []*ec2.Image{}
}

// apply regex Filters & sort by version desc
func regexFilterAndSortImages(images []*ec2.Image, platform, version string) []*ec2.Image {
	// apply regex Filters
	regexMap := imageFilterRegexMap()
	matching := make([]*ec2.Image, 0)
	for _, image := range images {
		m := FieldsMatch(image, regexMap)
		if m {
			p := ImagePlatform(image)
			platformMatch := strings.EqualFold(p, platform)
			if len(platform) == 0 || platformMatch {
				matching = append(matching, image)
			}

			v := ImageVersion(image)
			if len(version) > 0 && platformMatch && strings.EqualFold(v, version) {
				return []*ec2.Image{image} // found THE match!
			}
		}
	}
	// sort versions and give newest..
	sortByVersionDesc(matching)
	return matching
}

type Images []*ec2.Image

func (a Images) Len() int      { return len(a) }
func (a Images) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a Images) Less(i, j int) bool {
	verA := ImageVersion(a[i])
	verB := ImageVersion(a[j])
	return strings.Compare(verA, verB) == 1
}

func sortByVersionDesc(images []*ec2.Image) {
	sort.Sort(Images(images))
}

func ImageVersion(image *ec2.Image) string {
	verRe := regexp.MustCompile(viper.GetString("VersionRegex"))
	matches := verRe.FindAllStringSubmatch(*image.Name, -1)
	if len(matches) > 0 && len(matches[0]) >= 2 {
		return matches[0][1]
	}
	return ""
}

func ImagePlatform(image *ec2.Image) string {
	verRe := regexp.MustCompile(viper.GetString("PlatformRegex"))
	matches := verRe.FindAllStringSubmatch(*image.Name, -1)
	if len(matches) > 0 && len(matches[0]) >= 2 {
		return matches[0][1]
	}
	return ""
}

func FieldsMatch(image *ec2.Image, fieldFilters map[string]*regexp.Regexp) bool {
	if len(fieldFilters) == 0 {
		return true // nothing to match, so we're done.
	}

	match := false

	st := reflect.TypeOf(*image)
	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)
		if locationName, ok := field.Tag.Lookup("locationName"); ok {
			if v, found := fieldFilters[locationName]; found {
				fieldValue := reflect.ValueOf(*image).FieldByName(field.Name)
				val := fieldValue.Elem().String() // de-ref the *string
				match = match || v.MatchString(val)
			}
		}
	}
	return match
}
