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

	pkg "github.com/rancher-sandbox/hypper/internal/package"

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
				pkg.NewPkgMock("notinstalledbar", "1.0.0", "notinstalledtargetns", nil, nil, pkg.Unknown, pkg.Unknown),
				pkg.NewPkgMock("notinstalledbar", "2.0.0", "notinstalledtargetns", nil, nil, pkg.Unknown, pkg.Unknown),
				pkg.NewPkgMock("myawesomedep", "0.1.100", "myawesomedeptargetns", nil, nil, pkg.Unknown, pkg.Unknown),
				// toModify:
				pkg.NewPkgMock("wantedbaz", "1.0.0", "wantedbazns",
					[]*pkg.PkgRel{{
						BaseFingerprint: pkg.CreateBaseFingerPrint("myawesomedep", "myawesomedeptargetns"),
						SemverRange:     "~0.1.0",
					}},
					nil, pkg.Unknown, pkg.Present),
				// releases:
				pkg.NewPkgMock("installedfoo", "1.0.0", "installedns", nil, nil, pkg.Present, pkg.Unknown),
			},
			resultStatus: "SAT",
		},
		{
			name:   "install a pkg and dep, finding specific version",
			golden: "output/solve-sat-dependecy-specific-ver.txt",
			pkgs: []*pkg.Pkg{
				// dependency that doesn't match semver range:
				pkg.NewPkgMock("myawesomedep", "2.1.100", "myawesomedeptargetns", nil, nil, pkg.Unknown, pkg.Unknown),
				pkg.NewPkgMock("myawesomedep", "1.0.0", "myawesomedeptargetns", nil, nil, pkg.Unknown, pkg.Unknown),
				// dependency we want pulled:
				pkg.NewPkgMock("myawesomedep", "0.1.103", "myawesomedeptargetns", nil, nil, pkg.Unknown, pkg.Unknown),
				// toModify:
				pkg.NewPkgMock("wantedbaz", "1.0.0", "wantedbazns",
					[]*pkg.PkgRel{{
						BaseFingerprint: pkg.CreateBaseFingerPrint("myawesomedep", "myawesomedeptargetns"),
						SemverRange:     "0.1.103",
					}},
					nil, pkg.Unknown, pkg.Present),
			},
			resultStatus: "SAT",
		},
		{
			name:   "install a pkg and dep, finding minor version",
			golden: "output/solve-sat-dependecy-minor.txt",
			pkgs: []*pkg.Pkg{
				// dependency that doesn't match semver range:
				pkg.NewPkgMock("myawesomedep", "2.1.100", "myawesomedeptargetns", nil, nil, pkg.Unknown, pkg.Unknown),
				// dependency we want pulled:
				pkg.NewPkgMock("myawesomedep", "0.1.100", "myawesomedeptargetns", nil, nil, pkg.Unknown, pkg.Unknown),
				// toModify:
				pkg.NewPkgMock("wantedbaz", "1.0.0", "wantedbazns",
					[]*pkg.PkgRel{{
						BaseFingerprint: pkg.CreateBaseFingerPrint("myawesomedep", "myawesomedeptargetns"),
						SemverRange:     "~0.1.0",
					}},
					nil, pkg.Unknown, pkg.Present),
			},
			resultStatus: "SAT",
		},
		{
			name:   "install a pkg and dep, finding major version",
			golden: "output/solve-sat-dependecy-major.txt",
			pkgs: []*pkg.Pkg{
				// dependency that don't match semver range:
				pkg.NewPkgMock("myawesomedep", "2.0.0", "myawesomedeptargetns", nil, nil, pkg.Unknown, pkg.Unknown),
				pkg.NewPkgMock("myawesomedep", "0.1.100", "myawesomedeptargetns", nil, nil, pkg.Unknown, pkg.Unknown),
				// dependency we want pulled:
				pkg.NewPkgMock("myawesomedep", "1.9.0", "myawesomedeptargetns", nil, nil, pkg.Unknown, pkg.Unknown),
				// toModify:
				pkg.NewPkgMock("wantedbaz", "1.0.0", "wantedbazns",
					[]*pkg.PkgRel{{
						BaseFingerprint: pkg.CreateBaseFingerPrint("myawesomedep", "myawesomedeptargetns"),
						SemverRange:     "^1.2.0",
					}},
					nil, pkg.Unknown, pkg.Present),
			},
			resultStatus: "SAT",
		},
		{
			name:   "install a pkg and dep, being no matching version for the dep",
			golden: "output/solve-unsat-dependecy-no-version.txt",
			pkgs: []*pkg.Pkg{
				// no dependency satisfies the constraint:
				pkg.NewPkgMock("myawesomedep", "3.0.0", "myawesomedeptargetns", nil, nil, pkg.Unknown, pkg.Unknown),
				// toModify:
				pkg.NewPkgMock("wantedbaz", "1.0.0", "wantedbazns",
					[]*pkg.PkgRel{{
						BaseFingerprint: pkg.CreateBaseFingerPrint("myawesomedep", "myawesomedeptargetns"),
						SemverRange:     "^1.0.0",
					}},
					nil, pkg.Unknown, pkg.Present),
			},
			resultStatus: "UNSAT",
		},
		{
			name:   "install a pkg and dep, dependency not in db",
			golden: "output/solve-unsat-dependecy-not-known.txt",
			pkgs: []*pkg.Pkg{
				// dependency not in database (not in repos, for example)
				// toModify:
				pkg.NewPkgMock("wantedbaz", "1.0.0", "wantedbazns",
					[]*pkg.PkgRel{{
						BaseFingerprint: pkg.CreateBaseFingerPrint("myawesomedep", "myawesomedeptargetns"),
						SemverRange:     "^1.0.0",
					}},
					nil, pkg.Unknown, pkg.Present),
			},
			resultStatus: "UNSAT",
		},
		{
			name:   "unsatisfiable, remove a dependency",
			golden: "output/solve-unsat-remove-dep.txt",
			pkgs: []*pkg.Pkg{
				// release, to be removed:
				pkg.NewPkgMock("myawesomedep", "0.1.100", "myawesomedeptargetns", nil, nil, pkg.Present, pkg.Absent),
				// release, depends on pkg that is going to be removed:
				pkg.NewPkgMock("wantedbaz", "1.0.0", "wantedbazns",
					[]*pkg.PkgRel{{
						BaseFingerprint: pkg.CreateBaseFingerPrint("myawesomedep", "myawesomedeptargetns"),
						SemverRange:     "~0.1.0",
					}},
					nil, pkg.Unknown, pkg.Present),
			},
			resultStatus: "UNSAT",
		},
		{
			name:   "install several looped deps",
			golden: "output/solve-sat-loop-deps.txt",
			pkgs: []*pkg.Pkg{
				// package 1, depends on 2:
				pkg.NewPkgMock("wantedfoo", "1.0.0", "targetns",
					[]*pkg.PkgRel{{
						BaseFingerprint: pkg.CreateBaseFingerPrint("wantedbar", "targetns"),
						SemverRange:     "^1.0.0",
					}},
					nil, pkg.Absent, pkg.Present),
				// package 2, depends on 3:
				pkg.NewPkgMock("wantedbar", "1.0.0", "targetns",
					[]*pkg.PkgRel{{
						BaseFingerprint: pkg.CreateBaseFingerPrint("wantedbaz", "targetns"),
						SemverRange:     "^1.0.0",
					}},
					nil, pkg.Absent, pkg.Unknown),
				// package 1, depends on 1:
				pkg.NewPkgMock("wantedbaz", "1.0.0", "targetns",
					[]*pkg.PkgRel{{
						BaseFingerprint: pkg.CreateBaseFingerPrint("wantedfoo", "targetns"),
						SemverRange:     "^1.0.0",
					}},
					nil, pkg.Absent, pkg.Unknown),
			},
			resultStatus: "SAT",
		},
		{
			name:   "remove package",
			golden: "output/solve-sat-remove-package.txt",
			pkgs: []*pkg.Pkg{
				pkg.NewPkgMock("wantedbaz", "1.0.0", "wantedbazns", nil, nil, pkg.Present, pkg.Absent),
			},
			resultStatus: "SAT",
		},
		// {
		// 	name:   "update package",
		// 	golden: "output/solve-sat-update-package.txt",
		// 	pkgs: []*pkg.Pkg{
		// 		// releases:
		// 		pkg.NewPkgMock("toupdatebar", "1.0.0", "toupdatebarns", nil, nil, pkg.Present, pkg.Unknown),
		// 		pkg.NewPkgMock("installedfoo", "1.0.0", "installedns", nil, nil, pkg.Present, pkg.Unknown),
		// 		// package to update:
		// 		pkg.NewPkgMock("toupdatebar", "1.3.0", "toupdatebarns", nil, nil, pkg.Unknown, pkg.Present),
		// 	},
		// 	resultStatus: "SAT",
		// },
	} {
		t.Run(tcase.name, func(t *testing.T) {
			s := New()
			s.BuildWorldMock(tcase.pkgs)
			s.Solve()
			is := assert.New(t)
			is.Equal(tcase.resultStatus, s.PkgResultSet.Status)

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
			name:        "satisfiable, install 1",
			goldenYaml:  "output/format-sat-install1-yaml.txt",
			goldenJson:  "output/format-sat-install1-json.txt",
			goldenTable: "output/format-sat-install1-table.txt",
			pkgs: []*pkg.Pkg{
				pkg.NewPkgMock("bar", "1.0.0", "targetns", nil, nil, pkg.Absent, pkg.Present),
			},
		},
		{
			name:        "unsatisfiable, remove and install at the same time",
			goldenYaml:  "output/format-unsat-upgrade-yaml.txt",
			goldenJson:  "output/format-unsat-upgrade-json.txt",
			goldenTable: "output/format-unsat-upgrade-table.txt",
			pkgs: []*pkg.Pkg{
				pkg.NewPkgMock("bar", "1.0.0", "targetns", nil, nil, pkg.Present, pkg.Present),
			},
		},
		{
			name:        "unsatisfiable, nothing provides dep",
			goldenYaml:  "output/format-unsat-nothing-dep-yaml.txt",
			goldenJson:  "output/format-unsat-nothing-dep-json.txt",
			goldenTable: "output/format-unsat-nothing-dep-table.txt",
			pkgs: []*pkg.Pkg{
				pkg.NewPkgMock("wantedbaz", "1.0.0", "targetns",
					[]*pkg.PkgRel{{
						BaseFingerprint: pkg.CreateBaseFingerPrint("depfoo", "targetns"),
						SemverRange:     "^1.0.0",
					}},
					nil, pkg.Absent, pkg.Unknown),
			},
		},
	} {
		s := New()
		s.BuildWorldMock(tcase.pkgs)
		s.Solve()
		is := assert.New(t)
		is.Equal("SAT", s.PkgResultSet.Status)

		test.AssertGoldenString(t, s.FormatOutput(YAML), tcase.goldenYaml)
		test.AssertGoldenString(t, s.FormatOutput(JSON), tcase.goldenJson)
		test.AssertGoldenString(t, s.FormatOutput(Table), tcase.goldenTable)
	}
}
