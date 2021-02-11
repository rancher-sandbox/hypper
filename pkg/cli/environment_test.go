package cli

import (
	"os"
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
		debug    bool
		verbose  bool
		noColors bool
		noEmojis bool
	}{
		{
			name:     "defaults",
			debug:    false,
			noColors: false,
			noEmojis: false,
		},
		{
			name:     "with flags set",
			args:     "--debug --no-colors --no-emojis",
			debug:    true,
			verbose:  false,
			noColors: true,
			noEmojis: true,
		},
		{
			name:     "with flags and verbose set",
			args:     "--verbose --no-colors --no-emojis",
			debug:    false,
			verbose:  true,
			noColors: true,
			noEmojis: true,
		},
		{
			name:     "with envvars set",
			envvars:  map[string]string{"HYPPER_DEBUG": "true", "HYPPER_NOCOLORS": "true", "HYPPER_NOEMOJIS": "true"},
			debug:    true,
			verbose:  false,
			noColors: true,
			noEmojis: true,
		},
		{
			name:     "with args and envvars set",
			args:     "--debug --no-colors --no-emojis",
			envvars:  map[string]string{"HYPPER_DEBUG": "true", "HYPPER_NOCOLORS": "true", "HYPPER_NOEMOJIS": "true"},
			debug:    true,
			verbose:  false,
			noColors: true,
			noEmojis: true,
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

			if settings.Verbose != tt.verbose {
				t.Errorf("on test %q expected verbose %t, got %t", tt.name, tt.verbose, settings.Verbose)
			}
		})
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
