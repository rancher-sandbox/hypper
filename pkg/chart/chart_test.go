/*
Copyright SUSE LLC

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

package chart

import (
	"bytes"
	"testing"

	"github.com/Masterminds/log-go"
	logcli "github.com/Masterminds/log-go/impl/cli"
	"github.com/stretchr/testify/assert"
	helmChart "helm.sh/helm/v3/pkg/chart"
)

func TestGetSharedDeps(t *testing.T) {

	chrtWithSharedDeps := Mock(&MockChartOptions{
		Name:    "chartname",
		Version: "0.1.0",
	})
	chrtWithSharedDeps.Metadata.Annotations = make(map[string]string)
	chrtWithSharedDeps.Metadata.Annotations["hypper.cattle.io/shared-dependencies"] = "  - name: \"testdata/charts/shared-dep\"" + "\n" +
		"    version: \"0.1.0\"" + "\n" +
		"    repository: \"\"" + "\n"

	chrtWithSharedAndOptionalDeps := Mock(&MockChartOptions{
		Name:    "chartname",
		Version: "0.1.0",
	})
	chrtWithSharedAndOptionalDeps.Metadata.Annotations = make(map[string]string)
	chrtWithSharedAndOptionalDeps.Metadata.Annotations["hypper.cattle.io/shared-dependencies"] = "  - name: \"testdata/charts/shared-dep\"" + "\n" +
		"    version: \"0.1.0\"" + "\n" +
		"    repository: \"\"" + "\n"
	chrtWithSharedAndOptionalDeps.Metadata.Annotations["hypper.cattle.io/optional-dependencies"] = "  - name: \"testdata/charts/vanilla-helm\"" + "\n" +
		"    version: \"0.1.0\"" + "\n" +
		"    repository: \"\"" + "\n"

	for _, tcase := range []struct {
		name       string
		chart      *helmChart.Chart
		golden     string
		wantError  bool
		err        string
		wantDebug  bool
		debug      string
		depsNumber int
	}{
		{
			name: "chart has no shared-deps",
			chart: Mock(&MockChartOptions{
				Name:    "chartname",
				Version: "0.1.0",
			}),
			depsNumber: 0,
		},
		{
			name:       "chart has only shared-deps",
			chart:      chrtWithSharedDeps,
			depsNumber: 1,
		},
		{
			name:       "chart has shared and optional deps",
			chart:      chrtWithSharedAndOptionalDeps,
			depsNumber: 2,
		},
	} {
		is := assert.New(t)

		// create our own Logger that satisfies impl/cli.Logger, but with a buffer for tests
		buf := new(bytes.Buffer)
		logger := logcli.NewStandard()
		logger.InfoOut = buf
		logger.WarnOut = buf
		logger.ErrorOut = buf
		logger.DebugOut = buf
		if tcase.wantDebug {
			logger.Level = log.DebugLevel
		}
		log.Current = logger

		deps, err := GetSharedDeps(tcase.chart, logger)

		if (err != nil) != tcase.wantError {
			t.Errorf("expected error, got '%v'", err)
		}
		if tcase.wantError {
			is.Equal(tcase.err, err.Error())
		}

		is.Equal(tcase.depsNumber, len(deps))
	}
}
