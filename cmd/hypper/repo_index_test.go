package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/rancher-sandbox/hypper/internal/test/ensure"
	"helm.sh/helm/v3/pkg/repo"
)

func TestRepoIndexCmd(t *testing.T) {

	dir := ensure.TempDir(t)

	comp := filepath.Join(dir, "compressedchart-0.1.0.tgz")
	if err := linkOrCopy("testdata/testcharts/compressedchart-0.1.0.tgz", comp); err != nil {
		t.Fatal(err)
	}
	comp2 := filepath.Join(dir, "compressedchart-0.2.0.tgz")
	if err := linkOrCopy("testdata/testcharts/compressedchart-0.2.0.tgz", comp2); err != nil {
		t.Fatal(err)
	}

	buf := bytes.NewBuffer(nil)
	c := newRepoIndexCmd(buf)

	if err := c.RunE(c, []string{dir}); err != nil {
		t.Error(err)
	}

	destIndex := filepath.Join(dir, "index.yaml")

	index, err := repo.LoadIndexFile(destIndex)
	if err != nil {
		t.Fatal(err)
	}

	if len(index.Entries) != 1 {
		t.Errorf("expected 1 entry, got %d: %#v", len(index.Entries), index.Entries)
	}

	vs := index.Entries["compressedchart"]
	if len(vs) != 2 {
		t.Errorf("expected 2 versions, got %d: %#v", len(vs), vs)
	}

	expectedVersion := "0.2.0"
	if vs[0].Version != expectedVersion {
		t.Errorf("expected %q, got %q", expectedVersion, vs[0].Version)
	}

	// Test with `--merge`

	// Remove first two charts.
	if err := os.Remove(comp); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(comp2); err != nil {
		t.Fatal(err)
	}
	// Add a new chart and a new version of an existing chart
	if err := linkOrCopy("testdata/testcharts/reqtest-0.1.0.tgz", filepath.Join(dir, "reqtest-0.1.0.tgz")); err != nil {
		t.Fatal(err)
	}
	if err := linkOrCopy("testdata/testcharts/compressedchart-0.3.0.tgz", filepath.Join(dir, "compressedchart-0.3.0.tgz")); err != nil {
		t.Fatal(err)
	}

	c.ParseFlags([]string{"--merge", destIndex})
	if err := c.RunE(c, []string{dir}); err != nil {
		t.Error(err)
	}

	index, err = repo.LoadIndexFile(destIndex)
	if err != nil {
		t.Fatal(err)
	}

	if len(index.Entries) != 2 {
		t.Errorf("expected 2 entries, got %d: %#v", len(index.Entries), index.Entries)
	}

	vs = index.Entries["compressedchart"]
	if len(vs) != 3 {
		t.Errorf("expected 3 versions, got %d: %#v", len(vs), vs)
	}

	expectedVersion = "0.3.0"
	if vs[0].Version != expectedVersion {
		t.Errorf("expected %q, got %q", expectedVersion, vs[0].Version)
	}

	// test that index.yaml gets generated on merge even when it doesn't exist
	if err := os.Remove(destIndex); err != nil {
		t.Fatal(err)
	}

	c.ParseFlags([]string{"--merge", destIndex})
	if err := c.RunE(c, []string{dir}); err != nil {
		t.Error(err)
	}

	index, err = repo.LoadIndexFile(destIndex)
	if err != nil {
		t.Fatal(err)
	}

	// verify it didn't create an empty index.yaml and the merged happened
	if len(index.Entries) != 2 {
		t.Errorf("expected 2 entries, got %d: %#v", len(index.Entries), index.Entries)
	}

	vs = index.Entries["compressedchart"]
	if len(vs) != 1 {
		t.Errorf("expected 1 versions, got %d: %#v", len(vs), vs)
	}

	expectedVersion = "0.3.0"
	if vs[0].Version != expectedVersion {
		t.Errorf("expected %q, got %q", expectedVersion, vs[0].Version)
	}
}

func linkOrCopy(old, new string) error {
	if err := os.Link(old, new); err != nil {
		return copyFile(old, new)
	}

	return nil
}

func copyFile(dst, src string) error {
	i, err := os.Open(dst)
	if err != nil {
		return err
	}
	defer i.Close()

	o, err := os.Create(src)
	if err != nil {
		return err
	}
	defer o.Close()

	_, err = io.Copy(o, i)

	return err
}
