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

	"github.com/Masterminds/log-go"
	logcli "github.com/Masterminds/log-go/impl/cli"
	"github.com/rancher-sandbox/hypper/internal/test"
)

func newSharedDepFixture(t *testing.T, ns string) *SharedDependency {
	sd := NewSharedDependency(actionConfigFixture(t))
	sd.Namespace = ns
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

		sharedDepAction := newSharedDepFixture(t, "hypper")
		err := sharedDepAction.List(tcase.chart, log.Current)
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
	instAction.SetNamespace(chart, "defaultns")
	is.Equal("defaultns", instAction.Namespace)

	// hypper annotations have priority over fallback annotations
	instAction = installAction(t)
	chart = buildChart(withHypperAnnotations(), withFallbackAnnotations())
	instAction.SetNamespace(chart, "defaultns")
	is.Equal("hypper", instAction.Namespace)

	// fallback annotations have priority over default ns
	instAction = installAction(t)
	chart = buildChart(withFallbackAnnotations())
	instAction.SetNamespace(chart, "defaultns")
	is.Equal("fleet-system", instAction.Namespace)
}

func TestSharedDependencyStatus(t *testing.T) {
	is := assert.New(t)

	mk := func(name string, vers int, status release.Status, namespace string) *release.Release {
		return release.Mock(&release.MockReleaseOptions{
			Name:      name,
			Version:   vers,
			Status:    status,
			Namespace: namespace,
		})
	}

	// installed dep
	sharedDepAction := newSharedDepFixture(t, "hypper")
	dep := chart.Dependency{
		Name:       "mariadb",
		Version:    "10.5.9",
		Repository: "https://another.example.com/charts",
	}
	releases := []*release.Release{
		mk("mariadb", 3, release.StatusDeployed, "hypper"),
		mk("musketeers", 10, release.StatusSuperseded, "default"),
		mk("musketeers", 9, release.StatusSuperseded, "default"),
	}
	is.Equal("deployed", sharedDepAction.SharedDependencyStatus(&dep, releases))

	// print status of release matching dep
	sharedDepAction = newSharedDepFixture(t, "hypper")
	dep = chart.Dependency{
		Name:       "mariadb",
		Version:    "10.5.9",
		Repository: "https://another.example.com/charts",
	}
	releases = []*release.Release{
		mk("mariadb", 3, release.StatusPendingInstall, "hypper"),
	}
	is.Equal("pending-install", sharedDepAction.SharedDependencyStatus(&dep, releases))

	// not installed, but on the same ns
	sharedDepAction = newSharedDepFixture(t, "hypper")
	dep = chart.Dependency{
		Name:       "mariadb",
		Version:    "10.5.9",
		Repository: "https://another.example.com/charts",
	}
	releases = []*release.Release{
		mk("musketeers", 11, release.StatusDeployed, "hypper"),
		mk("musketeers", 10, release.StatusSuperseded, "hypper"),
		mk("carabins", 1, release.StatusSuperseded, "hypper"),
	}
	is.Equal("not-installed", sharedDepAction.SharedDependencyStatus(&dep, releases))

	// installed, but in a different namespace
	sharedDepAction = newSharedDepFixture(t, "other-ns")
	dep = chart.Dependency{
		Name:       "mariadb",
		Version:    "10.5.9",
		Repository: "https://another.example.com/charts",
	}
	releases = []*release.Release{
		mk("mariadb", 11, release.StatusDeployed, "hypper"),
		mk("musketeers", 10, release.StatusSuperseded, "hypper"),
		mk("carabins", 1, release.StatusSuperseded, "hypper"),
	}
	is.Equal("not-installed", sharedDepAction.SharedDependencyStatus(&dep, releases))

}
