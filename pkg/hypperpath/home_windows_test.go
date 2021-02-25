// +build windows

package hypperpath

import (
	"os"
	"testing"

	"github.com/rancher-sandbox/hypper/pkg/hypperpath/xdg"
)

func TestHelmHome(t *testing.T) {
	os.Setenv(xdg.CacheHomeEnvVar, "c:\\")
	os.Setenv(xdg.ConfigHomeEnvVar, "d:\\")
	os.Setenv(xdg.DataHomeEnvVar, "e:\\")
	isEq := func(t *testing.T, a, b string) {
		if a != b {
			t.Errorf("Expected %q, got %q", b, a)
		}
	}

	isEq(t, CachePath(), "c:\\hypper")
	isEq(t, ConfigPath(), "d:\\hypper")
	isEq(t, DataPath(), "e:\\hypper")

	// test to see if lazy-loading environment variables at runtime works
	os.Setenv(xdg.CacheHomeEnvVar, "f:\\")

	isEq(t, CachePath(), "f:\\hypper")
}
