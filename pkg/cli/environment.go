package cli

import (
	"os"
	"strconv"

	"github.com/spf13/pflag"
)

type EnvSettings struct {
	Debug    bool
	NoColors bool
}

func New() *EnvSettings {
	env := &EnvSettings{}
	env.Debug, _ = strconv.ParseBool(os.Getenv("HYPPER_DEBUG"))
	env.NoColors, _ = strconv.ParseBool(envOr("HYPPER_COLORS", "false"))

	return env
}

func (s *EnvSettings) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVarP(&s.Debug, "debug", "d", s.Debug, "enable verbose output")
	fs.BoolVarP(&s.NoColors, "no-colors", "n", s.NoColors, "disable colors")
}

func envOr(name, def string) string {
	if v, ok := os.LookupEnv(name); ok {
		return v
	}
	return def
}
