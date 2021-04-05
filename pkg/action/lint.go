/*
Copyright The Helm Authors, Suse LLC

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

package action

import (
	"github.com/rancher-sandbox/hypper/pkg/lint"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/lint/support"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// Lint is a composite type of Helm's Lint type
type Lint struct {
	*action.Lint
}

// NewLint creates a new Lint object with the given configuration.
func NewLint() *Lint {
	return &Lint{
		action.NewLint(),
	}
}

// Run executes 'helm Lint' against the given chart and then runs the chart against hypper lint rules
func (l *Lint) Run(paths []string, vals map[string]interface{}) *action.LintResult {
	lowestTolerance := support.ErrorSev
	if l.Strict {
		lowestTolerance = support.WarningSev
	}
	// run it against the helm linter rules
	result := l.Lint.Run(paths, vals)
	// run it against our linter rules
	for _, path := range paths {
		linter, err := lintChart(path)
		if err != nil {
			result.Errors = append(result.Errors, err)
			continue
		}

		result.Messages = append(result.Messages, linter.Messages...)
		result.TotalChartsLinted++
		for _, msg := range linter.Messages {
			if msg.Severity >= lowestTolerance {
				result.Errors = append(result.Errors, msg.Err)
			}
		}
	}
	return result
}

// lintChart extracts the chart to a temp dir, checks that is a valid chart (has a Chart.yaml) and runs our rules
// We dont add to errors if we cant open the file because the helm linter has already added those as it runs before
// us so no need to flag those errors or else we get them duplicated.
func lintChart(path string) (support.Linter, error) {
	var chartPath string
	linter := support.Linter{}

	if strings.HasSuffix(path, ".tgz") || strings.HasSuffix(path, ".tar.gz") {
		tempDir, err := ioutil.TempDir("", "hypper-lint")
		if err != nil {
			return linter, nil
		}
		defer os.RemoveAll(tempDir)

		file, err := os.Open(path)
		if err != nil {
			return linter, nil
		}
		defer file.Close()

		if err = chartutil.Expand(tempDir, file); err != nil {
			return linter, nil
		}

		files, err := ioutil.ReadDir(tempDir)
		if err != nil {
			return linter, nil
		}
		if !files[0].IsDir() {
			return linter, nil
		}

		chartPath = filepath.Join(tempDir, files[0].Name())
	} else {
		chartPath = path
	}

	// Guard: Error out if this is not a chart.
	if _, err := os.Stat(filepath.Join(chartPath, "Chart.yaml")); err != nil {
		return linter, nil
	}

	return lint.All(chartPath), nil
}
