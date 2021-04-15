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
	logio "github.com/Masterminds/log-go/io"
	"github.com/jinzhu/copier"
	"github.com/rancher-sandbox/hypper/pkg/cli"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/release"
)

// Install is a composite type of Helm's Install type
type Install struct {
	*action.Install

	// hypper specific
	NoSharedDeps bool

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
func (i *Install) Run(chrt *chart.Chart, vals map[string]interface{}, settings *cli.EnvSettings, logger log.Logger) (*release.Release, error) {

	if !i.NoSharedDeps {
		if err := i.InstallAllSharedDeps(chrt, settings, logger); err != nil {
			return nil, err
		}
	}

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

// InstallAllSharedDeps installs all shared dependencies listed in the passed
// chart.
//
// It will check for malformed chart.Metadata.Annotations, and skip those shared
// dependencies already deployed.
func (i *Install) InstallAllSharedDeps(chrt *chart.Chart, settings *cli.EnvSettings, logger log.Logger) error {

	if _, ok := chrt.Metadata.Annotations["hypper.cattle.io/shared-dependencies"]; !ok {
		logger.Debugf("No shared dependencies in %s\n", chrt.Name())
		return nil
	}

	depYaml := chrt.Metadata.Annotations["hypper.cattle.io/shared-dependencies"]
	var deps dependencies
	if err := yaml.UnmarshalStrict([]byte(depYaml), &deps); err != nil {
		logger.Errorf("Chart.yaml metadata is malformed for chart %s\n", chrt.Name())
		return err
	}

	clientList := NewList(i.Config)
	clientList.SetStateMask()
	releases, err := clientList.Run()
	if err != nil {
		return err
	}

	for _, dep := range deps {
		found := false
		for _, r := range releases {
			if r.Name == dep.Name {
				logger.Infof("Shared dependency %s already installed, not doing anything\n", dep.Name)
				found = true
				break // installed, don't keep looking
			}
		}
		if !found {
			if _, err = i.InstallSharedDep(dep, settings, logger); err != nil {
				return err
			}
		}
	}
	return nil
}

// InstallSharedDep installs a chart.Dependency using the provided settings.
//
// It does this by creating a new action.Install and setting it correctly,
// loading the chart, checking for constraints, and delegating the install.Run()
func (i *Install) InstallSharedDep(dep *chart.Dependency, settings *cli.EnvSettings, logger log.Logger) (*release.Release, error) {

	wInfo := logio.NewWriter(logger, log.InfoLevel)

	clientInstall := NewInstall(i.Config)
	// we need to automatically satisfy all install options (i.CreateNamespace,
	// i.DryRun, etc) when we are installing the dep using clientInstall. Doing
	// a shallow copy sounds like asking for trouble when the install struct
	// changes, so let's do a deep copy instead:
	if err := copier.Copy(&clientInstall, &i); err != nil {
		return nil, err
	}

	clientInstall.ChartPathOptions.RepoURL = dep.Repository
	cp, err := clientInstall.ChartPathOptions.LocateChart(dep.Name, settings.EnvSettings)
	if err != nil {
		return nil, err
	}

	logger.Debugf("CHART PATH: %s\n", cp)

	p := getter.All(settings.EnvSettings)
	vals := make(map[string]interface{}) // TODO calculate vals instead of {}

	chartRequested, err := loader.Load(cp)
	if err != nil {
		return nil, err
	}

	logger.Debugf("Original shared-dep chart version: %q", chartRequested.Metadata.Version)
	if clientInstall.Devel {
		logger.Debug("setting version to >0.0.0-0")
		clientInstall.Version = ">0.0.0-0"
	}
	// TODO check if chartRequested satisfies version range specified in dep

	// Set Namespace, Releasename for the install client without reevaluating them
	// from the dependent:
	SetNamespace(clientInstall, chartRequested, i.Namespace, true)
	clientInstall.ReleaseName, err = GetName(chartRequested, clientInstall.NameTemplate, dep.Name)
	if err != nil {
		return nil, err
	}

	if err := CheckIfInstallable(chartRequested); err != nil {
		return nil, err
	}

	if chartRequested.Metadata.Deprecated {
		logger.Warn("This chart is deprecated")
	}

	// Check chart dependencies to make sure all are present in /charts
	if req := chartRequested.Metadata.Dependencies; req != nil {
		// If CheckDependencies returns an error, we have unfulfilled dependencies.
		// As of Helm 2.4.0, this is treated as a stopping condition:
		// https://github.com/helm/helm/issues/2209
		if err := action.CheckDependencies(chartRequested, req); err != nil {
			if clientInstall.DependencyUpdate {
				man := &downloader.Manager{
					Out:              wInfo,
					ChartPath:        cp,
					Keyring:          clientInstall.ChartPathOptions.Keyring,
					SkipUpdate:       false,
					Getters:          p,
					RepositoryConfig: settings.RepositoryConfig,
					RepositoryCache:  settings.RepositoryCache,
					Debug:            settings.Debug,
				}
				if err := man.Update(); err != nil {
					return nil, err
				}
				// Reload the chart with the updated Chart.lock file.
				if chartRequested, err = loader.Load(cp); err != nil {
					return nil, errors.Wrap(err, "failed reloading chart after repo update")
				}
			} else {
				return nil, err
			}
		}
	}

	res, err := clientInstall.Run(chartRequested, vals, settings, logger)
	if err != nil {
		return res, err
	}
	return res, nil
}
