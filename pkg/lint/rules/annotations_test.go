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

package rules

import (
	"helm.sh/helm/v3/pkg/chart"
	"testing"
)

// Validation functions Test
func TestValidateChartHypperNamespace(t *testing.T) {
	annotations := map[string]string{
		"hypper.cattle.io/namespace": "namespaceTest",
	}
	chartMetadataGood := &chart.Metadata{Annotations: annotations}

	err := validateChartHypperNamespace(chartMetadataGood)
	if err != nil {
		t.Errorf("validateChartHypperNamespace to not return a linter error, got error")
	}

	annotationsBad := map[string]string{}

	chartMetadataBad := &chart.Metadata{Annotations: annotationsBad}

	err = validateChartHypperNamespace(chartMetadataBad)
	if err == nil {
		t.Errorf("validateChartHypperNamespace to return a linter warning, got no warning")
	}

}

func TestValidateChartHypperRelease(t *testing.T) {
	annotations := map[string]string{
		"hypper.cattle.io/release-name": "releaseTest",
	}
	chartMetadataGood := &chart.Metadata{Annotations: annotations}

	err := validateChartHypperRelease(chartMetadataGood)
	if err != nil {
		t.Errorf("validateChartHypperRelease to not return a linter error, got error")
	}

	annotationsBad := map[string]string{}

	chartMetadataBad := &chart.Metadata{Annotations: annotationsBad}

	err = validateChartHypperRelease(chartMetadataBad)
	if err == nil {
		t.Errorf("validateChartHypperRelease to return a linter warning, got no warning")
	}

}
