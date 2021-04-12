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

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func installAction(t *testing.T) *Install {
	config := actionConfigFixture(t)
	instAction := NewInstall(config)
	instAction.Namespace = "spaced"
	instAction.ReleaseName = "test-install-release"

	return instAction
}

func TestInstallSetNamespace(t *testing.T) {
	is := assert.New(t)

	// chart without annotations
	instAction := installAction(t)
	chart := buildChart()
	SetNamespace(instAction, chart, "defaultns", false)
	is.Equal("defaultns", instAction.Namespace)

	// hypper annotations have priority over fallback annotations
	instAction = installAction(t)
	chart = buildChart(withHypperAnnotations(), withFallbackAnnotations())
	SetNamespace(instAction, chart, "defaultns", false)
	is.Equal("hypper", instAction.Namespace)

	// fallback annotations have priority over default ns
	instAction = installAction(t)
	chart = buildChart(withFallbackAnnotations())
	SetNamespace(instAction, chart, "defaultns", false)
	is.Equal("fleet-system", instAction.Namespace)
}

func TestName(t *testing.T) {
	is := assert.New(t)

	// too many args
	chart := buildChart()
	_, err := GetName(chart, "", "name1", "chart-uri", "extraneous-arg")
	if err == nil {
		t.Fatal("expected an error")
	}
	is.Equal("expected at most two arguments, unexpected arguments: extraneous-arg", err.Error())

	// name and chart as args
	chart = buildChart()
	name, err := GetName(chart, "", "name1", "chart-uri")
	if err != nil {
		t.Fatal(err)
	}
	is.Equal("name1", name)

	// hypper annotations have priority over fallback annotations
	chart = buildChart(withHypperAnnotations(), withFallbackAnnotations())
	name, err = GetName(chart, "", "chart-uri")
	if err != nil {
		t.Fatal(err)
	}
	is.Equal("my-hypper-name", name)

	// fallback annotations
	chart = buildChart(withFallbackAnnotations())
	name, err = GetName(chart, "", "chart-uri")
	if err != nil {
		t.Fatal(err)
	}
	is.Equal("fleet", name)

	// no name or annotations present
	chart = buildChart()
	name, err = GetName(chart, "", "chart-uri")
	if err != nil {
		t.Fatal(err)
	}
	is.Equal("hello", name)
}

func TestChart(t *testing.T) {
	is := assert.New(t)
	instAction := installAction(t)

	// too many args
	_, err := instAction.Chart([]string{"name1", "chart-uri1", "extraneous-arg"})
	if err == nil {
		t.Fatal("expected an error")
	}
	is.Equal("expected at most two arguments, unexpected arguments: extraneous-arg", err.Error())

	// name and chart as args
	charturi, err := instAction.Chart([]string{"name2", "chart-uri2"})
	if err != nil {
		t.Fatal(err)
	}
	is.Equal("chart-uri2", charturi)

	// only chart as args
	charturi, err = instAction.Chart([]string{"chart-uri3"})
	if err != nil {
		t.Fatal(err)
	}
	is.Equal("chart-uri3", charturi)
}

func TestNameAndChart(t *testing.T) {
	is := assert.New(t)
	instAction := installAction(t)

	_, _, err := instAction.NameAndChart([]string{""})
	if err == nil {
		t.Fatal("expected an error")
	}
	is.Equal("NameAndChart() cannot be used", err.Error())
}

func TestCheckIfInstallable(t *testing.T) {
	is := assert.New(t)

	// Application chart type is installable
	err := CheckIfInstallable(buildChart(withTypeApplication()))
	is.NoError(err)

	// any other chart type besides Application is not installable
	err = CheckIfInstallable(buildChart(withTypeLibrary()))
	if err == nil {
		t.Fatal("expected an error")
	}
	is.Equal("library charts are not installable", err.Error())
}
