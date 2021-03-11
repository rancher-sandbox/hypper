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
	"time"

	"github.com/Masterminds/log-go"
	"github.com/rancher-sandbox/hypper/pkg/action"
	"github.com/rancher-sandbox/hypper/pkg/eyecandy"
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/cmd/helm/require"
)

var uninstallDesc = `remove a helm deployment by wrapping helm calls`

func newUninstallCmd(actionConfig *action.Configuration, logger log.Logger) *cobra.Command {
	client := action.NewUninstall(actionConfig)
	cmd := &cobra.Command{
		Use:        "uninstall [NAME]",
		Short:      "uninstall a deployment",
		Long:       uninstallDesc,
		Aliases:    []string{"del", "delete", "un"},
		SuggestFor: []string{"remove", "rm"},
		Args:       require.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			for i := 0; i < len(args); i++ {
				logger.Info(eyecandy.ESPrintf(settings.NoEmojis, ":fire: uninstalling %s", args[i]))
				client.Config.SetNamespace(settings.Namespace())
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
