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
			cmd:    "install zeppelin testdata/testcharts/hypper-annot -n=led",
			golden: "output/install-name-ns-args.txt",
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
