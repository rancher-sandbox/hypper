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

package action

import (
	"fmt"
	"path/filepath"

	"github.com/pkg/errors"
	"strings"

	"github.com/Masterminds/log-go"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/time"
)

// Install is a composite type of Helm's Install type
type Install struct {
	*action.Install
}

// NewInstall creates a new Install object with the given configuration,
// by wrapping action.NewInstall
func NewInstall(cfg *Configuration) *Install {
	return &Install{
		action.NewInstall(cfg.Configuration),
	}
}

// CheckDependencies checks the dependencies for a chart.
// by wrapping action.CheckDependencies
func CheckDependencies(ch *chart.Chart, reqs []*chart.Dependency) error {
	return action.CheckDependencies(ch, reqs)
}

// Run executes the installation
//
// If DryRun is set to true, this will prepare the release, but not install it
func (i *Install) Run(chrt *chart.Chart, vals map[string]interface{}) (*release.Release, error) {
	log.Infof("Installing chart \"%s\" in namespace \"%s\"…", i.ReleaseName, i.Namespace)
	helmInstall := i.Install
	rel, err := helmInstall.Run(chrt, vals) // wrap Helm's i.Run for now
	return rel, err
}

// SetNamespace sets the Namespace that should be used in action.Install
//
// This will read the chart annotations. If no annotations, it leave the existing ns in the action.
func (i *Install) SetNamespace(chart *chart.Chart, defaultns string) {
	i.Namespace = defaultns
	if chart.Metadata.Annotations != nil {
		if val, ok := chart.Metadata.Annotations["hypper.cattle.io/namespace"]; ok {
			i.Namespace = val
		} else {
			if val, ok := chart.Metadata.Annotations["catalog.cattle.io/namespace"]; ok {
				i.Namespace = val
			}
		}
	}
}

// Name returns the name that should be used.
//
// This will read the flags and handle name generation if necessary.
func (i *Install) Name(chart *chart.Chart, args []string) (string, error) {
	// args here will only be: [NAME] [CHART]
	// cobra flags have been already stripped

	flagsNotSet := func() error {
		if i.GenerateName {
			return errors.New("cannot set --generate-name and also specify a name")
		}
		if i.NameTemplate != "" {
			return errors.New("cannot set --name-template and also specify a name")
		}
		return nil
	}

	if len(args) > 2 {
		return args[0], errors.Errorf("expected at most two arguments, unexpected arguments: %v", strings.Join(args[2:], ", "))
	}

	if len(args) == 2 {
		return args[0], flagsNotSet()
	}

	if chart.Metadata.Annotations != nil {
		if val, ok := chart.Metadata.Annotations["hypper.cattle.io/release-name"]; ok {
			return val, nil
		}
		if val, ok := chart.Metadata.Annotations["catalog.cattle.io/release-name"]; ok {
			return val, nil
		}
	}

	if i.NameTemplate != "" {
		name, err := action.TemplateName(i.NameTemplate)
		return name, err
	}

	if i.ReleaseName != "" {
		return i.ReleaseName, nil
	}

	if !i.GenerateName {
		return "", errors.New("must either provide a name, set the correct chart annotations, or specify --generate-name")
	}

	base := filepath.Base(args[0])
	if base == "." || base == "" {
		base = "chart"
	}
	// if present, strip out the file extension from the name
	if idx := strings.Index(base, "."); idx != -1 {
		base = base[0:idx]
	}

	return fmt.Sprintf("%s-%d", base, time.Now().Unix()), nil
}

// Chart returns the chart that should be used.
//
// This will read the flags and skip args if necessary.
func (i *Install) Chart(args []string) (string, error) {
	if len(args) > 2 {
		return args[1], errors.Errorf("expected at most two arguments, unexpected arguments: %v", strings.Join(args[2:], ", "))
	}

	if len(args) == 2 {
		return args[1], nil
	}

	// len(args) == 1
	return args[0], nil
}

// NameAndChart overloads Helm's NameAndChart. It always fails.
//
// On Hypper, we need to read the chart annotations to know the correct release name.
// Therefore, it cannot happen in this function.
func (i *Install) NameAndChart(args []string) (string, string, error) {
	return "", "", errors.New("NameAndChart() cannot be used")
}
