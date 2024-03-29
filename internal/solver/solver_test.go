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
	"bytes"
	"testing"

	"github.com/Masterminds/log-go"
	logcli "github.com/Masterminds/log-go/impl/cli"
	pkg "github.com/rancher-sandbox/hypper/internal/package"

	"github.com/rancher-sandbox/hypper/internal/test"
	"github.com/stretchr/testify/assert"
)

func TestInstall(t *testing.T) {

	for _, tcase := range []struct {
		name         string
		wantedPkg    *pkg.Pkg
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
			wantedPkg: pkg.NewPkgMock("wantedbaz", "1.0.0", "wantedbazns",
				[]*pkg.PkgRel{{
					ReleaseName: "myawesomedep",
					Namespace:   "myawesomedeptargetns",
					SemverRange: "~0.1.0",
					ChartName:   "myawesomedep",
				}},
				nil, pkg.Unknown, pkg.Present),
			pkgs: []*pkg.Pkg{
				pkg.NewPkgMock("notinstalledbar", "1.0.0", "notinstalledtargetns", nil, nil, pkg.Unknown, pkg.Unknown),
				pkg.NewPkgMock("notinstalledbar", "2.0.0", "notinstalledtargetns", nil, nil, pkg.Unknown, pkg.Unknown),
				pkg.NewPkgMock("myawesomedep", "0.1.100", "myawesomedeptargetns", nil, nil, pkg.Unknown, pkg.Unknown),
				// toModify:
				pkg.NewPkgMock("wantedbaz", "1.0.0", "wantedbazns",
					[]*pkg.PkgRel{{
						ReleaseName: "myawesomedep",
						Namespace:   "myawesomedeptargetns",
						SemverRange: "~0.1.0",
						ChartName:   "myawesomedep",
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
			wantedPkg: pkg.NewPkgMock("wantedbaz", "1.0.0", "wantedbazns",
				[]*pkg.PkgRel{{
					ReleaseName: "myawesomedep",
					Namespace:   "myawesomedeptargetns",
					SemverRange: "0.1.103",
					ChartName:   "myawesomedep",
				}},
				nil, pkg.Unknown, pkg.Present),
			pkgs: []*pkg.Pkg{
				// dependency that doesn't match semver range:
				pkg.NewPkgMock("myawesomedep", "2.1.100", "myawesomedeptargetns", nil, nil, pkg.Unknown, pkg.Unknown),
				pkg.NewPkgMock("myawesomedep", "1.0.0", "myawesomedeptargetns", nil, nil, pkg.Unknown, pkg.Unknown),
				// dependency we want pulled:
				pkg.NewPkgMock("myawesomedep", "0.1.103", "myawesomedeptargetns", nil, nil, pkg.Unknown, pkg.Unknown),
				// toModify:
				pkg.NewPkgMock("wantedbaz", "1.0.0", "wantedbazns",
					[]*pkg.PkgRel{{
						ReleaseName: "myawesomedep",
						Namespace:   "myawesomedeptargetns",
						SemverRange: "0.1.103",
						ChartName:   "myawesomedep",
					}},
					nil, pkg.Unknown, pkg.Present),
			},
			resultStatus: "SAT",
		},
		{
			name:   "install a pkg and dep, finding minor version",
			golden: "output/solve-sat-dependecy-minor.txt",
			wantedPkg: pkg.NewPkgMock("wantedbaz", "1.0.0", "wantedbazns",
				[]*pkg.PkgRel{{
					ReleaseName: "myawesomedep",
					Namespace:   "myawesomedeptargetns",
					SemverRange: "~0.1.0",
					ChartName:   "myawesomedep",
				}},
				nil, pkg.Unknown, pkg.Present),
			pkgs: []*pkg.Pkg{
				// dependency that doesn't match semver range:
				pkg.NewPkgMock("myawesomedep", "2.1.100", "myawesomedeptargetns", nil, nil, pkg.Unknown, pkg.Unknown),
				// dependency we want pulled:
				pkg.NewPkgMock("myawesomedep", "0.1.100", "myawesomedeptargetns", nil, nil, pkg.Unknown, pkg.Unknown),
				// toModify:
				pkg.NewPkgMock("wantedbaz", "1.0.0", "wantedbazns",
					[]*pkg.PkgRel{{
						ReleaseName: "myawesomedep",
						Namespace:   "myawesomedeptargetns",
						SemverRange: "~0.1.0",
						ChartName:   "myawesomedep",
					}},
					nil, pkg.Unknown, pkg.Present),
			},
			resultStatus: "SAT",
		},
		{
			name:   "install a pkg and dep, finding major version",
			golden: "output/solve-sat-dependecy-major.txt",
			wantedPkg: pkg.NewPkgMock("wantedbaz", "1.0.0", "wantedbazns",
				[]*pkg.PkgRel{{
					ReleaseName: "myawesomedep",
					Namespace:   "myawesomedeptargetns",
					SemverRange: "^1.2.0",
					ChartName:   "myawesomedep",
				}},
				nil, pkg.Unknown, pkg.Present),
			pkgs: []*pkg.Pkg{
				// dependency that don't match semver range:
				pkg.NewPkgMock("myawesomedep", "2.0.0", "myawesomedeptargetns", nil, nil, pkg.Unknown, pkg.Unknown),
				pkg.NewPkgMock("myawesomedep", "0.1.100", "myawesomedeptargetns", nil, nil, pkg.Unknown, pkg.Unknown),
				// dependency we want pulled:
				pkg.NewPkgMock("myawesomedep", "1.9.0", "myawesomedeptargetns", nil, nil, pkg.Unknown, pkg.Unknown),
				// toModify:
				pkg.NewPkgMock("wantedbaz", "1.0.0", "wantedbazns",
					[]*pkg.PkgRel{{
						ReleaseName: "myawesomedep",
						Namespace:   "myawesomedeptargetns",
						SemverRange: "^1.2.0",
						ChartName:   "myawesomedep",
					}},
					nil, pkg.Unknown, pkg.Present),
			},
			resultStatus: "SAT",
		},
		{
			name:   "install a pkg and dep, being no matching version for the dep",
			golden: "output/solve-unsat-dependecy-no-version.txt",
			wantedPkg: pkg.NewPkgMock("wantedbaz", "1.0.0", "wantedbazns",
				[]*pkg.PkgRel{{
					ReleaseName: "myawesomedep",
					Namespace:   "myawesomedeptargetns",
					SemverRange: "^1.0.0",
					ChartName:   "myawesomedep",
				}},
				nil, pkg.Unknown, pkg.Present),
			pkgs: []*pkg.Pkg{
				// no dependency satisfies the constraint:
				pkg.NewPkgMock("myawesomedep", "3.0.0", "myawesomedeptargetns", nil, nil, pkg.Unknown, pkg.Unknown),
				// toModify:
				pkg.NewPkgMock("wantedbaz", "1.0.0", "wantedbazns",
					[]*pkg.PkgRel{{
						ReleaseName: "myawesomedep",
						Namespace:   "myawesomedeptargetns",
						SemverRange: "^1.0.0",
						ChartName:   "myawesomedep",
					}},
					nil, pkg.Unknown, pkg.Present),
			},
			resultStatus: "UNSAT",
		},
		{
			name:   "install a pkg and dep, dependency not in db",
			golden: "output/solve-unsat-dependecy-not-known.txt",
			wantedPkg: pkg.NewPkgMock("wantedbaz", "1.0.0", "wantedbazns",
				[]*pkg.PkgRel{{
					ReleaseName: "myawesomedep",
					Namespace:   "myawesomedeptargetns",
					SemverRange: "^1.0.0",
					ChartName:   "myawesomedep",
				}},
				nil, pkg.Unknown, pkg.Present),
			pkgs: []*pkg.Pkg{
				// dependency not in database (not in repos, for example)
				// toModify:
				pkg.NewPkgMock("wantedbaz", "1.0.0", "wantedbazns",
					[]*pkg.PkgRel{{
						ReleaseName: "myawesomedep",
						Namespace:   "myawesomedeptargetns",
						SemverRange: "^1.0.0",
						ChartName:   "myawesomedep",
					}},
					nil, pkg.Unknown, pkg.Present),
			},
			resultStatus: "UNSAT",
		},
		{
			name:      "unsatisfiable, remove a dependency",
			golden:    "output/solve-unsat-remove-dep.txt",
			wantedPkg: pkg.NewPkgMock("myawesomedep", "0.1.100", "myawesomedeptargetns", nil, nil, pkg.Present, pkg.Absent),
			pkgs: []*pkg.Pkg{
				// release, to be removed:
				pkg.NewPkgMock("myawesomedep", "0.1.100", "myawesomedeptargetns", nil, nil, pkg.Present, pkg.Absent),
				// release, depends on pkg that is going to be removed:
				pkg.NewPkgMock("wantedbaz", "1.0.0", "wantedbazns",
					[]*pkg.PkgRel{{
						ReleaseName: "myawesomedep",
						Namespace:   "myawesomedeptargetns",
						SemverRange: "~0.1.0",
						ChartName:   "myawesomedep",
					}},
					nil, pkg.Unknown, pkg.Present),
			},
			resultStatus: "UNSAT",
		},
		{
			name:   "install several looped deps",
			golden: "output/solve-sat-loop-deps.txt",
			wantedPkg: pkg.NewPkgMock("wantedfoo", "1.0.0", "targetns",
				[]*pkg.PkgRel{{
					ReleaseName: "wantedbar",
					Namespace:   "targetns",
					SemverRange: "^1.0.0",
					ChartName:   "wantedbar",
				}},
				nil, pkg.Absent, pkg.Present),
			pkgs: []*pkg.Pkg{
				// package 1, depends on 2:
				pkg.NewPkgMock("wantedfoo", "1.0.0", "targetns",
					[]*pkg.PkgRel{{
						ReleaseName: "wantedbar",
						Namespace:   "targetns",
						SemverRange: "^1.0.0",
						ChartName:   "wantedbar",
					}},
					nil, pkg.Absent, pkg.Present),
				// package 2, depends on 3:
				pkg.NewPkgMock("wantedbar", "1.0.0", "targetns",
					[]*pkg.PkgRel{{
						ReleaseName: "wantedbaz",
						Namespace:   "targetns",
						SemverRange: "^1.0.0",
						ChartName:   "wantedbaz",
					}},
					nil, pkg.Absent, pkg.Unknown),
				// package 1, depends on 1:
				pkg.NewPkgMock("wantedbaz", "1.0.0", "targetns",
					[]*pkg.PkgRel{{
						ReleaseName: "wantedfoo",
						Namespace:   "targetns",
						SemverRange: "^1.0.0",
						ChartName:   "wantedfoo",
					}},
					nil, pkg.Absent, pkg.Unknown),
			},
			resultStatus: "SAT",
		},
	} {
		t.Run(tcase.name, func(t *testing.T) {

			// create our own Logger that satisfies impl/cli.Logger, but with a buffer for tests
			buf := new(bytes.Buffer)
			logger := logcli.NewStandard()
			logger.InfoOut = buf
			logger.WarnOut = buf
			logger.ErrorOut = buf
			logger.DebugOut = buf
			log.Current = logger

			s := New(InstallOne, logger)
			s.BuildWorldMock(tcase.pkgs)
			s.Solve(tcase.wantedPkg)
			is := assert.New(t)
			is.Equal(tcase.resultStatus, s.PkgResultSet.Status)

			str := s.FormatOutput(YAML)
			if tcase.golden != "" {
				test.AssertGoldenString(t, str, tcase.golden)
			}
		})
	}
}

func TestFormatOutput(t *testing.T) {

	for _, tcase := range []struct {
		name        string
		wantedPkg   *pkg.Pkg
		pkgs        []*pkg.Pkg
		goldenYaml  string
		goldenJson  string
		goldenTable string
		status      string
	}{
		{
			name:        "empty world",
			goldenYaml:  "output/format-empty-yaml.txt",
			goldenJson:  "output/format-empty-json.txt",
			goldenTable: "output/format-empty-table.txt",
			wantedPkg:   nil,
			pkgs:        []*pkg.Pkg{},
			status:      "SAT",
		},
		{
			name:        "satisfiable, install 1",
			goldenYaml:  "output/format-sat-install1-yaml.txt",
			goldenJson:  "output/format-sat-install1-json.txt",
			goldenTable: "output/format-sat-install1-table.txt",
			wantedPkg:   pkg.NewPkgMock("bar", "1.0.0", "targetns", nil, nil, pkg.Absent, pkg.Present),
			pkgs: []*pkg.Pkg{
				pkg.NewPkgMock("bar", "1.0.0", "targetns", nil, nil, pkg.Absent, pkg.Present),
			},
			status: "SAT",
		},
		{
			name:        "satisfiable, upgrade a release",
			goldenYaml:  "output/format-sat-upgrade-yaml.txt",
			goldenJson:  "output/format-sat-upgrade-json.txt",
			goldenTable: "output/format-sat-upgrade-table.txt",
			wantedPkg:   pkg.NewPkgMock("bar", "1.0.0", "targetns", nil, nil, pkg.Present, pkg.Present),
			pkgs: []*pkg.Pkg{
				pkg.NewPkgMock("bar", "1.0.0", "targetns", nil, nil, pkg.Present, pkg.Present),
			},
			status: "SAT",
		},
		{
			name:        "unsatisfiable, nothing provides dep",
			goldenYaml:  "output/format-unsat-nothing-dep-yaml.txt",
			goldenJson:  "output/format-unsat-nothing-dep-json.txt",
			goldenTable: "output/format-unsat-nothing-dep-table.txt",
			wantedPkg: pkg.NewPkgMock("wantedbaz", "1.0.0", "targetns",
				[]*pkg.PkgRel{{
					ReleaseName: "depfoo",
					Namespace:   "targetns",
					SemverRange: "^1.0.0",
					ChartName:   "depfoo",
				}},
				nil, pkg.Absent, pkg.Present),
			pkgs: []*pkg.Pkg{
				pkg.NewPkgMock("wantedbaz", "1.0.0", "targetns",
					[]*pkg.PkgRel{{
						ReleaseName: "depfoo",
						Namespace:   "targetns",
						SemverRange: "^1.0.0",
						ChartName:   "depfoo",
					}},
					nil, pkg.Absent, pkg.Present),
			},
			status: "UNSAT",
		},
	} {
		// create our own Logger that satisfies impl/cli.Logger, but with a buffer for tests
		buf := new(bytes.Buffer)
		logger := logcli.NewStandard()
		logger.InfoOut = buf
		logger.WarnOut = buf
		logger.ErrorOut = buf
		logger.DebugOut = buf
		log.Current = logger

		s := New(InstallOne, logger)
		s.BuildWorldMock(tcase.pkgs)
		s.Solve(tcase.wantedPkg)
		is := assert.New(t)
		is.Equal(tcase.status, s.PkgResultSet.Status)

		test.AssertGoldenString(t, s.FormatOutput(YAML), tcase.goldenYaml)
		test.AssertGoldenString(t, s.FormatOutput(JSON), tcase.goldenJson)
		test.AssertGoldenString(t, s.FormatOutput(Table), tcase.goldenTable)
	}
}
