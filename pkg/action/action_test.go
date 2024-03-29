/*
Copyright The Helm Authors, SUSE LLC.

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
package action

import (
	"flag"
	"io/ioutil"
	"os"
	"testing"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	kubefake "helm.sh/helm/v3/pkg/kube/fake"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
)

var verbose = flag.Bool("test.log", false, "enable test logging")

func actionConfigFixture(t *testing.T) *Configuration {
	t.Helper()

	tdir, err := ioutil.TempDir("", "helm-action-test")
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { os.RemoveAll(tdir) })

	helmActionConfig := action.Configuration{
		Releases:     storage.Init(driver.NewMemory()),
		KubeClient:   &kubefake.FailingKubeClient{PrintingKubeClient: kubefake.PrintingKubeClient{Out: ioutil.Discard}},
		Capabilities: chartutil.DefaultCapabilities,
		Log: func(format string, v ...interface{}) {
			t.Helper()
			if *verbose {
				t.Logf(format, v...)
			}
		},
	}
	actionConfig := &Configuration{
		Configuration: &helmActionConfig,
	}
	return actionConfig
}

type chartOptions struct {
	*chart.Chart
}

type chartOption func(*chartOptions)

func buildChart(opts ...chartOption) *chart.Chart {
	c := &chartOptions{
		Chart: &chart.Chart{
			// TODO: This should be more complete.
			Metadata: &chart.Metadata{
				APIVersion: "v1",
				Name:       "hello",
				Version:    "0.1.0",
			},
			// This adds a basic template and hooks.
			Templates: []*chart.File{
				{Name: "templates/hello", Data: []byte("hello: world")},
			},
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c.Chart
}

func withHypperAnnotations() chartOption {
	return func(opts *chartOptions) {
		if opts.Chart.Metadata.Annotations == nil {
			opts.Chart.Metadata.Annotations = make(map[string]string)
		}
		opts.Chart.Metadata.Annotations["hypper.cattle.io/namespace"] = "hypper"
		opts.Chart.Metadata.Annotations["hypper.cattle.io/release-name"] = "my-hypper-name"
	}
}

func withFallbackAnnotations() chartOption {
	return func(opts *chartOptions) {
		if opts.Chart.Metadata.Annotations == nil {
			opts.Chart.Metadata.Annotations = make(map[string]string)
		}
		opts.Chart.Metadata.Annotations["catalog.cattle.io/namespace"] = "fleet-system"
		opts.Chart.Metadata.Annotations["catalog.cattle.io/release-name"] = "fleet"
	}
}

func withMalformedSharedDeps() chartOption {
	return func(opts *chartOptions) {
		if opts.Chart.Metadata.Annotations == nil {
			opts.Chart.Metadata.Annotations = make(map[string]string)
		}
		// this ought to have wrong indentation:
		opts.Chart.Metadata.Annotations["hypper.cattle.io/shared-dependencies"] = `- name: vanilla-helm
   version: "0.1.0"
 repository: "file://testdata/vanilla-helm"`
	}
}

func withSharedDeps() chartOption {
	return func(opts *chartOptions) {
		if opts.Chart.Metadata.Annotations == nil {
			opts.Chart.Metadata.Annotations = make(map[string]string)
		}
		opts.Chart.Metadata.Annotations["hypper.cattle.io/shared-dependencies"] = "  - name: \"testdata/charts/shared-dep\"" + "\n" +
			"    version: \"0.1.0\"" + "\n" +
			"    repository: \"\"" + "\n"
	}
}

func withOutOfRangeSharedDeps() chartOption {
	return func(opts *chartOptions) {
		if opts.Chart.Metadata.Annotations == nil {
			opts.Chart.Metadata.Annotations = make(map[string]string)
		}
		opts.Chart.Metadata.Annotations["hypper.cattle.io/shared-dependencies"] = "  - name: \"testdata/charts/shared-dep\"" + "\n" +
			"    version: \"1.1.0\"" + "\n" +
			"    repository: \"\"" + "\n"
	}
}

func withOptionalSharedDeps() chartOption {
	return func(opts *chartOptions) {
		if opts.Chart.Metadata.Annotations == nil {
			opts.Chart.Metadata.Annotations = make(map[string]string)
		}
		opts.Chart.Metadata.Annotations["hypper.cattle.io/namespace"] = "luke-skywalker"
		opts.Chart.Metadata.Annotations["hypper.cattle.io/optional-dependencies"] = "  - name: \"testdata/charts/vanilla-helm\"" + "\n" +
			"    version: \"0.1.0\"" + "\n" +
			"    repository: \"\"" + "\n"
	}
}

func withSharedDepsWithoutAnnotations() chartOption {
	return func(opts *chartOptions) {
		if opts.Chart.Metadata.Annotations == nil {
			opts.Chart.Metadata.Annotations = make(map[string]string)
		}
		opts.Chart.Metadata.Annotations["hypper.cattle.io/shared-dependencies"] = "  - name: \"testdata/charts/vanilla-helm\"" + "\n" +
			"    version: \"0.1.0\"" + "\n" +
			"    repository: \"\"" + "\n"
	}
}

func withSharedDepsFileRepo() chartOption {
	return func(opts *chartOptions) {
		if opts.Chart.Metadata.Annotations == nil {
			opts.Chart.Metadata.Annotations = make(map[string]string)
		}
		opts.Chart.Metadata.Annotations["hypper.cattle.io/shared-dependencies"] = "  - name: \"shared-dep-empty\"" + "\n" +
			"    version: \"0.1.0\"" + "\n" +
			"    repository: \"file://../shared-dep\"" + "\n"
	}
}

func withSharedDepsLoopedFileRepo() chartOption {
	return func(opts *chartOptions) {
		if opts.Chart.Metadata.Annotations == nil {
			opts.Chart.Metadata.Annotations = make(map[string]string)
		}
		opts.Chart.Metadata.Annotations["hypper.cattle.io/shared-dependencies"] = "  - name: \"local-dep-empty\"" + "\n" +
			"    version: \"0.1.0\"" + "\n" +
			"    repository: \"file://../dep-repo-local\"" + "\n"
	}
}

func withTypeApplication() chartOption {
	return func(opts *chartOptions) {
		opts.Chart.Metadata.Type = "application"
	}
}

func withTypeLibrary() chartOption {
	return func(opts *chartOptions) {
		opts.Chart.Metadata.Type = "library"
	}
}

func withHypperAnnotValues(name string, ns string) chartOption {
	return func(opts *chartOptions) {
		if opts.Chart.Metadata.Annotations == nil {
			opts.Chart.Metadata.Annotations = make(map[string]string)
		}
		opts.Chart.Metadata.Annotations["hypper.cattle.io/namespace"] = ns
		opts.Chart.Metadata.Annotations["hypper.cattle.io/release-name"] = name
	}
}

func withName(name string) chartOption {
	return func(opts *chartOptions) {
		opts.Metadata.Name = name
	}
}

func withChartVersion(semver string) chartOption {
	return func(opts *chartOptions) {
		opts.Metadata.Version = semver
	}
}
