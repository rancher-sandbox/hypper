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

package solver

import (
	"math"
	"sort"

	"github.com/Masterminds/log-go"
	pkg "github.com/rancher-sandbox/hypper/internal/package"

	"github.com/Masterminds/semver/v3"
)

// PkgDB implements a database of 2 keys (ID, fingerprint) and 1 value
// (*pkg.Pkg). Each package contains also a table of IDs for packages that
// are the similar to the current package and only differ in the version.
//
// The ID key starts at 1, as Pseudo-Boolean IDs cannot be 0.
//
// When packages get added to the database, they get assigned an ID.
//
// If a package is already present in the database, when adding, it would get
// merged with the existent entry in a way to complete unknown info of that
// package.
type PkgDB struct {
	mapFingerprintToPkg map[string]*pkg.Pkg
	// map: BaseFingerprint -> Semver version -> Fingerprint
	mapBaseFingerprintToVersions map[string]map[string]string
}

var PkgDBInstance *PkgDB

func (pkgdb *PkgDB) GetPackageByFingerprint(fp string) *pkg.Pkg {
	p, ok := pkgdb.mapFingerprintToPkg[fp]
	if !ok {
		return nil
	}
	return p
}

func (pkgdb *PkgDB) GetMapOfVersionsByBaseFingerPrint(basefp string) map[string]string {
	mapOfVersions, ok := pkgdb.mapBaseFingerprintToVersions[basefp]
	if !ok {
		// TODO what happens if there's no packages that satisfy the version range
		return map[string]string{}
	}
	return mapOfVersions
}

func (pkgdb *PkgDB) GetOrderedPackageFingerprintsThatDifferOnVersionByPackage(p *pkg.Pkg) (fps []string, weights []int) {
	mapOfVersions, ok := pkgdb.mapBaseFingerprintToVersions[p.GetBaseFingerPrint()]
	if !ok {
		// TODO what happens if there's no packages that satisfy the version range
		return fps, weights
	}
	for _, fp := range mapOfVersions {
		fps = append(fps, fp)
	}

	// Sort fps by weight
	sort.Slice(fps,
		func(i, j int) bool {
			semverI := CalculateSemverDistanceToZero(pkgdb.GetPackageByFingerprint(fps[i]).Version)
			semverJ := CalculateSemverDistanceToZero(pkgdb.GetPackageByFingerprint(fps[j]).Version)
			return semverI < semverJ
		},
	)
	// create corresponding weight slice
	for i := range fps {
		weights = append(weights, i+1)
	}

	return fps, weights
}

func CalculateSemverDistanceToZero(semversion string) (distance int) {

	sv := semver.MustParse(semversion)
	distance = (int(sv.Major())*int(math.Pow(10, 13)) + int(sv.Minor())*int(math.Pow(10, 9)) + int(sv.Patch()))
	return distance
}

func (pkgdb *PkgDB) DebugPrintDB(logger log.Logger) {
	logger.Debugf("Printing DB")
	for _, p := range pkgdb.mapFingerprintToPkg {
		logger.Debug(p.String())
	}
}

func CreatePkgDBInstance() *PkgDB {
	PkgDBInstance = &PkgDB{
		mapFingerprintToPkg:          make(map[string]*pkg.Pkg),
		mapBaseFingerprintToVersions: make(map[string]map[string]string),
	}
	return PkgDBInstance
}

func GetPkgdDBInstance() *PkgDB {
	return PkgDBInstance
}

// MergePkgs gives you a resulting package that is a copy of the new package,
// but making sure that unknown inthe known CurrenState and DesiredState is not changed,
// and only unknown info is getting filled.
func MergePkgs(old pkg.Pkg, new pkg.Pkg) (result *pkg.Pkg) {
	result = &old

	// Merge CurrentState and DesiredState
	// E.g:
	// PACKAGE                     CurrentState   DesiredState
	// old (coming from release)   installed      unknown
	// new (coming from tomodify)  unknown        removed
	// result                      installed      removed
	if old.CurrentState == pkg.Unknown {
		result.CurrentState = new.CurrentState
	}
	if old.DesiredState == pkg.Unknown {
		result.DesiredState = new.DesiredState
	}
	if old.PinnedVer == pkg.Unknown {
		result.PinnedVer = new.PinnedVer
	}

	// Merge Depends and DependsOptional slices
	if len(old.DependsRel) == 0 {
		result.DependsRel = new.DependsRel
	}
	if len(old.DependsOptionalRel) == 0 {
		result.DependsOptionalRel = new.DependsOptionalRel
	}

	return result
}

// Add adds a package to the database and returns it's ID. If a package was
// already present in the database, it makes sure to update it, in a way that
// only unknown info to that package is added.
func (pkgdb *PkgDB) Add(p *pkg.Pkg) {
	fp := p.GetFingerPrint()
	pInDB, ok := pkgdb.mapFingerprintToPkg[fp]
	if ok {
		// package already in DB, merge
		pkgdb.mapFingerprintToPkg[p.GetFingerPrint()] =
			MergePkgs(*pInDB, *p)
	} else {
		// package not there, add it
		pkgdb.mapFingerprintToPkg[fp] = p
	}

	// build map of same versions
	bfp := p.GetBaseFingerPrint()
	_, ok = pkgdb.mapBaseFingerprintToVersions[bfp]
	if !ok { // if pkg first of all packages that differ only in version
		pkgdb.mapBaseFingerprintToVersions[bfp] = make(map[string]string)
	}
	_, ok = pkgdb.mapBaseFingerprintToVersions[bfp][p.Version]
	if !ok {
		// add pkg to map of pkgs that differ only in version:
		pkgdb.mapBaseFingerprintToVersions[bfp][p.Version] = fp
	}
}
