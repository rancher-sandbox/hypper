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

package repo

import (
	"strings"
	"testing"

	"helm.sh/helm/v3/pkg/chart"
	helmRepo "helm.sh/helm/v3/pkg/repo"
)

const (
	testfile            = "testdata/local-index.yaml"
	annotationstestfile = "testdata/local-index-annotations.yaml"
	chartmuseumtestfile = "testdata/chartmuseum-index.yaml"
	unorderedTestfile   = "testdata/local-index-unordered.yaml"
	indexWithDuplicates = `
 apiVersion: v1
 entries:
   nginx:
     - urls:
         - https://charts.helm.sh/stable/nginx-0.2.0.tgz
       name: nginx
       description: string
       version: 0.2.0
       home: https://github.com/something/else
       digest: "sha256:1234567890abcdef"
   nginx:
     - urls:
         - https://charts.helm.sh/stable/alpine-1.0.0.tgz
         - http://storage2.googleapis.com/kubernetes-charts/alpine-1.0.0.tgz
       name: alpine
       description: string
       version: 1.0.0
       home: https://github.com/something
       digest: "sha256:1234567890abcdef"
 `
)

func TestIndexFile(t *testing.T) {
	i := NewIndexFile()
	for _, x := range []struct {
		md       *chart.Metadata
		filename string
		baseURL  string
		digest   string
	}{
		{&chart.Metadata{APIVersion: "v2", Name: "clipper", Version: "0.1.0"}, "clipper-0.1.0.tgz", "http://example.com/charts", "sha256:1234567890"},
		{&chart.Metadata{APIVersion: "v2", Name: "cutter", Version: "0.1.1"}, "cutter-0.1.1.tgz", "http://example.com/charts", "sha256:1234567890abc"},
		{&chart.Metadata{APIVersion: "v2", Name: "cutter", Version: "0.1.0"}, "cutter-0.1.0.tgz", "http://example.com/charts", "sha256:1234567890abc"},
		{&chart.Metadata{APIVersion: "v2", Name: "cutter", Version: "0.2.0"}, "cutter-0.2.0.tgz", "http://example.com/charts", "sha256:1234567890abc"},
		{&chart.Metadata{APIVersion: "v2", Name: "setter", Version: "0.1.9+alpha"}, "setter-0.1.9+alpha.tgz", "http://example.com/charts", "sha256:1234567890abc"},
		{&chart.Metadata{APIVersion: "v2", Name: "setter", Version: "0.1.9+beta"}, "setter-0.1.9+beta.tgz", "http://example.com/charts", "sha256:1234567890abc"},
	} {
		if err := i.MustAdd(x.md, x.filename, x.baseURL, x.digest); err != nil {
			t.Errorf("unexpected error adding to index: %s", err)
		}
	}
	i.SortEntries()
	if i.APIVersion != APIVersionV1 {
		t.Error("Expected API version v1")
	}
	if len(i.Entries) != 3 {
		t.Errorf("Expected 3 charts. Got %d", len(i.Entries))
	}
	if i.Entries["clipper"][0].Name != "clipper" {
		t.Errorf("Expected clipper, got %s", i.Entries["clipper"][0].Name)
	}
	if len(i.Entries["cutter"]) != 3 {
		t.Error("Expected three cutters.")
	}
	// Test that the sort worked. 0.2 should be at the first index for Cutter.
	if v := i.Entries["cutter"][0].Version; v != "0.2.0" {
		t.Errorf("Unexpected first version: %s", v)
	}
	cv, err := i.Get("setter", "0.1.9")
	if err == nil && !strings.Contains(cv.Metadata.Version, "0.1.9") {
		t.Errorf("Unexpected version: %s", cv.Metadata.Version)
	}
	cv, err = i.Get("setter", "0.1.9+alpha")
	if err != nil || cv.Metadata.Version != "0.1.9+alpha" {
		t.Errorf("Expected version: 0.1.9+alpha")
	}
}

func TestLoadIndex(t *testing.T) {
	tests := []struct {
		Name     string
		Filename string
	}{
		{
			Name:     "regular index file",
			Filename: testfile,
		},
		{
			Name:     "chartmuseum index file",
			Filename: chartmuseumtestfile,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			i, err := LoadIndexFile(tc.Filename)
			if err != nil {
				t.Fatal(err)
			}
			verifyLocalIndex(t, i)
		})
	}
}

// TestLoadIndex_Duplicates is a regression to make sure that we don't non-deterministically allow duplicate packages.
func TestLoadIndex_Duplicates(t *testing.T) {
	if _, err := loadIndex([]byte(indexWithDuplicates), "indexWithDuplicates"); err == nil {
		t.Errorf("Expected an error when duplicate entries are present")
	}
}
func TestLoadIndexFileAnnotations(t *testing.T) {
	i, err := LoadIndexFile(annotationstestfile)
	if err != nil {
		t.Fatal(err)
	}
	verifyLocalIndex(t, i)
	if len(i.Annotations) != 1 {
		t.Fatalf("Expected 1 annotation but got %d", len(i.Annotations))
	}
	if i.Annotations["helm.sh/test"] != "foo bar" {
		t.Error("Did not get expected value for helm.sh/test annotation")
	}
}
func TestLoadUnorderedIndex(t *testing.T) {
	i, err := LoadIndexFile(unorderedTestfile)
	if err != nil {
		t.Fatal(err)
	}
	verifyLocalIndex(t, i)
}

func verifyLocalIndex(t *testing.T, i *IndexFile) {
	numEntries := len(i.Entries)
	if numEntries != 3 {
		t.Errorf("Expected 3 entries in index file but got %d", numEntries)
	}
	alpine, ok := i.Entries["alpine"]
	if !ok {
		t.Fatalf("'alpine' section not found.")
	}
	if l := len(alpine); l != 1 {
		t.Fatalf("'alpine' should have 1 chart, got %d", l)
	}
	nginx, ok := i.Entries["nginx"]
	if !ok || len(nginx) != 2 {
		t.Fatalf("Expected 2 nginx entries")
	}
	expects := []*helmRepo.ChartVersion{
		{
			Metadata: &chart.Metadata{
				APIVersion:  "v2",
				Name:        "alpine",
				Description: "string",
				Version:     "1.0.0",
				Keywords:    []string{"linux", "alpine", "small", "sumtin"},
				Home:        "https://github.com/something",
			},
			URLs: []string{
				"https://charts.helm.sh/stable/alpine-1.0.0.tgz",
				"http://storage2.googleapis.com/kubernetes-charts/alpine-1.0.0.tgz",
			},
			Digest: "sha256:1234567890abcdef",
		},
		{
			Metadata: &chart.Metadata{
				APIVersion:  "v2",
				Name:        "nginx",
				Description: "string",
				Version:     "0.2.0",
				Keywords:    []string{"popular", "web server", "proxy"},
				Home:        "https://github.com/something/else",
			},
			URLs: []string{
				"https://charts.helm.sh/stable/nginx-0.2.0.tgz",
			},
			Digest: "sha256:1234567890abcdef",
		},
		{
			Metadata: &chart.Metadata{
				APIVersion:  "v2",
				Name:        "nginx",
				Description: "string",
				Version:     "0.1.0",
				Keywords:    []string{"popular", "web server", "proxy"},
				Home:        "https://github.com/something",
			},
			URLs: []string{
				"https://charts.helm.sh/stable/nginx-0.1.0.tgz",
			},
			Digest: "sha256:1234567890abcdef",
		},
	}
	tests := []*helmRepo.ChartVersion{alpine[0], nginx[0], nginx[1]}
	for i, tt := range tests {
		expect := expects[i]
		if tt.Name != expect.Name {
			t.Errorf("Expected name %q, got %q", expect.Name, tt.Name)
		}
		if tt.Description != expect.Description {
			t.Errorf("Expected description %q, got %q", expect.Description, tt.Description)
		}
		if tt.Version != expect.Version {
			t.Errorf("Expected version %q, got %q", expect.Version, tt.Version)
		}
		if tt.Digest != expect.Digest {
			t.Errorf("Expected digest %q, got %q", expect.Digest, tt.Digest)
		}
		if tt.Home != expect.Home {
			t.Errorf("Expected home %q, got %q", expect.Home, tt.Home)
		}
		for i, url := range tt.URLs {
			if url != expect.URLs[i] {
				t.Errorf("Expected URL %q, got %q", expect.URLs[i], url)
			}
		}
		for i, kw := range tt.Keywords {
			if kw != expect.Keywords[i] {
				t.Errorf("Expected keywords %q, got %q", expect.Keywords[i], kw)
			}
		}
	}
}
