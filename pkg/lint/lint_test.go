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

package lint

import (
	"strings"
	"testing"

	"helm.sh/helm/v3/pkg/lint/support"
)

const badChartDir = "rules/testdata/badchart"
const badChartDirWithBrokenHypperDeps = "rules/testdata/badchartbrokenhypperdeps"
const goodChartDir = "rules/testdata/goodchart"

func TestBadChart(t *testing.T) {
	m := All(badChartDir).Messages
	if len(m) != 3 {
		t.Errorf("Number of errors %v", len(m))
		t.Errorf("All didn't fail with expected errors, got %#v", m)
	}
	// There should be 3 WARNINGs, check for them
	var w1, w2, w3 bool
	for _, msg := range m {
		if msg.Severity == support.WarningSev {
			if strings.Contains(msg.Err.Error(), "Setting hypper.cattle.io/release-name in annotations is recommended") {
				w1 = true
			}
		}
		if msg.Severity == support.WarningSev {
			if strings.Contains(msg.Err.Error(), "Setting hypper.cattle.io/namespace in annotations is recommended") {
				w2 = true
			}
		}
		if msg.Severity == support.WarningSev {
			if strings.Contains(msg.Err.Error(), "Setting hypper.cattle.io/shared-dependencies in annotations is recommended") {
				w3 = true
			}
		}
	}
	if !w1 || !w2 || !w3 {
		t.Errorf("Didn't find all the expected errors, got %#v", m)
	}
}

func TestBadChartBrokenDeps(t *testing.T) {
	m := All(badChartDirWithBrokenHypperDeps).Messages
	if len(m) != 1 {
		t.Errorf("Number of errors %v", len(m))
		t.Errorf("All didn't fail with expected errors, got %#v", m)
	}
	// There should be 1 Error, check for it
	var e1 bool
	for _, msg := range m {
		if msg.Severity == support.ErrorSev {
			if strings.Contains(msg.Err.Error(), "Shared dependencies list is broken, please check the correct format") {
				e1 = true
			}
		}
	}
	if !e1 {
		t.Errorf("Didn't find all the expected errors, got %#v", m)
	}
}

func TestGoodChart(t *testing.T) {
	m := All(goodChartDir).Messages
	if len(m) != 0 {
		t.Error("All returned linter messages when it shouldn't have")
		for i, msg := range m {
			t.Logf("Message %d: %s", i, msg)
		}
	}
}
