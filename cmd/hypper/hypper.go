package main

import (
	"fmt"
	"log"
	"os"

	"github.com/mattfarina/hypper/pkg/cli"
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/action"
)

var settings = cli.New()

func debug(format string, v ...interface{}) {
	if settings.Debug {
		format = fmt.Sprintf("[debug] %s\n", format)
		log.Output(2, fmt.Sprintf(format, v...))
	}
}

func main() {
	actionConfig := new(action.Configuration)
	cmd, err := newRootCmd(actionConfig, os.Stdout, os.Args[1:])
	if err != nil {
		debug("%v", err)
		os.Exit(1)
	}

	cobra.OnInitialize(func() {
	})

	if err := cmd.Execute(); err != nil {
		debug("%+v", err)
		os.Exit(1)
	}
}
