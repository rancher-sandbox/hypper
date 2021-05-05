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
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/Masterminds/log-go"
	logio "github.com/Masterminds/log-go/io"
	"github.com/Masterminds/semver/v3"
	"github.com/jinzhu/copier"

	"github.com/rancher-sandbox/hypper/pkg/chart"
	"github.com/rancher-sandbox/hypper/pkg/cli"
	"github.com/rancher-sandbox/hypper/pkg/eyecandy"

	"helm.sh/helm/v3/pkg/action"
	helmChart "helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/release"

	"github.com/manifoldco/promptui"
)

// Install is a composite type of Helm's Install type
type Install struct {
	*action.Install

	// hypper specific
	NoSharedDeps bool
	OptionalDeps string

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
func CheckDependencies(ch *helmChart.Chart, reqs []*helmChart.Dependency) error {
	return action.CheckDependencies(ch, reqs)
}

// Run executes the installation
//
// If DryRun is set to true, this will prepare the release, but not install it
// lvl is used for printing nested stagered output on recursion. Starts at 0.
func (i *Install) Run(chrt *helmChart.Chart, vals map[string]interface{}, settings *cli.EnvSettings, logger log.Logger, lvl int) (*release.Release, error) {

	if lvl >= 10 {
		return nil, errors.Errorf("ABORTING: Nested recursion #%d. we don't have a SAT solver yet, chances are you are in a cycle!", lvl)
	}

	if !i.NoSharedDeps {
		if err := i.InstallAllSharedDeps(chrt, settings, logger, lvl); err != nil {
			return nil, err
		}
	}

	prefix := ""
	if lvl > 0 {
		prefix = fmt.Sprintf("%*s", lvl*2, "- ")
	}
	logger.Infof(eyecandy.ESPrintf(settings.NoEmojis, ":cruise_ship: %sInstalling chart \"%s\" as \"%s\" in namespace \"%s\"…", prefix, chrt.Name(), i.ReleaseName, i.Namespace))
	helmInstall := i.Install
	i.Config.SetNamespace(i.Namespace)
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
func CheckIfInstallable(ch *helmChart.Chart) error {
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
// lvl is used for printing nested stagered output on recursion. Starts at 0.
func (i *Install) InstallAllSharedDeps(parentChart *helmChart.Chart, settings *cli.EnvSettings, logger log.Logger, lvl int) error {

	sharedDeps, err := chart.GetSharedDeps(parentChart, logger)
	if err == nil && len(sharedDeps) == 0 {
		return nil
	}

	logger.Infof(eyecandy.ESPrintf(settings.NoEmojis, ":cruise_ship: %sInstalling shared dependencies for chart \"%s\":", strings.Repeat("  ", lvl), parentChart.Name()))
	if err != nil {
		return err
	}

	// increment padding of output
	lvl++
	prefix := ""
	if lvl > 0 {
		prefix = fmt.Sprintf("%*s", lvl*2, "- ")
	}

	for _, dep := range sharedDeps {
		switch i.OptionalDeps {
		case "all":
			// install all deps, optional or not
		case "none":
			if dep.IsOptional {
				continue
			}
		case "ask":
			if dep.IsOptional {
				prompt := promptui.Prompt{
					Label: fmt.Sprintf("Install optional shared dependency \"%s\" ?", dep.Name),
					Validate: func(input string) error {
						if strings.ToLower(input) != "y" &&
							strings.ToLower(input) != "n" &&
							strings.ToLower(input) != "yes" &&
							strings.ToLower(input) != "no" {
							return errors.New("Invalid input")
						}
						return nil
					},
				}
				result, err := prompt.Run()
				if err != nil {
					logger.Errorf("Prompt failed %v\n", err)
				}
				if result != "y" {
					continue
				}
			}
		default:
			return errors.New("Incorrect value for --optional-deps. Valid values: [default=ask|all|none]")
		}

		found := false
		depChart, err := i.LoadChartFromDep(dep, settings, logger)
		if err != nil {
			return err
		}

		// calculate which ns corresponds to the dependency
		var ns string
		if settings.NamespaceFromFlag {
			ns = settings.Namespace()
		} else {
			// either shared-dep has annotations, or the parent has, or we use the default ns
			ns = GetNamespace(depChart, GetNamespace(parentChart, settings.Namespace()))
		}
		i.Config.SetNamespace(ns)

		name, err := GetName(depChart, "")
		if err != nil {
			return err
		}

		// obtain the releases for the specific ns that we are searching into
		clientList := NewList(i.Config)
		clientList.SetStateMask()
		releases, err := clientList.Run()
		if err != nil {
			return err
		}

		// create constraint for version checking
		constraint, err := semver.NewConstraint(dep.Version)
		if err != nil {
			return err
		}

		for _, r := range releases {
			if r.Name == name && r.Namespace == ns {
				v, err := semver.NewVersion(r.Chart.Metadata.Version)
				if err != nil {
					return err
				}
				if b, errs := constraint.Validate(v); !b {
					logger.Errorf(eyecandy.ESPrintf(settings.NoEmojis, ":x: %sShared dependency chart \"%s\" already installed in an unsatisfiable version, aborting\n", prefix, dep.Name))
					err := errors.New("Shared dep version out of range")
					for _, e := range errs {
						err = fmt.Errorf("%w; %s", err, e)
					}
					return err
				}
				logger.Infof(eyecandy.ESPrintf(settings.NoEmojis, ":information_source: %sShared dependency chart \"%s\" already installed, skipping\n", prefix, dep.Name))
				found = true
				break // installed, don't keep looking
			}
		}
		if !found {
			if _, err = i.InstallSharedDep(dep, settings, logger, lvl); err != nil {
				return err
			}
		}
	}
	return nil
}

func (i *Install) LoadChartFromDep(dep *chart.Dependency, settings *cli.EnvSettings, logger log.Logger) (*helmChart.Chart, error) {
	i.ChartPathOptions.RepoURL = dep.Repository
	cp, err := i.ChartPathOptions.LocateChart(dep.Name, settings.EnvSettings)
	if err != nil && !dep.IsOptional {
		return nil, err
	}

	logger.Debugf("CHART PATH: %s\n", cp)

	chartRequested, err := loader.Load(cp)
	if err != nil {
		return nil, err
	}
	return chartRequested, nil
}

// InstallSharedDep installs a chart.Dependency using the provided settings.
//
// It does this by creating a new action.Install and setting it correctly,
// loading the chart, checking for constraints, and delegating the install.Run()
// lvl is used for printing nested stagered output on recursion. Starts at 0.
func (i *Install) InstallSharedDep(dep *chart.Dependency, settings *cli.EnvSettings, logger log.Logger, lvl int) (*release.Release, error) {

	wInfo := logio.NewWriter(logger, log.InfoLevel)

	clientInstall := NewInstall(i.Config)
	// we need to automatically satisfy all install options (i.CreateNamespace,
	// i.DryRun, etc) when we are installing the dep using clientInstall. Doing
	// a shallow copy sounds like asking for trouble when the install struct
	// changes, so let's do a deep copy instead:
	if err := copier.Copy(&clientInstall, &i); err != nil {
		return nil, err
	}

	chartRequested, err := clientInstall.LoadChartFromDep(dep, settings, logger)
	if err != nil {
		return nil, err
	}

	p := getter.All(settings.EnvSettings)
	vals := make(map[string]interface{}) // TODO calculate vals instead of {}

	logger.Debugf("Original shared-dep chart version: %q", chartRequested.Metadata.Version)
	if clientInstall.Devel {
		logger.Debug("setting version to >0.0.0-0")
		clientInstall.Version = ">0.0.0-0"
	}

	// check if chartRequested satisfies version range specified in dep
	if chartRequested.Metadata.Version != dep.Version {
		constraint, err := semver.NewConstraint(dep.Version)
		if err != nil {
			return nil, err
		}

		v, err := semver.NewVersion(chartRequested.Metadata.Version)
		if err != nil {
			return nil, err
		}

		if b, errs := constraint.Validate(v); !b {
			logger.Errorf(eyecandy.ESPrintf(settings.NoEmojis, ":x: Satisfiable version for chart \"%s\" not found, aborting\n", dep.Name))
			err := errors.New("Satisfiable chart version not found")
			for _, e := range errs {
				err = fmt.Errorf("%w; %s", err, e)
			}
			return nil, err
		}
	}

	// Set Namespace, Releasename for the install client without reevaluating them
	// from the dependent:
	SetNamespace(clientInstall, chartRequested, i.Namespace, false)
	clientInstall.ReleaseName, err = GetName(chartRequested, clientInstall.NameTemplate, dep.Name)
	if err != nil {
		return nil, err
	}

	if err := CheckIfInstallable(chartRequested); err != nil {
		return nil, err
	}

	if chartRequested.Metadata.Deprecated {
		logger.Warnf("Chart \"$s\" is deprecated", chartRequested.Name())
	}

	// re-obtain the cp again, for Metadata.Dependencies
	// FIXME deduplicate
	i.ChartPathOptions.RepoURL = dep.Repository
	cp, err := i.ChartPathOptions.LocateChart(dep.Name, settings.EnvSettings)
	if err != nil {
		return nil, err
	}

	logger.Debugf("CHART PATH: %s\n", cp)

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

	res, err := clientInstall.Run(chartRequested, vals, settings, logger, lvl)
	if err != nil {
		return res, err
	}
	return res, nil
}
