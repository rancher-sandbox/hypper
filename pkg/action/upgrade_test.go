/*
Copyright The Helm Authors, SUSE LLC.

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

import "github.com/stretchr/testify/assert"
import "testing"

func upgradeAction(t *testing.T) *Upgrade {
	config := actionConfigFixture(t)
	upgrAction := NewUpgrade(config)

	return upgrAction
}

func TestUpgradeSetNamespace(t *testing.T) {
	is := assert.New(t)

	// chart without annotations
	upgrAction := upgradeAction(t)
	chart := buildChart()
	upgrAction.SetNamespace(chart, "defaultns")
	is.Equal("defaultns", upgrAction.Namespace)

	// hypper annotations have priority over fallback annotations
	upgrAction = upgradeAction(t)
	chart = buildChart(withHypperAnnotations(), withFallbackAnnotations())
	upgrAction.SetNamespace(chart, "defaultns")
	is.Equal("hypper", upgrAction.Namespace)

	// fallback annotations have priority over default ns
	upgrAction = upgradeAction(t)
	chart = buildChart(withFallbackAnnotations())
	upgrAction.SetNamespace(chart, "defaultns")
	is.Equal("fleet-system", upgrAction.Namespace)
}

func TestUpgradeName(t *testing.T) {
	is := assert.New(t)

	// hypper annotations have priority over fallback annotations
	upgrAction := upgradeAction(t)
	chart := buildChart(withHypperAnnotations(), withFallbackAnnotations())
	name, err := upgrAction.Name(chart)
	if err != nil {
		t.Fatal(err)
	}
	is.Equal("my-hypper-name", name)

	// fallback annotations
	upgrAction = upgradeAction(t)
	chart = buildChart(withFallbackAnnotations())
	name, err = upgrAction.Name(chart)
	if err != nil {
		t.Fatal(err)
	}
	is.Equal("fleet", name)

	// no annotations should default to the chart name
	upgrAction = upgradeAction(t)
	chart = buildChart()
	name, err = upgrAction.Name(chart)
	if err != nil {
		t.Fatal(err)
	}
	is.Equal("hello", name)
}
