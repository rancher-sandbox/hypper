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
