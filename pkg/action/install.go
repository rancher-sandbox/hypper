package action

import (
	"strings"

	"github.com/pkg/errors"

	"github.com/Masterminds/log-go"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
)

// Install is a composite type of Helm's Install type
type Install struct {
	*action.Install
}

// NewInstall creates a new Install object with the given configuration,
// by wrapping action.NewInstall
func NewInstall(cfg *Configuration) *Install {
	return &Install{
		action.NewInstall(cfg.Configuration),
	}
}

// CheckDependencies checks the dependencies for a chart.
// by wrapping action.CheckDependencies
func CheckDependencies(ch *chart.Chart, reqs []*chart.Dependency) error {
	return action.CheckDependencies(ch, reqs)
}

// Run executes the installation
//
// If DryRun is set to true, this will prepare the release, but not install it
func (i *Install) Run(chrt *chart.Chart, vals map[string]interface{}) (*release.Release, error) {
	log.Infof("Installing chart \"%s\" in namespace \"%s\"…", i.ReleaseName, i.Namespace)
	helmInstall := i.Install
	rel, err := helmInstall.Run(chrt, vals) // wrap Helm's i.Run for now
	return rel, err
}
