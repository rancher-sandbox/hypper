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

package action

import (
	"github.com/Masterminds/log-go"
	pkg "github.com/rancher-sandbox/hypper/internal/package"
	solver "github.com/rancher-sandbox/hypper/internal/solver"
	"github.com/rancher-sandbox/hypper/pkg/cli"
	"github.com/rancher-sandbox/hypper/pkg/repo"

	"github.com/rancher-sandbox/hypper/pkg/chart"
	helmAction "helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	helmLoader "helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/release"
)

// FIXME assume all charts come from just 1 repo. We will generalize later.
// TODO add chart.Hash into index file, to not need to pull hypper charts.
// Note that we will still need to pull helm charts to calculate its chart.Hash
func BuildWorld(s *solver.Solver, repoEntriesSlice []*map[string]repo.ChartVersions,
	releases []*release.Release, toModify []*pkg.Pkg,
	settings *cli.EnvSettings, logger *log.Logger) (err error) {
	// add repos to db
	// for all repos:
	for _, repoEntries := range repoEntriesSlice {
		// for all chart entries in the repo:
		for _, chartVersions := range *repoEntries {
			// for all versions of a single chart:
			for _, chartVersion := range chartVersions.ChartVersions {

				// obtain the chart (needed for pkg.ChartHash)
				chart, err := loader.Load(chartVersion.URLs[0])
				if err != nil {
					return err
				}

				// add chart to db
				ns := GetNamespace(chart, "") //TODO figure out the default ns for bare helm charts, and honour kubectl ns and flag
				p, err := pkg.NewPkgFromChart(chart, chart.Name(), ns, pkg.Unknown)
				if err != nil {
					return err
				}
				s.PkgDB.Add(p)
			}
		}
	}

	// add releases to db
	for _, r := range releases {
		p, err := pkg.NewPkgFromRelease(r)
		if err != nil {
			return err
		}
		s.PkgDB.Add(p)
	}

	// add toModify to db
	for _, p := range toModify {
		s.PkgDB.Add(p)
	}

	// create dependency relations in all packages:
	for i := 1; i <= s.PkgDB.Size(); i++ { // IDs start with 1
		p := s.PkgDB.GetPackageByPbID(i)
		createDependencyRelations(p, settings, logger)
	}

	return nil
}

func createDependencyRelations(p *pkg.Pkg, settings *cli.EnvSettings, logger *log.Logger) error {

	// don't check error, dependencies come from repo, they are correctly formed
	sharedDeps, _ := chart.GetSharedDeps(p.Chart, logger)

	for _, dep := range sharedDeps {
		// from chart -> obtain list of deps -> obtain default
		// ns,version,release, and build relation.

		// pull chart:
		chartPathOptions := helmAction.ChartPathOptions{}
		chartPathOptions.RepoURL = dep.Repository
		cp, err := chartPathOptions.LocateChart(dep.Name, settings.EnvSettings)
		if err != nil {
			return err
		}
		depChart, err := helmLoader.Load(cp)
		if err != nil {
			return err
		}

		// Obtain fingerprint and semver for relation:
		depNS := GetNamespace(p.Chart, "") //TODO figure out the default ns for bare helm charts, and honour kubectl ns and flag
		baseFP := pkg.CreateBaseFingerPrint(depChart.Name(), depNS)

		// build relation:
		if dep.IsOptional {
			p.DependsOptionalRel = append(p.DependsOptionalRel, &pkg.PkgRel{
				BaseFingerprint: baseFP,
				SemverRange:     dep.Version,
			})
		} else {
			p.DependsRel = append(p.DependsRel, &pkg.PkgRel{
				BaseFingerprint: baseFP,
				SemverRange:     dep.Version,
			})
		}
	}
	return nil
}
