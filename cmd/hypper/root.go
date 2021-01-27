package main

import (
	"os"

	"github.com/fatih/color"
	"github.com/mattfarina/log"
	"github.com/spf13/cobra"
	helmAction "helm.sh/helm/v3/pkg/action"
)

var globalUsage = `Usage: hypper command

A package manager built on Helm charts and Helm itself.
`

func newRootCmd(actionConfig *helmAction.Configuration, logger log.Logger, args []string) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:          "hypper",
		Short:        "A package manager built on Helm charts and Helm itself",
		Long:         globalUsage,
		SilenceUsage: false,
	}

	flags := cmd.PersistentFlags()
	settings.AddFlags(flags)

	cmd.AddCommand(
		newInstallCmd(actionConfig, logger),
	)
	err := flags.Parse(args)

	if err != nil {
		log.Errorf("failed while parsing flags for %s", args)
		os.Exit(1)
	}

	if settings.NoColors {
		color.NoColor = true // disable colorized output
	}
	return cmd, nil
}
