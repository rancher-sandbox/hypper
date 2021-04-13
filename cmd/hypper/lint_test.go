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
	"fmt"
	"runtime"
	"testing"
)

func TestLintCmdWithSubchartsFlag(t *testing.T) {
	testChart := "testdata/testcharts/chart-with-bad-subcharts"
	testLintGoodChartBadSubchartWithFlag := cmdTestCase{
		name:      "lint good chart with bad subcharts using --with-subcharts flag",
		cmd:       fmt.Sprintf("lint --with-subcharts %s", testChart),
		golden:    "output/lint-chart-with-bad-subcharts-with-subcharts.txt",
		wantError: true,
	}

	// If your golden file output contains a path, you are gonna have a bad time! in windows it
	// will show as \path\to\whatever so it doesnt match the golden files.
	if runtime.GOOS == "windows" {
		testLintGoodChartBadSubchartWithFlag.golden = "output/lint-chart-with-bad-subcharts-with-subcharts-windows.txt"
	}

	tests := []cmdTestCase{
		{
			name:      "lint good chart with bad subcharts",
			cmd:       fmt.Sprintf("lint %s", testChart),
			golden:    "output/lint-chart-with-bad-subcharts.txt",
			wantError: true,
		},
		testLintGoodChartBadSubchartWithFlag,
	}

	runTestCmd(t, tests)
}
