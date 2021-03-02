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
