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
	"helm.sh/helm/v3/pkg/chart"
)

// SetNamespace sets the Namespace that should be used based on annotations or fallback to default
// both on the action and on the storage.
// if setDefault is true then it will just set the default namespace
// This will read the chart annotations. If no annotations, it leave the existing ns in the action.
// targetNS can be either the default namesapce (usually "default") or the namespace passed via
// cli flag
func SetNamespace(x interface{}, chart *chart.Chart, targetNS string, NamespaceFromFlag bool) {
	namespace := targetNS
	// NamespaceFromFlag is mainly used when we set the namespace via the cli flag --namespace
	// and it has priority over everything else
	if NamespaceFromFlag {
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
	}
}
