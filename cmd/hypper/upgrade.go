/*
Copyright The Helm Authors, SUSE LLC

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
	"time"

	"github.com/Masterminds/log-go"
	logio "github.com/Masterminds/log-go/io"
	"github.com/rancher-sandbox/hypper/pkg/eyecandy"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/rancher-sandbox/hypper/pkg/action"
	"helm.sh/helm/v3/cmd/helm/require"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli/output"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/storage/driver"
)

const upgradeDesc = `
This command upgrades a release to a new version of a chart.

The upgrade arguments must be a chart. The chart
argument can be either: a chart reference('example/mariadb'), a path to a chart directory,
a packaged chart, or a fully qualified URL. For chart references, the latest
version will be specified unless the '--version' flag is set.

There are four different ways you can select the release name and namespace
where the chart will be upgraded. By priority order:

1. By the args passed from the CLI: hypper upgrade --release-name mariadb example/mariadb --namespace system
2. By using hypper.cattle.io annotations in the Chart.yaml
3. By using catalog.cattle.io annotations in the Chart.yaml
4. By using the current namespace as configured with the kubeconfig

To override values in a chart, use either the '--values' flag and pass in a file
or use the '--set' flag and pass configuration from the command line, to force string
values, use '--set-string'. In case a value is large and therefore
you want not to use neither '--values' nor '--set', use '--set-file' to read the
single large value from file.

You can specify the '--values'/'-f' flag multiple times. The priority will be given to the
last (right-most) file specified. For example, if both myvalues.yaml and override.yaml
contained a key called 'Test', the value set in override.yaml would take precedence:

    $ helm upgrade -f myvalues.yaml -f override.yaml ./redis

You can specify the '--set' flag multiple times. The priority will be given to the
last (right-most) set specified. For example, if both 'bar' and 'newbar' values are
set for a key called 'foo', the 'newbar' value would take precedence:

    $ helm upgrade --set foo=bar --set foo=newbar ./redis
`

func newUpgradeCmd(cfg *action.Configuration, logger log.Logger) *cobra.Command {
	client := action.NewUpgrade(cfg)
	valueOpts := &values.Options{}
	var outfmt output.Format
	var createNamespace bool

	cmd := &cobra.Command{
		Use:   "upgrade [CHART]",
		Short: "upgrade a chart",
		Long:  upgradeDesc,
		Args:  require.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client.Namespace = settings.Namespace()

			wInfo := logio.NewWriter(logger, log.InfoLevel)

			chartPath, err := client.ChartPathOptions.LocateChart(args[0], settings.EnvSettings)
			if err != nil {
				return err
			}

			// Check chart dependencies to make sure all are present in /charts
			ch, err := loader.Load(chartPath)
			if err != nil {
				return err
			}
			if req := ch.Metadata.Dependencies; req != nil {
				if err := action.CheckDependencies(ch, req); err != nil {
					return err
				}
			}

			if ch.Metadata.Deprecated {
				logger.Warn(eyecandy.ESPrint(settings.NoEmojis, "This chart has been deprecated :exclamation:"))

			}
			if client.ReleaseName == "" {
				client.ReleaseName, err = action.GetName(ch, "")
				if err != nil {
					return err
				}
				logger.Debugf("client.ReleaseName was empty, setting release name to %s", client.ReleaseName)
			}

			// Fixes helm#7002 - Support reading values from STDIN for `upgrade` command
			// Must load values AFTER determining if we have to call install so that values loaded from stdin are are not read twice
			if client.Install {
				// If a release does not exist, install it.
				histClient := action.NewHistory(cfg)
				histClient.Max = 1
				if _, err := histClient.Run(client.ReleaseName); err == driver.ErrReleaseNotFound {
					// Only print this to stdout for table output
					if outfmt == output.Table {
						logger.Info(eyecandy.ESPrintf(settings.NoEmojis, "Release %q does not exist. Installing it now. :toolbox:\n", client.ReleaseName))
					}
					instClient := action.NewInstall(cfg)
					// Set namespace for the install client
					action.SetNamespace(client, ch, settings.Namespace(), settings.NamespaceFromFlag)

					instClient.CreateNamespace = createNamespace
					instClient.ChartPathOptions = client.ChartPathOptions
					instClient.DryRun = client.DryRun
					instClient.DisableHooks = client.DisableHooks
					instClient.SkipCRDs = client.SkipCRDs
					instClient.Timeout = client.Timeout
					instClient.Wait = client.Wait
					instClient.WaitForJobs = client.WaitForJobs
					instClient.Devel = client.Devel
					instClient.Namespace = client.Namespace
					instClient.Atomic = client.Atomic
					instClient.PostRenderer = client.PostRenderer
					instClient.DisableOpenAPIValidation = client.DisableOpenAPIValidation
					instClient.SubNotes = client.SubNotes
					instClient.Description = client.Description
					instClient.ReleaseName = client.ReleaseName

					rel, err := runInstall(args, instClient, valueOpts, logger, settings)
					if err != nil {
						return err
					}
					return outfmt.Write(wInfo, &statusPrinter{rel, settings.Debug, false})
				} else if err != nil {
					return err
				}
			}

			if client.Version == "" && client.Devel {
				logger.Debug("setting version to >0.0.0-0")
				client.Version = ">0.0.0-0"
			}

			vals, err := valueOpts.MergeValues(getter.All(settings.EnvSettings))
			if err != nil {
				return err
			}
			// Set namespace for the upgrade client
			action.SetNamespace(client, ch, settings.Namespace(), settings.NamespaceFromFlag)

			rel, err := client.Run(client.ReleaseName, ch, vals)
			if err != nil {
				return errors.Wrap(err, "UPGRADE FAILED")
			}

			if outfmt == output.Table {
				logger.Info(eyecandy.ESPrintf(settings.NoEmojis, "Release %q has been upgraded :partying_face:", client.ReleaseName))
			}

			return outfmt.Write(wInfo, &statusPrinter{rel, settings.Debug, false})

		},
	}

	f := cmd.Flags()
	f.BoolVar(&createNamespace, "create-namespace", false, "if --install is set, create the release namespace if not present")
	f.BoolVarP(&client.Install, "install", "i", false, "if a release by this name doesn't already exist, run an install")
	f.BoolVar(&client.Devel, "devel", false, "use development versions, too. Equivalent to version '>0.0.0-0'. If --version is set, this is ignored")
	f.BoolVar(&client.DryRun, "dry-run", false, "simulate an upgrade")
	f.BoolVar(&client.Recreate, "recreate-pods", false, "performs pods restart for the resource if applicable")
	_ = f.MarkDeprecated("recreate-pods", "functionality will no longer be updated. Consult the documentation for other methods to recreate pods")
	f.BoolVar(&client.Force, "force", false, "force resource updates through a replacement strategy")
	f.BoolVar(&client.DisableHooks, "no-hooks", false, "disable pre/post upgrade hooks")
	f.BoolVar(&client.DisableOpenAPIValidation, "disable-openapi-validation", false, "if set, the upgrade process will not validate rendered templates against the Kubernetes OpenAPI Schema")
	f.BoolVar(&client.SkipCRDs, "skip-crds", false, "if set, no CRDs will be installed when an upgrade is performed with install flag enabled. By default, CRDs are installed if not already present, when an upgrade is performed with install flag enabled")
	f.DurationVar(&client.Timeout, "timeout", 300*time.Second, "time to wait for any individual Kubernetes operation (like Jobs for hooks)")
	f.BoolVar(&client.ResetValues, "reset-values", false, "when upgrading, reset the values to the ones built into the chart")
	f.BoolVar(&client.ReuseValues, "reuse-values", false, "when upgrading, reuse the last release's values and merge in any overrides from the command line via --set and -f. If '--reset-values' is specified, this is ignored")
	f.BoolVar(&client.Wait, "wait", false, "if set, will wait until all Pods, PVCs, Services, and minimum number of Pods of a Deployment, StatefulSet, or ReplicaSet are in a ready state before marking the release as successful. It will wait for as long as --timeout")
	f.BoolVar(&client.WaitForJobs, "wait-for-jobs", false, "if set and --wait enabled, will wait until all Jobs have been completed before marking the release as successful. It will wait for as long as --timeout")
	f.BoolVar(&client.Atomic, "atomic", false, "if set, upgrade process rolls back changes made in case of failed upgrade. The --wait flag will be set automatically if --atomic is used")
	f.IntVar(&client.MaxHistory, "history-max", settings.MaxHistory, "limit the maximum number of revisions saved per release. Use 0 for no limit")
	f.BoolVar(&client.CleanupOnFail, "cleanup-on-fail", false, "allow deletion of new resources created in this upgrade when upgrade fails")
	f.BoolVar(&client.SubNotes, "render-subchart-notes", false, "if set, render subchart notes along with the parent")
	f.StringVar(&client.Description, "description", "", "add a custom description")
	f.StringVar(&client.ReleaseName, "release-name", "", "add a custom release name, overrides annotations")
	addChartPathOptionsFlags(f, &client.ChartPathOptions)
	addValueOptionsFlags(f, valueOpts)
	bindOutputFlag(cmd, &outfmt)
	bindPostRenderFlag(cmd, &client.PostRenderer)

	return cmd
}
