/*
Copyright The Helm Authors, SUSE LLC

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

package main

import (
	"bytes"
	"fmt"
	"github.com/Masterminds/log-go"
	logio "github.com/Masterminds/log-go/io"
	"github.com/rancher-sandbox/hypper/internal/version"
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/cmd/helm/require"
	"io"
	"text/template"
)

const versionDesc = `
Show the version for Hypper.

This will print a representation the version of Hypper.
The output will look something like this:

version.BuildInfo{Version:"v3.2.1", GitCommit:"fe51cd1e31e6a202cba7dead9552a6d418ded79a", GitTreeState:"clean", GoVersion:"go1.13.10"}

- Version is the semantic version of the release.
- GitCommit is the SHA for the commit that this version was built from.
- GitTreeState is "clean" if there are no local code changes when this binary was
  built, and "dirty" if the binary was built from locally modified code.
- GoVersion is the version of Go that was used to compile Helm.

When using the --template flag the following properties are available to use in
the template:

- .Version contains the semantic version of Helm
- .GitCommit is the git commit
- .GoVersion contains the version of Go that Helm was compiled with
`

type versionOptions struct {
	short    bool
	template string
}

func newVersionCmd(logger log.Logger) *cobra.Command {
	o := &versionOptions{}

	cmd := &cobra.Command{
		Use:   "version",
		Short: "print the client version information",
		Long:  versionDesc,
		Args:  require.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			wInfo := logio.NewWriter(logger, log.InfoLevel)
			return o.run(wInfo)
		},
	}
	f := cmd.Flags()
	f.BoolVar(&o.short, "short", false, "print the version number")
	f.StringVar(&o.template, "template", "", "template for version string format")
	f.BoolP("client", "c", true, "display client version information")
	_ = f.MarkHidden("client")

	return cmd
}

func (o *versionOptions) run(wr io.Writer) error {
	if o.template != "" {
		tt, err := template.New("_").Parse(o.template)
		if err != nil {
			return err
		}
		buf := &bytes.Buffer{}
		_ = tt.Execute(buf, version.Get())
		_, _ = io.Copy(wr, buf)
		return nil

	}
	_, _ = fmt.Fprintln(wr, formatVersion(o.short))
	return nil
}

func formatVersion(short bool) string {
	v := version.Get()
	if short {
		if len(v.GitCommit) >= 7 {
			return fmt.Sprintf("%s+g%s", v.Version, v.GitCommit[:7])
		}
		return version.GetVersion()
	}
	return fmt.Sprintf("%#v", v)
}
