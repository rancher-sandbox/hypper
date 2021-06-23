/*
Copyright SUSE LLC.

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

package pkg

import (
	"github.com/rancher-sandbox/hypper/pkg/action"
	helmChart "helm.sh/helm/v3/pkg/chart"
	helmRelease "helm.sh/helm/v3/pkg/release"
)

// NewPkgFromChart creates a Pkg struct from chart, without filling the
// dependencies and optional dependencies information yet. This needs to happen
// when all packages are already on the db with IDs.
func NewPkgFromChart(chart *helmChart.Chart,
	dependsRel, dependsOptionalRel []string, desiredState tristate) *Pkg {

	ns := action.GetNamespace(chart, "") //TODO figure out the default ns for bare helm charts, and honour kubectl ns and flag

	return NewPkg(chart.Name(), chart.Metadata.Version,
		ns, dependsRel, dependsOptionalRel,
		Unknown, desiredState,
		chart)
}

// NewPkgFromRelease creates a Pkg struct from release, without filling the
// dependencies and optional dependencies information yet. This needs to happen
// when all packages are already on the db with IDs.
func NewPkgFromRelease(release *helmRelease.Release) *Pkg {

	return NewPkg(release.Name, release.Chart.Metadata.Version,
		release.Namespace, nil, nil,
		Present, Unknown,
		release.Chart)
}
