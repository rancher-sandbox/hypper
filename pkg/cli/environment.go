/*
Copyright The Helm Authors, SUSE

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

package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"

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

	// hypper specific
	Verbose  bool
	NoColors bool
	NoEmojis bool
}

// New is a constructor of EnvSettings
func New() *EnvSettings {
	helmEnv := cli.New()
	env := &EnvSettings{
		EnvSettings: helmEnv,
		namespace:   os.Getenv("HYPPER_NAMESPACE"),
		Verbose:     false,
		NoColors:    false,
		NoEmojis:    false,
	}
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
}

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

// Namespace gets the namespace from the configuration
func (s *EnvSettings) Namespace() string {
	if ns, _, err := s.config.ToRawKubeConfigLoader().Namespace(); err == nil {
		return ns
	}
	return "default"
}
