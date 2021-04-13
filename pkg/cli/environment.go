/*
Copyright The Helm Authors, SUSE LLC.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/*Package cli describes the operating environment for the hypper CLI.
hypper's environment encapsulates all of the service dependencies hypper has.
These dependencies are expressed as interfaces so that alternate implementations
(mocks, etc.) can be easily generated.
Hypper inherits most of the environment values from helm and expands on top of them
*/
package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/rancher-sandbox/hypper/pkg/hypperpath"
	"github.com/spf13/pflag"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/helmpath"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// defaultMaxHistory sets the maximum number of releases to 0: unlimited
const defaultMaxHistory = 10

// EnvSettings is a composite type of helm.pkg.env.EnvSettings.
// describes all of the environment settings.
type EnvSettings struct {

	// we use Helm's for all exported fields
	*cli.EnvSettings

	// unexported Helm's cli.EnvSettings fields are here:
	namespace string
	config    *genericclioptions.ConfigFlags

	// KubeConfig is the path to the kubeconfig file
	KubeConfig string
	// KubeContext is the name of the kubeconfig context.
	KubeContext string
	// Bearer KubeToken used for authentication
	KubeToken string
	// Username to impersonate for the operation
	KubeAsUser string
	// Groups to impersonate for the operation, multiple groups parsed from a comma delimited list
	KubeAsGroups []string
	// Kubernetes API Server Endpoint for authentication
	KubeAPIServer string
	// Custom certificate authority file.
	KubeCaFile string
	// Debug indicates whether or not Helm is running in Debug mode.
	Debug bool
	// RegistryConfig is the path to the registry config file.
	RegistryConfig string
	// RepositoryConfig is the path to the repositories file.
	RepositoryConfig string
	// RepositoryCache is the path to the repository cache directory.
	RepositoryCache string
	// PluginsDirectory is the path to the plugins directory.
	PluginsDirectory string
	// MaxHistory is the max release history maintained.
	MaxHistory int

	// hypper specific
	Verbose           bool
	NoColors          bool
	NoEmojis          bool
	NamespaceFromFlag bool
}

// New is a constructor of EnvSettings
func New() *EnvSettings {
	// TODO check with reflection that we are not missing any field
	env := &EnvSettings{
		EnvSettings: nil, // this is nil it gets overwritten below to ensure that it does not touch Helm's stuff

		namespace:        os.Getenv("HYPPER_NAMESPACE"),
		MaxHistory:       envIntOr("HYPPER_MAX_HISTORY", defaultMaxHistory),
		KubeContext:      os.Getenv("HYPPER_KUBECONTEXT"),
		KubeToken:        os.Getenv("HYPPER_KUBETOKEN"),
		KubeAsUser:       os.Getenv("HYPPER_KUBEASUSER"),
		KubeAsGroups:     envCSV("HYPPER_KUBEASGROUPS"),
		KubeAPIServer:    os.Getenv("HYPPER_KUBEAPISERVER"),
		KubeCaFile:       os.Getenv("HYPPER_KUBECAFILE"),
		PluginsDirectory: envOr("HYPPER_PLUGINS", hypperpath.DataPath("plugins")),
		RegistryConfig:   envOr("HYPPER_REGISTRY_CONFIG", hypperpath.ConfigPath("registry.json")),
		RepositoryConfig: envOr("HYPPER_REPOSITORY_CONFIG", hypperpath.ConfigPath("repositories.yaml")),
		RepositoryCache:  envOr("HYPPER_REPOSITORY_CACHE", hypperpath.CachePath("repository")),

		Verbose:  false,
		NoColors: false,
		NoEmojis: false,
	}

	os.Setenv("HELM_NAMESPACE", env.namespace) // WHY???

	// Create default helm settings and propagate our default hypper values
	env.EnvSettings = cli.New()

	env.FillHelmSettings()

	env.Debug, _ = strconv.ParseBool(os.Getenv("HYPPER_DEBUG"))
	env.Verbose, _ = strconv.ParseBool(os.Getenv("HYPPER_TRACE"))
	env.NoColors, _ = strconv.ParseBool(os.Getenv("HYPPER_NOCOLORS"))
	env.NoEmojis, _ = strconv.ParseBool(os.Getenv("HYPPER_NOEMOJIS"))

	// bind to kubernetes config flags
	env.config = &genericclioptions.ConfigFlags{
		Namespace:        &env.namespace,
		Context:          &env.KubeContext,
		BearerToken:      &env.KubeToken,
		APIServer:        &env.KubeAPIServer,
		CAFile:           &env.KubeCaFile,
		KubeConfig:       &env.KubeConfig,
		Impersonate:      &env.KubeAsUser,
		ImpersonateGroup: &env.KubeAsGroups,
	}

	return env
}

// FillHelmSettings will fill the helm cli.Settings with the proper values from CLI flags
// This has to be called after the flags are set, otherwise the flags are not propagated to Helm cli.Settings
func (s *EnvSettings) FillHelmSettings() {
	s.EnvSettings.MaxHistory = s.MaxHistory
	s.EnvSettings.KubeContext = s.KubeContext
	s.EnvSettings.KubeToken = s.KubeToken
	s.EnvSettings.KubeAsUser = s.KubeAsUser
	s.EnvSettings.KubeAsGroups = s.KubeAsGroups
	s.EnvSettings.KubeAPIServer = s.KubeAPIServer
	s.EnvSettings.KubeCaFile = s.KubeCaFile
	s.EnvSettings.PluginsDirectory = s.PluginsDirectory
	s.EnvSettings.RegistryConfig = s.RegistryConfig
	s.EnvSettings.RepositoryCache = s.RepositoryCache
	s.EnvSettings.RepositoryConfig = s.RepositoryConfig
}

// AddFlags binds flags to the given flagset.
func (s *EnvSettings) AddFlags(fs *pflag.FlagSet) {
	// Hypper specific
	fs.StringVarP(&s.namespace, "namespace", "n", s.namespace, "namespace scope for this request")
	fs.BoolVar(&s.Debug, "debug", s.Debug, "enable verbose output")
	fs.BoolVar(&s.NoColors, "no-colors", s.NoColors, "disable colors")
	fs.BoolVar(&s.NoEmojis, "no-emojis", s.NoEmojis, "disable emojis")

	// imported from Helm
	fs.StringVar(&s.KubeConfig, "kubeconfig", "", "path to the kubeconfig file")
	fs.StringVar(&s.KubeContext, "kube-context", s.KubeContext, "name of the kubeconfig context to use")
	fs.StringVar(&s.KubeToken, "kube-token", s.KubeToken, "bearer token used for authentication")
	fs.StringVar(&s.KubeAsUser, "kube-as-user", s.KubeAsUser, "username to impersonate for the operation")
	fs.StringArrayVar(&s.KubeAsGroups, "kube-as-group", s.KubeAsGroups, "group to impersonate for the operation, this flag can be repeated to specify multiple groups.")
	fs.StringVar(&s.KubeAPIServer, "kube-apiserver", s.KubeAPIServer, "the address and the port for the Kubernetes API server")
	fs.StringVar(&s.KubeCaFile, "kube-ca-file", s.KubeCaFile, "the certificate authority file for the Kubernetes API server connection")

	fs.StringVar(&s.RegistryConfig, "registry-config", s.RegistryConfig, "path to the registry config file")
	fs.StringVar(&s.RepositoryConfig, "repository-config", s.RepositoryConfig, "path to the file containing repository names and URLs")
	fs.StringVar(&s.RepositoryCache, "repository-cache", s.RepositoryCache, "path to the file containing cached repository indexes")

}

func envOr(name, def string) string {
	if v, ok := os.LookupEnv(name); ok {
		return v
	}
	return def
}

func envIntOr(name string, def int) int {
	if name == "" {
		return def
	}
	envVal := envOr(name, strconv.Itoa(def))
	ret, err := strconv.Atoi(envVal)
	if err != nil {
		return def
	}
	return ret
}

func envCSV(name string) (ls []string) {
	trimmed := strings.Trim(os.Getenv(name), ", ")
	if trimmed != "" {
		ls = strings.Split(trimmed, ",")
	}
	return
}

// EnvVars reads the hypper and helm enviromental variables and populates
// values in the EnvSettings caller.
func (s *EnvSettings) EnvVars() map[string]string {
	envvars := map[string]string{
		"HELM_BIN":               os.Args[0],
		"HELM_CACHE_HOME":        helmpath.CachePath(""),
		"HELM_CONFIG_HOME":       helmpath.ConfigPath(""),
		"HELM_DATA_HOME":         helmpath.DataPath(""),
		"HELM_DEBUG":             fmt.Sprint(s.Debug),
		"HELM_PLUGINS":           s.PluginsDirectory,
		"HELM_REGISTRY_CONFIG":   s.RegistryConfig,
		"HELM_REPOSITORY_CACHE":  s.RepositoryCache,
		"HELM_REPOSITORY_CONFIG": s.RepositoryConfig,
		"HELM_NAMESPACE":         s.Namespace(),
		"HELM_MAX_HISTORY":       strconv.Itoa(s.MaxHistory),

		// broken, these are populated from helm flags and not kubeconfig.
		"HELM_KUBECONTEXT":   s.KubeContext,
		"HELM_KUBETOKEN":     s.KubeToken,
		"HELM_KUBEASUSER":    s.KubeAsUser,
		"HELM_KUBEASGROUPS":  strings.Join(s.KubeAsGroups, ","),
		"HELM_KUBEAPISERVER": s.KubeAPIServer,
		"HELM_KUBECAFILE":    s.KubeCaFile,

		//hypper specific
		"HYPPER_VERBOSE":  fmt.Sprint(s.Verbose),
		"HYPPER_NOCOLORS": fmt.Sprint(s.NoColors),
		"HYPPER_NOEMOJIS": fmt.Sprint(s.NoEmojis),
	}
	if s.KubeConfig != "" {
		envvars["KUBECONFIG"] = s.KubeConfig
	}
	return envvars
}

// Namespace gets the namespace from the configuration or the config flag
func (s *EnvSettings) Namespace() string {
	if s.NamespaceFromFlag {
		return s.namespace
	}
	if ns, _, err := s.config.ToRawKubeConfigLoader().Namespace(); err == nil {
		return ns
	}
	return "default"
}
