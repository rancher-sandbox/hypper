package main

import (
	"bytes"
	"fmt"

	"github.com/Masterminds/log-go"
	logcli "github.com/Masterminds/log-go/impl/cli"

	pkg "github.com/rancher-sandbox/hypper/internal/package"
	"github.com/rancher-sandbox/hypper/internal/solver"
)

func main() {

	pkgs := []*pkg.Pkg{
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
	}

	// create our own Logger that satisfies impl/cli.Logger, but with a buffer for tests
	buf := new(bytes.Buffer)
	logger := logcli.NewStandard()
	logger.Level = 1
	logger.InfoOut = buf
	logger.WarnOut = buf
	logger.ErrorOut = buf
	logger.DebugOut = buf
	log.Current = logger

	s := solver.New()

	s.PkgDB.DebugPrintDB(logger)
	s.BuildWorldMock(pkgs)

	s.Solve(logger)
	s.PkgDB.DebugPrintDB(logger)

	fmt.Println(s.FormatOutput(solver.YAML))
}
