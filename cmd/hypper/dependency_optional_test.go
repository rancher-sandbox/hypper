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
	"testing"

	"helm.sh/helm/v3/pkg/release"
)

func TestOptionalDependencyListCmd(t *testing.T) {
	noSuchChart := cmdTestCase{
		name:      "No such chart",
		cmd:       "shared-deps list /no/such/chart",
		golden:    "output/shared-deps-list-no-chart-linux.txt",
		wantError: true,
	}

	optionalDependencies := cmdTestCase{
		name:   "Optional dependencies installed in correct ns",
		cmd:    "optional-dep list testdata/testcharts/optional-deps",
		golden: "output/optional-deps-list-installed.txt",
		rels:   []*release.Release{release.Mock(&release.MockReleaseOptions{Name: "my-shared-dep", Namespace: "my-shared-dep-ns"})},
	}

	sharedAndOptionalDependencies := cmdTestCase{
		name:   "Optional dependencies installed in correct ns",
		cmd:    "optional-dep list testdata/testcharts/shared-and-optional-deps",
		golden: "output/shared-and-optional-deps-list-installed.txt",
		rels:   []*release.Release{release.Mock(&release.MockReleaseOptions{Name: "my-shared-dep", Namespace: "my-shared-dep-ns"})},
	}

	tests := []cmdTestCase{noSuchChart,
		optionalDependencies,
		sharedAndOptionalDependencies,
	}
	runTestCmd(t, tests)
}
