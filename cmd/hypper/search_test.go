/*
Copyright The Helm Authors.

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

func TestSearchRepositoriesCmd(t *testing.T) {
	repoFile := "testdata/hypperhome/hypper/repositories.yaml"
	repoCache := "testdata/hypperhome/hypper/repository"

	tests := []cmdTestCase{{
		name:   "search for 'alpine', expect one match with latest stable version",
		cmd:    "search repo alpine",
		golden: "output/search-multiple-stable-release.txt",
	}, {
		name:   "search for 'alpine', expect one match with newest development version",
		cmd:    "search repo alpine --devel",
		golden: "output/search-multiple-devel-release.txt",
	}, {
		name:   "search for 'alpine' with versions, expect three matches",
		cmd:    "search repo alpine --versions",
		golden: "output/search-multiple-versions.txt",
	}, {
		name:   "search for 'alpine' with version constraint, expect one match with version 0.1.0",
		cmd:    "search repo alpine --version '>= 0.1, < 0.2'",
		golden: "output/search-constraint.txt",
	}, {
		name:   "search for 'alpine' with version constraint, expect one match with version 0.1.0",
		cmd:    "search repo alpine --versions --version '>= 0.1, < 0.2'",
		golden: "output/search-versions-constraint.txt",
	}, {
		name:   "search for 'alpine' with version constraint, expect one match with version 0.2.0",
		cmd:    "search repo alpine --version '>= 0.1'",
		golden: "output/search-constraint-single.txt",
	}, {
		name:   "search for 'alpine' with version constraint and --versions, expect two matches",
		cmd:    "search repo alpine --versions --version '>= 0.1'",
		golden: "output/search-multiple-versions-constraints.txt",
	}, {
		name:   "search for 'syzygy', expect no matches",
		cmd:    "search repo syzygy",
		golden: "output/search-not-found.txt",
	}, {
		name:   "search for 'alp[a-z]+', expect two matches",
		cmd:    "search repo alp[a-z]+ --regexp",
		golden: "output/search-regex.txt",
	}, {
		name:      "search for 'alp[', expect failure to compile regexp",
		cmd:       "search repo alp[ --regexp",
		wantError: true,
	}, {
		name:   "search for 'maria', expect valid json output",
		cmd:    "search repo maria --output json",
		golden: "output/search-output-json.txt",
	}, {
		name:   "search for 'alpine', expect valid yaml output",
		cmd:    "search repo alpine --output yaml",
		golden: "output/search-output-yaml.txt",
	}, {
		name:   "search for hypper, expect latest version 0.2.0",
		cmd:    "search repo hypper --output yaml",
		golden: "output/search-hypper-latest.txt",
	}, { // I use yaml for the output because its easier to create the golden files, sue me :p
		// Should find version 0.1.0 as that is the only one with that annotation
		name:   "search for hypper with annotation hypper.cattle.io/namespace:hyppernamespace",
		cmd:    "search repo hypper -a 'hypper.cattle.io/namespace=hyppernamespace' --output yaml",
		golden: "output/search-hypper-annotation.txt",
	}, {
		// Should find latest version 0.2.0 with that annotation
		name:   "search for hypper with regexp annotation hypper.cattle.io/release-name:hypper*",
		cmd:    "search repo hypper --regexp -a 'hypper.cattle.io/release-name=hypper*' --output yaml",
		golden: "output/search-hypper-annotation-regexp.txt",
	}, {
		// Should find that exact version with the annotation, not the latest
		name:   "search for hypper with version and annotation hypper.cattle.io/release-name:hyppername",
		cmd:    "search repo hypper --version 0.1.0 -a 'hypper.cattle.io/release-name=hyppername' --output yaml",
		golden: "output/search-hypper-annotation-version.txt",
	}, {
		// Now everything at the same time! regexp! version! annotations!
		name:   "search for hypper with version and annotation hypper.cattle.io/release-name:hyppername",
		cmd:    "search repo hypper --regexp --version 0.1.0 -a 'hypper.cattle.io/release-name=^hypper' --output yaml",
		golden: "output/search-hypper-annotation-version-regexp.txt",
	},
	}

	settings.Debug = true
	defer func() { settings.Debug = false }()

	for i := range tests {
		tests[i].cmd += " --repository-config " + repoFile
		tests[i].cmd += " --repository-cache " + repoCache
	}
	runTestCmd(t, tests)
}
