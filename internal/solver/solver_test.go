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
	"testing"

	"github.com/rancher-sandbox/hypper/internal/pkg"

	"github.com/rancher-sandbox/hypper/internal/test"
	"github.com/stretchr/testify/assert"
)

func TestSolver(t *testing.T) {

	// test cases have hardcoded IDs for now

	dep := pkg.NewPkgMock(3, "myawesomedep", "0.1.100", "myawesomedependencytargetns", nil, nil, pkg.Unknown, pkg.Unknown)
	deps := []*pkg.Pkg{dep}

	// repo
	repo := []*pkg.Pkg{
		pkg.NewPkgMock(1, "notinstalledbar", "1.0.0", "notinstalledtargetns", nil, nil, pkg.Unknown, pkg.Unknown),
		pkg.NewPkgMock(2, "notinstalledbar", "2.0.0", "notinstalledtargetns", nil, nil, pkg.Unknown, pkg.Unknown),
		dep,
		pkg.NewPkgMock(4, "wantedbaz", "1.0.0", "wantedbazns", deps, nil, pkg.Unknown, pkg.Unknown),
	}

	// releases
	releases := []*pkg.Pkg{
		pkg.NewPkgMock(5, "installedfoo", "1.0.0", "installedns", nil, nil, pkg.Present, pkg.Unknown),
	}

	// list of packages to modify
	toModify := []*pkg.Pkg{
		pkg.NewPkgMock(4, "wantedbaz", "1.0.0", "wantedbazns", deps, nil, pkg.Unknown, pkg.Present),
	}

	pkgsMain := append([]*pkg.Pkg{}, repo...)
	pkgsMain = append(pkgsMain, releases...)
	pkgsMain = append(pkgsMain, toModify...)

	/////////////////////////////////////////////////////////////////////////////

	loopPkgs := []*pkg.Pkg{
		pkg.NewPkgMock(1, "wantedfoo", "1.0.0", "targetns", nil, nil, pkg.Absent, pkg.Present),
		pkg.NewPkgMock(2, "wantedbar", "1.0.0", "targetns", nil, nil, pkg.Absent, pkg.Unknown),
		pkg.NewPkgMock(3, "wantedbaz", "1.0.0", "targetns", nil, nil, pkg.Absent, pkg.Unknown),
	}
	loopPkgs[0].Depends = []*pkg.Pkg{loopPkgs[1]}
	loopPkgs[1].Depends = []*pkg.Pkg{loopPkgs[2]}
	loopPkgs[2].Depends = []*pkg.Pkg{loopPkgs[0]}
	for _, p := range loopPkgs {
		pkg.UpdatePkgRel(p)
	}

	/////////////////////////////////////////////////////////////////////////////

	unsatDep := pkg.NewPkgMock(1, "myawesomedep", "0.1.100", "myawesomedependencytargetns", nil, nil, pkg.Present, pkg.Absent)
	unsatDeps := []*pkg.Pkg{unsatDep}

	unsatDependencies := []*pkg.Pkg{
		pkg.NewPkgMock(1, "wantedbaz", "1.0.0", "wantedbazns", unsatDeps, nil, pkg.Unknown, pkg.Present),
	}

	/////////////////////////////////////////////////////////////////////////////

	for _, tcase := range []struct {
		name         string
		pkgs         []*pkg.Pkg
		golden       string
		resultStatus string
	}{
		{
			name:         "empty world",
			golden:       "output/solve-empty.txt",
			pkgs:         []*pkg.Pkg{},
			resultStatus: "SAT",
		},
		{
			name:         "solve for the example in main",
			golden:       "output/solve-main.txt",
			pkgs:         pkgsMain,
			resultStatus: "SAT",
		},
		{
			name:         "unsatisfiable, remove a dependency",
			golden:       "output/solve-unsat-remove-dep.txt",
			pkgs:         unsatDependencies,
			resultStatus: "UNSAT",
		},
		{
			name:   "unsatisfiable, missing deps",
			golden: "output/solve-unsat-missing-deps.txt",
			pkgs: []*pkg.Pkg{
				pkg.NewPkgMock(1, "wantedbar", "1.0.0", "targetns", deps, nil, pkg.Absent, pkg.Present),
			},
			resultStatus: "UNSAT",
		},
		{
			name:         "install several looped deps",
			golden:       "output/solve-sat-loop-deps.txt",
			pkgs:         loopPkgs,
			resultStatus: "SAT",
		},
	} {
		t.Run(tcase.name, func(t *testing.T) {
			s := New()
			s.BuildWorldMock(tcase.pkgs)
			s.Solve()
			is := assert.New(t)
			is.Equal(s.pkgResultSet.Status, tcase.resultStatus)

			str := s.FormatOutput(YAML)
			if tcase.golden != "" {
				test.AssertGoldenString(t, str, tcase.golden)
			}
		})
	}
}

// TestBuildWorld {
//}

func TestFormatOutput(t *testing.T) {

	for _, tcase := range []struct {
		name        string
		pkgs        []*pkg.Pkg
		goldenYaml  string
		goldenJson  string
		goldenTable string
	}{
		{
			name:        "empty world",
			goldenYaml:  "output/format-empty-yaml.txt",
			goldenJson:  "output/format-empty-json.txt",
			goldenTable: "output/format-empty-table.txt",
			pkgs:        []*pkg.Pkg{},
		},
		{
			name:        "unsatisfiable, remove and install at the same time",
			goldenYaml:  "output/format-unsat-yaml.txt",
			goldenJson:  "output/format-unsat-json.txt",
			goldenTable: "output/format-unsat-table.txt",
			pkgs: []*pkg.Pkg{
				pkg.NewPkgMock(1, "bar", "1.0.0", "targetns", nil, nil, pkg.Present, pkg.Absent),
				pkg.NewPkgMock(2, "bar", "1.0.0", "targetns", nil, nil, pkg.Absent, pkg.Present),
			},
		},
	} {
		s := New()
		s.BuildWorldMock(tcase.pkgs)
		s.Solve()

		test.AssertGoldenString(t, s.FormatOutput(YAML), tcase.goldenYaml)
		test.AssertGoldenString(t, s.FormatOutput(JSON), tcase.goldenJson)
		test.AssertGoldenString(t, s.FormatOutput(Table), tcase.goldenTable)
	}
}
