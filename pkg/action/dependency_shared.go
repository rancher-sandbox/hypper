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
	"github.com/Masterminds/log-go"
	"github.com/Masterminds/semver/v3"
	"github.com/gosuri/uitable"
	"github.com/pkg/errors"
	"github.com/rancher-sandbox/hypper/pkg/chart"
	"github.com/rancher-sandbox/hypper/pkg/cli"

	"helm.sh/helm/v3/pkg/action"
	helmChart "helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
)

// SharedDependency is the action for building a given chart's shared dependency tree.
//
// It provides the implementation of 'hypper shared-dependency' and its respective subcommands.
type SharedDependency struct {
	*action.Dependency

	// hypper specific:
	Config *Configuration
}

// NewSharedDependency creates a new SharedDependency object with the given configuration.
func NewSharedDependency(cfg *Configuration) *SharedDependency {
	return &SharedDependency{
		action.NewDependency(),
		cfg,
	}
}

// List executes 'hypper shared-dep list'.
func (d *SharedDependency) List(chartpath string, settings *cli.EnvSettings, logger log.Logger) error {

	c, err := loader.Load(chartpath)
	if err != nil {
		return err
	}

	sharedDeps, err := chart.GetSharedDeps(c, logger)
	if err != nil {
		return err
	}

	return d.printSharedDependencies(chartpath, logger, sharedDeps, settings)
}

// SharedDependencyStatus returns a string describing the status of a dependency
// viz a viz the releases in depNS context.
func (d *SharedDependency) SharedDependencyStatus(depChart *helmChart.Chart, depNS string, depVersion string) (string, error) {

	// TODO refactor GetName() into GetName(){ret error} and GetNameFromAnnot()
	depName, err := GetName(depChart, "")
	if err != nil {
		return "", err
	}

	clientList := NewList(d.Config)
	clientList.SetStateMask()
	releases, err := clientList.Run()
	if err != nil {
		return "", err
	}

	for _, r := range releases {
		if r.Name == depName && r.Namespace == depNS {
			if r.Chart.Metadata.Version != depVersion {
				constraint, err := semver.NewConstraint(depVersion)
				if err != nil {
					return "", errors.New("dependency version not parseable")
				}

				v, _ := semver.NewVersion(r.Chart.Metadata.Version)
				// not needed to check err, gets validated on chart creation
				if !constraint.Check(v) {
					return "out-of-range", nil
				}
			}
			return r.Info.Status.String(), nil
		}
	}

	return "not-installed", nil
}

// printSharedDependencies prints all of the shared dependencies in the yaml file.
// It will respect settings.NamespaceFromFlag when iterating through releases.
func (d *SharedDependency) printSharedDependencies(chartpath string, logger log.Logger, deps []*chart.Dependency, settings *cli.EnvSettings) error {

	table := uitable.New()
	table.MaxColWidth = 80
	table.AddRow("NAME", "VERSION", "REPOSITORY", "STATUS", "NAMESPACE")
	for _, dep := range deps {
		chartPathOptions := action.ChartPathOptions{}
		chartPathOptions.RepoURL = dep.Repository
		cp, err := chartPathOptions.LocateChart(dep.Name, settings.EnvSettings)
		if err != nil {
			return err
		}
		logger.Debugf("CHART PATH: %s\n", cp)

		depChart, err := loader.Load(cp)
		if err != nil {
			return err
		}

		// calculate which ns corresponds to the dependency
		var depNS string
		if settings.NamespaceFromFlag {
			depNS = settings.Namespace()
		} else {
			// either shared-dep has annotations, or the parent has, or we use the default ns
			depNS = GetNamespace(depChart, GetNamespace(depChart, settings.Namespace()))
		}
		d.Config.SetNamespace(depNS)

		if settings.NamespaceFromFlag && settings.Namespace() != depNS {
			// skip listing this dep, it's not in the same namespace as the flag
			continue
		}

		depStatus, err := d.SharedDependencyStatus(depChart, depNS, dep.Version)
		if err != nil {
			return err
		}
		table.AddRow(depChart.Name(), dep.Version, dep.Repository, depStatus, depNS)
	}
	log.Infof(table.String())
	return nil
}
