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
	"fmt"
	"math"

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
	mapFingerprintToPkg  map[string]*pkg.Pkg
	mapFingerprintToPbID map[string]int
	mapPbIDToFingerprint map[int]string
	// map: BaseFingerprint -> Semver version -> Fingerprint
	mapBaseFingerprintToVersions map[string]map[string]string
	// struct {
	// 	semver string
	// 	semverDistToZero int // major * 10^6 + minor * 10^4 + patch
	// }
	lastElem int
	// MaxSemverDistance
}

var PkgDBInstance *PkgDB

func (pkgdb *PkgDB) GetPackageByPbID(ID int) *pkg.Pkg {
	fp, ok := pkgdb.mapPbIDToFingerprint[ID]
	if !ok {
		return nil
	}
	return pkgdb.GetPackageByFingerprint(fp)
}

func (pkgdb *PkgDB) GetPbIDByFingerprint(fp string) (id int) {
	id, ok := pkgdb.mapFingerprintToPbID[fp]
	if !ok {
		return -1
	}
	return id
}

func (pkgdb *PkgDB) GetPackageByFingerprint(fp string) *pkg.Pkg {
	p, ok := pkgdb.mapFingerprintToPkg[fp]
	if !ok {
		return nil
	}
	return p
}

func (pkgdb *PkgDB) GetIDByPackage(p *pkg.Pkg) (id int) {
	fp := p.GetFingerPrint()

	id, ok := pkgdb.mapFingerprintToPbID[fp]
	if !ok {
		return -1
	}
	return id
}

func (pkgdb *PkgDB) GetMapOfVersionsByBaseFingerPrint(basefp string) map[string]string {
	mapOfVersions, ok := pkgdb.mapBaseFingerprintToVersions[basefp]
	if !ok {
		// TODO what happens if there's no packages that satisfy the version range
		return map[string]string{}
	}
	return mapOfVersions
}

func (pkgdb *PkgDB) GetPackageIDsThatDifferOnVersionByPackage(p *pkg.Pkg) (ids []int, weights []int) {
	mapOfVersions, ok := pkgdb.mapBaseFingerprintToVersions[p.GetBaseFingerPrint()]
	if !ok {
		// TODO what happens if there's no packages that satisfy the version range
		return ids, weights
	}
	for semver, fp := range mapOfVersions {
		ids = append(ids, pkgdb.GetPbIDByFingerprint(fp))
		weights = append(weights, CalculateSemverDistanceToZero(semver))
	}
	return ids, weights
}

func CalculateSemverDistanceToZero(semversion string) (distance int) {

	sv := semver.MustParse(semversion)
	// if sv.Major() > uint64(10^4) || sv.Minor() > uint64(10^4) || sv.Patch() > uint64(10^4) {
	// 	fmt.Printf("\n%v\n", sv)
	// 	panic("Semver out of range")
	// }
	// return 10^14 - (int(sv.Major())*10 ^ 13 + int(sv.Minor())*10 ^ 9 + int(sv.Patch()))
	// distance = int(math.Pow(10, 16)) - (int(sv.Major())*int(math.Pow(10, 13)) + int(sv.Minor())*int(math.Pow(10, 9)) + int(sv.Patch()))
	distance = (int(sv.Major())*int(math.Pow(10, 13)) + int(sv.Minor())*int(math.Pow(10, 9)) + int(sv.Patch()))
	fmt.Println(distance)
	return distance
}

// // needed to find if a pkg is a dependency, to skip when i.NoSharedDeps
// func (pkgdb *PkgDB) IsDependency(fp string) bool {
// 	for _, p := range pkgdb.mapFingerprintToPkg {
// 		for _, depfp := range p.DependsRel {
// 			if fp == depfp {
// 				return true
// 			}
// 		}
// 	}
// 	return false
// }

// // needed to find if a pkg is an optional dependency, to skip when i.NoSharedDeps
// func (pkgdb *PkgDB) IsDependencyOptional(fp string) bool {
// 	for _, p := range pkgdb.mapFingerprintToPkg {
// 		for _, depfp := range p.DependsOptionalRel {
// 			if fp == depfp {
// 				return true
// 			}
// 		}
// 	}
// 	return false
// }

func CreatePkgDBInstance() *PkgDB {
	PkgDBInstance = &PkgDB{
		mapFingerprintToPkg:          make(map[string]*pkg.Pkg),
		mapFingerprintToPbID:         make(map[string]int),
		mapPbIDToFingerprint:         make(map[int]string),
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
	result = &new

	// Merge CurrentState and DesiredState
	// E.g:
	// PACKAGE                     CurrentState   DesiredState
	// old (coming from release)   installed      unknown
	// new (coming from tomodify)  unknown        removed
	// result                      installed      removed
	if result.CurrentState == pkg.Unknown {
		result.CurrentState = old.CurrentState
	}
	if result.DesiredState == pkg.Unknown {
		result.DesiredState = old.DesiredState
	}

	// Merge Depends and DependsOptional slices
	if old.DependsRel == nil {
		result.DependsRel = new.DependsRel
	}
	if old.DependsOptionalRel == nil {
		result.DependsOptionalRel = new.DependsOptionalRel
	}

	return result
}

// Add adds a package to the database and returns it's ID. If a package was
// already present in the database, it makes sure to update it, in a way that
// only unknown info to that package is added.
func (pkgdb *PkgDB) Add(p *pkg.Pkg) (ID int) {
	id := pkgdb.GetIDByPackage(p)
	if id > -1 {
		// package already there, merge
		pkgdb.mapFingerprintToPkg[p.GetFingerPrint()] =
			MergePkgs(*pkgdb.GetPackageByPbID(id), *p)

		return id
	}
	// package not there, add it
	fp := p.GetFingerPrint()
	if pkgdb.lastElem == int(^uint(0)>>1) {
		panic("Attempting to add too many packages.")
	}
	pkgdb.lastElem = pkgdb.lastElem + 1
	pkgdb.mapFingerprintToPkg[fp] = p
	pkgdb.mapFingerprintToPbID[fp] = pkgdb.lastElem
	pkgdb.mapPbIDToFingerprint[pkgdb.lastElem] = fp

	// build map of same versions
	// TODO this is broken, depending in the order of pkgs being added
	bfp := p.GetBaseFingerPrint()
	_, ok := pkgdb.mapBaseFingerprintToVersions[bfp]
	if !ok { // if pkg first of all packages that differ only in version
		pkgdb.mapBaseFingerprintToVersions[bfp] = make(map[string]string)
	}
	_, ok = pkgdb.mapBaseFingerprintToVersions[bfp][p.Version]
	if !ok {
		// add pkg to map of pkgs that differ only in version:
		pkgdb.mapBaseFingerprintToVersions[bfp][p.Version] = fp
		// TODO calculate maximum semver distance between this package and installed package?
		// if pkgdb.maxSemverDistance < distance {
		// 	pkgdb.maxSemverDistance = distance
		// }
	}

	// TODO no need to return lastElem, either nothing or fingerprint
	return pkgdb.lastElem
}

func (pkgdb *PkgDB) Size() int {
	return pkgdb.lastElem
}
