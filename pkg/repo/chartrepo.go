/*
Copyright SUSE LLC.

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

package repo

import (
	"net/url"

	"github.com/pkg/errors"
	"github.com/rancher-sandbox/hypper/pkg/hypperpath"
	"helm.sh/helm/v3/pkg/getter"
	helmRepo "helm.sh/helm/v3/pkg/repo"
)

// ChartRepository represents a chart repository
//
// ChartRepository is a composite type of Helm's repo.ChartRepository
type ChartRepository struct {
	helmRepo.ChartRepository
}

// NewChartRepository constructs ChartRepository
func NewChartRepository(cfg *helmRepo.Entry, getters getter.Providers) (*ChartRepository, error) {
	u, err := url.Parse(cfg.URL)
	if err != nil {
		return nil, errors.Errorf("invalid chart URL format: %s", cfg.URL)
	}

	client, err := getters.ByScheme(u.Scheme)
	if err != nil {
		return nil, errors.Errorf("could not find protocol handler for: %s", u.Scheme)
	}
	return &ChartRepository{
		helmRepo.ChartRepository{
			Config:    cfg,
			IndexFile: helmRepo.NewIndexFile(),
			Client:    client,
			CachePath: hypperpath.CachePath("repository"),
		},
	}, nil

}
