package getter

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/rancher-sandbox/hypper/internal/helm/plugin"
	"github.com/rancher-sandbox/hypper/pkg/cli"
)

// collectPlugins scans for getter plugins.
// This will load plugins according to the cli.
func collectPlugins(settings *cli.EnvSettings) (Providers, error) {
	plugins, err := plugin.FindPlugins(settings.PluginsDirectory)
	if err != nil {
		return nil, err
	}
	var result Providers
	for _, plugin := range plugins {
		for _, downloader := range plugin.Metadata.Downloaders {
			result = append(result, Provider{
				Schemes: downloader.Protocols,
				New: NewPluginGetter(
					downloader.Command,
					settings,
					plugin.Metadata.Name,
					plugin.Dir,
				),
			})
		}
	}
	return result, nil
}

// pluginGetter is a generic type to invoke custom downloaders,
// implemented in plugins.
type pluginGetter struct {
	command  string
	settings *cli.EnvSettings
	name     string
	base     string
	opts     options
}

// Get runs downloader plugin command
func (p *pluginGetter) Get(href string, options ...Option) (*bytes.Buffer, error) {
	for _, opt := range options {
		opt(&p.opts)
	}
	commands := strings.Split(p.command, " ")
	argv := append(commands[1:], p.opts.certFile, p.opts.keyFile, p.opts.caFile, href)
	prog := exec.Command(filepath.Join(p.base, commands[0]), argv...)
	plugin.SetupPluginEnv(p.settings, p.name, p.base)
	prog.Env = os.Environ()
	buf := bytes.NewBuffer(nil)
	prog.Stdout = buf
	prog.Stderr = os.Stderr
	if err := prog.Run(); err != nil {
		if eerr, ok := err.(*exec.ExitError); ok {
			os.Stderr.Write(eerr.Stderr)
			return nil, errors.Errorf("plugin %q exited with error", p.command)
		}
		return nil, err
	}
	return buf, nil
}

// NewPluginGetter constructs a valid plugin getter
func NewPluginGetter(command string, settings *cli.EnvSettings, name, base string) Constructor {
	return func(options ...Option) (Getter, error) {
		result := &pluginGetter{
			command:  command,
			settings: settings,
			name:     name,
			base:     base,
		}
		for _, opt := range options {
			opt(&result.opts)
		}
		return result, nil
	}
}
