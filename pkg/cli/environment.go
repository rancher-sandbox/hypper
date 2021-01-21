package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/pflag"
)

// EnvSettings describes all of the environment settings.
type EnvSettings struct {
	Debug    bool
	NoColors bool
}

// New is a constructor of EnvSettings
func New() *EnvSettings {
	env := &EnvSettings{}
	env.Debug, _ = strconv.ParseBool(os.Getenv("HYPPER_DEBUG"))
	env.NoColors, _ = strconv.ParseBool(os.Getenv("HYPPER_COLORS"))

	return env
}

// AddFlags binds flags to the given flagset.
func (s *EnvSettings) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVarP(&s.Debug, "debug", "d", s.Debug, "enable verbose output")
	fs.BoolVarP(&s.NoColors, "no-colors", "n", s.NoColors, "disable colors")
}

// EnvVars gives you a string map of the environment variables
func (s *EnvSettings) EnvVars() map[string]string {
	return map[string]string{
		"HYPPER_DEBUG":    fmt.Sprint(s.Debug),
		"HYPPER_NOCOLORS": fmt.Sprint(s.NoColors),
	}
}
