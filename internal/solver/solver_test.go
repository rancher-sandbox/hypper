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
			name:   "solve for the example in main",
			golden: "output/solve-main.txt",
			pkgs: []*pkg.Pkg{
				pkg.NewPkgMock(1, "notinstalledbar", "1.0.0", "notinstalledtargetns", nil, nil, pkg.Unknown, pkg.Unknown),
				pkg.NewPkgMock(2, "notinstalledbar", "2.0.0", "notinstalledtargetns", nil, nil, pkg.Unknown, pkg.Unknown),
				pkg.NewPkgMock(3, "myawesomedep", "0.1.100", "myawesomedeptargetns", nil, nil, pkg.Unknown, pkg.Unknown),
				// toModify:
				pkg.NewPkgMock(4, "wantedbaz", "1.0.0", "wantedbazns",
					[]*pkg.PkgRel{
						{
							TargetID:  3,
							Name:      "myawesomedep",
							Version:   "0.1.100",
							Namespace: "myawesomedeptargetns",
						},
					},
					nil, pkg.Unknown, pkg.Present),
				// releases:
				pkg.NewPkgMock(5, "installedfoo", "1.0.0", "installedns", nil, nil, pkg.Present, pkg.Unknown),
			},
			resultStatus: "SAT",
		},
		{
			name:   "unsatisfiable, remove a dependency",
			golden: "output/solve-unsat-remove-dep.txt",
			pkgs: []*pkg.Pkg{
				// release, to be removed:
				pkg.NewPkgMock(1, "myawesomedep", "0.1.100", "myawesomedependencytargetns", nil, nil, pkg.Present, pkg.Absent),
				// release, depends on pkg that is going to be removed:
				pkg.NewPkgMock(2, "wantedbaz", "1.0.0", "wantedbazns",
					[]*pkg.PkgRel{ // dependencies:
						{
							TargetID:  1,
							Name:      "myawesomedep",
							Version:   "0.1.100",
							Namespace: "myawesomedeptargetns",
						},
					},
					nil, pkg.Unknown, pkg.Present),
			},
			resultStatus: "UNSAT",
		},
		{
			name:   "install several looped deps",
			golden: "output/solve-sat-loop-deps.txt",
			pkgs: []*pkg.Pkg{
				// package 1, depends on 2:
				pkg.NewPkgMock(1, "wantedfoo", "1.0.0", "targetns",
					[]*pkg.PkgRel{
						{
							TargetID:  2,
							Name:      "wantedbar",
							Version:   "1.0.0",
							Namespace: "targetns",
						},
					},
					nil, pkg.Absent, pkg.Present),
				// package 2, depends on 3:
				pkg.NewPkgMock(2, "wantedbar", "1.0.0", "targetns",
					[]*pkg.PkgRel{
						{
							TargetID:  3,
							Name:      "wantedbaz",
							Version:   "1.0.0",
							Namespace: "targetns",
						},
					},
					nil, pkg.Absent, pkg.Unknown),
				// package 1, depends on 1:
				pkg.NewPkgMock(3, "wantedbaz", "1.0.0", "targetns",
					[]*pkg.PkgRel{
						{
							TargetID:  1,
							Name:      "wantedfoo",
							Version:   "1.0.0",
							Namespace: "targetns",
						},
					},
					nil, pkg.Absent, pkg.Unknown),
			},
			resultStatus: "SAT",
		},
		{
			name:   "remove package",
			golden: "output/solve-sat-remove-package.txt",
			pkgs: []*pkg.Pkg{
				pkg.NewPkgMock(1, "wantedbaz", "1.0.0", "wantedbazns", nil, nil, pkg.Present, pkg.Absent),
			},
			resultStatus: "SAT",
		},
		{
			name:   "update package",
			golden: "output/solve-sat-update-package.txt",
			pkgs: []*pkg.Pkg{
				// releases:
				pkg.NewPkgMock(1, "toupdatebar", "1.0.0", "toupdatebarns", nil, nil, pkg.Present, pkg.Unknown),
				pkg.NewPkgMock(2, "installedfoo", "1.0.0", "installedns", nil, nil, pkg.Present, pkg.Unknown),
				// package to update:
				pkg.NewPkgMock(3, "toupdatebar", "1.3.0", "toupdatebarns", nil, nil, pkg.Unknown, pkg.Present),
			},
			resultStatus: "SAT",
		},
	} {
		t.Run(tcase.name, func(t *testing.T) {
			s := New()
			s.BuildWorldMock(tcase.pkgs)
			s.Solve()
			is := assert.New(t)
			is.Equal(tcase.resultStatus, s.pkgResultSet.Status)

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
				pkg.NewPkgMock(1, "bar", "1.0.0", "targetns", nil, nil, pkg.Present, pkg.Present),
				pkg.NewPkgMock(2, "baz", "1.0.0", "targetns", nil, nil, pkg.Absent, pkg.Present),
				pkg.NewPkgMock(3, "foo", "1.0.0", "targetns", nil, nil, pkg.Present, pkg.Absent),
			},
		},
	} {
		s := New()
		s.BuildWorldMock(tcase.pkgs)
		s.Solve()
		is := assert.New(t)
		is.Equal("SAT", s.pkgResultSet.Status)

		test.AssertGoldenString(t, s.FormatOutput(YAML), tcase.goldenYaml)
		test.AssertGoldenString(t, s.FormatOutput(JSON), tcase.goldenJson)
		test.AssertGoldenString(t, s.FormatOutput(Table), tcase.goldenTable)
	}
}
