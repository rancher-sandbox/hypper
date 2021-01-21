package cli

import (
	"os"
	"strings"
	"testing"

	"github.com/spf13/pflag"
)

func TestEmojis(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "plain string",
			input:    "this is a plain string",
			expected: "this is a plain string",
		},
		{
			name:     "plain utf-8 string",
			input:    "\u2000-\u3300",
			expected: "\u2000-\u3300",
		},
		{
			name:     "pure emoji string",
			input:    ":pizza::beer:",
			expected: "",
		},
		{
			name:     "mixed string",
			input:    ":pizza: beer",
			expected: " beer",
		},
		{
			name:     "weird edge cases",
			input:    ":pizza: :beer: :pizza beer:",
			expected: "  :pizza beer:",
		},
		{
			name:     "double colon",
			input:    ":: test",
			expected: ":: test",
		},
		{
			name:     "double colon emoji mixup",
			input:    "::pizza::",
			expected: "::",
		},
		{
			name:     "do not touch native utf-8 inside colons",
			input:    ":\u2000-\u3300:",
			expected: ":\u2000-\u3300:",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := RemoveEmojiFromString(tt.input)
			if tt.expected != s {
				t.Errorf("expected %s got %s", tt.expected, s)
			}
		})
	}
}

func TestEnvSettings(t *testing.T) {
	tests := []struct {
		name string

		// input
		args    string
		envvars map[string]string

		// expected values
		debug    bool
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
			noColors: true,
			noEmojis: true,
		},
		{
			name:     "with envvars set",
			envvars:  map[string]string{"HYPPER_DEBUG": "true", "HYPPER_NOCOLORS": "true", "HYPPER_NOEMOJIS": "true"},
			debug:    true,
			noColors: true,
			noEmojis: true,
		},
		{
			name:     "with args and envvars set",
			args:     "--debug --no-colors --no-emojis",
			envvars:  map[string]string{"HYPPER_DEBUG": "true", "HYPPER_NOCOLORS": "true", "HYPPER_NOEMOJIS": "true"},
			debug:    true,
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
				t.Errorf("expected debug %t, got %t", tt.debug, settings.Debug)
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
