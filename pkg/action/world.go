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
	"strings"

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
func (i *Install) BuildWorld(pkgdb *solver.PkgDB, repositories []*helmRepo.Entry,
	releases []*release.Release,
	toModify *pkg.Pkg, toModifyChart *helmChart.Chart,
	settings *cli.EnvSettings, logger log.Logger) (err error) {

	logger.Debug("Building package DB…")

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
			ns := GetNamespaceFromAnnot(chrtVer.Annotations, settingsNS)
			relName := GetNameFromAnnot(chrtVer.Annotations, chrtVer.Metadata.Name)
			repo := chrtVersions.url
			p := pkg.NewPkg(relName, chrtName, chrtVer.Version, ns,
				pkg.Unknown, pkg.Unknown, pkg.Unknown, repo, "")

			// fill dep relations
			if err := i.CreateDepRelsFromAnnot(p, chrtVer.Annotations, repoEntries,
				pkgdb, settings, logger); err != nil {
				return err
			}

			// add chart to db:
			pkgdb.Add(p)
		}
	}

	// add releases to db
	for _, r := range releases {
		fp := pkg.CreateFingerPrint(r.Name, r.Chart.Metadata.Version, r.Namespace, r.Chart.Metadata.Name)
		p := pkgdb.GetPackageByFingerprint(fp)
		if p != nil {
			// release is in repos, hence it was added to db. Modify directly:
			p.CurrentState = pkg.Present
		} else {
			// release is not in repos
			// we don't know the repo where the release has originally been
			// installed from, so we add it as stale package with an empty repo
			// string
			p := pkg.NewPkg(r.Name, r.Chart.Name(), r.Chart.Metadata.Version, r.Namespace,
				pkg.Present, pkg.Unknown, pkg.Present, "", "")
			// fill dep relations:
			if err := i.CreateDepRelsFromAnnot(p, r.Chart.Metadata.Annotations, repoEntries,
				pkgdb, settings, logger); err != nil {
				return err
			}
			pkgdb.Add(p)
		}
	}

	// calculate dep rels for toModify
	// fill dep relations
	if err := i.CreateDepRelsFromAnnot(toModify, toModifyChart.Metadata.Annotations, repoEntries,
		pkgdb, settings, logger); err != nil {
		return err
	}

	// add toModify to db
	pkgdb.Add(toModify)

	return nil
}

// CreateDepRelsFromAnnot fills the p.DepRel and p.DepOptionalRel of a package,
// by unmarshalling and checking the Metadata.Annotations of the chart that
// corresponds to that package.
//
// For local local charts (repository starts with `file://`), it will finish
// without doing anything if they are already present in the DB (have been
// processed), or recursively call itself to create deps from annot and add
// those charts to the DB.
func (i *Install) CreateDepRelsFromAnnot(p *pkg.Pkg,
	chartAnnot map[string]string, repoEntries map[string]chrtEntry,
	pkgdb *solver.PkgDB,
	settings *cli.EnvSettings, logger log.Logger) (err error) {

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
			var depNS, depRelName string
			// find dependency:
			depChrtVer, depInRepo := repoEntries[dep.Name]
			if !depInRepo {
				// pull chart to obtain default ns
				log.Debugf("Dependency \"%s\" not found in repos, loading chart", dep.Name)
				depChart, err := i.LoadChart(dep.Name, p.ParentChartPath, dep.Repository, dep.Version, settings, logger)
				if err != nil {
					return err
				}
				// obtain default ns and release name of dep:
				depNS = GetNamespaceFromAnnot(depChart.Metadata.Annotations, settings.Namespace())
				depRelName = GetNameFromAnnot(depChart.Metadata.Annotations, depChart.Name())

				depP := pkg.NewPkg(depRelName, dep.Name, depChart.Metadata.Version, depNS,
					pkg.Unknown, pkg.Unknown, pkg.Unknown, dep.Repository, p.ParentChartPath)

				if strings.HasPrefix(dep.Repository, "file://") /* depP local */ {
					// if depP is local, it can depend on local charts too: check recursively,
					// but break loops by not recurse into charts already processed.

					if depPinDB := pkgdb.GetPackageByFingerprint(depP.GetFingerPrint()); depPinDB == nil {
						// first time we process depP

						// Add dep to DB, marking it as processed
						pkgdb.Add(depP)

						// Create depP dependency relations, and recursively add any
						// deps depP may have.
						if err := i.CreateDepRelsFromAnnot(depP, depChart.Metadata.Annotations, repoEntries,
							pkgdb, settings, logger); err != nil {
							return err
						}
					}
				} else {
					// depP is not a local chart
					// Add dep to DB
					pkgdb.Add(depP)
				}

			} else {
				// obtain default ns and release name of dep:
				depNS = GetNamespaceFromAnnot(depChrtVer.chartVersions[0].Annotations, settings.Namespace())
				depRelName = GetNameFromAnnot(depChrtVer.chartVersions[0].Annotations, dep.Name)
			}

			// TODO each version can have a different default ns

			switch c {
			case "hypper.cattle.io/shared-dependencies":
				//add relation to pkg
				p.DependsRel = append(p.DependsRel, &pkg.PkgRel{
					ReleaseName: depRelName,
					Namespace:   depNS,
					SemverRange: dep.Version,
					ChartName:   dep.Name,
				})
			case "hypper.cattle.io/optional-dependencies":
				p.DependsOptionalRel = append(p.DependsOptionalRel, &pkg.PkgRel{
					ReleaseName: depRelName,
					Namespace:   depNS,
					SemverRange: dep.Version,
					ChartName:   dep.Name,
				})
			}
		}
	}
	return nil
}
