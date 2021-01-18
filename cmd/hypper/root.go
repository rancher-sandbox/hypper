package main

import (
	"io"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var globalUsage = `Usage: hypper command

A package manager built on Helm charts and Helm itself.
`

func newRootCmd(out io.Writer, args []string) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:          "hypper",
		Short:        "A package manager built on Helm charts and Helm itself",
		Long:         globalUsage,
		SilenceUsage: false,
	}

	flags := cmd.PersistentFlags()
	settings.AddFlags(flags)

	cmd.AddCommand(
	// list of newsubcommandcmd from files
	)
	flags.Parse(args)

	if settings.NoColors {
		color.NoColor = true // disable colorized output
	}
	return cmd, nil
}
