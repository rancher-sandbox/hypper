package main

import (
	"testing"
)

func TestInstallCmd(t *testing.T) {
	tests := []cmdTestCase{
		// Install, base case
		{
			name: "basic install",
			cmd:  "install empty testdata/testcharts/empty",
		},
		{
			name:      "install with no chart specified",
			cmd:       "install",
			wantError: true,
		},
	}
	runTestActionCmd(t, tests)
}
