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
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"

	"github.com/Masterminds/log-go"
	logcli "github.com/Masterminds/log-go/impl/cli"
	"github.com/rancher-sandbox/hypper/internal/test"
	hypperChart "github.com/rancher-sandbox/hypper/pkg/chart"
	"github.com/rancher-sandbox/hypper/pkg/cli"
)

func newSharedDepFixture(t *testing.T, ns string) *SharedDependency {
	sd := NewSharedDependency(actionConfigFixture(t))
	sd.Config.SetNamespace(ns)
	return sd
}

func TestSharedDepsList(t *testing.T) {
	for _, tcase := range []struct {
		chart     string
		golden    string
		wantError bool
	}{
		{
			chart:     "no/such/chart",
			wantError: true,
		},
		{
			chart:  "testdata/charts/vanilla-helm",
			golden: "output/shared-deps-no-deps.txt",
		},
		{
			chart:     "testdata/charts/shared-deps-malformed",
			golden:    "output/shared-deps-malformed.txt",
			wantError: true,
		},
		{
			chart:  "testdata/charts/shared-deps",
			golden: "output/shared-deps-some-deps.txt",
		},
	} {
		// create our own Logger that satisfies impl/cli.Logger, but with a buffer for tests
		buf := new(bytes.Buffer)
		logger := logcli.NewStandard()
		logger.InfoOut = buf
		logger.WarnOut = buf
		logger.ErrorOut = buf
		log.Current = logger

		settings := cli.New()

		sharedDepAction := newSharedDepFixture(t, "hypper")
		err := sharedDepAction.List(tcase.chart, settings, log.Current)
		if (err != nil) != tcase.wantError {
			t.Errorf("expected error, got '%v'", err)
		}
		if tcase.golden != "" {
			test.AssertGoldenBytes(t, buf.Bytes(), tcase.golden)
		}
	}
}

func TestSharedDepsSetNamespace(t *testing.T) {
	is := assert.New(t)

	// chart without annotations
	instAction := installAction(t)
	chart := buildChart()
	SetNamespace(instAction, chart, "defaultns", false)
	is.Equal("defaultns", instAction.Namespace)

	// hypper annotations have priority over fallback annotations
	instAction = installAction(t)
	chart = buildChart(withHypperAnnotations(), withFallbackAnnotations())
	SetNamespace(instAction, chart, "defaultns", false)
	is.Equal("hypper", instAction.Namespace)

	// fallback annotations have priority over default ns
	instAction = installAction(t)
	chart = buildChart(withFallbackAnnotations())
	SetNamespace(instAction, chart, "defaultns", false)
	is.Equal("fleet-system", instAction.Namespace)
}

func TestSharedDependencyStatus(t *testing.T) {
	mk := func(name string, vers int, status release.Status, namespace string, chrtVer string) *release.Release {
		return release.Mock(&release.MockReleaseOptions{
			Name:      name,
			Version:   vers,
			Status:    status,
			Namespace: namespace,
			Chart: hypperChart.Mock(&hypperChart.MockChartOptions{
				Name:    name,
				Version: chrtVer,
			}),
		})
	}

	releasesFixture := []*release.Release{
		mk("my-hypper-name", 10, release.StatusDeployed, "hypper", "0.1.0"),
		mk("musketeers", 10, release.StatusPendingInstall, "hypper", "0.1.0"),
		mk("dartagnan", 9, release.StatusFailed, "default", "0.1.0"),
		mk("aramis", 3, release.StatusDeployed, "other-ns", "0.1.0"),
	}

	for _, tcase := range []struct {
		name       string
		depChart   *chart.Chart
		depNS      string
		depVersion string
		output     string
		wantError  bool
		err        string
		releases   []*release.Release
	}{
		{
			name:       "shared dep is installed and found",
			depChart:   buildChart(withHypperAnnotations(), withChartVersion("0.1.0")),
			depNS:      "hypper",
			depVersion: "~0.1.0",
			output:     "deployed",
			releases:   releasesFixture,
		},
		{
			name:       "shared dep not installed",
			depChart:   buildChart(withHypperAnnotValues("cow", "hypper")),
			depNS:      "hypper",
			depVersion: "~0.1.0",
			output:     "not-installed",
			releases:   releasesFixture,
		},
		{
			name:       "shared dep without hypper annot, uses default ns",
			depChart:   buildChart(withName("dartagnan"), withChartVersion("0.1.0")),
			depNS:      "default",
			depVersion: "0.1.0",
			output:     "failed",
			releases:   releasesFixture,
		},
		{
			name:       "shared dep version not parseable",
			depChart:   buildChart(withName("dartagnan")),
			depNS:      "default",
			depVersion: "foo0.1.0",
			wantError:  true,
			err:        "dependency version not parseable",
			releases:   releasesFixture,
		},
		{
			name:       "shared dep version out-of-range",
			depChart:   buildChart(withName("dartagnan")),
			depNS:      "default",
			depVersion: "~1.1.0",
			output:     "out-of-range",
			releases:   releasesFixture,
		},
	} {
		is := assert.New(t)
		sharedDepAction := newSharedDepFixture(t, tcase.depNS)
		store := storage.Init(driver.NewMemory())
		sharedDepAction.Config.Releases = store

		for _, r := range tcase.releases {
			if err := store.Create(r); err != nil {
				t.Fatal(err)
			}
		}

		if mem, ok := store.Driver.(*driver.Memory); ok {
			mem.SetNamespace(tcase.depNS)
		}

		depStatus, err := sharedDepAction.SharedDependencyStatus(tcase.depChart, tcase.depNS, tcase.depVersion)
		if (err != nil) != tcase.wantError {
			t.Errorf("expected error, got '%v'", err)
		}
		if tcase.wantError {
			is.Equal(tcase.err, err.Error())
		}
		if tcase.output != "" {
			is.Equal(tcase.output, depStatus)
		}
	}
}
