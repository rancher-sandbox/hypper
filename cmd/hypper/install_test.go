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
			cmd:    "install zeppelin testdata/testcharts/hypper-annot -n led",
			golden: "output/install-name-ns-args.txt",
		},
		// Install, name and namespace as args, create ns
		{
			name:   "install, name and ns as args",
			cmd:    "install purple testdata/testcharts/hypper-annot --namespace deep --create-namespace",
			golden: "output/install-create-namespace.txt",
		},
		// Install, hypper annot have priority over fallback annot
		{
			name:   "install, hypper annot have priority over fallback annot",
			cmd:    "install testdata/testcharts/hypper-annot",
			golden: "output/install-hypper-annot.txt",
		},
		// Install, fallback annotations
		{
			name:   "install, fallback annotations",
			cmd:    "install testdata/testcharts/fallback-annot",
			golden: "output/install-fallback-annot.txt",
		},
		// Install, annotations have priority over generate-name
		{
			name:   "install, annot have priority over generate-name",
			cmd:    "install testdata/testcharts/hypper-annot --generate-name",
			golden: "output/install-hypper-annot.txt",
		},
		// Install, no name or annotations specified
		{
			name:      "install, with no name or annot specified",
			cmd:       "install testdata/testcharts/vanilla-helm",
			golden:    "output/install-no-name-or-annot.txt",
			wantError: true,
		},
	}
	runTestActionCmd(t, tests)
}
