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
	"io"

	"github.com/Masterminds/log-go"
	logio "github.com/Masterminds/log-go/io"
	"github.com/gosuri/uitable"
	"gopkg.in/yaml.v2"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"

	"helm.sh/helm/v3/pkg/release"
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
func (d *SharedDependency) List(chartpath string, logger log.Logger) error {

	wInfo := logio.NewWriter(logger, log.InfoLevel)
	wWarn := logio.NewWriter(logger, log.WarnLevel)
	wError := logio.NewWriter(logger, log.ErrorLevel)

	c, err := loader.Load(chartpath)
	if err != nil {
		fmt.Fprintf(wWarn, "No shared dependencies in %s\n", chartpath)
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

	clientList := action.NewList(d.Config.Configuration)

	clientList.SetStateMask()

	releases, err := clientList.Run()
	if err != nil {
		return err
	}

	d.printSharedDependencies(chartpath, wInfo, deps, releases)
	return nil
}

// SharedDependencyStatus returns a string describing the status of a dependency viz a viz the releases in context.
func (d *SharedDependency) SharedDependencyStatus(dep *chart.Dependency, releases []*release.Release) string {
	// For now, this is all we can check:
	// Releases don't contain semver or repository info,
	// checking RBAC to compare ns and error is not possible, as deps don't
	// record ns and use the dependee namespace. So either we see them installed, or we don't.
	for _, v := range releases {
		if v.Name == dep.Name && v.Namespace == d.Namespace {
			return v.Info.Status.String()
		}
	}

	return "not-installed"
}

// printSharedDependencies prints all of the shared dependencies in the yaml file.
func (d *SharedDependency) printSharedDependencies(chartpath string, out io.Writer, deps dependencies, releases []*release.Release) {

	table := uitable.New()
	table.MaxColWidth = 80
	table.AddRow("NAME", "VERSION", "REPOSITORY", "STATUS")
	for _, v := range deps {
		table.AddRow(v.Name, v.Version, v.Repository, d.SharedDependencyStatus(v, releases))
	}
	fmt.Fprintln(out, table)
}
