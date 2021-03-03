package action

import (
	"helm.sh/helm/v3/pkg/action"
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
