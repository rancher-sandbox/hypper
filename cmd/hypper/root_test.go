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
