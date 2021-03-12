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

package main

import (
	"path/filepath"

	"github.com/Masterminds/log-go"
	"github.com/spf13/cobra"

	"github.com/rancher-sandbox/hypper/pkg/action"
	"helm.sh/helm/v3/cmd/helm/require"
	"helm.sh/helm/v3/pkg/chart/loader"
)

const sharedDependencyDesc = `
Manage the shared dependencies of a chart.

Helm charts store their shared dependencies in
'annotations.hypper.cattle.io/shared-dependencies' in Chart.yaml.

For example, this Chart.yaml declares two shared dependencies:

    # Chart.yaml
    annotations:
      hypper.cattle.io/shared-dependencies: |
	- name: prometheus
	  version: "13.3.1"
      repository: "https://example.com/charts"
	- name: postgresql
	  version: "10.3.11"
      repository: "https://another.example.com/charts"


The 'name' should be the name of a chart, where that name must match the name
in that chart's 'Chart.yaml' file.

The 'version' field should contain a semantic version or version range.

The 'repository' URL should point to a Chart Repository. Hypper expects that by
appending '/index.yaml' to the URL, it should be able to retrieve the chart
repository's index. Note: 'repository' can be an alias. The alias must start
with 'alias:' or '@'.

The repository can also be defined as the path to the directory of the
dependency charts stored locally. The path should start with a prefix of
"file://". For example,

    # Chart.yaml
    dependencies:
    - name: nginx
      version: "1.2.3"
      repository: "file://../dependent_chart/nginx"

If the dependency chart is retrieved locally, it is not required to have the
repository added to hypper by "hypper add repo". Version matching is also
supported for this case.
`

const sharedDependencyListDesc = `
List all of the shared dependencies declared in a chart, showing their statuses.

This can take chart archives and chart directories as input. It will not alter
the contents of a chart.

This will produce an error if the chart cannot be loaded, or the YAML annotations
of the shared dependencies is malformed.
`

func newSharedDependencyCmd(cfg *action.Configuration, logger log.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "shared-dep list",
		Aliases: []string{"shared-deps", "shared-dependencies"},
		Short:   "manage a chart's shared dependencies",
		Long:    sharedDependencyDesc,
		Args:    require.NoArgs,
	}

	cmd.AddCommand(newSharedDependencyListCmd(cfg, logger))

	return cmd
}

func newSharedDependencyListCmd(cfg *action.Configuration, logger log.Logger) *cobra.Command {
	client := action.NewSharedDependency(cfg)

	cmd := &cobra.Command{
		Use:     "list CHART",
		Aliases: []string{"ls"},
		Short:   "list the dependencies for the given chart",
		Long:    sharedDependencyListDesc,
		Args:    require.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(args, client, logger)
		},
	}
	return cmd
}

func runList(args []string, client *action.SharedDependency, logger log.Logger) error {
	chartpath := "."
	if len(args) > 0 {
		chartpath = filepath.Clean(args[0])
	}

	c, err := loader.Load(chartpath)
	if err != nil {
		return err
	}

	if settings.NamespaceFromFlag {
		client.Namespace = settings.Namespace()
	} else {
		client.SetNamespace(c, settings.Namespace())
	}

	client.Config.SetNamespace(client.Namespace)

	return client.List(chartpath, logger)
}
