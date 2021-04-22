/*
Copyright The Helm Authors, SUSE LLC.

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

package cli

import (
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/cli"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/pflag"
)

func TestEnvSettings(t *testing.T) {
	tests := []struct {
		name string

		// input
		args    string
		envvars map[string]string

		// expected values
		ns, kcontext string
		debug        bool
		noColors     bool
		noEmojis     bool
		maxhistory   int
		kAsUser      string
		kAsGroups    []string
		kCaFile      string
	}{
		{
			debug:      false,
			noColors:   false,
			noEmojis:   false,
			name:       "defaults",
			ns:         "default",
			maxhistory: defaultMaxHistory,
		},
		{
			debug:      true,
			noColors:   true,
			noEmojis:   true,
			name:       "with flags set",
			args:       "--debug --no-colors --no-emojis --namespace=myns --kube-as-user=poro --kube-as-group=admins --kube-as-group=teatime --kube-as-group=snackeaters --kube-ca-file=/tmp/ca.crt",
			ns:         "myns",
			maxhistory: defaultMaxHistory,
			kAsUser:    "poro",
			kAsGroups:  []string{"admins", "teatime", "snackeaters"},
			kCaFile:    "/tmp/ca.crt",
		},
		{
			debug:      true,
			noColors:   true,
			noEmojis:   true,
			name:       "with envvars set",
			envvars:    map[string]string{"HYPPER_DEBUG": "true", "HYPPER_NOCOLORS": "true", "HYPPER_NOEMOJIS": "true", "HYPPER_NAMESPACE": "yourns", "HYPPER_KUBEASUSER": "pikachu", "HYPPER_KUBEASGROUPS": ",,,operators,snackeaters,partyanimals", "HYPPER_MAX_HISTORY": "5", "HYPPER_KUBECAFILE": "/tmp/ca.crt"},
			ns:         "yourns",
			maxhistory: 5,
			kAsUser:    "pikachu",
			kAsGroups:  []string{"operators", "snackeaters", "partyanimals"},
			kCaFile:    "/tmp/ca.crt",
		},
		{
			debug:      true,
			noColors:   true,
			noEmojis:   true,
			name:       "with flags and envvars set",
			args:       "--debug --no-colors --no-emojis --namespace=myns --kube-as-user=poro --kube-as-group=admins --kube-as-group=teatime --kube-as-group=snackeaters --kube-ca-file=/my/ca.crt",
			envvars:    map[string]string{"HYPPER_DEBUG": "true", "HYPPER_NOCOLORS": "true", "HYPPER_NOEMOJIS": "false", "HYPPER_NAMESPACE": "myns", "HYPPER_KUBEASUSER": "pikachu", "HYPPER_KUBEASGROUPS": ",,,operators,snackeaters,partyanimals", "HYPPER_MAX_HISTORY": "5", "HYPPER_KUBECAFILE": "/tmp/ca.crt"},
			ns:         "myns",
			maxhistory: 5,
			kAsUser:    "poro",
			kAsGroups:  []string{"admins", "teatime", "snackeaters"},
			kCaFile:    "/my/ca.crt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer resetEnv()()

			for k, v := range tt.envvars {
				os.Setenv(k, v)
			}

			flags := pflag.NewFlagSet("testing", pflag.ContinueOnError)

			settings := New()
			settings.AddFlags(flags)
			err := flags.Parse(strings.Split(tt.args, " "))
			if err != nil {
				t.Errorf("failed while parsing flags for %s", tt.args)
			}

			if settings.Debug != tt.debug {
				t.Errorf("on test %q expected debug %t, got %t", tt.name, tt.debug, settings.Debug)
			}
		})
	}
}

func TestHelmFields(t *testing.T) {

	helmSettings := reflect.TypeOf(cli.EnvSettings{})
	helmNumFields := helmSettings.NumField()
	hypperSettings := reflect.TypeOf(EnvSettings{})
	for i := 0; i < helmNumFields; i++ {
		field := helmSettings.Field(i)
		_, found := hypperSettings.FieldByName(field.Name)
		if !found {
			t.Fatalf("field %v from Helm cli.Settings not found in our settings", field.Name)
		}
	}

	// Composite type of helm EnvSettings plus an extra field
	type FakeSettings struct {
		*cli.EnvSettings
		missingField string
	}

	helmSettings = reflect.TypeOf(FakeSettings{})
	helmNumFields = helmSettings.NumField()
	hypperSettings = reflect.TypeOf(EnvSettings{})
	for i := 0; i < helmNumFields; i++ {
		field := helmSettings.Field(i)
		_, found := hypperSettings.FieldByName(field.Name)
		if !found {
			// If we fail, make sure that the error is due to our known missing field
			assert.Equal(t, field.Name, "missingField")
		}
	}
}

func resetEnv() func() {
	origEnv := os.Environ()

	// ensure any local envvars do not hose us
	for e := range New().EnvVars() {
		os.Unsetenv(e)
	}

	return func() {
		for _, pair := range origEnv {
			kv := strings.SplitN(pair, "=", 2)
			os.Setenv(kv[0], kv[1])
		}
	}
}
