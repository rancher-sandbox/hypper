package action

import (
	"helm.sh/helm/v3/pkg/action"
)

// List is a composite type of Helm's List type
type List struct {
	*action.List
	cfg *Configuration
}

// NewList constructs a new *List by embedding action.List
func NewList(cfg *Configuration) *List {
	return &List{
		action.NewList(cfg.Configuration),
		cfg,
	}
}
