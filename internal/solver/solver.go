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
	"sort"
	"strings"
	"sync"

	"github.com/Masterminds/log-go"
	"github.com/Masterminds/semver/v3"
	"github.com/crillab/gophersat/maxsat"
	pkg "github.com/rancher-sandbox/hypper/internal/package"
	"gopkg.in/yaml.v2"
)

type SolverStrategy int

const (
	InstallOne SolverStrategy = iota
	// Install 1 package. If chart to be installed has a releaseName and NS
	// already in use, result is UNSAT, and informs that you may be wanting to
	// do upgrade.

	UpgradeOne
	// Don't add a constraintPresent for the specific release/package to be upgraded.
	// We get a package to be installed, which conflicts with a release. Find release,
	// mark as DesiredState:Absent. Which means, checking the releases before s.Solve(),
	// because an upgrade of a chart that is not a release cannot be performed.

	// TODO
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
)

// Solver performs SAT solving of dependency problems. It codifies the state of
// the world into packages, saved into a package database. It gets created with
// a specific SolverStrategy, and contains the results in PkgResultSet.
type Solver struct {
	PkgDB        *PkgDB       // DB containing packages
	PkgResultSet PkgResultSet // outcome of sat solving
	Strategy     SolverStrategy
	logger       log.Logger
	model        maxsat.Model
}

// PkgTree is a polytree (directed, acyclic graph) of packages.
// Used for storing the tree of packages to be installed.
type PkgTree struct {
	Node      *pkg.Pkg
	Relations []*PkgTree
}

// PkgResultSet contains the status outcome of solving, and the different sets of
// packages derived from the outcome.
// It will be marshalled into Yaml and Json.
type PkgResultSet struct {
	PresentUnchanged []*pkg.Pkg
	ToInstall        *PkgTree
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
func New(strategy SolverStrategy, logger log.Logger) (s *Solver) {
	s = &Solver{
		PkgDB:        CreatePkgDBInstance(),
		PkgResultSet: PkgResultSet{},
		Strategy:     strategy,
		logger:       logger,
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

func (s *Solver) Solve(wantedPkg *pkg.Pkg) {
	// generate constraints for all packages
	s.logger.Debug("Building constraints…")
	var (
		mu      = &sync.Mutex{}
		constrs = make([]maxsat.Constr, 0)
	)
	var waitgroup sync.WaitGroup
	for _, p := range s.PkgDB.mapFingerprintToPkg {
		waitgroup.Add(1)
		go func(p *pkg.Pkg) {
			defer waitgroup.Done()
			tmpConstrs := s.BuildConstraints(p)
			mu.Lock()
			constrs = append(constrs, tmpConstrs...)
			mu.Unlock()
		}(p)
	}
	waitgroup.Wait()

	// s.logger.Debug("Constraints:")
	// for _, c := range constrs {
	// 	s.logger.Debugf("    %v\n", c)
	// }

	s.logger.Debug("Solving…")

	// create problem with constraints, and solve
	problem := maxsat.New(constrs...)
	s.model, _ = problem.Solve()

	if s.model != nil { // SAT
		//	there is a result model, generate pkg sets then:
		s.GeneratePkgSets(wantedPkg)
		s.PkgResultSet.Status = "SAT"
		s.logger.Debug("Result: SAT\n")
	} else {
		s.PkgResultSet.Status = "UNSAT"
		s.logger.Debug("Result: UNSAT\n")
	}

	// s.logger.Debugf("Result %v\n", s.model)
}

func (s *Solver) IsSAT() bool {
	return s.PkgResultSet.Status == "SAT"
}

// GeneratePkgSets obtains back the sets of packages from IDs.
func (s *Solver) GeneratePkgSets(wantedPkg *pkg.Pkg) {

	s.PkgResultSet.ToRemove = []*pkg.Pkg{}
	s.PkgResultSet.PresentUnchanged = []*pkg.Pkg{}

	// iterate through the model:
	for fp, pkgResult := range s.model {
		p := s.PkgDB.GetPackageByFingerprint(fp)
		// segregate packages into PkgResultSet:
		if pkgResult && p.CurrentState == pkg.Present {
			s.PkgResultSet.PresentUnchanged = append(s.PkgResultSet.PresentUnchanged, p)
		} else if !pkgResult && p.CurrentState == pkg.Present {
			s.PkgResultSet.ToRemove = append(s.PkgResultSet.ToRemove, p)
		}
	}

	if s.Strategy == InstallOne {
		s.PkgResultSet.ToInstall = &PkgTree{}
		visited := map[string]bool{}
		// add dependencies of wantedPkg
		s.PkgResultSet.ToInstall = s.recBuildTree(wantedPkg, visited)
	}
}

func (s *Solver) recBuildTree(p *pkg.Pkg, visited map[string]bool) *PkgTree {
	if p == nil {
		// we are a leaf, stop
		return nil
	}

	// create tree with p:
	tr := &PkgTree{
		Node:      p,
		Relations: []*PkgTree{},
	}

	// add p to visited:
	visited[p.GetFingerPrint()] = true

	// recursively create trees with dependencies of p:
	for _, depRel := range p.DependsRel {
		depBFP := pkg.CreateBaseFingerPrint(depRel.ReleaseName, depRel.Namespace)
		// see if dependency is in the model:
		for modelFP, pkgResult := range s.model {
			modelP := s.PkgDB.GetPackageByFingerprint(modelFP)
			if !pkgResult {
				continue // skip pkgs in model that are not be installed
			}
			modelBFP := modelP.GetBaseFingerPrint()
			if modelBFP == depBFP { // found our dependency in the model
				// if dependency was already visited, skip:
				if visited[modelFP] {
					continue
				}
				// if dependency is not known, add to tree and recursive call:
				tr.Relations = append(tr.Relations, s.recBuildTree(modelP, visited))
			}
		}
	}
	return tr
}

func (s *Solver) SortPkgSets() {
	sort.Strings(s.PkgResultSet.Inconsistencies)
	sort.SliceStable(s.PkgResultSet.PresentUnchanged, func(i, j int) bool {
		return s.PkgResultSet.PresentUnchanged[i].ChartName < s.PkgResultSet.PresentUnchanged[j].ChartName
	})
	sort.SliceStable(s.PkgResultSet.ToRemove, func(i, j int) bool {
		return s.PkgResultSet.ToRemove[i].ChartName < s.PkgResultSet.ToRemove[j].ChartName
	})
}

func PrintPkgTree(tr *PkgTree, lvl int) (output string) {
	var sb strings.Builder
	if tr == nil {
		return ""
	}
	sb.WriteString(fmt.Sprintf("%s\t%s\n", tr.Node.ReleaseName, tr.Node.Version))
	for _, rel := range tr.Relations {
		sb.WriteString(fmt.Sprintf("%s%s\n", strings.Repeat("\t", lvl), PrintPkgTree(rel, lvl+1)))
	}
	return sb.String()
}

func (s *Solver) FormatOutput(t OutputMode) (output string) {
	var sb strings.Builder
	switch t {
	case Table:
		// TODO: Refurbish this to create some fancy emoji/table output
		sb.WriteString(fmt.Sprintf("Status: %s\n", s.PkgResultSet.Status))
		sb.WriteString("Packages to be installed:\n")
		sb.WriteString(PrintPkgTree(s.PkgResultSet.ToInstall, 0))
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
		sb.WriteString("Inconsistencies:\n")
		for _, incos := range s.PkgResultSet.Inconsistencies {
			sb.WriteString(fmt.Sprintf("\t%s\n", incos))
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

// buildConstraintPresent returns a constraint specifycing that package p is to
// be present in result
func (s *Solver) buildConstraintPresent(p *pkg.Pkg) (constr []maxsat.Constr) {
	// Boolean equation:
	// packageA == true (packageA installed)

	// create lit for solver:
	lit := maxsat.Lit{
		Var:     p.GetFingerPrint(),
		Negated: false, // installed
	}

	sliceConstr := maxsat.HardClause(lit)
	constr = append(constr, sliceConstr)

	return constr
}

func (s *Solver) buildConstraintToModify(p *pkg.Pkg) (constr []maxsat.Constr) {

	if p.CurrentState == pkg.Present { // if is a release
		fps, _ := s.PkgDB.GetOrderedPackageFingerprintsThatDifferOnVersionByPackage(p)
		for _, fp := range fps { // for all the packages that only differ in version
			pkgDifferVersion := s.PkgDB.GetPackageByFingerprint(fp)
			if pkgDifferVersion.DesiredState == pkg.Present {
				if s.Strategy == InstallOne {
					// package is scheduled for an upgrade, this is not possible
					// as we aren't separating install and upgrade implementation yet
					incons := fmt.Sprintf("Package %s is scheduled for upgrade, did you mean \"hypper upgrade\" instead of \"hypper install\"\n",
						p.GetFingerPrint())
					s.PkgResultSet.Inconsistencies = append(s.PkgResultSet.Inconsistencies, incons)
					break
				}
				if s.Strategy == UpgradeOne {
					// we have found a package, pkgDifferVersion, to upgrade the release in p.
					// Add constraint for new package.
					lit := []maxsat.Lit{{
						Var:     pkgDifferVersion.GetFingerPrint(),
						Negated: false, // installed
					}}
					sliceConstr := maxsat.HardPBConstr(lit, nil, 1)
					constr = append(constr, sliceConstr)
				}
			}
		}
	}

	// if package not release and we want to install
	if p.CurrentState != pkg.Present && p.DesiredState == pkg.Present {

		if p.PinnedVer == pkg.Present {
			// we only want 1 version, the current one:

			lit := []maxsat.Lit{{
				Var:     p.GetFingerPrint(),
				Negated: false, // installed
			}}
			sliceConstr := maxsat.HardPBConstr(lit, nil, 1)
			constr = append(constr, sliceConstr)

		} else {
			// atLeast 1 of all versions

			// obtain all fps for the packages that only differ in version
			fps, _ := s.PkgDB.GetOrderedPackageFingerprintsThatDifferOnVersionByPackage(p)
			lits := []maxsat.Lit{}
			coeffs := []int{}
			for _, fp := range fps { // for all the packages that only differ in version
				pkgDifferVersion := s.PkgDB.GetPackageByFingerprint(fp)
				// create lit for solver:
				lit := maxsat.Lit{
					Var:     pkgDifferVersion.GetFingerPrint(),
					Negated: false, // installed
				}

				coeffs = append(coeffs, 1)
				lits = append(lits, lit)
			}
			sliceConstr := maxsat.HardPBConstr(lits, coeffs, 1)
			constr = append(constr, sliceConstr)
		}

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

		sliceConstr := maxsat.HardClause(lit)
		constr = append(constr, sliceConstr)
	}

	return constr
}

func (s *Solver) buildConstraintRelations(p *pkg.Pkg) (constr []maxsat.Constr) {
	// E.g: A depends on B,~1.0.0, with B having several or zero versions to
	// chose from.
	// Constraints:
	//
	// A depends on any versions of B:  not(A) or B
	// And least 1 satisfying version of B:  B-1.0.0 + ... + B-1.5.0 >= 1
	// These two constraints are equivalent to:
	// A depends on at least 1 version of B:
	//    not(A) or atleast1(B-1.0.0, B-2.0.0 ..., B-3.0.0) ==
	//    atLeast1( not(A) or B-1.0.0 or B-2.0.0 ... B-3.0.0 )
	//
	// At most 1 of all the possible versions, satisfying semver range or not,
	// as they all share releaseName and namespace  (added outside of this function)
	//     B-1.0.0 + ... + B-1.5.0 + B-3.0.0 <= 1

	// build constraints for 'Depends' relations
	for _, deprel := range p.DependsRel {
		// obtain all IDs for the packages that only differ in version
		mapOfVersions := s.PkgDB.GetMapOfVersionsByBaseFingerPrint(pkg.CreateBaseFingerPrint(deprel.ReleaseName, deprel.Namespace))
		satisfyingVersions := []string{} // slice of fingerprints
		for depVersion, depFingerprint := range mapOfVersions {
			// build list of packages that differ only in version and that satisfy semver
			if semverSatisfies(deprel.SemverRange, depVersion) {
				// efficiently build a slice of version IDs for use in the constraint:
				satisfyingVersions = append(satisfyingVersions, depFingerprint)
			}
		}

		// build lits:  not(A) , B1, B2, B3, B4
		lits := []maxsat.Lit{}

		// create parent lit: not(A)
		litParent := maxsat.Lit{
			Var:     p.GetFingerPrint(),
			Negated: true, // not installed
		}
		lits = append(lits, litParent)

		for _, fp := range satisfyingVersions {
			// create lit for solver:
			lit := maxsat.Lit{
				Var:     fp,
				Negated: false, // installed
			}

			lits = append(lits, lit)
		}

		if len(satisfyingVersions) == 0 {
			// there are no packages that match the version we depend on, add
			// that to inconsistencies
			incons := fmt.Sprintf("Chart \"%s\" depends on \"%s\" in namespace \"%s\", semver \"%s\", but nothing satisfies it",
				p.ChartName, deprel.ReleaseName, deprel.Namespace, deprel.SemverRange)
			s.PkgResultSet.Inconsistencies = append(s.PkgResultSet.Inconsistencies, incons)
		}

		// at least 1 of all the versions that satisfy semver, and not(A)
		//
		// for cases where the dependency package is not in repos (and we don't
		// know its default ns), this keeps doing the correct thing. As asking for:
		//   A == true
		//   not(A) + (no known B verions) == true
		// will result in UNSAT.
		sliceConstr := maxsat.HardPBConstr(lits, nil, 1)
		constr = append(constr, sliceConstr)
	}
	return constr
}

func (s *Solver) buildConstraintAtMost1(p *pkg.Pkg) (constr []maxsat.Constr) {
	// E.g: B having several versions: B-1.0.0, B-2.0.0, B-3.0.0
	// Only one can be installed, as they all share releaseName and ns.
	//
	// Add constraint:
	//
	//   B-1.0.0 + ... + B-3.0.0 <= 1  (at most 1). If there are no versions of B,
	//   it's SAT.
	//
	// This is equivalent to:
	//
	//   not(A) or not(B) or ... not(C) where at least numPackages -1 need to be
	//   true, to only select 1 package
	//
	// In case that there's only 1 version of B, we can skip adding a constraint

	// obtain all fps, weights, for the packages that only differ in version
	fps, coeffs := PkgDBInstance.GetOrderedPackageFingerprintsThatDifferOnVersionByPackage(p)

	if len(fps) == 1 {
		// there is only one package on that releaseName and Namespace. No need
		// to create the constraint.
		return []maxsat.Constr{}
	}

	lits := []maxsat.Lit{}
	for i, fp := range fps { // for all the packages that only differ in version
		pkgDifferVersion := s.PkgDB.GetPackageByFingerprint(fp)

		// create lit for atleast(not(B1), not(B2),  num -1):
		lit := maxsat.Lit{
			Var:     pkgDifferVersion.GetFingerPrint(),
			Negated: true, // not installed
		}
		lits = append(lits, lit)

		// create lit for weighting semver distances:
		weightedLit := []maxsat.Lit{{
			Var:     pkgDifferVersion.GetFingerPrint(),
			Negated: false, // installed
		}}

		if pkgDifferVersion.DesiredState == pkg.Present {
			// add weighted constraints to select newest version
			sliceConstr := maxsat.WeightedClause(weightedLit, coeffs[i])
			constr = append(constr, sliceConstr)
		}
	}
	atLeast := len(lits) - 1
	sliceConstr := maxsat.HardPBConstr(lits, nil, atLeast)
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
