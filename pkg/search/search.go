/*
Copyright The Helm Authors, Suse LLC.

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

/* Package search implements the search for charts in repos but extracts it into a package
so it can be reused and composed over
Currently the helm search is implemented under the cmd dir which means that most
of its options cant be used in a composite struct to build on top of them
*/
package search

import (
	"fmt"
	"github.com/Masterminds/log-go"
	logio "github.com/Masterminds/log-go/io"
	"github.com/Masterminds/semver/v3"
	"github.com/gosuri/uitable"
	"github.com/pkg/errors"
	"github.com/rancher-sandbox/hypper/pkg/hypperpath"
	"github.com/rancher-sandbox/hypper/pkg/repo"
	"helm.sh/helm/v3/pkg/cli/output"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// searchMaxScore suggests that any score higher than this is not considered a match.
const searchMaxScore = 25

func isNotExist(err error) bool {
	return os.IsNotExist(errors.Cause(err))
}

// RepoOptions is the struct used to search, and stores the different options to filter and configure the output
type RepoOptions struct {
	Versions     bool
	Regexp       bool
	Devel        bool
	Version      string
	MaxColWidth  uint
	RepoFile     string
	RepoCacheDir string
	OutputFormat output.Format
}

// Run searchs and prints the found charts based of the filters
func (o *RepoOptions) Run(logger log.Logger, args []string) error {
	wInfo := logio.NewWriter(logger, log.InfoLevel)
	o.setupSearchedVersion()

	index, err := o.buildIndex()
	if err != nil {
		return err
	}

	var res []*Result
	if len(args) == 0 {
		res = index.All()
	} else {
		q := strings.Join(args, " ")
		res, err = index.Search(q, searchMaxScore, o.Regexp)
		if err != nil {
			return err
		}
	}

	SortScore(res)
	data, err := o.applyConstraint(res)
	if err != nil {
		return err
	}

	return o.OutputFormat.Write(wInfo, &repoSearchWriter{data, o.MaxColWidth})
}

// setupSearchedVersion sets the version to search the chart
// if no version is set it sets it to any semver version higher than 0.0.0
func (o *RepoOptions) setupSearchedVersion() {
	log.Debug("Original chart version: %q", o.Version)

	if o.Version != "" {
		return
	}

	if o.Devel { // search for releases and prereleases (alpha, beta, and release candidate releases).
		log.Debug("setting version to >0.0.0-0")
		o.Version = ">0.0.0-0"
	} else { // search only for stable releases, prerelease versions will be skip
		log.Debug("setting version to >0.0.0")
		o.Version = ">0.0.0"
	}
}

// applyConstraint get a result list and filters it based on the version constraint set
func (o *RepoOptions) applyConstraint(res []*Result) ([]*Result, error) {
	if o.Version == "" {
		return res, nil
	}

	constraint, err := semver.NewConstraint(o.Version)
	if err != nil {
		return res, errors.Wrap(err, "an invalid version/constraint format")
	}

	data := res[:0]
	foundNames := map[string]bool{}
	for _, r := range res {
		// if not returning all versions and already have found a result,
		// you're done!
		if !o.Versions && foundNames[r.Name] {
			continue
		}
		v, err := semver.NewVersion(r.Chart.Version)
		if err != nil {
			continue
		}
		if constraint.Check(v) {
			data = append(data, r)
			foundNames[r.Name] = true
		}
	}

	return data, nil
}

// buildIndex loads the repos to add them to the index search
func (o *RepoOptions) buildIndex() (*Index, error) {
	// Load the repositories.yaml
	rf, err := repo.LoadFile(o.RepoFile)
	if isNotExist(err) || len(rf.Repositories) == 0 {
		return nil, errors.New("no repositories configured")
	}

	i := NewIndex()
	for _, re := range rf.Repositories {
		n := re.Name
		f := filepath.Join(o.RepoCacheDir, hypperpath.CacheIndexFile(n))
		ind, err := repo.LoadIndexFile(f)
		if err != nil {
			log.Warn("Repo %q is corrupt or missing. Try 'hypper repo update'.", n)
			log.Warn("%s", err)
			continue
		}

		i.AddRepo(n, ind, o.Versions || len(o.Version) > 0)
	}
	return i, nil
}

// repoChartElement is used to store the final chart values that will get printed
type repoChartElement struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	AppVersion  string `json:"app_version"`
	Description string `json:"description"`
}

// repoSearchWriter is used to store and print the search results
type repoSearchWriter struct {
	results     []*Result
	columnWidth uint
}

// WriteTable writes the results as a table
func (r *repoSearchWriter) WriteTable(out io.Writer) error {
	if len(r.results) == 0 {
		_, err := out.Write([]byte("No results found\n"))
		if err != nil {
			return fmt.Errorf("unable to write results: %s", err)
		}
		return nil
	}
	table := uitable.New()
	table.MaxColWidth = r.columnWidth
	table.AddRow("NAME", "CHART VERSION", "APP VERSION", "DESCRIPTION")
	for _, r := range r.results {
		table.AddRow(r.Name, r.Chart.Version, r.Chart.AppVersion, r.Chart.Description)
	}
	return output.EncodeTable(out, table)
}

// WriteJSON prints the results as a json
func (r *repoSearchWriter) WriteJSON(out io.Writer) error {
	return r.encodeByFormat(out, output.JSON)
}

// WriteYAML prints the results as a yaml
func (r *repoSearchWriter) WriteYAML(out io.Writer) error {
	return r.encodeByFormat(out, output.YAML)
}

// encodeByFormat creates the final Chartlist that will get formatted into the final results
func (r *repoSearchWriter) encodeByFormat(out io.Writer, format output.Format) error {
	// Initialize the array so no results returns an empty array instead of null
	chartList := make([]repoChartElement, 0, len(r.results))

	for _, r := range r.results {
		chartList = append(chartList, repoChartElement{r.Name, r.Chart.Version, r.Chart.AppVersion, r.Chart.Description})
	}

	switch format {
	case output.JSON:
		return output.EncodeJSON(out, chartList)
	case output.YAML:
		return output.EncodeYAML(out, chartList)
	}

	// Because this is a non-exported function and only called internally by
	// WriteJSON and WriteYAML, we shouldn't get invalid types
	return nil
}
