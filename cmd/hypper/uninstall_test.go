package main

import (
	"testing"

	"helm.sh/helm/v3/pkg/release"
)

func TestUninstall(t *testing.T) {
	tests := []cmdTestCase{
		{
			name:   "basic uninstall",
			cmd:    "uninstall aeneas",
			golden: "output/uninstall.txt",
			rels:   []*release.Release{release.Mock(&release.MockReleaseOptions{Name: "aeneas"})},
		},
		{
			name:   "multiple uninstall",
			cmd:    "uninstall aeneas aeneas2",
			golden: "output/uninstall-multiple.txt",
			rels: []*release.Release{
				release.Mock(&release.MockReleaseOptions{Name: "aeneas"}),
				release.Mock(&release.MockReleaseOptions{Name: "aeneas2"}),
			},
		},
		{
			name:   "uninstall with timeout",
			cmd:    "uninstall aeneas --timeout 120s",
			golden: "output/uninstall-timeout.txt",
			rels:   []*release.Release{release.Mock(&release.MockReleaseOptions{Name: "aeneas"})},
		},
		{
			name:   "uninstall without hooks",
			cmd:    "uninstall aeneas --no-hooks",
			golden: "output/uninstall-no-hooks.txt",
			rels:   []*release.Release{release.Mock(&release.MockReleaseOptions{Name: "aeneas"})},
		},
		{
			name:   "keep history",
			cmd:    "uninstall aeneas --keep-history",
			golden: "output/uninstall-keep-history.txt",
			rels:   []*release.Release{release.Mock(&release.MockReleaseOptions{Name: "aeneas"})},
		},
		{
			name:      "uninstall without release",
			cmd:       "uninstall",
			golden:    "output/uninstall-no-args.txt",
			wantError: true,
		},
	}
	runTestCmd(t, tests)
}
