package main

import (
	"bytes"
	"os"

	"github.com/mattn/go-shellwords"
	"github.com/spf13/cobra"
)

//TODO re-enable when subcommands present
/*
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
*/

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

//TODO re-enable when subcommands present
/*
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
*/

//TODO re-enable when subcommands present
// cmdTestCase describes a test case that works with releases.
/*
type cmdTestCase struct {
	name      string
	cmd       string
	wantError bool
	// Number of repeats (in case a feature was previously flaky and the test checks
	// it's now stably producing identical results). 0 means test is run exactly once.
	repeat int
}
*/
