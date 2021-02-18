package action

import (
	"helm.sh/helm/v3/pkg/action"
)

// Uninstall is a composite type of Helm's Uninstall type
type Uninstall struct {
	*action.Uninstall
}

// NewUninstall creates a new Uninstall by embedding action.Uninstall
func NewUninstall(cfg *Configuration) *Uninstall {
	return &Uninstall{
		action.NewUninstall(cfg.Configuration),
	}
}
