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
	"path/filepath"

	"github.com/Masterminds/log-go"
	"gopkg.in/yaml.v2"

	pkg "github.com/rancher-sandbox/hypper/internal/package"
	solver "github.com/rancher-sandbox/hypper/internal/solver"
	"github.com/rancher-sandbox/hypper/pkg/cli"
	"github.com/rancher-sandbox/hypper/pkg/repo"

	helmChart "helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/helmpath"
	"helm.sh/helm/v3/pkg/release"
	helmRepo "helm.sh/helm/v3/pkg/repo"
)

// chrtEntry is a helper to iterate through all versions of a chart in the repo
// index.yaml
type chrtEntry struct {
	chartVersions []*helmRepo.ChartVersion
	url           string
}

// BuildWorld adds all known charts to the package database:
//
// - For all the repos, it iterates through the chart entries and adds a package
//   to the DB for each version of the chart.
// - For all releases and wanted packages, it adds a package or updates a
//   present package in the DB.
func BuildWorld(pkgdb *solver.PkgDB, repositories []*helmRepo.Entry,
	releases []*release.Release,
	toModify *pkg.Pkg, toModifyChart *helmChart.Chart,
	settings *cli.EnvSettings, logger log.Logger) (err error) {

	// concatenate all index entries from all repositories:
	repoEntries := make(map[string]chrtEntry)
	for _, r := range repositories {
		idxFilepath := filepath.Join(settings.RepositoryCache, helmpath.CacheIndexFile(r.Name))
		// obtain repo index file from cache:
		index, err := repo.LoadIndexFile(idxFilepath)
		if err != nil {
			return err
		}
		for chrtName, chrtVers := range index.Entries {
			// TODO this overrides other repos, if charts are in both
			repoEntries[chrtName] = chrtEntry{
				chartVersions: chrtVers,
				url:           r.URL,
			}
		}
	}

	// save ns from kube client, for performance reasons
	settingsNS := settings.Namespace()

	// add repos to db
	// for all chart entries in repos:
	for chrtName, chrtVersions := range repoEntries {
		// for all the versions of a chart:
		for _, chrtVer := range chrtVersions.chartVersions {

			// create pkg:
			ns := GetNamespaceFromAnnot(chrtVer.Annotations, settingsNS)            //TODO figure out the default ns for bare helm charts, and honour kubectl ns and flag
			relName := GetNameFromAnnot(chrtVer.Annotations, chrtVer.Metadata.Name) // TODO default name for helm repos
			repo := chrtVersions.url
			p := pkg.NewPkg(relName, chrtName, chrtVer.Version, ns, pkg.Unknown, pkg.Unknown, pkg.Unknown, repo)

			// fill dep relations
			if err := CreateDepRelsFromAnnot(p, chrtVer.Annotations, repoEntries); err != nil {
				return err
			}

			// add chart to db:
			pkgdb.Add(p)
		}
	}

	// add releases to db
	for _, r := range releases {
		fp := pkg.CreateFingerPrint(r.Name, r.Chart.Metadata.Version, r.Namespace)
		p := pkgdb.GetPackageByFingerprint(fp)
		if p != nil {
			// release is in repos, hence it was added to db. Modify directly:
			p.CurrentState = pkg.Present
		} else {
			// release is not in repos
			// we don't know the repo where the release has originally been
			// installed from, so we add it as stale package with an empty repo
			// string
			p := pkg.NewPkg(r.Name, r.Chart.Name(), r.Chart.Metadata.Version, r.Namespace, pkg.Present, pkg.Unknown, pkg.Present, "")
			// fill dep relations:
			// FIXME releases depending on charts not on the repos are not done
			// yet, need to pull the chart to obtain default ns
			if err := CreateDepRelsFromAnnot(p, r.Chart.Metadata.Annotations, repoEntries); err != nil {
				return err
			}
			pkgdb.Add(p)
		}
	}

	// calculate dep rels for toModify
	// fill dep relations
	if err := CreateDepRelsFromAnnot(toModify, toModifyChart.Metadata.Annotations, repoEntries); err != nil {
		return err
	}

	// add toModify to db
	pkgdb.Add(toModify)

	return nil
}

// CreateDepRelsFromAnnot fills the p.DepRel and p.DepOptionalRel of a package,
// by unmarshalling and checking the Metadata.Annotations of the chart that
// corresponds to that package.
func CreateDepRelsFromAnnot(p *pkg.Pkg, chartAnnot map[string]string, repoEntries map[string]chrtEntry) (err error) {

	// unmarshal dependencies:
	cases := []string{"hypper.cattle.io/shared-dependencies", "hypper.cattle.io/optional-dependencies"}

	for _, c := range cases {
		sharedDepsYaml := chartAnnot[c]
		var sharedDeps []*helmChart.Dependency
		// unmarshalling Helm's Dependency because gopkg.in/yaml.v2 doesn't do composite types
		if err := yaml.UnmarshalStrict([]byte(sharedDepsYaml), &sharedDeps); err != nil {
			log.Errorf("Chart.yaml metadata is malformed for repo entry \"%s\", \"%s\"\n", p.ChartName, p.Version)
			return err
		}

		// for all deps in chrtVer.Annotations, find the dep in repo entries
		// to obtain its default ns:
		for _, dep := range sharedDeps {
			// find dependency:
			depChrtVer, ok := repoEntries[dep.Name]
			if !ok {
				log.Warnf("Dependency \"%s\" not found in repos, skipping it", dep.Name)
				continue
			}

			// TODO each version can have a different default ns

			// obtain default ns of dep
			depNS := GetNamespaceFromAnnot(depChrtVer.chartVersions[0].Annotations, "") //TODO figure out the default ns for bare helm charts, and honour kubectl ns and flag
			depName := GetNameFromAnnot(depChrtVer.chartVersions[0].Annotations, "")    // TODO default name for helm repos

			// FIXME if dependency chart not in repo, we are defaulting to
			// default release name and namespace

			switch c {
			case "hypper.cattle.io/shared-dependencies":
				//add relation to pkg
				p.DependsRel = append(p.DependsRel, &pkg.PkgRel{
					BaseFingerprint: pkg.CreateBaseFingerPrint(depName, depNS),
					SemverRange:     dep.Version,
				})
			case "hypper.cattle.io/optional-dependencies":
				p.DependsOptionalRel = append(p.DependsOptionalRel, &pkg.PkgRel{
					BaseFingerprint: pkg.CreateBaseFingerPrint(depName, depNS),
					SemverRange:     dep.Version,
				})
			}
		}
	}
	return nil
}
