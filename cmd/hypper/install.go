/*
Copyright SUSE LLC.

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
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"helm.sh/helm/v3/cmd/helm/require"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli/output"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/release"

	"github.com/Masterminds/log-go"
	logio "github.com/Masterminds/log-go/io"
	"github.com/rancher-sandbox/hypper/pkg/action"
	"github.com/rancher-sandbox/hypper/pkg/eyecandy"
)

const installDesc = `
This command installs a chart.

The install argument must be a chart reference, a path to a packaged chart,
a path to an unpacked chart directory or a URL.

There are four different ways you can select the release name and namespace
where the chart will be installed. By priority order:

1. By the args passed from the CLI: hypper install mymaria example/mariadb -n system
2. By using hypper.cattle.io annotations in the Chart.yaml
3. By using catalog.cattle.io annotations in the Chart.yaml
4. By using the chart name from the Chart.yaml if nothing else is specified
`

func newInstallCmd(actionConfig *action.Configuration, logger log.Logger) *cobra.Command {
	client := action.NewInstall(actionConfig)
	valueOpts := &values.Options{}
	var outfmt output.Format

	cmd := &cobra.Command{
		Use:   "install [NAME] [CHART]",
		Short: "install a chart",
		Long:  installDesc,
		Args:  require.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			// TODO decide how to use returned rel:
			_, err := runInstall(args, client, valueOpts, logger)
			if err != nil {
				return err
			}
			logger.Info(eyecandy.ESPrint(settings.NoEmojis, "Done! :clapping_hands:"))
			return nil
		},
	}
	f := cmd.Flags()
	addInstallFlags(cmd, f, client, valueOpts)
	addValueOptionsFlags(f, valueOpts)
	addChartPathOptionsFlags(f, &client.ChartPathOptions)
	bindOutputFlag(cmd, &outfmt)
	return cmd
}

func addInstallFlags(cmd *cobra.Command, f *pflag.FlagSet, client *action.Install, valueOpts *values.Options) {
	f.BoolVar(&client.CreateNamespace, "create-namespace", false, "create the release namespace if not present")
}

func runInstall(args []string, client *action.Install, valueOpts *values.Options, logger log.Logger) (*release.Release, error) {

	// Get an io.Writer compliant logger instance at the info level.
	wInfo := logio.NewWriter(logger, log.InfoLevel)

	logger.Debugf("Original chart version: %q", client.Version)
	if client.Version == "" && client.Devel {
		logger.Debug("setting version to >0.0.0-0")
		client.Version = ">0.0.0-0"
	}

	chart, err := client.Chart(args)
	if err != nil {
		return nil, err
	}

	cp, err := client.ChartPathOptions.LocateChart(chart, settings.EnvSettings)
	if err != nil {
		return nil, err
	}

	logger.Debugf("CHART PATH: %s\n", cp)

	p := getter.All(settings.EnvSettings)
	vals, err := valueOpts.MergeValues(p)
	if err != nil {
		return nil, err
	}

	// Check chart dependencies to make sure all are present in /charts
	chartRequested, err := loader.Load(cp)
	if err != nil {
		return nil, err
	}

	// Set namespace for the install client
	action.SetNamespace(client, chartRequested, settings.Namespace(), settings.NamespaceFromFlag)

	if client.ReleaseName == "" {
		client.ReleaseName, err = action.GetName(chartRequested, client.NameTemplate, args...)
		if err != nil {
			return nil, err
		}
	}

	if err := checkIfInstallable(chartRequested); err != nil {
		return nil, err
	}

	if chartRequested.Metadata.Deprecated {
		logger.Warn("This chart is deprecated")
	}

	if req := chartRequested.Metadata.Dependencies; req != nil {
		// If CheckDependencies returns an error, we have unfulfilled dependencies.
		// As of Helm 2.4.0, this is treated as a stopping condition:
		// https://github.com/helm/helm/issues/2209
		if err := action.CheckDependencies(chartRequested, req); err != nil {
			if client.DependencyUpdate {
				man := &downloader.Manager{
					Out:              wInfo,
					ChartPath:        cp,
					Keyring:          client.ChartPathOptions.Keyring,
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

	return client.Run(chartRequested, vals)
}

// checkIfInstallable validates if a chart can be installed
//
// Application chart type is only installable
func checkIfInstallable(ch *chart.Chart) error {
	switch ch.Metadata.Type {
	case "", "application":
		return nil
	}
	return errors.Errorf("%s charts are not installable", ch.Metadata.Type)
}
