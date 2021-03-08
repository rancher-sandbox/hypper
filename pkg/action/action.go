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
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/kube"
	"helm.sh/helm/v3/pkg/time"
)

// Timestamper is a function capable of producing a timestamp.Timestamper.
//
// By default, this is a time.Time function from the Helm time package. This can
// be overridden for testing though, so that timestamps are predictable.
var Timestamper = time.Now

// Configuration is a composite type of Helm's Configuration type
type Configuration struct {
	*action.Configuration
}

// SetNamespace sets the namespace on the kubeclient
func (c *Configuration) SetNamespace(namespace string) {
	switch i := c.KubeClient.(type) {
	case *kube.Client:
		i.Namespace = namespace
		c.KubeClient = i
	}
}
