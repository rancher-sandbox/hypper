package main

import (
	"fmt"
	"log"
	"os"

	"github.com/fatih/color"
	"github.com/mattfarina/hypper/pkg/cli"
	"github.com/spf13/cobra"
)

var settings = cli.New()

var red = color.New(color.FgRed).SprintFunc()
var yellow = color.New(color.FgYellow).SprintFunc()
var blue = color.New(color.FgBlue).SprintFunc()
var magenta = color.New(color.FgMagenta).SprintFunc()

func debug(format string, v ...interface{}) {
	if settings.Debug {
		format = fmt.Sprintf("[debug] %s\n", magenta(format))
		log.Output(2, fmt.Sprintf(format, v...))
	}
}

func main() {

	cmd, err := newRootCmd(os.Stdout, os.Args[1:])
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
