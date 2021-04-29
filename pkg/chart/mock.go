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
	"helm.sh/helm/v3/pkg/chart"
)

// MockChartOptions allows for user-configurable options on mock chart objects.
type MockChartOptions struct {
	Name    string
	Version string
}

// Mock creates a mock chart object based on options set by MockChartOptions.
// This function should typically not be used outside of testing.
func Mock(opts *MockChartOptions) *chart.Chart {
	return &chart.Chart{
		Metadata: &chart.Metadata{
			Name:       opts.Name,
			Version:    opts.Version,
			AppVersion: "1.0",
		},
		Templates: []*chart.File{
			{Name: "templates/hello", Data: []byte("hello: world")},
		},
	}
}
