/*
Copyright The Helm Authors, SUSE

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
	"testing"

	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/time"
)

func installAction(t *testing.T) *Install {
	config := actionConfigFixture(t)
	instAction := NewInstall(config)
	instAction.Namespace = "spaced"
	instAction.ReleaseName = "test-install-release"

	return instAction
}

func TestSetNamespace(t *testing.T) {
	is := assert.New(t)

	// chart without annotations
	instAction := installAction(t)
	chart := buildChart()
	instAction.SetNamespace(chart, "defaultns")
	is.Equal("defaultns", instAction.Namespace)

	// hypper annotations have priority over fallback annotations
	instAction = installAction(t)
	chart = buildChart(withHypperAnnotations())
	instAction.SetNamespace(chart, "defaultns")
	is.Equal("hypper", instAction.Namespace)

	// fallback annotations have priority over default ns
	instAction = installAction(t)
	chart = buildChart(withFallbackAnnotations())
	instAction.SetNamespace(chart, "defaultns")
	is.Equal("fleet-system", instAction.Namespace)
}

func TestName(t *testing.T) {
	is := assert.New(t)

	// too many args
	instAction := installAction(t)
	chart := buildChart()
	_, err := instAction.Name(chart, []string{"name1", "chart-uri", "extraneous-arg"})
	if err == nil {
		t.Fatal("expected an error")
	}
	is.Equal("expected at most two arguments, unexpected arguments: extraneous-arg", err.Error())

	// name and chart as args
	instAction = installAction(t)
	chart = buildChart()
	name, err := instAction.Name(chart, []string{"name1", "chart-uri"})
	if err != nil {
		t.Fatal(err)
	}
	is.Equal("name1", name)

	// hypper annotations have priority over fallback annotations
	instAction = installAction(t)
	chart = buildChart(withHypperAnnotations())
	name, err = instAction.Name(chart, []string{"chart-uri"})
	if err != nil {
		t.Fatal(err)
	}
	is.Equal("my-hypper-name", name)

	// fallback annotations
	instAction = installAction(t)
	chart = buildChart(withFallbackAnnotations())
	name, err = instAction.Name(chart, []string{"chart-uri"})
	if err != nil {
		t.Fatal(err)
	}
	is.Equal("fleet", name)

	// annotations have priority over generate-name
	instAction = installAction(t)
	instAction.GenerateName = true
	chart = buildChart(withFallbackAnnotations())
	name, err = instAction.Name(chart, []string{"chart-uri"})
	if err != nil {
		t.Fatal(err)
	}
	is.Equal("fleet", name)

	// --generate-name while specifying a name
	instAction = installAction(t)
	instAction.GenerateName = true
	chart = buildChart()
	_, err = instAction.Name(chart, []string{"name1", "chart-uri"})
	if err == nil {
		t.Fatal(err)
	}
	is.Equal("cannot set --generate-name and also specify a name", err.Error())

	// no name or annotations present
	instAction = installAction(t)
	instAction.ReleaseName = ""
	chart = buildChart()
	_, err = instAction.Name(chart, []string{"chart-uri"})
	if err == nil {
		t.Fatal("expected an error")
	}
	is.Equal("must either provide a name, set the correct chart annotations, or specify --generate-name", err.Error())
}

func TestNameGenerateName(t *testing.T) {
	is := assert.New(t)
	instAction := installAction(t)

	instAction.ReleaseName = ""
	instAction.GenerateName = true
	chart := buildChart()

	tests := []struct {
		Name         string
		Chart        string
		ExpectedName string
	}{
		{
			"local filepath",
			"./chart",
			fmt.Sprintf("chart-%d", time.Now().Unix()),
		},
		{
			"dot filepath",
			".",
			fmt.Sprintf("chart-%d", time.Now().Unix()),
		},
		{
			"empty filepath",
			"",
			fmt.Sprintf("chart-%d", time.Now().Unix()),
		},
		{
			"packaged chart",
			"chart.tgz",
			fmt.Sprintf("chart-%d", time.Now().Unix()),
		},
		{
			"packaged chart with .tar.gz extension",
			"chart.tar.gz",
			fmt.Sprintf("chart-%d", time.Now().Unix()),
		},
		{
			"packaged chart with local extension",
			"./chart.tgz",
			fmt.Sprintf("chart-%d", time.Now().Unix()),
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			name, err := instAction.Name(chart, []string{tc.Chart})
			if err != nil {
				t.Fatal(err)
			}

			is.Equal(tc.ExpectedName, name)
		})
	}
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
