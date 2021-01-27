package main

import (
	"fmt"
	"log"
	"os"

	"github.com/fatih/color"
	"github.com/kyokomi/emoji/v2"
	"github.com/mattfarina/hypper/pkg/cli"
	"github.com/spf13/cobra"
	helmAction "helm.sh/helm/v3/pkg/action"
)

var settings = cli.New()

var blue = color.New(color.FgBlue).SprintFunc()
var magenta = color.New(color.FgMagenta).SprintFunc()

func debug(format string, v ...interface{}) {
	if settings.Debug {
		format = fmt.Sprintf("[debug] %s\n", magenta(format))
		_ = log.Output(2, fmt.Sprintf(format, v...))
	}
}

func warning(format string, v ...interface{}) {
	// TODO missing colors
	format = fmt.Sprintf("WARNING: %s\n", format)
	fmt.Fprintf(os.Stderr, format, v...)
}

func esPrintf(format string, v ...interface{}) string {
	if settings.NoEmojis {
		return fmt.Sprintf(cli.RemoveEmojiFromString(format), v...)
	}
	return emoji.Sprintf(format, v...)
}

func esPrint(s string) string {
	if settings.NoEmojis {
		return fmt.Sprint(cli.RemoveEmojiFromString(s))
	}
	return emoji.Sprint(s)
}

func main() {
	actionConfig := new(helmAction.Configuration)
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
	fmt.Println((esPrint(":wave: Welcome to") + esPrintf(" %s", blue("Hypper"))))
}
