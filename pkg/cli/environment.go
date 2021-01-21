package cli

import (
	"fmt"
	"os"
	"regexp"
	"strconv"

	"github.com/spf13/pflag"
)

// EnvSettings describes all of the environment settings.
type EnvSettings struct {
	Debug    bool
	NoColors bool
	NoEmojis bool
}

// New is a constructor of EnvSettings
func New() *EnvSettings {
	env := &EnvSettings{}
	env.Debug, _ = strconv.ParseBool(os.Getenv("HYPPER_DEBUG"))
	env.NoColors, _ = strconv.ParseBool(os.Getenv("HYPPER_NOCOLORS"))
	env.NoEmojis, _ = strconv.ParseBool(os.Getenv("HYPPER_NOEMOJIS"))
	return env
}

// AddFlags binds flags to the given flagset.
func (s *EnvSettings) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVarP(&s.Debug, "debug", "d", s.Debug, "enable verbose output")
	fs.BoolVar(&s.NoColors, "no-colors", s.NoColors, "disable colors")
	fs.BoolVar(&s.NoEmojis, "no-emojis", s.NoEmojis, "disable emojis")
}

// EnvVars gives you a string map of the environment variables
func (s *EnvSettings) EnvVars() map[string]string {
	return map[string]string{
		"HYPPER_DEBUG":    fmt.Sprint(s.Debug),
		"HYPPER_NOCOLORS": fmt.Sprint(s.NoColors),
		"HYPPER_NOEMOJIS": fmt.Sprint(s.NoEmojis),
	}
}

func RemoveEmojiFromString(s string) string {
	re := regexp.MustCompile(`:[a-zA-Z0-9-_+]+?:`)
	return re.ReplaceAllString(s, "")
}
