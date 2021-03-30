/*
Copyright SUSE LLC.

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
	"strings"

	"github.com/Masterminds/log-go"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
)

// SetNamespace sets the Namespace that should be used based on annotations or fallback to default
// both on the action and on the storage.
// if setDefault is true then it will just set the default namespace
// This will read the chart annotations. If no annotations, it leave the existing ns in the action.
// targetNS can be either the default namespace (usually "default") or the namespace passed via
// cli flag
func SetNamespace(x interface{}, chart *chart.Chart, targetNS string, setDefault bool) {
	namespace := targetNS
	// setDefault is mainly used when we set the namespace via the cli flag --namespace
	// and it has priority over everything else
	if setDefault {
		namespace = targetNS
	} else {
		if chart.Metadata.Annotations != nil {
			if val, ok := chart.Metadata.Annotations["hypper.cattle.io/namespace"]; ok {
				namespace = val
			} else {
				if val, ok := chart.Metadata.Annotations["catalog.cattle.io/namespace"]; ok {
					namespace = val
				}
			}
		}
	}

	switch i := x.(type) {
	case *Install:
		i.Namespace = namespace
		i.Config.SetNamespace(namespace)
	case *Upgrade:
		i.Namespace = namespace
		i.Config.SetNamespace(namespace)
	default:
		// No namespace was set because the type was unknown to set the
		// namespace on. This is an error in the use of the function in the
		// source so a panic is thrown.
		log.Panic("SetNamespace called on unknown type")
	}
}

// Name returns the name that should be used based of annotations
func GetName(chart *chart.Chart, nameTemplate string, args ...string) (string, error) {
	// args here could be: [NAME] [CHART]
	// cobra flags have been already stripped

	flagsNotSet := func() error {
		if nameTemplate != "" {
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

	if nameTemplate != "" {
		name, err := action.TemplateName(nameTemplate)
		return name, err
	}

	return chart.Name(), nil
}
