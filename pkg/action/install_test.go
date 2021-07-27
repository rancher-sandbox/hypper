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
	"bufio"
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/Masterminds/log-go"
	logcli "github.com/Masterminds/log-go/impl/cli"
	"github.com/rancher-sandbox/hypper/internal/solver"
	"github.com/rancher-sandbox/hypper/internal/test"
	"github.com/rancher-sandbox/hypper/pkg/chart"
	"github.com/rancher-sandbox/hypper/pkg/cli"
	"github.com/stretchr/testify/assert"

	helmChart "helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/time"
)

func installAction(t *testing.T) *Install {
	config := actionConfigFixture(t)
	instAction := NewInstall(config)
	instAction.Namespace = "spaced"
	instAction.ReleaseName = "test-install-release"

	return instAction
}

func TestInstallRun(t *testing.T) {

	for _, tcase := range []struct {
		name                  string
		chart                 *helmChart.Chart
		golden                string
		wantError             bool
		error                 string
		wantDebug             bool
		debug                 string
		addRelStub            bool
		optionalDeps          optionalDepsStrategy
		wantNSFromFlag        string
		numReturnedRels       int
		wantDryRun            bool
		skipActionReleaseName bool
		NoRepo                bool
	}{
		{
			name:            "non existent cache and repo.yaml, chart with no deps",
			chart:           buildChart(withHypperAnnotations()),
			golden:          "output/install-no-shared-deps.txt",
			numReturnedRels: 1,
			NoRepo:          true,
		},
		{
			name:            "chart has no shared-deps",
			chart:           buildChart(withHypperAnnotations()),
			golden:          "output/install-no-shared-deps.txt",
			addRelStub:      true,
			numReturnedRels: 1,
		},
		{
			name:                  "chart has no shared-deps and no action releaseName",
			chart:                 buildChart(withHypperAnnotations()),
			golden:                "output/install-no-action-release-name.txt",
			addRelStub:            true,
			numReturnedRels:       1,
			skipActionReleaseName: true,
		},
		{
			name:            "chart metadata has malformed yaml",
			chart:           buildChart(withMalformedSharedDeps()),
			golden:          "output/install-malformed-shared-deps.txt",
			error:           "yaml: line 2: mapping values are not allowed in this context",
			wantError:       true,
			numReturnedRels: 0,
		},
		{
			name:            "dependencies get correctly installed",
			chart:           buildChart(withHypperAnnotations(), withSharedDeps()),
			golden:          "output/install-correctly-shared-deps.txt",
			numReturnedRels: 2,
		},
		{
			name:            "dependencies without annotations get correctly installed",
			chart:           buildChart(withHypperAnnotations(), withSharedDepsWithoutAnnotations()),
			golden:          "output/install-shared-deps-without-annotations.txt",
			numReturnedRels: 2,
		},
		{
			name:            "dependencies with NamespaceFromFlag get correctly installed",
			chart:           buildChart(withHypperAnnotations(), withSharedDeps()),
			golden:          "output/install-shared-deps-with-ns-from-flag.txt",
			wantNSFromFlag:  "ns-from-flag",
			numReturnedRels: 2,
		},
		{
			name:            "dependencies with repo file:// correctly installed",
			chart:           buildChart(withHypperAnnotations(), withSharedDepsFileRepo()),
			golden:          "output/install-correctly-shared-deps-repo-file.txt",
			numReturnedRels: 2,
		},
		{
			name:            "looped dependencies with repo file:// correctly installed",
			chart:           buildChart(withHypperAnnotations(), withSharedDepsLoopedFileRepo()),
			golden:          "output/install-correctly-shared-deps-looped-repo-file.txt",
			numReturnedRels: 2,
		},
		{
			name:            "dependencies are already installed",
			chart:           buildChart(withHypperAnnotations(), withSharedDeps()),
			golden:          "output/install-shared-dep-installed.txt",
			addRelStub:      true,
			numReturnedRels: 1,
		},
		{
			name:            "dependencies are already installed in out-of-range ver",
			chart:           buildChart(withHypperAnnotations(), withOutOfRangeSharedDeps()),
			golden:          "output/install-shared-dep-installed-out-of-range.txt",
			addRelStub:      true,
			wantError:       true,
			error:           "Chart \"hello\" depends on \"my-shared-dep\" in namespace \"my-shared-dep-ns\", semver \"1.1.0\", but nothing satisfies it",
			numReturnedRels: 0,
		},
		{
			name:            "optional dependencies get correctly installed",
			chart:           buildChart(withHypperAnnotations(), withOptionalSharedDeps()),
			golden:          "output/install-correctly-optional-shared-deps.txt",
			optionalDeps:    OptionalDepsAll,
			numReturnedRels: 2,
		},
		{
			name:            "optional dependencies get correctly skipped",
			chart:           buildChart(withHypperAnnotations(), withOptionalSharedDeps()),
			golden:          "output/skip-optional-shared-deps.txt",
			optionalDeps:    OptionalDepsNone,
			numReturnedRels: 1,
		},
		{
			name:            "install chart and dependency with --dry-run",
			chart:           buildChart(withHypperAnnotations(), withSharedDeps()),
			golden:          "output/install-shared-deps-dry-run.txt",
			numReturnedRels: 2,
			wantDryRun:      true,
		},
	} {
		t.Run(tcase.name, func(t *testing.T) {

			var settings *cli.EnvSettings
			if tcase.wantNSFromFlag != "" {
				settings = cli.NewWithNamespace(tcase.wantNSFromFlag)
				settings.NamespaceFromFlag = true
			} else {
				settings = cli.New()
			}
			if tcase.NoRepo {
				settings.RepositoryCache = "non-existent-dir"
				settings.RepositoryConfig = "non-existent-dir/repositories.yaml"
			} else {
				settings.RepositoryCache = "testdata/hypperhome/hypper/repository"
				settings.RepositoryConfig = "testdata/hypperhome/hypper/repositories.yaml"
			}
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
			instAction.OptionalDeps = tcase.optionalDeps
			instAction.DryRun = tcase.wantDryRun

			if tcase.skipActionReleaseName {
				instAction.ReleaseName = ""
			}

			if tcase.addRelStub {
				now := time.Now()
				rel := &release.Release{
					Name: "my-shared-dep",
					Info: &release.Info{
						FirstDeployed: now,
						LastDeployed:  now,
						Status:        release.StatusDeployed,
						Description:   "Named Release Stub",
					},
					Version:   1,
					Namespace: "my-shared-dep-ns",
					Chart:     buildChart(withName("testdata/charts/shared-dep"), withChartVersion("0.1.0")),
				}
				instAction.Config.SetNamespace("spaced")
				err := instAction.Config.Releases.Create(rel)
				if err != nil {
					t.Fatalf("Failed creating rel stub: %s", err)
				}
			}
			is := assert.New(t)

			rels, err := instAction.Run(solver.InstallOne, tcase.chart, map[string]interface{}{}, settings, log.Current)
			is.Equal(tcase.numReturnedRels, len(rels))

			if (err != nil) && !tcase.wantError {
				t.Errorf("on test %q, got unexpected error '%v'", tcase.name, err)
			}

			if tcase.wantError {
				is.Equal(tcase.error, err.Error())
			} else {
				if tcase.wantDryRun {
					for _, r := range rels {
						is.Equal("pending-install", r.Info.Status.String(), "Expected status of the installed dependency.")
					}
				}
			}

			if tcase.golden != "" {
				test.AssertGoldenBytes(t, buf.Bytes(), tcase.golden)
			}

		})

	}
}

// func TestInstallPkg(t *testing.T) {
// }

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

func TestPromptBool(t *testing.T) {
	defaultDep := &chart.Dependency{
		Dependency: &helmChart.Dependency{
			Name:       "testdata/charts/vanilla-helm",
			Version:    "^0.1.0",
			Repository: "",
		},
	}
	for _, tcase := range []struct {
		name      string
		dep       *chart.Dependency
		input     string
		doInstall bool
	}{
		{
			name:      "prompt for yes",
			dep:       defaultDep,
			input:     "yes",
			doInstall: true,
		},
		{
			name:      "prompt for y",
			dep:       defaultDep,
			input:     "y",
			doInstall: true,
		},
		{
			name:      "prompt for no",
			dep:       defaultDep,
			input:     "no",
			doInstall: false,
		},
		{
			name:      "prompt for n",
			dep:       defaultDep,
			input:     "n",
			doInstall: false,
		},
		{
			name:      "prompt for yEs",
			dep:       defaultDep,
			input:     "yEs",
			doInstall: true,
		},
		{
			name:      "prompt for enter",
			dep:       defaultDep,
			input:     "",
			doInstall: true,
		},
	} {
		is := assert.New(t)

		// create our own Logger that satisfies impl/cli.Logger, but with a buffer for tests
		buf := new(bytes.Buffer)
		logger := logcli.NewStandard()
		logger.InfoOut = buf
		log.Current = logger

		reader := bufio.NewReader(strings.NewReader(tcase.input + "\n"))
		question := fmt.Sprintf("Install optional shared dependency \"%s\" ?", tcase.dep.Name)

		is.Equal(tcase.doInstall, promptBool(question, reader, logger))
	}
}
