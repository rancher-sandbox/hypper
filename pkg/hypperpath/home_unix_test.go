// +build !windows

package hypperpath

import (
	"os"
	"runtime"
	"testing"

	"github.com/rancher-sandbox/hypper/pkg/hypperpath/xdg"
)

func TestHelmHome(t *testing.T) {
	os.Setenv(xdg.CacheHomeEnvVar, "/cache")
	os.Setenv(xdg.ConfigHomeEnvVar, "/config")
	os.Setenv(xdg.DataHomeEnvVar, "/data")
	isEq := func(t *testing.T, got, expected string) {
		t.Helper()
		if expected != got {
			t.Error(runtime.GOOS)
			t.Errorf("Expected %q, got %q", expected, got)
		}
	}

	isEq(t, CachePath(), "/cache/hypper")
	isEq(t, ConfigPath(), "/config/hypper")
	isEq(t, DataPath(), "/data/hypper")

	// test to see if lazy-loading environment variables at runtime works
	os.Setenv(xdg.CacheHomeEnvVar, "/cache2")

	isEq(t, CachePath(), "/cache2/hypper")
}
