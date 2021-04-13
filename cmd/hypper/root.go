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
	"errors"
	"os"

	"github.com/Masterminds/log-go"
	"github.com/fatih/color"
	"github.com/rancher-sandbox/hypper/pkg/action"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var globalUsage = `Usage: hypper command

A package manager built on Helm charts and Helm itself.
`

func newRootCmd(actionConfig *action.Configuration, logger log.Logger, args []string) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:          "hypper",
		Short:        "A package manager built on Helm charts and Helm itself",
		Long:         globalUsage,
		SilenceUsage: true,
	}

	flags := cmd.PersistentFlags()
	settings.AddFlags(flags)

	cmd.AddCommand(
		newInstallCmd(actionConfig, logger),
		newUninstallCmd(actionConfig, logger),
		newListCmd(actionConfig, logger),
		newStatusCmd(actionConfig, logger),
		newRepoCmd(logger),
		newUpgradeCmd(actionConfig, logger),
		newSharedDependencyCmd(actionConfig, logger),
		newVersionCmd(logger),
		newLintCmd(logger),
		newSearchCmd(logger),
	)

	flags.ParseErrorsWhitelist.UnknownFlags = true
	err := flags.Parse(args)

	// Flags are parsed, lets fill the helm cli.Settings with our current settings
	settings.FillHelmSettings()

	if err != nil && !errors.Is(err, pflag.ErrHelp) {
		log.Errorf("failed while parsing flags for %s: %s", args, err)

		os.Exit(1)
	}

	flags.Visit(func(f *pflag.Flag) {
		if f.Name == "namespace" || f.Name == "n" {
			settings.NamespaceFromFlag = true
		}
	})

	if settings.NoColors {
		color.NoColor = true // disable colorized output
	}

	return cmd, nil
}
