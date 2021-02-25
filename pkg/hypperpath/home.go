// Package helmpath calculates filesystem paths to Helm's configuration, cache and data.
package hypperpath

// This helper builds paths to Helm's configuration, cache and data paths.
const lp = lazypath("hypper")

// ConfigPath returns the path where Helm stores configuration.
func ConfigPath(elem ...string) string { return lp.configPath(elem...) }

// CachePath returns the path where Helm stores cached objects.
func CachePath(elem ...string) string { return lp.cachePath(elem...) }

// DataPath returns the path where Helm stores data.
func DataPath(elem ...string) string { return lp.dataPath(elem...) }

// CacheIndexFile returns the path to an index for the given named repository.
func CacheIndexFile(name string) string {
	if name != "" {
		name += "-"
	}
	return name + "index.yaml"
}

// CacheChartsFile returns the path to a text file listing all the charts
// within the given named repository.
func CacheChartsFile(name string) string {
	if name != "" {
		name += "-"
	}
	return name + "charts.txt"
}
