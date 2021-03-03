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
	)

	flags.ParseErrorsWhitelist.UnknownFlags = true
	err := flags.Parse(args)

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
