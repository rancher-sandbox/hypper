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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetNamespace(t *testing.T) {
	// SetNamespace is tested in the install/upgrade tests. What is not tested
	// is the case where a struct it can't set the namespace for is handled.
	assert.Panics(t, func() {
		act := &nmspc{}
		chart := buildChart()
		SetNamespace(act, chart, "defaultns", false)
	})
}

type nmspc struct{}
