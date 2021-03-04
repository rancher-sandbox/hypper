// +build darwin

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

	expected := filepath.Join(homedir.HomeDir(), "Library", appName, testFile)

	if lazy.dataPath(testFile) != expected {
		t.Errorf("expected '%s', got '%s'", expected, lazy.dataPath(testFile))
	}

	os.Setenv(xdg.DataHomeEnvVar, "/tmp")

	expected = filepath.Join("/tmp", appName, testFile)

	if lazy.dataPath(testFile) != expected {
		t.Errorf("expected '%s', got '%s'", expected, lazy.dataPath(testFile))
	}
}

func TestConfigPath(t *testing.T) {
	os.Unsetenv(xdg.ConfigHomeEnvVar)

	expected := filepath.Join(homedir.HomeDir(), "Library", "Preferences", appName, testFile)

	if lazy.configPath(testFile) != expected {
		t.Errorf("expected '%s', got '%s'", expected, lazy.configPath(testFile))
	}

	os.Setenv(xdg.ConfigHomeEnvVar, "/tmp")

	expected = filepath.Join("/tmp", appName, testFile)

	if lazy.configPath(testFile) != expected {
		t.Errorf("expected '%s', got '%s'", expected, lazy.configPath(testFile))
	}
}

func TestCachePath(t *testing.T) {
	os.Unsetenv(xdg.CacheHomeEnvVar)

	expected := filepath.Join(homedir.HomeDir(), "Library", "Caches", appName, testFile)

	if lazy.cachePath(testFile) != expected {
		t.Errorf("expected '%s', got '%s'", expected, lazy.cachePath(testFile))
	}

	os.Setenv(xdg.CacheHomeEnvVar, "/tmp")

	expected = filepath.Join("/tmp", appName, testFile)

	if lazy.cachePath(testFile) != expected {
		t.Errorf("expected '%s', got '%s'", expected, lazy.cachePath(testFile))
	}
}
