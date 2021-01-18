package main

import (
	"io"

	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/action"
)

var globalUsage = `A package manager built on Helm charts and Helm itself.
`

func newRootCmd(actionConfig *action.Configuration, out io.Writer, args []string) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:          "hypper",
		Short:        "A package manager built on Helm charts and Helm itself",
		Long:         globalUsage,
		SilenceUsage: true,
	}

	cmd.AddCommand(
	// List of newCommandCmd from files
	)

	return cmd, nil
}
