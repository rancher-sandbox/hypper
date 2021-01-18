package cli

import (
	"os"
	"strconv"
)

type EnvSettings struct {
	Debug bool
}

func New() *EnvSettings {
	env := &EnvSettings{}
	env.Debug, _ = strconv.ParseBool(os.Getenv("HYPPER_DEBUG"))
	return env
}
