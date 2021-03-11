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
	"fmt"
	"os"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/kube"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
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

		// When actionConfig.Init is called it sets up the driver with a
		// namespace. We need to change the namespace for the driver because
		// we know a new location. Here we detect what driver is already in
		// use and recreate it with the new namespace.
		lazyClient := &lazyClient{
			namespace: namespace,
			clientFn:  i.Factory.KubernetesClientSet,
		}

		// Note, there is no default case at the end so that test drivers are
		// left alone.
		switch c.Releases.Driver.Name() {
		case "Secret":
			d := driver.NewSecrets(newSecretClient(lazyClient))
			d.Log = c.Log
			c.Releases = storage.Init(d)
		case "ConfigMap":
			d := driver.NewConfigMaps(newConfigMapClient(lazyClient))
			d.Log = c.Log
			c.Releases = storage.Init(d)
		case "Memory":
			var d *driver.Memory
			if c.Releases != nil {
				if mem, ok := c.Releases.Driver.(*driver.Memory); ok {
					d = mem
				}
			}
			if d == nil {
				d = driver.NewMemory()
			}
			d.SetNamespace(namespace)
			c.Releases = storage.Init(d)
		case "SQL":
			d, err := driver.NewSQL(
				os.Getenv("HELM_DRIVER_SQL_CONNECTION_STRING"),
				c.Log,
				namespace,
			)
			if err != nil {
				panic(fmt.Sprintf("Unable to instantiate SQL driver: %v", err))
			}
			c.Releases = storage.Init(d)
		}
	// tester Client:
	default:
		var d *driver.Memory
		if c.Releases != nil {
			if mem, ok := c.Releases.Driver.(*driver.Memory); ok {
				d = mem
			}
		}
		if d == nil {
			d = driver.NewMemory()
		}
		d.SetNamespace(namespace)
		c.Releases = storage.Init(d)
	}

}
