package main

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/mattfarina/hypper/pkg/cli"
	"github.com/mattfarina/hypper/pkg/eyecandy"
	"github.com/mattfarina/log-go"
	logcli "github.com/mattfarina/log-go/impl/cli"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	helmAction "helm.sh/helm/v3/pkg/action"
	kubefake "helm.sh/helm/v3/pkg/kube/fake"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
)

var logger = logcli.NewStandard()
var settings = cli.New(logger)

func main() {
	log.Current = logger
	actionConfig := new(helmAction.Configuration)
	cmd, err := newRootCmd(actionConfig, logger, os.Args[1:])
	if settings.Debug {
		logger.Level = log.DebugLevel
	}
	if settings.Verbose {
		logger.Level = log.TraceLevel

		// When verbose is enabled then debug is also enabled
		settings.Debug = true
	}

	if err != nil {
		logger.Debug(eyecandy.Magenta("%v"), err)
		os.Exit(1)
	}

	cobra.OnInitialize(func() {
		helmDriver := os.Getenv("HELM_DRIVER")
		if err := actionConfig.Init(settings.HelmSettings.RESTClientGetter(), settings.HelmSettings.Namespace(), helmDriver, logger.Debugf); err != nil {
			log.Fatal(err)
		}
		if helmDriver == "memory" {
			loadReleasesInMemory(actionConfig)
		}
	})

	if err := cmd.Execute(); err != nil {
		logger.Debug(eyecandy.Magenta("%v"), err)
		os.Exit(1)
	}
}

// This function loads releases into the memory storage if the
// environment variable is properly set.
func loadReleasesInMemory(actionConfig *helmAction.Configuration) {
	filePaths := strings.Split(os.Getenv("HELM_MEMORY_DRIVER_DATA"), ":")
	if len(filePaths) == 0 {
		return
	}

	store := actionConfig.Releases
	mem, ok := store.Driver.(*driver.Memory)
	if !ok {
		// For an unexpected reason we are not dealing with the memory storage driver.
		return
	}

	actionConfig.KubeClient = &kubefake.PrintingKubeClient{Out: ioutil.Discard}

	for _, path := range filePaths {
		b, err := ioutil.ReadFile(path)
		if err != nil {
			log.Fatal("Unable to read memory driver data", err)
		}

		releases := []*release.Release{}
		if err := yaml.Unmarshal(b, &releases); err != nil {
			log.Fatal("Unable to unmarshal memory driver data: ", err)
		}

		for _, rel := range releases {
			if err := store.Create(rel); err != nil {
				log.Fatal(err)
			}
		}
	}
	// Must reset namespace to the proper one
	mem.SetNamespace(settings.HelmSettings.Namespace())
}
