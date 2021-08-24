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
	"fmt"
	"testing"
)

func TestInstallCmd(t *testing.T) {

	repoCache := "testdata/testcharts"

	repoConfig := repoCache + "/repositories.yaml"

	tests := []cmdTestCase{
		// Install, no chart specified
		{
			name:      "install, no chart specified",
			cmd:       "install",
			golden:    "output/install-no-chart.txt",
			wantError: true,
		},
		// Install, name and namespace as args
		{
			name:   "install, name and ns as args",
			cmd:    fmt.Sprintf("install zeppelin testdata/testcharts/hypper-annot -n led --no-shared-deps --repository-config %s --repository-cache %s", repoConfig, repoCache),
			golden: "output/install-name-ns-args.txt",
		},
		// Install, name and namespace as args, no create ns
		{
			name:   "install, name and ns as args, don't create ns",
			cmd:    fmt.Sprintf("install purple testdata/testcharts/hypper-annot --namespace deep --no-create-namespace --no-shared-deps --repository-config %s --repository-cache %s", repoConfig, repoCache),
			golden: "output/install-no-create-namespace.txt",
			// wantError: true, there's no error, the client allows targetting any ns
		},
		// Install, hypper annot have priority over fallback annot
		{
			name:   "install, hypper annot have priority over fallback annot",
			cmd:    fmt.Sprintf("install testdata/testcharts/hypper-annot --no-shared-deps --repository-config %s --repository-cache %s", repoConfig, repoCache),
			golden: "output/install-hypper-annot.txt",
		},
		// Install, fallback annotations
		{
			name:   "install, fallback annotations",
			cmd:    fmt.Sprintf("install testdata/testcharts/fallback-annot --no-shared-deps --repository-config %s --repository-cache %s", repoConfig, repoCache),
			golden: "output/install-fallback-annot.txt",
		},
		// Install, annotations have priority over default name from Chart.yml
		{
			name:   "install, annot have priority over default name from Chart.yml",
			cmd:    fmt.Sprintf("install testdata/testcharts/hypper-annot --no-shared-deps --repository-config %s --repository-cache %s", repoConfig, repoCache),
			golden: "output/install-hypper-annot.txt",
		},
		// Install, no name or annotations specified
		{
			name:   "install, with no name or annot specified",
			cmd:    fmt.Sprintf("install testdata/testcharts/vanilla-helm --no-shared-deps --repository-config %s --repository-cache %s", repoConfig, repoCache),
			golden: "output/install-no-name-or-annot.txt",
		},
		// Install, with shared deps
		{
			name:   "install, with shared deps",
			cmd:    fmt.Sprintf("install testdata/testcharts/shared-deps --repository-config %s --repository-cache %s", repoConfig, repoCache),
			golden: "output/install-with-shared-deps.txt",
		},
		// Install, with shared deps out-of-range
		{
			name:      "install, with shared deps out-of-range",
			cmd:       fmt.Sprintf("install testdata/testcharts/shared-deps-out-of-range --repository-config %s --repository-cache %s", repoConfig, repoCache),
			golden:    "output/install-with-shared-deps-out-of-range.txt",
			wantError: true,
		},
		// Install, with all optional shared deps
		{
			name:   "install, with all optional shared deps",
			cmd:    fmt.Sprintf("install testdata/testcharts/shared-and-optional-deps --optional-deps all --repository-config %s --repository-cache %s", repoConfig, repoCache),
			golden: "output/install-with-all-optional-deps.txt",
		},
		// Install, skip optional shared deps
		{
			name:   "install, skip optional shared deps",
			cmd:    fmt.Sprintf("install testdata/testcharts/shared-and-optional-deps --optional-deps none --repository-config %s --repository-cache %s", repoConfig, repoCache),
			golden: "output/install-skip-all-optional-deps.txt",
		},

		// Install, ask for optional shared deps (default), tested in pkg/action/install_test.go

		// Install, incorrect flag value for optional shared deps
		{
			name:      "install, incorrect flag value for optional shared deps",
			cmd:       "install testdata/testcharts/shared-and-optional-deps --optional-deps foo",
			golden:    "output/install-incorrect-flag-optional-deps.txt",
			wantError: true,
		},

		// dry-run, with all optional shared deps
		{
			name:   "install dry-run, with all optional shared deps",
			cmd:    fmt.Sprintf("install testdata/testcharts/shared-and-optional-deps --optional-deps all --dry-run --repository-config %s --repository-cache %s", repoConfig, repoCache),
			golden: "output/install-dry-run-with-all-optional-deps.txt",
		},
	}
	runTestActionCmd(t, tests)
}
