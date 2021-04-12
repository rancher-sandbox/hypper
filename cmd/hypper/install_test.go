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
	"testing"
)

func TestInstallCmd(t *testing.T) {
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
			cmd:    "install zeppelin testdata/testcharts/hypper-annot -n led --no-shared-deps",
			golden: "output/install-name-ns-args.txt",
		},
		// Install, name and namespace as args, create ns
		{
			name:   "install, name and ns as args",
			cmd:    "install purple testdata/testcharts/hypper-annot --namespace deep --create-namespace --no-shared-deps",
			golden: "output/install-create-namespace.txt",
		},
		// Install, hypper annot have priority over fallback annot
		{
			name:   "install, hypper annot have priority over fallback annot",
			cmd:    "install testdata/testcharts/hypper-annot --no-shared-deps",
			golden: "output/install-hypper-annot.txt",
		},
		// Install, fallback annotations
		{
			name:   "install, fallback annotations",
			cmd:    "install testdata/testcharts/fallback-annot --no-shared-deps",
			golden: "output/install-fallback-annot.txt",
		},
		// Install, annotations have priority over default name from Chart.yml
		{
			name:   "install, annot have priority over default name from Chart.yml",
			cmd:    "install testdata/testcharts/hypper-annot --no-shared-deps",
			golden: "output/install-hypper-annot.txt",
		},
		// Install, no name or annotations specified
		{
			name:   "install, with no name or annot specified",
			cmd:    "install testdata/testcharts/vanilla-helm --no-shared-deps",
			golden: "output/install-no-name-or-annot.txt",
		},
		// Install, with shared deps
		{
			name:   "install, with shared deps",
			cmd:    "install testdata/testcharts/shared-deps",
			golden: "output/install-with-shared-deps.txt",
		},
	}
	runTestActionCmd(t, tests)
}
