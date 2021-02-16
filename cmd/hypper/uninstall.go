package main

import (
	"time"

	"github.com/Masterminds/log-go"
	"github.com/rancher-sandbox/hypper/pkg/eyecandy"
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/cmd/helm/require"
	helmAction "helm.sh/helm/v3/pkg/action"
)

var uninstallDesc = `remove a helm deployment by wrapping helm calls`

func newUninstallCmd(actionConfig *helmAction.Configuration, logger log.Logger) *cobra.Command {
	client := helmAction.NewUninstall(actionConfig)
	cmd := &cobra.Command{
		Use:   "uninstall [NAME]",
		Short: "uninstall a deployment",
		Long:  uninstallDesc,
		Args:  require.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			for i := 0; i < len(args); i++ {
				logger.Info(eyecandy.ESPrintf(settings.NoEmojis, ":fire: uninstalling %s", args[i]))
				res, err := client.Run(args[i])
				if err != nil {
					return err
				}
				if res != nil && res.Info != "" {
					logger.Info(res.Info)
				}
				logger.Info(eyecandy.ESPrintf(settings.NoEmojis, ":fire: release \"%s\" uninstalled", args[i]))
			}
			return nil
		},
	}

	f := cmd.Flags()
	f.BoolVar(&client.DryRun, "dry-run", false, "simulate a uninstall")
	f.BoolVar(&client.DisableHooks, "no-hooks", false, "prevent hooks from running during uninstallation")
	f.BoolVar(&client.KeepHistory, "keep-history", false, "remove all associated resources and mark the release as deleted, but retain the release history")
	f.DurationVar(&client.Timeout, "timeout", 300*time.Second, "time to wait for any individual Kubernetes operation (like Jobs for hooks)")
	f.StringVar(&client.Description, "description", "", "add a custom description")

	return cmd
}
