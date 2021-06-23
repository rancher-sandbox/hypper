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
Package solver provides operations on packages: install, remove, upgrade,
check integrity and others.
*/

package solver

import (
	"encoding/json"
	"fmt"
	"strings"

	gsolver "github.com/crillab/gophersat/solver"
	"github.com/rancher-sandbox/hypper/internal/pkg"
	"github.com/rancher-sandbox/hypper/pkg/repo"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/release"
)

type Solver struct {
	PkgDB        *PkgDB       // DB containing packages
	pkgResultSet PkgResultSet // outcome of sat solving
	// Strategy
	//
	// Install1: tries to install, if UNSAT, tells why, and that you may be wanting to do upgrade.
	//           Or, check before s.Solve() by querying that the chart to be installed
	//           doesn't have a releasename already in use.
	// Upgrade1: don't add a constraintPresent for the specific release/package to be upgraded.
	//           We get a package to be installed, which conflicts with a release. Find release,
	//           mark as DesiredState:Absent. Which means, checking the releases before s.Solve(),
	//           because an upgrade of a chart that is not a release cannot be performed.
	// Upgrade1ToMajor: same as Upgrade, but tune semver distances.
	// Upgrade1ToMinor: same as Upgrade, but tune semver distances.
	// UpgradeAll: don't add a constraintPresent for all the current releases,
	//             with the AtMost1 of all the versions, plus semver distance, it's ok.
	// UpgradeAllToMajor: same as UpgradeAll, but tune semver distances.
	// UpgradeAllToMinor: same as UpgradeAll, but tune semver distances.
	// Remove1: don't add a constraintPresent for the specific release/package to be removed.
	//          CurrentStatus:Present and DesiredStatus:Absent
	// CheckAll: add everything with desiredState:Unknown, check SAT or UNSAT.
	// AutoremoveAll: all packages not marked as autoinstalled can be dropped. Not
	//                enough with setting their desiredState to absent, they may be
	//                dependencies.
}

// PkgResultSet contains the status outcome of solving, and the different sets of
// packages derived from the outcome.
// It will be marshalled into Yaml and Json.
type PkgResultSet struct {
	PresentUnchanged []*pkg.Pkg
	ToInstall        []*pkg.Pkg
	ToRemove         []*pkg.Pkg
	Status           string
	//Incosistencies []
}

type OutputMode int

const (
	JSON OutputMode = iota
	YAML
	Table
)

// New creates a new Solver, initializing its database.
func New() (s *Solver) {
	return &Solver{
		PkgDB:        CreatePkgDBInstance(),
		pkgResultSet: PkgResultSet{},
	}
}

// FIXME assume all charts come from just 1 repo. We will generalize later.
// TODO add chart.Hash into index file, to not need to pull hypper charts.
// Note that we will still need to pull helm charts to calculate its chart.Hash
func (s *Solver) BuildWorld(repoEntriesSlice []*map[string]repo.ChartVersions, releases []*release.Release, toModify []*pkg.Pkg) (err error) {
	//first, for all charts and releases create a package and assign it an ID

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
				p := pkg.NewPkgFromChart(chart, nil, nil, pkg.Unknown)
				s.PkgDB.Add(p)
			}
		}
	}

	// add releases to db
	for _, r := range releases {
		s.PkgDB.Add(pkg.NewPkgFromRelease(r))
	}

	// add toModify to db
	for _, p := range toModify {
		s.PkgDB.Add(p)
	}

	// second, fill all the packages with information about dependency
	// relations, now that all packages are in the db and they have assigned IDs
	for _, p := range s.PkgDB.mapFingerprintToPkg {
		s.PkgDB.UpdateDeps(p)
	}

	return nil
}

// BuildWorldMock fills the database with pkgs instead of releases, charts from
// repositories, and so.
// Useful for testing.
func (s *Solver) BuildWorldMock(pkgs []*pkg.Pkg) {
	for _, p := range pkgs {
		s.PkgDB.Add(p)
	}
}

// BuildConstraints generates all constraints for package p with ID
func (s *Solver) BuildConstraints(p *pkg.Pkg) (constrs []gsolver.PBConstr) {

	// add constraints for relationships
	packageConstrs := s.buildConstraintRelations(p)
	constrs = append(constrs, packageConstrs...)

	if p.CurrentState == pkg.Present && p.DesiredState != pkg.Absent {
		// p is a release, and is not going to be changed
		packageConstrs := buildConstraintPresent(p)
		constrs = append(constrs, packageConstrs...)
		// TODO p is a release, and is not going to be upgraded
	}

	if p.DesiredState != pkg.Unknown {
		// p is going to be installed, or removed (and is a release)
		packageConstrs := buildConstraintToModify(p)
		constrs = append(constrs, packageConstrs...)
	}

	// TODO don't hardcode desired version of a dependency, accept a version interval
	// depending on the semver range

	// TODO the rest of constraints get duplicated several times, when they
	// should only be added once. We should only iterate the
	// packages-differing-in-version once (which means having a several
	// database).

	// add constraints for all packages that are the same and differ only in version
	packageConstrs = buildConstraintAtMost1(p)
	constrs = append(constrs, packageConstrs...)

	// TODO semvers for all packages
	// 	constraint to take the newest: min(semver distance)

	return constrs
}

func (s *Solver) Solve() {
	// TODO grab lock when creating world, release after solving (maybe better
	// in pkg/action)

	// generate constraints for all packages
	constrs := []gsolver.PBConstr{}
	for _, p := range s.PkgDB.mapFingerprintToPkg {
		constrs = append(constrs, s.BuildConstraints(p)...)
	}

	// create problem with constraints, and solve
	pb := gsolver.ParsePBConstrs(constrs)

	sp := gsolver.New(pb)
	// sp.Verbose = true

	// result.model is a [id]bool, saying if the package should be present or not
	result := sp.Optimal(nil, nil)
	s.pkgResultSet.Status = result.Status.String()

	s.GeneratePkgSets(result.Model)
}

// GeneratePkgSets obtains back the sets of packages from IDs.
func (s *Solver) GeneratePkgSets(model []bool) {

	s.pkgResultSet.ToInstall = []*pkg.Pkg{}
	s.pkgResultSet.ToRemove = []*pkg.Pkg{}
	s.pkgResultSet.PresentUnchanged = []*pkg.Pkg{}

	for id, pkgResult := range model {
		p := s.PkgDB.GetPackageByPbID(id + 1) // model starts at 0, IDs at 1

		if pkgResult && p.CurrentState == pkg.Present {
			s.pkgResultSet.PresentUnchanged = append(s.pkgResultSet.PresentUnchanged, p)
		} else if pkgResult && p.CurrentState != pkg.Present {
			s.pkgResultSet.ToInstall = append(s.pkgResultSet.ToInstall, p)
		} else if !pkgResult && p.CurrentState == pkg.Present {
			s.pkgResultSet.ToRemove = append(s.pkgResultSet.ToRemove, p)
		}
	}
}

func (s *Solver) FormatOutput(t OutputMode) (output string) {
	var sb strings.Builder
	switch t {
	case Table:
		// TODO: Refurbish this to create some fancy emoji/table output
		sb.WriteString(fmt.Sprintf("Status: %s\n", s.pkgResultSet.Status))
		sb.WriteString("Packages to be installed:\n")
		for _, p := range s.pkgResultSet.ToInstall {
			sb.WriteString(fmt.Sprintf("%s\t%s\n", p.Name, p.Version))
		}
		sb.WriteString("\n")
		sb.WriteString("Packages to be removed:\n")
		for _, p := range s.pkgResultSet.ToRemove {
			sb.WriteString(fmt.Sprintf("%s\t%s\n", p.Name, p.Version))
		}
		sb.WriteString("\n")
		sb.WriteString("Releases already in the system:\n")
		for _, p := range s.pkgResultSet.PresentUnchanged {
			sb.WriteString(fmt.Sprintf("%s\t%s\n", p.Name, p.Version))
		}
	case YAML:
		o, _ := yaml.Marshal(s.pkgResultSet)
		sb.WriteString(string(o))
	case JSON:
		o, err := json.Marshal(s.pkgResultSet)
		fmt.Println(err)
		sb.WriteString(string(o))
	}
	return sb.String()
}

// operations to provide:
// install(pkg...)
// upgradeToMinor(pkg...)
// upgradeToMajor(pkg...)
// uninstall(pkg...)
// integrityCheck()

// TODO maybe in the future:
// CRDs
// values.yaml
// autoremove

func buildConstraintPresent(p *pkg.Pkg) (constr []gsolver.PBConstr) {
	constr = []gsolver.PBConstr{}
	// obtain ID to use in constraints
	id := PkgDBInstance.GetIDByPackage(p)

	// build constraint if package is installed
	if p.CurrentState == pkg.Present {
		// Pseudo-Boolean equation:
		// package1 == 1 (package1 installed)
		sliceConstr := gsolver.Eq([]int{id}, []int{1}, 1)
		constr = append(constr, sliceConstr...)
	}

	return constr
}

func buildConstraintToModify(p *pkg.Pkg) (constr []gsolver.PBConstr) {
	constr = []gsolver.PBConstr{}
	// obtain ID to use in constraints
	id := PkgDBInstance.GetIDByPackage(p)

	// build constraint if package is desired installed
	if p.DesiredState == pkg.Present {
		// Pseudo-Boolean equation:
		// packageA == 1 (packageA installed)
		// E.g:
		// a         == 1  satisfiable?
		// true      1     yes
		// false     0     no
		sliceConstr := gsolver.Eq([]int{id}, []int{1}, 1)
		constr = append(constr, sliceConstr...)
	}

	// build constraint if package is desired removed
	if p.DesiredState == pkg.Absent {
		// Pseudo-Boolean equation:
		// packageA == 0 (packageA installed)
		// E.g:
		// a         == 0  satisfiable?
		// true      0     no
		// false     1     yes
		sliceConstr := gsolver.Eq([]int{id}, []int{1}, 0)
		constr = append(constr, sliceConstr...)
	}

	return constr
}

func (s *Solver) buildConstraintRelations(p *pkg.Pkg) (constr []gsolver.PBConstr) {
	constr = []gsolver.PBConstr{}
	// obtain ID to use in constraints
	id := PkgDBInstance.GetIDByPackage(p)

	// build constraints for 'Depends' relations
	for _, depfp := range p.DependsRel {
		// Pseudo-Boolean equation:
		// a depends on b and on c: b - a >= 0 ; c - a >= 0
		// E.g:
		// b     -     a       >= 0   satisfiable?
		// true        false   1      yes
		// false       false   0      yes
		// true        true    0      yes
		// false       true    -1     no
		// weirdly, the lib needs a GtEq(x,y,1) instead of 0
		sliceConstr := gsolver.GtEq([]int{s.PkgDB.mapFingerprintToPbID[depfp], -1 * id}, []int{1, 1}, 1)
		constr = append(constr, sliceConstr)
	}

	// build constraints for 'Optional-Depends' relations
	for _, depfp := range p.DependsOptionalRel {
		// Pseudo-Boolean equation:
		// same as example above
		// weirdly, the lib needs a GtEq(x,y,1) instead of 0
		sliceConstr := gsolver.GtEq([]int{s.PkgDB.mapFingerprintToPbID[depfp], -1 * id}, []int{1, 1}, 1)
		constr = append(constr, sliceConstr)
	}

	return constr
}

func buildConstraintAtMost1(p *pkg.Pkg) (constr []gsolver.PBConstr) {
	// obtain all IDs for the packages that only differ in version
	ids := PkgDBInstance.GetPackageIDsThatDifferOnVersionByPackage(p)

	// at most 1 of all the IDs is allowed
	// Pseudo-Boolean equation:
	// a + b + ... + c == 1
	// a       b      c        == 1  satisfiable?
	// true    false  false    1     yes
	// false   true   true     2     no
	sliceConstr := gsolver.AtMost(ids, 1)
	constr = append(constr, sliceConstr)

	return constr
}
