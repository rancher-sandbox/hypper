package action

import (
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"strings"
)

type Upgrade struct {
	*action.Upgrade
	cfg *Configuration
}

// NewUpgrade creates a new Upgrade object with the given configuration.
func NewUpgrade(cfg *Configuration) *Upgrade {
	return &Upgrade{
		Upgrade: action.NewUpgrade(cfg.Configuration),
		cfg:     cfg,
	}
}

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
