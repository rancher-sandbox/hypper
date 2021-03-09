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
	"io/ioutil"
	"os"
	"strings"

	"github.com/Masterminds/log-go"
	logcli "github.com/Masterminds/log-go/impl/cli"
	"github.com/rancher-sandbox/hypper/pkg/action"
	"github.com/rancher-sandbox/hypper/pkg/cli"
	"github.com/rancher-sandbox/hypper/pkg/eyecandy"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	helmAction "helm.sh/helm/v3/pkg/action"
	kubefake "helm.sh/helm/v3/pkg/kube/fake"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
)

var settings = cli.New()

func main() {
	logger := logcli.NewStandard()
	log.Current = logger

	helmActionConfig := new(helmAction.Configuration)
	actionConfig := &action.Configuration{
		Configuration: helmActionConfig,
	}

	cmd, err := newRootCmd(actionConfig, log.Current, os.Args[1:])
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
		if err := actionConfig.Init(settings.RESTClientGetter(), settings.Namespace(), helmDriver, logger.Debugf); err != nil {
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
func loadReleasesInMemory(actionConfig *action.Configuration) {
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
	mem.SetNamespace(settings.Namespace())
}

// This re-Inits the config so we can set the Storage Driver namespace to what is obtained
// from the chart.
// This is because on helm you can only set the namespace via the --namespace flag which gets into
// the envsettings.namespace and the helm.go loads that var into the Configuration object during
// cobra.OnInitialize. To maintain compativility we do the same but then during the command execution
// we read the chart annotations and change the release namespace on the fly. This is ok for the
// namespace install but the Storage Driver, which is already loaded, has a different namespace (usually default)
// and the Release gets stored in the wrong namespace (stored *for* helm, for k8s is installed properly)
// This results into a disconnect between where the release is installed and where hypper/helm looks in the
// Storage Driver to search for it. With this below we re-Init the configuration and hance the Storage Driver to
// have the namespace correctly set in sync with the Release
func ReinitConfigForNamespace(cfg *action.Configuration, namespace string, logger log.Logger) {
	_ = cfg.Init(settings.RESTClientGetter(), namespace, os.Getenv("HELM_DRIVER"), logger.Debugf)
}
