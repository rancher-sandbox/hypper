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

package main

import (
	"runtime"
	"testing"

	"github.com/rancher-sandbox/hypper/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
)

func TestSharedDependencyListCmd(t *testing.T) {
	noSuchChart := cmdTestCase{
		name:      "No such chart",
		cmd:       "shared-deps list /no/such/chart",
		golden:    "output/shared-deps-list-no-chart-linux.txt",
		wantError: true,
	}

	noSharedDependencies := cmdTestCase{
		name:   "Chart doesn't have shared dependencies",
		cmd:    "shared-deps list testdata/testcharts/vanilla-helm",
		golden: "output/shared-deps-list-no-shared-deps-linux.txt",
	}

	noSharedDependenciesInstalled := cmdTestCase{
		name:   "No Shared dependencies installed",
		cmd:    "shared-deps list testdata/testcharts/shared-deps",
		golden: "output/shared-deps-list-not-installed.txt",
	}

	sharedDependenciesInstalled := cmdTestCase{
		name:   "Shared dependencies installed in correct ns",
		cmd:    "shared-deps list testdata/testcharts/shared-deps",
		golden: "output/shared-deps-list-installed.txt",
		rels: []*release.Release{release.Mock(&release.MockReleaseOptions{
			Name:      "my-shared-dep",
			Namespace: "my-shared-dep-ns",
			Chart: chart.Mock(&chart.MockChartOptions{
				Name:    "shared-dep-empty",
				Version: "0.1.1",
			}),
		})},
	}

	sharedDependenciesInstalledDiffNS := cmdTestCase{
		name:   "Shared dependencies installed in different ns, and not finding them",
		cmd:    "shared-deps list testdata/testcharts/shared-deps",
		golden: "output/shared-deps-list-installed-diff-ns.txt",
		rels: []*release.Release{release.Mock(&release.MockReleaseOptions{
			Name:      "my-shared-dep",
			Namespace: "other-ns",
			Chart: chart.Mock(&chart.MockChartOptions{
				Name:    "shared-dep-empty",
				Version: "0.1.1",
			}),
		})},
	}

	sharedDependenciesInstalledDiffNSFlag := cmdTestCase{
		name:   "Shared dependencies installed in different ns, passing -n flag to find them",
		cmd:    "shared-deps list testdata/testcharts/shared-deps -n ns-from-flag",
		golden: "output/shared-deps-list-installed-diff-ns-found.txt",
		rels: []*release.Release{release.Mock(&release.MockReleaseOptions{
			Name:      "my-shared-dep",
			Namespace: "ns-from-flag",
			Chart: chart.Mock(&chart.MockChartOptions{
				Name:    "shared-dep-empty",
				Version: "0.1.1",
			}),
		})},
	}

	sharedDependenciesInstalledDiffNSFlagNotFound := cmdTestCase{
		name:   "Shared dependencies installed in different ns, passing -n flag with incorrect ns and not finding them",
		cmd:    "shared-deps list testdata/testcharts/shared-deps -n ns-from-flag",
		golden: "output/shared-deps-list-installed-diff-ns-not-found.txt",
		rels: []*release.Release{release.Mock(&release.MockReleaseOptions{
			Name:      "my-shared-dep",
			Namespace: "ns-from-flag-not-found",
			Chart: chart.Mock(&chart.MockChartOptions{
				Name:    "shared-dep-empty",
				Version: "0.1.1",
			}),
		})},
	}

	sharedDependenciesInstalledVersionOutOfRange := cmdTestCase{
		name:   "Shared dependencies installed, semver out of range",
		cmd:    "shared-deps list testdata/testcharts/shared-deps -n my-shared-dep-ns",
		golden: "output/shared-deps-list-installed-out-of-range.txt",
		rels: []*release.Release{release.Mock(&release.MockReleaseOptions{
			Name:      "my-shared-dep",
			Namespace: "my-shared-dep-ns",
			Chart: chart.Mock(&chart.MockChartOptions{
				Name:    "my-shared-dep",
				Version: "3.1.0",
			}),
		})},
	}

	if runtime.GOOS == "windows" {
		noSuchChart.golden = "output/shared-deps-list-no-chart-windows.txt"
		noSharedDependencies.golden = "output/shared-deps-list-no-shared-deps-windows.txt"
	}

	tests := []cmdTestCase{
		noSuchChart,
		noSharedDependencies,
		noSharedDependenciesInstalled,
		sharedDependenciesInstalled,
		sharedDependenciesInstalledDiffNS,
		sharedDependenciesInstalledDiffNSFlag,
		sharedDependenciesInstalledDiffNSFlagNotFound,
		sharedDependenciesInstalledVersionOutOfRange,
	}
	runTestCmd(t, tests)
}
