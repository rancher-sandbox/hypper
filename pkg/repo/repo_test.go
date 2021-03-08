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

package repo

import (
	"strings"
	"testing"

	helmRepo "helm.sh/helm/v3/pkg/repo"
)

const testRepositoriesFile = "testdata/repositories.yaml"

func TestFile(t *testing.T) {
	rf := NewFile()
	rf.Add(
		&helmRepo.Entry{
			Name: "stable",
			URL:  "https://example.com/stable/charts",
		},
		&helmRepo.Entry{
			Name: "incubator",
			URL:  "https://example.com/incubator",
		},
	)

	if len(rf.Repositories) != 2 {
		t.Fatal("Expected 2 repositories")
	}

	if rf.Has("nosuchrepo") {
		t.Error("Found nonexistent repo")
	}
	if !rf.Has("incubator") {
		t.Error("incubator repo is missing")
	}

	stable := rf.Repositories[0]
	if stable.Name != "stable" {
		t.Error("stable is not named stable")
	}
	if stable.URL != "https://example.com/stable/charts" {
		t.Error("Wrong URL for stable")
	}
}

func TestNewFile(t *testing.T) {
	expects := NewFile()
	expects.Add(
		&helmRepo.Entry{
			Name: "stable",
			URL:  "https://example.com/stable/charts",
		},
		&helmRepo.Entry{
			Name: "incubator",
			URL:  "https://example.com/incubator",
		},
	)

	file, err := LoadFile(testRepositoriesFile)
	if err != nil {
		t.Errorf("%q could not be loaded: %s", testRepositoriesFile, err)
	}

	if len(expects.Repositories) != len(file.Repositories) {
		t.Fatalf("Unexpected repo data: %#v", file.Repositories)
	}

	for i, expect := range expects.Repositories {
		got := file.Repositories[i]
		if expect.Name != got.Name {
			t.Errorf("Expected name %q, got %q", expect.Name, got.Name)
		}
		if expect.URL != got.URL {
			t.Errorf("Expected url %q, got %q", expect.URL, got.URL)
		}
	}
}

func TestRepoNotExists(t *testing.T) {
	if _, err := LoadFile("/this/path/does/not/exist.yaml"); err == nil {
		t.Errorf("expected err to be non-nil when path does not exist")
	} else if !strings.Contains(err.Error(), "couldn't load repositories file") {
		t.Errorf("expected prompt `couldn't load repositories file`")
	}
}
