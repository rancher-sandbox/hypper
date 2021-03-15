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
)

// Uninstall is a composite type of Helm's Uninstall type
type Uninstall struct {
	*action.Uninstall
	Config *Configuration
}

// NewUninstall creates a new Uninstall by embedding action.Uninstall
func NewUninstall(cfg *Configuration) *Uninstall {
	return &Uninstall{
		action.NewUninstall(cfg.Configuration),
		cfg,
	}
}
