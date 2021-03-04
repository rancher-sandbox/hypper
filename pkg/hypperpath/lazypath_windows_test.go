// +build windows

package hypperpath

import (
	"os"
	"path/filepath"
	"testing"

	"k8s.io/client-go/util/homedir"

	"github.com/rancher-sandbox/hypper/pkg/hypperpath/xdg"
)

const (
	appName  = "hypper"
	testFile = "test.txt"
	lazy     = lazypath(appName)
)

func TestDataPath(t *testing.T) {
	os.Unsetenv(xdg.DataHomeEnvVar)
	os.Setenv("APPDATA", filepath.Join(homedir.HomeDir(), "foo"))

	expected := filepath.Join(homedir.HomeDir(), "foo", appName, testFile)

	if lazy.dataPath(testFile) != expected {
		t.Errorf("expected '%s', got '%s'", expected, lazy.dataPath(testFile))
	}

	os.Setenv(xdg.DataHomeEnvVar, filepath.Join(homedir.HomeDir(), "xdg"))

	expected = filepath.Join(homedir.HomeDir(), "xdg", appName, testFile)

	if lazy.dataPath(testFile) != expected {
		t.Errorf("expected '%s', got '%s'", expected, lazy.dataPath(testFile))
	}
}

func TestConfigPath(t *testing.T) {
	os.Unsetenv(xdg.ConfigHomeEnvVar)
	os.Setenv("APPDATA", filepath.Join(homedir.HomeDir(), "foo"))

	expected := filepath.Join(homedir.HomeDir(), "foo", appName, testFile)

	if lazy.configPath(testFile) != expected {
		t.Errorf("expected '%s', got '%s'", expected, lazy.configPath(testFile))
	}

	os.Setenv(xdg.ConfigHomeEnvVar, filepath.Join(homedir.HomeDir(), "xdg"))

	expected = filepath.Join(homedir.HomeDir(), "xdg", appName, testFile)

	if lazy.configPath(testFile) != expected {
		t.Errorf("expected '%s', got '%s'", expected, lazy.configPath(testFile))
	}
}

func TestCachePath(t *testing.T) {
	os.Unsetenv(xdg.CacheHomeEnvVar)
	os.Setenv("TEMP", filepath.Join(homedir.HomeDir(), "foo"))

	expected := filepath.Join(homedir.HomeDir(), "foo", appName, testFile)

	if lazy.cachePath(testFile) != expected {
		t.Errorf("expected '%s', got '%s'", expected, lazy.cachePath(testFile))
	}

	os.Setenv(xdg.CacheHomeEnvVar, filepath.Join(homedir.HomeDir(), "xdg"))

	expected = filepath.Join(homedir.HomeDir(), "xdg", appName, testFile)

	if lazy.cachePath(testFile) != expected {
		t.Errorf("expected '%s', got '%s'", expected, lazy.cachePath(testFile))
	}
}
