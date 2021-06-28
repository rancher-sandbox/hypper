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

	"github.com/Masterminds/semver/v3"
	gsolver "github.com/crillab/gophersat/solver"
	pkg "github.com/rancher-sandbox/hypper/internal/package"
	"gopkg.in/yaml.v2"
)

type Solver struct {
	PkgDB        *PkgDB       // DB containing packages
	PkgResultSet PkgResultSet // outcome of sat solving
	// TODO Strategy
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
	Inconsistencies  []string
}

type OutputMode int

const (
	JSON OutputMode = iota
	YAML
	Table
)

// New creates a new Solver, initializing its database.
func New() (s *Solver) {
	s = &Solver{
		PkgDB:        CreatePkgDBInstance(),
		PkgResultSet: PkgResultSet{},
	}
	s.PkgResultSet.Inconsistencies = []string{}
	return s
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
	// generate constraints for all packages
	constrs := []gsolver.PBConstr{}
	for _, p := range s.PkgDB.mapFingerprintToPkg {
		constrs = append(constrs, s.BuildConstraints(p)...)
	}

	// create problem with constraints, and solve
	pb := gsolver.ParsePBConstrs(constrs)

	sp := gsolver.New(pb)

	// result.model is a [id]bool, saying if the package should be present or not
	result := sp.Optimal(nil, nil)
	s.PkgResultSet.Status = result.Status.String()

	s.GeneratePkgSets(result.Model)
}

func (s *Solver) IsSAT() bool {
	return s.PkgResultSet.Status == "SAT"
}

// GeneratePkgSets obtains back the sets of packages from IDs.
func (s *Solver) GeneratePkgSets(model []bool) {

	s.PkgResultSet.ToInstall = []*pkg.Pkg{}
	s.PkgResultSet.ToRemove = []*pkg.Pkg{}
	s.PkgResultSet.PresentUnchanged = []*pkg.Pkg{}

	for id, pkgResult := range model {
		p := s.PkgDB.GetPackageByPbID(id + 1) // model starts at 0, IDs at 1

		if pkgResult && p.CurrentState == pkg.Present {
			s.PkgResultSet.PresentUnchanged = append(s.PkgResultSet.PresentUnchanged, p)
		} else if pkgResult && p.CurrentState != pkg.Present {
			s.PkgResultSet.ToInstall = append(s.PkgResultSet.ToInstall, p)
		} else if !pkgResult && p.CurrentState == pkg.Present {
			s.PkgResultSet.ToRemove = append(s.PkgResultSet.ToRemove, p)
		}
	}
}

func (s *Solver) FormatOutput(t OutputMode) (output string) {
	var sb strings.Builder
	switch t {
	case Table:
		// TODO: Refurbish this to create some fancy emoji/table output
		sb.WriteString(fmt.Sprintf("Status: %s\n", s.PkgResultSet.Status))
		sb.WriteString("Packages to be installed:\n")
		for _, p := range s.PkgResultSet.ToInstall {
			sb.WriteString(fmt.Sprintf("%s\t%s\n", p.Name, p.Version))
		}
		sb.WriteString("\n")
		sb.WriteString("Packages to be removed:\n")
		for _, p := range s.PkgResultSet.ToRemove {
			sb.WriteString(fmt.Sprintf("%s\t%s\n", p.Name, p.Version))
		}
		sb.WriteString("\n")
		sb.WriteString("Releases already in the system:\n")
		for _, p := range s.PkgResultSet.PresentUnchanged {
			sb.WriteString(fmt.Sprintf("%s\t%s\n", p.Name, p.Version))
		}
	case YAML:
		o, _ := yaml.Marshal(s.PkgResultSet)
		sb.WriteString(string(o))
	case JSON:
		o, err := json.Marshal(s.PkgResultSet)
		fmt.Println(err)
		sb.WriteString(string(o))
	}
	return sb.String()
}

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
	// E.g: A depends on B,~1.0.0, with B having several or zero versions to
	// chose from.
	// Constraints:
	//     B-1.0.0 + ... + B-1.5.0 - A >= 0
	//     B-1.0.0 + ... + B-1.5.0 >= 1  (at most 1, added outside of this function)

	constr = []gsolver.PBConstr{}
	// obtain ID to use in constraints
	parentID := PkgDBInstance.GetIDByPackage(p)

	// build constraints for 'Depends' relations
	for _, deprel := range p.DependsRel {
		// obtain all IDs for the packages that only differ in version
		mapOfVersions := s.PkgDB.GetMapOfVersionsByBaseFingerPrint(deprel.BaseFingerprint)
		matchingVersionIDs := []int{}
		for depVersion, depFingerprint := range mapOfVersions {
			depID := PkgDBInstance.GetPbIDByFingerprint(depFingerprint)

			// add a constraint to install those versions that satisfy semver range
			if semverSatisfies(deprel.SemverRange, depVersion) {
				// efficiently build a slice of version IDs for use in the constraint:
				matchingVersionIDs = append(matchingVersionIDs, depID)
			}
		}
		if len(matchingVersionIDs) == 0 {
			// there are no packages that match the version we depend on, add
			// that to inconsistencies
			// TODO create acyclic graph of result instead
			incos := fmt.Sprintf("Package %s depends on %s, semver %s, but nothing satisfies it\n",
				p.GetFingerPrint(), deprel.BaseFingerprint, deprel.SemverRange)
			s.PkgResultSet.Inconsistencies = append(s.PkgResultSet.Inconsistencies, incos)
		}

		// A depends on all valid versions of B.
		// Pseudo-Boolean equation:
		// B-1.0.0 + ... + B-1.5.0 - A >= 0   satisfiable?
		// true            false   - 1    0    yes, 1 package satisfies dependency
		// false           true      0    0    yes, A is not being installed
		// true            false     0    0    yes, A is not being installed
		// false           false   - 1   -1    no, no package satisfies dependency

		// build []lits and []weights:
		lits := append(matchingVersionIDs, -1*parentID) // B1 + ... + B2 - A
		weights := make([]int, len(lits))
		for i := range weights {
			weights[i] = 1
		}
		// weirdly, the lib needs a GtEq(x,y,1) instead of 0
		sliceConstr := gsolver.GtEq(lits, weights, 1)
		constr = append(constr, sliceConstr)
	}

	// build constraints for 'Optional-Depends' relations
	for _, deprel := range p.DependsOptionalRel {
		// obtain all IDs for the packages that only differ in version
		mapOfVersions := s.PkgDB.GetMapOfVersionsByBaseFingerPrint(deprel.BaseFingerprint)
		matchingVersionIDs := []int{}
		for depVersion, depFingerprint := range mapOfVersions {
			depID := PkgDBInstance.GetPbIDByFingerprint(depFingerprint)

			// add a constraint to install those versions that satisfy semver range
			if semverSatisfies(deprel.SemverRange, depVersion) {
				// efficiently build a slice of version IDs for use in the constraint:
				matchingVersionIDs = append(matchingVersionIDs, depID)
			}
		}

		// A depends on all valid versions of B.
		// Pseudo-Boolean equation:
		// B-1.0.0 + ... + B-1.5.0 - A >= 0   satisfiable?
		// true            false   - 1    0    yes, 1 package satisfies dependency
		// false           true      0    0    yes, A is not being installed
		// true            false     0    0    yes, A is not being installed
		// false           false   - 1   -1    no, no package satisfies dependency

		// build []lits and []weights:
		lits := append(matchingVersionIDs, -1*parentID) // B1 + ... + B2 - A
		weights := make([]int, len(lits))
		for i := range weights {
			weights[i] = 1
		}
		// weirdly, the lib needs a GtEq(x,y,1) instead of 0
		sliceConstr := gsolver.GtEq(lits, weights, 1)
		constr = append(constr, sliceConstr)
	}

	return constr
}

func buildConstraintAtMost1(p *pkg.Pkg) (constr []gsolver.PBConstr) {
	// E.g: B having several versions: B-1.0.0, B-2.0.0, B-3.0.0
	// Only one can be installed, as they all share releaseName and ns.
	//
	// Add constraint:
	// B-1.3.0 + ... + B-1.2.0 <= 1  (at most 1). If there are no versions of B,
	// it's SAT.

	// obtain all IDs for the packages that only differ in version
	ids, _ := PkgDBInstance.GetPackageIDsThatDifferOnVersionByPackage(p)

	// at most 1 of all the IDs is allowed
	// Pseudo-Boolean equation:
	// a*X + b*Y + ... + c*Z == 1
	// a       b      c        == 1  satisfiable?
	// true    false  false    1     yes
	// false   true   true     2     no
	// weirdly, the lib needs a GtEq(x,y,1) instead of 0
	//TODO sliceConstr := gsolver.LtEq(ids, weights, 1)
	sliceConstr := gsolver.AtMost(ids, 1)
	constr = append(constr, sliceConstr)

	return constr
}

func semverSatisfies(semverRange string, ourSemver string) bool {

	// generate semver constraint and check:
	c, err := semver.NewConstraint(semverRange)
	if err != nil {
		// TODO Handle constraint not being parseable.
		return false
	}

	v, err := semver.NewVersion(ourSemver)
	if err != nil {
		// TODO Handle version not being parseable.
		return false
	}
	// Check if the version meets the constraints. The a variable will be true.
	return c.Check(v)
}
