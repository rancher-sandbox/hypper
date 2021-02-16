package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rancher-sandbox/hypper/internal/helm/pkg/getter"
	"github.com/rancher-sandbox/hypper/internal/helm/pkg/repo"
	"github.com/rancher-sandbox/hypper/internal/helm/pkg/repo/repotest"
	"github.com/rancher-sandbox/hypper/internal/test/ensure"
)

func TestUpdateCmd(t *testing.T) {
	var out bytes.Buffer
	// Instead of using the HTTP updater, we provide our own for this test.
	// The TestUpdateCharts test verifies the HTTP behavior independently.
	updater := func(repos []*repo.ChartRepository, out io.Writer) {
		for _, re := range repos {
			fmt.Fprintln(out, re.Config.Name)
		}
	}
	o := &repoUpdateOptions{
		update:   updater,
		repoFile: "testdata/repositories.yaml",
	}
	if err := o.run(&out); err != nil {
		t.Fatal(err)
	}

	if got := out.String(); !strings.Contains(got, "charts") {
		t.Errorf("Expected 'charts' got %q", got)
	}
}

func TestUpdateCustomCacheCmd(t *testing.T) {
	rootDir := ensure.TempDir(t)
	cachePath := filepath.Join(rootDir, "updcustomcache")
	os.Mkdir(cachePath, os.ModePerm)
	defer os.RemoveAll(cachePath)

	ts, err := repotest.NewTempServerWithCleanup(t, "testdata/testserver/*.*")
	if err != nil {
		t.Fatal(err)
	}
	defer ts.Stop()

	o := &repoUpdateOptions{
		update:    updateCharts,
		repoFile:  filepath.Join(ts.Root(), "repositories.yaml"),
		repoCache: cachePath,
	}
	b := ioutil.Discard
	if err := o.run(b); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(cachePath, "test-index.yaml")); err != nil {
		t.Fatalf("error finding created index file in custom cache: %v", err)
	}
}

func TestUpdateCharts(t *testing.T) {
	defer resetEnv()()
	defer ensure.HelmHome(t)()

	ts, err := repotest.NewTempServerWithCleanup(t, "testdata/testserver/*.*")
	if err != nil {
		t.Fatal(err)
	}
	defer ts.Stop()

	r, err := repo.NewChartRepository(&repo.Entry{
		Name: "charts",
		URL:  ts.URL(),
	}, getter.All(settings))
	if err != nil {
		t.Error(err)
	}

	b := bytes.NewBuffer(nil)
	updateCharts([]*repo.ChartRepository{r}, b)

	got := b.String()
	if strings.Contains(got, "Unable to get an update") {
		t.Errorf("Failed to get a repo: %q", got)
	}
	if !strings.Contains(got, "Update Complete.") {
		t.Error("Update was not successful")
	}
}
