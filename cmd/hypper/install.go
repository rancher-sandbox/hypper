package main

import (
	"github.com/pkg/errors"
	"github.com/rancher-sandbox/hypper/pkg/eyecandy"
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/cmd/helm/require"
	helmAction "helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	helmCli "helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/release"

	"github.com/Masterminds/log-go"
	logio "github.com/Masterminds/log-go/io"
)

var installDesc = `install a helm chart by wrapping helm calls (for now)`

func newInstallCmd(actionConfig *helmAction.Configuration, logger log.Logger) *cobra.Command {
	client := helmAction.NewInstall(actionConfig)
	valuesOpts := &values.Options{}
	cmd := &cobra.Command{
		Use:   "install [NAME] [CHART]",
		Short: "install a chart",
		Long:  installDesc,
		Args:  require.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			logger.Info(eyecandy.ESPrintf(settings.NoEmojis, ":cruise_ship: Installing %s…", args[0]))
			// TODO decide how to use returned rel:
			_, err := runInstall(args, client, valuesOpts, logger)
			if err != nil {
				return err
			}
			logger.Info(eyecandy.ESPrint(settings.NoEmojis, "Done! :clapping_hands:"))
			return nil
		},
	}
	return cmd
}

func runInstall(args []string, client *helmAction.Install, valueOpts *values.Options, logger log.Logger) (*release.Release, error) {
	helmSettings := helmCli.New()
	// TODO add hypper specific code here

	// Get an io.Writer compliant logger instance at the info level.
	wInfo := logio.NewWriter(logger, log.InfoLevel)

	if client.Version == "" && client.Devel {
		logger.Debug("setting version to >0.0.0-0")
		client.Version = ">0.0.0-0"
	}

	name, chart, err := client.NameAndChart(args)
	if err != nil {
		return nil, err
	}
	client.ReleaseName = name

	cp, err := client.ChartPathOptions.LocateChart(chart, helmSettings)
	if err != nil {
		return nil, err
	}

	logger.Debugf("CHART PATH: %s\n", cp)

	p := getter.All(helmSettings)
	vals, err := valueOpts.MergeValues(p)
	if err != nil {
		return nil, err
	}

	// Check chart dependencies to make sure all are present in /charts
	chartRequested, err := loader.Load(cp)
	if err != nil {
		return nil, err
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
		if err := helmAction.CheckDependencies(chartRequested, req); err != nil {
			if client.DependencyUpdate {
				man := &downloader.Manager{
					Out:              wInfo,
					ChartPath:        cp,
					Keyring:          client.ChartPathOptions.Keyring,
					SkipUpdate:       false,
					Getters:          p,
					RepositoryConfig: helmSettings.RepositoryConfig,
					RepositoryCache:  helmSettings.RepositoryCache,
					Debug:            helmSettings.Debug,
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

	client.Namespace = helmSettings.Namespace()
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
