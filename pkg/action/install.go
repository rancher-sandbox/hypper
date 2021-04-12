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

	// Config stores the actionconfig so it can be retrieved and used again
	Config *Configuration
}

// NewInstall creates a new Install object with the given configuration,
// by wrapping action.NewInstall
func NewInstall(cfg *Configuration) *Install {
	return &Install{
		Install: action.NewInstall(cfg.Configuration),
		Config:  cfg,
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

// Chart returns the chart that should be used.
//
// This will read the flags and skip args if necessary.
func (i *Install) Chart(args []string) (string, error) {
	if len(args) > 2 {
		return args[1], errors.Errorf("expected at most two arguments, unexpected arguments: %v", strings.Join(args[2:], ", "))
	}

	if len(args) == 2 {
		return args[1], nil
	}

	// len(args) == 1
	return args[0], nil
}

// NameAndChart overloads Helm's NameAndChart. It always fails.
//
// On Hypper, we need to read the chart annotations to know the correct release name.
// Therefore, it cannot happen in this function.
func (i *Install) NameAndChart(args []string) (string, string, error) {
	return "", "", errors.New("NameAndChart() cannot be used")
}

// checkIfInstallable validates if a chart can be installed
//
// Application chart type is only installable
func CheckIfInstallable(ch *chart.Chart) error {
	switch ch.Metadata.Type {
	case "", "application":
		return nil
	}
	return errors.Errorf("%s charts are not installable", ch.Metadata.Type)
}
