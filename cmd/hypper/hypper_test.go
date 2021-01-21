package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/cloudfoundry/bosh-cli/release"
	"github.com/mattfarina/hypper/pkg/cli"
	"github.com/mattn/go-shellwords"
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/time"
)

// TODO: questionable code, remove or test
func testTimestamper() time.Time { return time.Unix(242085845, 0).UTC() }

// TODO: questionable code, remove or test
func init() {
	action.Timestamper = testTimestamper
}

// TODO: questionable code, remove or test
func runTestCmd(t *testing.T, tests []cmdTestCase) {
	t.Helper()
	for _, tt := range tests {
		for i := 0; i <= tt.repeat; i++ {
			t.Run(tt.name, func(t *testing.T) {
				defer resetEnv()()

				t.Logf("running cmd (attempt %d): %s", i+1, tt.cmd)
				_, _, err := executeCommandStdinC(tt.cmd)
				if (err != nil) != tt.wantError {
					t.Errorf("expected error, got '%v'", err)
				}
			})
		}
	}
}

func executeCommandStdinC(cmd string) (*cobra.Command, string, error) {

	args, err := shellwords.Parse(cmd)

	if err != nil {
		return nil, "", err
	}

	buf := new(bytes.Buffer)
	root, err := newRootCmd(buf, args)
	if err != nil {
		return nil, "", err
	}

	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	oldStdin := os.Stdin

	c, err := root.ExecuteC()
	result := buf.String()
	os.Stdin = oldStdin

	return c, result, err
}

func resetEnv() func() {
	origEnv := os.Environ()
	return func() {
		os.Clearenv()
		for _, pair := range origEnv {
			kv := strings.SplitN(pair, "=", 2)
			os.Setenv(kv[0], kv[1])
		}
		settings = cli.New()
	}
}

// cmdTestCase describes a test case that works with releases.
type cmdTestCase struct {
	name      string
	cmd       string
	golden    string
	wantError bool
	// Rels are the available releases at the start of the test.
	rels []*release.Release
	// Number of repeats (in case a feature was previously flaky and the test checks
	// it's now stably producing identical results). 0 means test is run exactly once.
	repeat int
}
