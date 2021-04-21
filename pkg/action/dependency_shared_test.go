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

	"github.com/Masterminds/log-go"
	logcli "github.com/Masterminds/log-go/impl/cli"
	"github.com/rancher-sandbox/hypper/internal/test"
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
			chart:  "testdata/charts/hypper-annot",
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

// func TestSharedDependencyStatus(t *testing.T) {
// 	mk := func(name string, vers int, status release.Status, namespace string) *release.Release {
// 		return release.Mock(&release.MockReleaseOptions{
// 			Name:      name,
// 			Version:   vers,
// 			Status:    status,
// 			Namespace: namespace,
// 		})
// 	}

// 	releasesFixture := []*release.Release{
// 		mk("my-hypper-name", 3, release.StatusDeployed, "hypper"),
// 		mk("musketeers", 10, release.StatusPendingInstall, "hypper"),
// 		mk("dartagnan", 9, release.StatusSuperseded, "default"),
// 	}

// 	for _, tcase := range []struct {
// 		name      string
// 		chart     *chart.Chart
// 		ns        string
// 		output    string
// 		wantError bool
// 		error     string
// 		releases  []*rspb.Release
// 	}{
// 		{
// 			name:     "shared dep is installed and found",
// 			chart:    buildChart(withHypperAnnotations()),
// 			ns:       "hypper",
// 			output:   "deployed",
// 			releases: releasesFixture,
// 		},
// 		{
// 			name:     "shared dep not installed",
// 			chart:    buildChart(withHypperAnnotValues("cow", "other-ns")),
// 			ns:       "hypper",
// 			output:   "not-installed",
// 			releases: releasesFixture,
// 		},
// 		{
// 			name:     "shared dep without hypper annot, uses default ns",
// 			chart:    buildChart(withName("dartagnan")),
// 			ns:       "default",
// 			output:   "superseeded",
// 			releases: releasesFixture,
// 		},
// 	} {
// 		is := assert.New(t)
// 		sharedDepAction := newSharedDepFixture(t, tcase.ns)

// 		storage := storage.Init(driver.NewMemory())
// 		for _, r := range tcase.releases {
// 			if err := storage.Create(r); err != nil {
// 				t.Fatal(err)
// 			}
// 		}

// 		// sharedDepAction.Config.SetNamespace(tcase.ns) // from the other day, but not needed, sharedDepAction should do it on its own
// 		sharedDepAction.Config.Releases = storage

// 		depStatus, err := sharedDepAction.SharedDependencyStatus(tcase.chart, tcase.ns)
// 		if (err != nil) != tcase.wantError {
// 			t.Errorf("expected error, got '%v'", err)
// 		}
// 		if tcase.output != "" {
// 			is.Equal(tcase.output, depStatus)
// 		}
// 	}
// }
