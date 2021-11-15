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
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"helm.sh/helm/v3/cmd/helm/require"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli/output"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/release"

	"github.com/Masterminds/log-go"
	logio "github.com/Masterminds/log-go/io"
	"github.com/rancher-sandbox/hypper/internal/solver"
	"github.com/rancher-sandbox/hypper/pkg/action"
	"github.com/rancher-sandbox/hypper/pkg/eyecandy"
	"github.com/thediveo/enumflag"
)

type OptionalDepsMode enumflag.Flag

// define enum values of --optional-deps flag
const (
	OptionalDepsAsk OptionalDepsMode = iota
	OptionalDepsAll
	OptionalDepsNone
)

// map enum values of --optional-deps flag to string representation
var OptionalDepsModeIds = map[OptionalDepsMode][]string{
	OptionalDepsAsk:  {"ask"},
	OptionalDepsAll:  {"all"},
	OptionalDepsNone: {"none"},
}

var optionaldepsmode = OptionalDepsAsk

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
			_, err := runInstall(solver.InstallOne, args, client, valueOpts, logger)
			if err != nil {
				// Capturing a specific error message, when a chart in a repo
				// was called for but the repo was never added. Adding more
				// context for the user.
				if strings.HasPrefix(err.Error(), "unable to load dependency") {
					logger.Info("Please add any missing repositories, if necessary")
				}
				err = errors.New(eyecandy.ESPrintf(settings.NoEmojis, ":x: %s", err))
				return err
			}
			logger.Info(eyecandy.ESPrint(settings.NoEmojis, ":clapping_hands:Done!"))
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
	f.BoolVar(&client.NoCreateNamespace, "no-create-namespace", false, "don't create the release namespace if not present")
	f.BoolVar(&client.NoSharedDeps, "no-shared-deps", false, "skip installation of shared dependencies")
	f.Var(enumflag.New(&optionaldepsmode, "option", OptionalDepsModeIds, enumflag.EnumCaseInsensitive),
		"optional-deps", "install optional shared dependencies [ask|all|none]")
	f.BoolVar(&client.DryRun, "dry-run", false, "simulate an install")
	f.DurationVar(&client.Timeout, "timeout", 300*time.Second, "time to wait for any individual Kubernetes operation (like Jobs for hooks)")
	f.BoolVar(&client.Wait, "wait", false, "if set, will wait until all Pods, PVCs, Services, and minimum number of Pods of a Deployment, StatefulSet, or ReplicaSet are in a ready state before marking the release as successful. It will wait for as long as --timeout")
	f.BoolVar(&client.WaitForJobs, "wait-for-jobs", false, "if set and --wait enabled, will wait until all Jobs have been completed before marking the release as successful. It will wait for as long as --timeout")
}

func runInstall(strategy solver.SolverStrategy, args []string, client *action.Install, valueOpts *values.Options, logger log.Logger) ([]*release.Release, error) {

	// Get an io.Writer compliant logger instance at the info level.
	wInfo := logio.NewWriter(logger, log.InfoLevel)

	logger.Debugf("Original chart version: %q", client.Version)
	if client.Version == "" && client.Devel {
		logger.Debug("setting version to >0.0.0-0")
		client.Version = ">0.0.0-0"
	}

	// map flag to action.OptionalDeps strategy
	switch optionaldepsmode {
	case OptionalDepsAsk:
		client.OptionalDeps = action.OptionalDepsAsk
	case OptionalDepsAll:
		client.OptionalDeps = action.OptionalDepsAll
	case OptionalDepsNone:
		client.OptionalDeps = action.OptionalDepsNone
	}

	// map hypper's NoCreateNamespace to Helm's CreateNamespace
	client.CreateNamespace = !client.NoCreateNamespace

	chartName, err := client.Chart(args)
	if err != nil {
		return nil, err
	}

	chartPath, err := client.ChartPathOptions.LocateChart(chartName, settings.EnvSettings)
	if err != nil {
		return nil, err
	}

	logger.Debugf("CHART PATH: %s\n", chartPath)

	p := getter.All(settings.EnvSettings)
	vals, err := valueOpts.MergeValues(p)
	if err != nil {
		return nil, err
	}

	// Check chart dependencies to make sure all are present in /charts
	chartRequested, err := loader.Load(chartPath)
	if err != nil {
		return nil, err
	}

	// Set namespace for the install client
	action.SetNamespace(client, chartRequested, settings.Namespace(), settings.NamespaceFromFlag)

	if client.ReleaseName == "" {
		// calculate releaseName either from args, metadata, or chart name:
		client.ReleaseName, err = action.GetName(chartRequested, client.NameTemplate, args...)
		if err != nil {
			return nil, err
		}
	}

	if err := action.CheckIfInstallable(chartRequested); err != nil {
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
					ChartPath:        chartPath,
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
				if chartRequested, err = loader.Load(chartPath); err != nil {
					return nil, errors.Wrap(err, "failed reloading chart after repo update")
				}
			} else {
				return nil, err
			}
		}
	}

	return client.Run(solver.InstallOne, chartRequested, chartPath, vals, settings, logger)
}
