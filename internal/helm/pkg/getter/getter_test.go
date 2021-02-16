package getter

import (
	"testing"

	"github.com/rancher-sandbox/hypper/pkg/cli"
)

const pluginDir = "testdata/plugins"

func TestProvider(t *testing.T) {
	p := Provider{
		[]string{"one", "three"},
		func(_ ...Option) (Getter, error) { return nil, nil },
	}

	if !p.Provides("three") {
		t.Error("Expected provider to provide three")
	}
}

func TestProviders(t *testing.T) {
	ps := Providers{
		{[]string{"one", "three"}, func(_ ...Option) (Getter, error) { return nil, nil }},
		{[]string{"two", "four"}, func(_ ...Option) (Getter, error) { return nil, nil }},
	}

	if _, err := ps.ByScheme("one"); err != nil {
		t.Error(err)
	}
	if _, err := ps.ByScheme("four"); err != nil {
		t.Error(err)
	}

	if _, err := ps.ByScheme("five"); err == nil {
		t.Error("Did not expect handler for five")
	}
}

func TestAll(t *testing.T) {
	env := cli.New()
	env.PluginsDirectory = pluginDir

	all := All(env)
	if len(all) != 4 {
		t.Errorf("expected 4 providers (default plus three plugins), got %d", len(all))
	}

	if _, err := all.ByScheme("test2"); err != nil {
		t.Error(err)
	}
}

func TestByScheme(t *testing.T) {
	env := cli.New()
	env.PluginsDirectory = pluginDir

	g := All(env)
	if _, err := g.ByScheme("test"); err != nil {
		t.Error(err)
	}
	if _, err := g.ByScheme("https"); err != nil {
		t.Error(err)
	}
}
