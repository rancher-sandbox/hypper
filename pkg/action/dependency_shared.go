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

	"github.com/Masterminds/log-go"
	logio "github.com/Masterminds/log-go/io"
	"github.com/gosuri/uitable"
	"github.com/rancher-sandbox/hypper/pkg/cli"
	"gopkg.in/yaml.v2"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
)

// SharedDependency is the action for building a given chart's shared dependency tree.
//
// It provides the implementation of 'hypper shared-dependency' and its respective subcommands.
type SharedDependency struct {
	*action.Dependency

	// hypper specific:
	Namespace string
	Config    *Configuration
}

// NewSharedDependency creates a new SharedDependency object with the given configuration.
func NewSharedDependency(cfg *Configuration) *SharedDependency {
	return &SharedDependency{
		action.NewDependency(),
		"", //namespace, to be filled when we have the chart
		cfg,
	}
}

type dependencies []*chart.Dependency

// SetNamespace sets the Namespace that should be used in action.SharedDependency
//
// This will read the chart annotations. If no annotations, it leave the existing ns in the action.
func (d *SharedDependency) SetNamespace(chart *chart.Chart, defaultns string) {
	d.Namespace = defaultns
	if chart.Metadata.Annotations != nil {
		if val, ok := chart.Metadata.Annotations["hypper.cattle.io/namespace"]; ok {
			d.Namespace = val
		} else {
			if val, ok := chart.Metadata.Annotations["catalog.cattle.io/namespace"]; ok {
				d.Namespace = val
			}
		}
	}
}

// List executes 'hypper shared-dep list'.
func (d *SharedDependency) List(chartpath string, settings *cli.EnvSettings, logger log.Logger) error {

	wWarn := logio.NewWriter(logger, log.WarnLevel)
	wError := logio.NewWriter(logger, log.ErrorLevel)

	c, err := loader.Load(chartpath)
	if err != nil {
		return err
	}

	_, ok := c.Metadata.Annotations["hypper.cattle.io/shared-dependencies"]
	if !ok {
		fmt.Fprintf(wWarn, "No shared dependencies in %s\n", chartpath)
		return nil
	}

	depYaml := c.Metadata.Annotations["hypper.cattle.io/shared-dependencies"]
	var deps dependencies
	if err = yaml.UnmarshalStrict([]byte(depYaml), &deps); err != nil {
		fmt.Fprintf(wError, "Chart.yaml metadata is malformed for chart %s\n", chartpath)
		return err
	}

	return d.printSharedDependencies(chartpath, logger, deps, settings)
}

// SharedDependencyStatus returns a string describing the status of a dependency viz a viz the releases in context.
func (d *SharedDependency) SharedDependencyStatus(depChart *chart.Chart, settings *cli.EnvSettings) (string, error) {

	// obtain the dep ns: either shared-dep has annotations, or the parent has, or we use the default ns
	depNS := GetNamespace(depChart, GetNamespace(depChart, settings.Namespace()))

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
		// For now, this is all we can check:
		// Releases don't contain semver or repository info,
		// checking RBAC to compare ns and error is not possible, as deps don't
		// record ns and use the dependee namespace. So either we see them installed, or we don't.
		if r.Name == depName && r.Namespace == depNS {
			return r.Info.Status.String(), nil
		}
	}

	return "not-installed", nil
}

// printSharedDependencies prints all of the shared dependencies in the yaml file.
// It will respect settings.NamespaceFromFlag when iterating through releases.
func (d *SharedDependency) printSharedDependencies(chartpath string, logger log.Logger, deps dependencies, settings *cli.EnvSettings) error {

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

		// obtain the dep ns: either shared-dep has annotations, or the parent has, or we use the default ns
		depNS := GetNamespace(depChart, GetNamespace(depChart, settings.Namespace()))

		if settings.NamespaceFromFlag {
			d.Namespace = settings.Namespace()
		} else {
			// look for releases in the specific ns that we are searching into
			d.Config.SetNamespace(depNS)
		}
		d.Config.SetNamespace(d.Namespace)

		if settings.NamespaceFromFlag && d.Namespace != depNS {
			// skip listing this dep, it's not in the same name as the flag
			continue
		}

		depStatus, err := d.SharedDependencyStatus(depChart, settings)
		if err != nil {
			return err
		}
		table.AddRow(depChart.Name(), dep.Version, dep.Repository, depStatus, depNS)
	}
	log.Infof(table.String())
	return nil
}
