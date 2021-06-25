// Copyright SUSE LLC.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package chart

import (
	"github.com/Masterminds/log-go"
	"github.com/mitchellh/hashstructure/v2"
	"gopkg.in/yaml.v2"
	helmChart "helm.sh/helm/v3/pkg/chart"
)

// Dependency is a composite type of Helm's chart.Dependency
type Dependency struct {
	*helmChart.Dependency
	// hypper specific
	IsOptional bool `json:"-"`
}

// GetSharedDeps returns a *[] of all shared and optional dependencies in a
// chart, read from annotations.
func GetSharedDeps(c *helmChart.Chart, logger log.Logger) ([]*Dependency, error) {

	sharedDeps := make([]*Dependency, 0)

	var helmSharedDeps []*helmChart.Dependency
	_, ok := c.Metadata.Annotations["hypper.cattle.io/shared-dependencies"]
	if !ok {
		log.Debugf("No shared dependencies in \"%s\"\n", c.Name())
	} else {
		sharedDepYaml := c.Metadata.Annotations["hypper.cattle.io/shared-dependencies"]
		if err := yaml.UnmarshalStrict([]byte(sharedDepYaml), &helmSharedDeps); err != nil {
			log.Errorf("Chart.yaml metadata is malformed for chart \"%s\"\n", c.Name())
			return nil, err
		}
		// unmarshalling Helm's Dependency because gopkg.in/yaml.v2 doesn't do composite types
		for _, helmDep := range helmSharedDeps {
			sharedDeps = append(sharedDeps, &Dependency{Dependency: helmDep, IsOptional: false})
		}
	}

	var helmOptionalDeps []*helmChart.Dependency
	_, ok = c.Metadata.Annotations["hypper.cattle.io/optional-dependencies"]
	if !ok {
		log.Debugf("No optional shared dependencies in \"%s\"\n", c.Name())
	} else {
		optDepYaml := c.Metadata.Annotations["hypper.cattle.io/optional-dependencies"]
		if err := yaml.UnmarshalStrict([]byte(optDepYaml), &helmOptionalDeps); err != nil {
			log.Errorf("Chart.yaml metadata is malformed for chart \"%s\"\n", c.Name())
			return nil, err
		}
		// unmarshalling Helm's Dependency because gopkg.in/yaml.v2 doesn't do composite types
		for _, helmDep := range helmOptionalDeps {
			sharedDeps = append(sharedDeps, &Dependency{Dependency: helmDep, IsOptional: true})
		}
	}

	return sharedDeps, nil
}

func Hash(c *helmChart.Chart) uint64 {
	hash, err := hashstructure.Hash(c, hashstructure.FormatV2, nil)
	if err != nil {
		panic(err)
	}
	return hash
}
