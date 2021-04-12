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

	"github.com/Masterminds/log-go"
	logcli "github.com/Masterminds/log-go/impl/cli"
	"github.com/rancher-sandbox/hypper/internal/test"
	"github.com/rancher-sandbox/hypper/pkg/cli"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/chart"
)

func installAction(t *testing.T) *Install {
	config := actionConfigFixture(t)
	instAction := NewInstall(config)
	instAction.Namespace = "spaced"
	instAction.ReleaseName = "test-install-release"

	return instAction
}

func TestInstallAllSharedDeps(t *testing.T) {
	for _, tcase := range []struct {
		name      string
		chart     *chart.Chart
		golden    string
		wantError bool
		error     string
		wantDebug bool
		debug     string
	}{
		{
			name:      "chart has no shared-deps",
			chart:     buildChart(withHypperAnnotations()),
			golden:    "output/install-no-shared-deps.txt",
			wantDebug: true,
		},
		{
			name:      "chart metadata has malformed yaml",
			chart:     buildChart(withMalformedSharedDeps()),
			golden:    "output/install-malformed-shared-deps.txt",
			wantError: true,
			error:     "yaml: line 2: mapping values are not allowed in this context",
		},
		{
			name:      "dependencies get correctly installed",
			chart:     buildChart(withHypperAnnotations(), withSharedDeps()),
			golden:    "output/install-correctly-shared-deps.txt",
			wantDebug: true,
		},
		{
			name:   "dependencies are already installed",
			chart:  buildChart(withHypperAnnotations(), withSharedDeps()),
			golden: "output/install-no-shared-metadata.txt",
		},
	} {
		settings := cli.New()
		settings.Debug = tcase.wantDebug

		// create our own Logger that satisfies impl/cli.Logger, but with a buffer for tests
		buf := new(bytes.Buffer)
		logger := logcli.NewStandard()
		logger.InfoOut = buf
		logger.WarnOut = buf
		logger.ErrorOut = buf
		logger.DebugOut = buf
		if tcase.wantDebug {
			logger.Level = log.DebugLevel
		}
		log.Current = logger

		instAction := installAction(t)

		err := instAction.InstallAllSharedDeps(tcase.chart, settings, log.Current)
		if (err != nil) != tcase.wantError {
			t.Errorf("on test %q expected error, got '%v'", tcase.name, err)
		}
		if tcase.wantError {
			is := assert.New(t)
			is.Equal(tcase.error, err.Error())
		}
		if tcase.golden != "" {
			test.AssertGoldenBytes(t, buf.Bytes(), tcase.golden)
		}
	}
}

func TestInstallSharedDep(t *testing.T) {
	is := assert.New(t)

	// create our own Logger that satisfies impl/cli.Logger, but with a buffer for tests
	buf := new(bytes.Buffer)
	logger := logcli.NewStandard()
	logger.InfoOut = buf
	logger.WarnOut = buf
	logger.ErrorOut = buf
	log.Current = logger

	settings := cli.New()

	instAction := installAction(t)

	// TODO dependency version is not satisfied

	// check that install options such as DryRun are passed
	instAction.DryRun = true
	dep := &chart.Dependency{
		Name:       "testdata/charts/vanilla-helm",
		Repository: "",
		Version:    "0.1.0",
	}
	res, err := instAction.InstallSharedDep(dep, settings, log.Current)
	instAction.DryRun = false
	if err != nil {
		t.Fatalf("Failed install: %s", err)
	}
	is.Equal(res.Info.Status.String(), "pending-install", "Expected status of the installed dependency.")

	// install dependency correctly
	dep = &chart.Dependency{
		Name:       "testdata/charts/vanilla-helm",
		Repository: "",
		Version:    "0.1.0",
	}
	res, err = instAction.InstallSharedDep(dep, settings, log.Current)
	if err != nil {
		t.Fatalf("Failed install: %s", err)
	}
	is.Equal(res.Name, "empty", "Expected release name from dependency.")
	is.Equal(res.Namespace, "spaced", "Expected parent ns.")
	is.Equal(res.Info.Status.String(), "deployed", "Expected status of the installed dependency.")

	// install non-existent dependency
	dep = &chart.Dependency{
		Name:       "nonexistent-chart",
		Repository: "",
		Version:    "0.1.0",
	}
	_, err = instAction.InstallSharedDep(dep, settings, log.Current)
	if err == nil {
		t.Fatal(err)
	}
	is.Equal("failed to download \"nonexistent-chart\" (hint: running `helm repo update` may help)", err.Error())
}

func TestInstallSetNamespace(t *testing.T) {
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

func TestName(t *testing.T) {
	is := assert.New(t)

	// too many args
	chart := buildChart()
	_, err := GetName(chart, "", "name1", "chart-uri", "extraneous-arg")
	if err == nil {
		t.Fatal("expected an error")
	}
	is.Equal("expected at most two arguments, unexpected arguments: extraneous-arg", err.Error())

	// name and chart as args
	chart = buildChart()
	name, err := GetName(chart, "", "name1", "chart-uri")
	if err != nil {
		t.Fatal(err)
	}
	is.Equal("name1", name)

	// hypper annotations have priority over fallback annotations
	chart = buildChart(withHypperAnnotations(), withFallbackAnnotations())
	name, err = GetName(chart, "", "chart-uri")
	if err != nil {
		t.Fatal(err)
	}
	is.Equal("my-hypper-name", name)

	// fallback annotations
	chart = buildChart(withFallbackAnnotations())
	name, err = GetName(chart, "", "chart-uri")
	if err != nil {
		t.Fatal(err)
	}
	is.Equal("fleet", name)

	// no name or annotations present
	chart = buildChart()
	name, err = GetName(chart, "", "chart-uri")
	if err != nil {
		t.Fatal(err)
	}
	is.Equal("hello", name)
}

func TestChart(t *testing.T) {
	is := assert.New(t)
	instAction := installAction(t)

	// too many args
	_, err := instAction.Chart([]string{"name1", "chart-uri1", "extraneous-arg"})
	if err == nil {
		t.Fatal("expected an error")
	}
	is.Equal("expected at most two arguments, unexpected arguments: extraneous-arg", err.Error())

	// name and chart as args
	charturi, err := instAction.Chart([]string{"name2", "chart-uri2"})
	if err != nil {
		t.Fatal(err)
	}
	is.Equal("chart-uri2", charturi)

	// only chart as args
	charturi, err = instAction.Chart([]string{"chart-uri3"})
	if err != nil {
		t.Fatal(err)
	}
	is.Equal("chart-uri3", charturi)
}

func TestNameAndChart(t *testing.T) {
	is := assert.New(t)
	instAction := installAction(t)

	_, _, err := instAction.NameAndChart([]string{""})
	if err == nil {
		t.Fatal("expected an error")
	}
	is.Equal("NameAndChart() cannot be used", err.Error())
}

func TestCheckIfInstallable(t *testing.T) {
	is := assert.New(t)

	// Application chart type is installable
	err := CheckIfInstallable(buildChart(withTypeApplication()))
	is.NoError(err)

	// any other chart type besides Application is not installable
	err = CheckIfInstallable(buildChart(withTypeLibrary()))
	if err == nil {
		t.Fatal("expected an error")
	}
	is.Equal("library charts are not installable", err.Error())
}
