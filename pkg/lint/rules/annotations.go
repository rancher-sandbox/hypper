/*
Copyright The Helm Authors, SUSE LLC

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

/*
Package rules contains all the rules that hypper will run against a chart when
hypper lint is run. This rules are run on top of the default helm rules to cover as much
as possible
*/
package rules

import (
	"path/filepath"

	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	helmChart "helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/lint/support"
)

// Annotations runs a set of linter rules related to chart annotations
func Annotations(linter *support.Linter) {
	chartFileName := "Chart.yaml"
	chartPath := filepath.Join(linter.ChartDir, chartFileName)
	chartFile, _ := chartutil.LoadChartfile(chartPath)
	linter.RunLinterRule(support.InfoSev, chartFileName, validateChartHypperRelease(chartFile))
	linter.RunLinterRule(support.InfoSev, chartFileName, validateChartHypperNamespace(chartFile))
	linter.RunLinterRule(support.InfoSev, chartFileName, validateChartHypperSharedDeps(chartFile))
	linter.RunLinterRule(support.InfoSev, chartFileName, validateChartHypperOptionalSharedDeps(chartFile))
	// If we have shared deps annotations then check its correct formatting
	// This is an Error severity check!
	if _, ok := chartFile.Annotations["hypper.cattle.io/shared-dependencies"]; ok {
		linter.RunLinterRule(support.ErrorSev, chartFileName, validateChartHypperSharedDepsCorrect(chartFile))
	}
	if _, ok := chartFile.Annotations["hypper.cattle.io/optional-dependencies"]; ok {
		linter.RunLinterRule(support.ErrorSev, chartFileName, validateChartHypperOptionalSharedDepsCorrect(chartFile))
	}
}

// validateChartHypperRelease checks that hypper release-name annotation is set
func validateChartHypperRelease(chart *helmChart.Metadata) error {
	if _, ok := chart.Annotations["hypper.cattle.io/release-name"]; !ok {
		return errors.New("Setting hypper.cattle.io/release-name in annotations is recommended")
	}
	return nil
}

// validateChartHypperNamespace checks that hypper namespace annotation is set
func validateChartHypperNamespace(chart *helmChart.Metadata) error {
	if _, ok := chart.Annotations["hypper.cattle.io/namespace"]; !ok {
		return errors.New("Setting hypper.cattle.io/namespace in annotations is recommended")
	}
	return nil
}

// validateChartHypperSharedDeps checks that shared-dependencies hypper annotation is set
func validateChartHypperSharedDeps(chart *helmChart.Metadata) error {
	if _, ok := chart.Annotations["hypper.cattle.io/shared-dependencies"]; !ok {
		return errors.New("Setting hypper.cattle.io/shared-dependencies in annotations is optional")
	}
	return nil
}

// validateChartHypperOptionalSharedDeps checks that optiona-dependencies hypper annotation is set
func validateChartHypperOptionalSharedDeps(chart *helmChart.Metadata) error {
	if _, ok := chart.Annotations["hypper.cattle.io/optional-dependencies"]; !ok {
		return errors.New("Setting hypper.cattle.io/optional-dependencies in annotations is optional")
	}
	return nil
}

// validateChartHypperSharedDepsCorrect checks that shared deps are in the correct format
func validateChartHypperSharedDepsCorrect(chart *helmChart.Metadata) error {
	depYaml := chart.Annotations["hypper.cattle.io/shared-dependencies"]
	var deps []*helmChart.Dependency
	if err := yaml.UnmarshalStrict([]byte(depYaml), &deps); err != nil {
		return errors.New("Shared dependencies list is broken, please check the correct format")
	}
	for _, d := range deps {
		if err := validateSharedDepVersion(d.Version); err != nil {
			return err
		}
	}
	return nil
}

// validateChartHypperOptionalSharedDepsCorrect checks that optional shared deps are in the correct format
func validateChartHypperOptionalSharedDepsCorrect(chart *helmChart.Metadata) error {
	depYaml := chart.Annotations["hypper.cattle.io/optional-dependencies"]
	var deps []*helmChart.Dependency
	if err := yaml.UnmarshalStrict([]byte(depYaml), &deps); err != nil {
		return errors.New("Optional shared dependencies list is broken, please check the correct format")
	}
	for _, d := range deps {
		if err := validateSharedDepVersion(d.Version); err != nil {
			return err
		}
	}
	return nil
}

// validateSharedDepVersion checks that the shared dep version is an actual semver range
func validateSharedDepVersion(depVersion string) error {
	if depVersion == "" {
		return errors.New("version is required")
	}

	_, err := semver.NewConstraint(depVersion)
	if err != nil {
		return errors.Wrap(err, "Shared dependency version is broken")
	}

	return nil
}
