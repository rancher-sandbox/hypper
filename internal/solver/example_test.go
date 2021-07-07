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
	"bytes"
	"fmt"

	"github.com/Masterminds/log-go"
	logcli "github.com/Masterminds/log-go/impl/cli"

	pkg "github.com/rancher-sandbox/hypper/internal/package"
)

func ExampleSolver() {

	// Create a slice of mock packages:
	pkgs := []*pkg.Pkg{
		pkg.NewPkgMock("notinstalledbar", "1.0.0", "notinstalledtargetns", nil, nil, pkg.Unknown, pkg.Unknown),
		pkg.NewPkgMock("notinstalledbar", "2.0.0", "notinstalledtargetns", nil, nil, pkg.Unknown, pkg.Unknown),
		pkg.NewPkgMock("myawesomedep", "0.1.100", "myawesomedeptargetns", nil, nil, pkg.Unknown, pkg.Unknown),
		// package to modify (install, in this case, see pkg.DesiredState set to Present):
		pkg.NewPkgMock("wantedbaz", "1.0.0", "wantedbazns",
			// dependency relations of wantedbaz:
			[]*pkg.PkgRel{{
				ReleaseName: "myawesomedep",
				Namespace:   "myawesomedeptargetns",
				SemverRange: "~0.1.0",
			}},
			nil, pkg.Unknown, pkg.Present),
		// releases already in the cluster:
		pkg.NewPkgMock("installedfoo", "1.0.0", "installedns", nil, nil, pkg.Present, pkg.Unknown),
	}

	// create our own Logger that satisfies impl/cli.Logger, but with a buffer for tests
	buf := new(bytes.Buffer)
	logger := logcli.NewStandard()
	logger.InfoOut = buf
	logger.WarnOut = buf
	logger.ErrorOut = buf
	logger.DebugOut = buf
	log.Current = logger
	// logger.Level = log.DebugLevel

	s := New(InstallOne, logger)

	// Fill the DB with our mock packages:
	s.BuildWorldMock(pkgs)
	s.PkgDB.DebugPrintDB(logger)

	// Call the solver
	s.Solve()

	fmt.Println(s.FormatOutput(YAML))

	// Output:
	// presentunchanged:
	// - releasename: installedfoo
	//   version: 1.0.0
	//   namespace: installedns
	//   chartname: installedfoo
	//   dependsrel: []
	//   dependsoptionalrel: []
	//   repository: ourrepo
	//   currentstate: 1
	//   desiredstate: 0
	//   pinnedver: 0
	// toinstall:
	// - releasename: myawesomedep
	//   version: 0.1.100
	//   namespace: myawesomedeptargetns
	//   chartname: myawesomedep
	//   dependsrel: []
	//   dependsoptionalrel: []
	//   repository: ourrepo
	//   currentstate: 0
	//   desiredstate: 0
	//   pinnedver: 0
	// - releasename: wantedbaz
	//   version: 1.0.0
	//   namespace: wantedbazns
	//   chartname: wantedbaz
	//   dependsrel:
	//   - releasename: myawesomedep
	//     namespace: myawesomedeptargetns
	//     semverrange: ~0.1.0
	//   dependsoptionalrel: []
	//   repository: ourrepo
	//   currentstate: 0
	//   desiredstate: 1
	//   pinnedver: 0
	// toremove: []
	// status: SAT
	// inconsistencies: []
}
