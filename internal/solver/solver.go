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
	"github.com/crillab/gophersat/maxsat"
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

// BuildConstraints generates all constraints for package p
func (s *Solver) BuildConstraints(p *pkg.Pkg) (constrs []maxsat.Constr) {

	// add constraints for relationships
	packageConstrs := s.buildConstraintRelations(p)
	constrs = append(constrs, packageConstrs...)

	if p.CurrentState == pkg.Present && p.DesiredState != pkg.Absent {
		// p is a release, and is not going to be changed
		packageConstrs := s.buildConstraintPresent(p)
		constrs = append(constrs, packageConstrs...)
		// TODO p is a release, and is not going to be upgraded
	}

	if p.DesiredState != pkg.Unknown {
		// p is going to be installed, or removed (and is a release)
		packageConstrs := s.buildConstraintToModify(p)
		constrs = append(constrs, packageConstrs...)
	}

	// add constraints for all packages that are the same and differ only in version
	packageConstrs = s.buildConstraintAtMost1(p)
	constrs = append(constrs, packageConstrs...)

	return constrs
}

func (s *Solver) Solve() {
	// generate constraints for all packages
	constrs := []maxsat.Constr{}
	for _, p := range s.PkgDB.mapFingerprintToPkg {
		constrs = append(constrs, s.BuildConstraints(p)...)
	}

	// create problem with constraints, and solve
	problem := maxsat.New(constrs...)
	solver := problem.Solver()
	result := solver.Optimal(nil, nil)

	// result.Weight is 0, means all constraints could be solved
	if result.Status.String() == "SAT" && result.Weight == 0 {
		s.PkgResultSet.Status = "SAT"
		// there is a result.model, generate pkg sets then:
		s.GeneratePkgSets(result.Model)
	}

}

func (s *Solver) IsSAT() bool {
	return s.PkgResultSet.Status == "SAT"
}

// GeneratePkgSets obtains back the sets of packages from IDs.
func (s *Solver) GeneratePkgSets(model []bool) {

	s.PkgResultSet.ToInstall = []*pkg.Pkg{}
	s.PkgResultSet.ToRemove = []*pkg.Pkg{}
	s.PkgResultSet.PresentUnchanged = []*pkg.Pkg{}

	// iterate through the db:
	for _, p := range s.PkgDB.mapFingerprintToPkg {
		// obtain pkgResult from model:
		pkgResult := model[p.ID-1]

		// segregate packages into PkgResultSet:
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
		if s.IsSAT() {
			sb.WriteString(fmt.Sprintf("Status: %s\n", s.PkgResultSet.Status))
			sb.WriteString("Packages to be installed:\n")
			for _, p := range s.PkgResultSet.ToInstall {
				sb.WriteString(fmt.Sprintf("%s\t%s\n", p.ReleaseName, p.Version))
			}
			sb.WriteString("\n")
			sb.WriteString("Packages to be removed:\n")
			for _, p := range s.PkgResultSet.ToRemove {
				sb.WriteString(fmt.Sprintf("%s\t%s\n", p.ReleaseName, p.Version))
			}
			sb.WriteString("\n")
			sb.WriteString("Releases already in the system:\n")
			for _, p := range s.PkgResultSet.PresentUnchanged {
				sb.WriteString(fmt.Sprintf("%s\t%s\n", p.ReleaseName, p.Version))
			}
		} else {
			sb.WriteString("Inconsistencies:\n")
			for _, incos := range s.PkgResultSet.Inconsistencies {
				sb.WriteString(fmt.Sprintf("\t%s\n", incos))
			}
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

func (s *Solver) buildConstraintPresent(p *pkg.Pkg) (constr []maxsat.Constr) {
	constr = []maxsat.Constr{}

	// build constraint if package is installed
	if p.CurrentState == pkg.Present {
		// Boolean equation:
		// package1 == true (package1 installed)
		// create lit for solver:
		lit := maxsat.Lit{
			Var:     p.GetFingerPrint(),
			Negated: false, // installed
		}
		// assign the package an id, to recover from model later:
		if p.ID == -1 {
			// first time we use this package, set id
			s.PkgDB.lastElem++
			p.ID = s.PkgDB.lastElem
		}

		sliceConstr := maxsat.HardClause(lit)
		constr = append(constr, sliceConstr)
	}

	return constr
}

func (s *Solver) buildConstraintToModify(p *pkg.Pkg) (constr []maxsat.Constr) {

	if p.CurrentState == pkg.Present { // if is a release
		fps, _ := s.PkgDB.GetPackageFingerprintsThatDifferOnVersionByPackage(p)
		for _, fp := range fps { // for all the packages that only differ in version
			pkgDifferVersion := s.PkgDB.GetPackageByFingerprint(fp)
			if pkgDifferVersion.DesiredState == pkg.Present {
				// package is scheduled for an upgrade, this is not possible
				// as we aren't separating install and upgrade implementation yet
				incons := fmt.Sprintf("Package %s is scheduled for upgrade, did you mean \"hypper upgrade\" instead of \"hypper install\"\n",
					p.GetFingerPrint())
				s.PkgResultSet.Inconsistencies = append(s.PkgResultSet.Inconsistencies, incons)
				break
			}
		}
	}

	// if package not release and we want to install
	if p.CurrentState != pkg.Present && p.DesiredState == pkg.Present {
		// obtain all fps for the packages that only differ in version
		fps, _ := s.PkgDB.GetPackageFingerprintsThatDifferOnVersionByPackage(p)

		// atLeast 1 of all versions
		lits := []maxsat.Lit{}
		coeffs := []int{}
		for _, fp := range fps { // for all the packages that only differ in version
			pkgDifferVersion := s.PkgDB.GetPackageByFingerprint(fp)
			// create lit for solver:
			lit := maxsat.Lit{
				Var:     pkgDifferVersion.GetFingerPrint(),
				Negated: false, // installed
			}
			// assign the package an id, to recover from model later:
			if pkgDifferVersion.ID == -1 {
				// first time we use this package, set id
				s.PkgDB.lastElem++
				pkgDifferVersion.ID = s.PkgDB.lastElem
			}

			coeffs = append(coeffs, 1)
			lits = append(lits, lit)
		}
		sliceConstr := maxsat.HardPBConstr(lits, coeffs, 1)
		constr = append(constr, sliceConstr)
	}

	// build constraint if package is desired removed
	if p.DesiredState == pkg.Absent {
		// Pseudo-Boolean equation:
		// packageA == 0 (packageA installed)
		// E.g:
		// a         == 0  satisfiable?
		// true      0     no
		// false     1     yes
		// create lit for solver:
		lit := maxsat.Lit{
			Var:     p.GetFingerPrint(),
			Negated: true, // not installed
		}
		// assign the package an id, to recover from model later:
		if p.ID == -1 {
			// first time we use this package, set id
			s.PkgDB.lastElem++
			p.ID = s.PkgDB.lastElem
		}

		sliceConstr := maxsat.HardClause(lit)
		constr = append(constr, sliceConstr)
	}

	return constr
}

func (s *Solver) buildConstraintRelations(p *pkg.Pkg) (constr []maxsat.Constr) {
	// E.g: A depends on B,~1.0.0, with B having several or zero versions to
	// chose from.
	// Constraints:
	//     B-1.0.0 + ... + B-1.5.0 - A >= 0
	//     B-1.0.0 + ... + B-1.5.0 >= 1  (at most 1, added outside of this function)

	// build constraints for 'Depends' relations
	for _, deprel := range p.DependsRel {
		// obtain all IDs for the packages that only differ in version
		mapOfVersions := s.PkgDB.GetMapOfVersionsByBaseFingerPrint(deprel.BaseFingerprint)
		matchingVersions := []string{} // slice of fingerprints
		for depVersion, depFingerprint := range mapOfVersions {

			// build list of packages that differ only in version and that satisfy semver
			if semverSatisfies(deprel.SemverRange, depVersion) {
				// efficiently build a slice of version IDs for use in the constraint:
				matchingVersions = append(matchingVersions, depFingerprint)
			}
			// assign the parent package an id, to recover from model later:
			if p.ID == -1 {
				// first time we use this package, set id
				s.PkgDB.lastElem++
				p.ID = s.PkgDB.lastElem
			}
			// assign the dep package an id, to recover from model later:
			depP := s.PkgDB.GetPackageByFingerprint(depFingerprint)
			if depP.ID == -1 {
				// first time we use this package, set id
				s.PkgDB.lastElem++
				depP.ID = s.PkgDB.lastElem
			}

			// A depends on B:  not(A) or B
			sliceConstr := maxsat.HardClause(maxsat.Not(p.GetFingerPrint()), maxsat.Var(depFingerprint))
			constr = append(constr, sliceConstr)
		}

		if len(matchingVersions) == 0 {
			// there are no packages that match the version we depend on, add
			// that to inconsistencies
			// TODO create acyclic graph of result instead
			incons := fmt.Sprintf("Package %s depends on %s, semver %s, but nothing satisfies it\n",
				p.GetFingerPrint(), deprel.BaseFingerprint, deprel.SemverRange)
			s.PkgResultSet.Inconsistencies = append(s.PkgResultSet.Inconsistencies, incons)
			break
		} else {
			// for all deps that satisfy the semver range, use at least 1
			// build lits:
			lits := []maxsat.Lit{}
			coeffs := []int{}
			for _, fp := range matchingVersions {
				// create lit for solver:
				lit := maxsat.Lit{
					Var:     fp,
					Negated: false, // installed
				}
				// assing the package an id, to recover from model later:
				depP := s.PkgDB.GetPackageByFingerprint(fp)
				if depP.ID == -1 {
					// first time we use this package, set id
					s.PkgDB.lastElem++
					depP.ID = s.PkgDB.lastElem
				}

				lits = append(lits, lit)
				coeffs = append(coeffs, 1)
			}
			// at least 1 of all the versions that satisfy semver
			sliceConstr := maxsat.HardPBConstr(lits, coeffs, 1)
			constr = append(constr, sliceConstr)
		}

	}
	return constr
}

func (s *Solver) buildConstraintAtMost1(p *pkg.Pkg) (constr []maxsat.Constr) {
	// E.g: B having several versions: B-1.0.0, B-2.0.0, B-3.0.0
	// Only one can be installed, as they all share releaseName and ns.
	//
	// Add constraint:
	// B-1.3.0 + ... + B-1.2.0 <= 1  (at most 1). If there are no versions of B,
	// it's SAT.

	// obtain all fps, weights, for the packages that only differ in version
	fps, coeffs := PkgDBInstance.GetPackageFingerprintsThatDifferOnVersionByPackage(p)

	// at most 1 of all the IDs is allowed
	// xor(A,B, ... C) == not(A) or not(B) or ... not(C), at least num -1 need to be true
	lits := []maxsat.Lit{}
	for _, fp := range fps { // for all the packages that only differ in version
		pkgDifferVersion := s.PkgDB.GetPackageByFingerprint(fp)
		// create lit for solver:
		lit := maxsat.Lit{
			Var:     pkgDifferVersion.GetFingerPrint(),
			Negated: true, // not installed
		}
		lits = append(lits, lit)
		// assign the package an id, to recover from model later:
		if pkgDifferVersion.ID == -1 {
			// first time we use this package, set id
			s.PkgDB.lastElem++
			pkgDifferVersion.ID = s.PkgDB.lastElem
		}
	}
	sliceConstr := maxsat.HardPBConstr(lits, coeffs, len(lits)-1)
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
	// Check if the version meets the constraints.
	return c.Check(v)
}
