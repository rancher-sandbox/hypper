/*
Copyright The Helm Authors, SUSE LLC.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package action

import (
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"strings"
)

// Upgrade is a composite type of Helm's Upgrade type
type Upgrade struct {
	*action.Upgrade
	cfg         *Configuration
	ReleaseName string
}

// NewUpgrade creates a new Upgrade object with the given configuration.
func NewUpgrade(cfg *Configuration) *Upgrade {
	return &Upgrade{
		Upgrade: action.NewUpgrade(cfg.Configuration),
		cfg:     cfg,
	}
}

// Name returns the name that should be used.
func (i *Upgrade) Name(chart *chart.Chart, args []string) (string, error) {
	// args here will only be: [CHART]
	// cobra flags have been already stripped

	if len(args) > 2 {
		return args[0], errors.Errorf("expected at most two arguments, unexpected arguments: %v", strings.Join(args[2:], ", "))
	}

	if chart.Metadata.Annotations != nil {
		if val, ok := chart.Metadata.Annotations["hypper.cattle.io/release-name"]; ok {
			return val, nil
		}
		if val, ok := chart.Metadata.Annotations["catalog.cattle.io/release-name"]; ok {
			return val, nil
		}
	}

	// If we dont have our annotations then return the base name
	return chart.Metadata.Name, nil
}
