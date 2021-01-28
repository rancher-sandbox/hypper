package main

import (
	"os"
	"testing"
)

func TestRootCmd(t *testing.T) {
	tests := []struct {
		name, args string
		envvars    map[string]string
	}{
		{
			name: "defaults",
			args: "", //run default without any arguments
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.envvars {
				os.Setenv(k, v)
			}
			if _, _, err := executeCommandStdinC(tt.args); err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
		})
	}
}

func TestUnknownSubCmd(t *testing.T) {
	_, _, err := executeCommandStdinC("foobar")

	if err == nil || err.Error() != `unknown command "foobar" for "hypper"` {
		t.Errorf("Expect unknown command error, got %q", err)
	}
}
