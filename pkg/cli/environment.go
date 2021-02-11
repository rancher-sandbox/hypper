package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/pflag"
	helmCli "helm.sh/helm/v3/pkg/cli"
)

// EnvSettings describes all of the environment settings.
type EnvSettings struct {
	HelmSettings *helmCli.EnvSettings
	Debug        bool
	Verbose      bool
	NoColors     bool
	NoEmojis     bool
}

// New is a constructor of EnvSettings
func New() *EnvSettings {
	env := &EnvSettings{}
	env.Debug, _ = strconv.ParseBool(os.Getenv("HYPPER_DEBUG"))
	env.Verbose, _ = strconv.ParseBool(os.Getenv("HYPPER_TRACE"))
	env.NoColors, _ = strconv.ParseBool(os.Getenv("HYPPER_NOCOLORS"))
	env.NoEmojis, _ = strconv.ParseBool(os.Getenv("HYPPER_NOEMOJIS"))
	env.HelmSettings = helmCli.New()
	return env
}

// AddFlags binds flags to the given flagset.
func (s *EnvSettings) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVarP(&s.Debug, "debug", "d", s.Debug, "enable debug output")
	fs.BoolVarP(&s.Verbose, "verbose", "v", s.Verbose, "enable verbose output (more verbose than debug)")
	fs.BoolVar(&s.NoColors, "no-colors", s.NoColors, "disable colors")
	fs.BoolVar(&s.NoEmojis, "no-emojis", s.NoEmojis, "disable emojis")
}

// EnvVars gives you a string map of the environment variables
func (s *EnvSettings) EnvVars() map[string]string {
	return map[string]string{
		"HYPPER_DEBUG":    fmt.Sprint(s.Debug),
		"HYPPER_VERBOSE":  fmt.Sprint(s.Verbose),
		"HYPPER_NOCOLORS": fmt.Sprint(s.NoColors),
		"HYPPER_NOEMOJIS": fmt.Sprint(s.NoEmojis),
	}
}
