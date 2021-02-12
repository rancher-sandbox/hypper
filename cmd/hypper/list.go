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

package main

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"

	"helm.sh/helm/v3/cmd/helm/require"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli/output"
	"helm.sh/helm/v3/pkg/release"

	"github.com/mattfarina/log-go"
	logio "github.com/mattfarina/log-go/io"
)

var listHelp = `
List all of the releases for a specified namespace, by wrapping Helm for now.
`

func newListCmd(cfg *action.Configuration, logger log.Logger) *cobra.Command {
	client := action.NewList(cfg)
	var outfmt output.Format

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "list releases",
		Long:    listHelp,
		Aliases: []string{"ls"},
		Args:    require.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if client.AllNamespaces {
				if err := cfg.Init(settings.RESTClientGetter(), "", os.Getenv("HELM_DRIVER"), logger.Debugf); err != nil {
					return err
				}
			}
			client.SetStateMask()

			results, err := client.Run()
			if err != nil {
				return err
			}

			// Get an io.Writer compliant logger instance at the info level.
			wInfo := logio.NewWriter(logger, log.InfoLevel)

			if client.Short {

				names := make([]string, 0)
				for _, res := range results {
					names = append(names, res.Name)
				}

				outputFlag := cmd.Flag("output")

				switch outputFlag.Value.String() {
				case "json":
					err = output.EncodeJSON(wInfo, names)
					if err != nil {
						logger.Error(err)
					}
					return nil
				case "yaml":
					err = output.EncodeYAML(wInfo, names)
					if err != nil {
						logger.Error(err)
					}
					return nil
				case "table":
					for _, res := range results {
						logger.Info(res.Name)
					}
					return nil
				default:
					return outfmt.Write(wInfo, newReleaseListWriter(results, client.TimeFormat))
				}
			}

			return outfmt.Write(wInfo, newReleaseListWriter(results, client.TimeFormat))
		},
	}

	f := cmd.Flags()
	f.BoolVarP(&client.Short, "short", "q", false, "output short (quiet) listing format")
	f.StringVar(&client.TimeFormat, "time-format", "", "format time. Example: --time-format 2009-11-17 20:34:10 +0000 UTC")
	f.BoolVarP(&client.ByDate, "date", "t", false, "sort by release date")
	f.BoolVarP(&client.SortReverse, "reverse", "r", false, "reverse the sort order")
	f.BoolVarP(&client.All, "all", "a", false, "show all releases without any filter applied")
	f.BoolVar(&client.Uninstalled, "uninstalled", false, "show uninstalled releases (if 'helm uninstall --keep-history' was used)")
	f.BoolVar(&client.Superseded, "superseded", false, "show superseded releases")
	f.BoolVar(&client.Uninstalling, "uninstalling", false, "show releases that are currently being uninstalled")
	f.BoolVar(&client.Deployed, "deployed", false, "show deployed releases. If no other is specified, this will be automatically enabled")
	f.BoolVar(&client.Failed, "failed", false, "show failed releases")
	f.BoolVar(&client.Pending, "pending", false, "show pending releases")
	f.BoolVarP(&client.AllNamespaces, "all-namespaces", "A", false, "list releases across all namespaces")
	f.IntVarP(&client.Limit, "max", "m", 256, "maximum number of releases to fetch")
	f.IntVar(&client.Offset, "offset", 0, "next release name in the list, used to offset from start value")
	f.StringVarP(&client.Filter, "filter", "f", "", "a regular expression (Perl compatible). Any releases that match the expression will be included in the results")
	f.StringVarP(&client.Selector, "selector", "l", "", "Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2). Works only for secret(default) and configmap storage backends.")
	bindOutputFlag(cmd, &outfmt)

	return cmd
}

type releaseElement struct {
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	Revision   string `json:"revision"`
	Updated    string `json:"updated"`
	Status     string `json:"status"`
	Chart      string `json:"chart"`
	AppVersion string `json:"app_version"`
}

type releaseListWriter struct {
	releases []releaseElement
}

func newReleaseListWriter(releases []*release.Release, timeFormat string) *releaseListWriter {
	// Initialize the array so no results returns an empty array instead of null
	elements := make([]releaseElement, 0, len(releases))
	for _, r := range releases {
		element := releaseElement{
			Name:       r.Name,
			Namespace:  r.Namespace,
			Revision:   strconv.Itoa(r.Version),
			Status:     r.Info.Status.String(),
			Chart:      fmt.Sprintf("%s-%s", r.Chart.Metadata.Name, r.Chart.Metadata.Version),
			AppVersion: r.Chart.Metadata.AppVersion,
		}

		t := "-"
		if tspb := r.Info.LastDeployed; !tspb.IsZero() {
			if timeFormat != "" {
				t = tspb.Format(timeFormat)
			} else {
				t = tspb.String()
			}
		}
		element.Updated = t

		elements = append(elements, element)
	}
	return &releaseListWriter{elements}
}

func (r *releaseListWriter) WriteTable(out io.Writer) error {
	table := uitable.New()
	table.AddRow("NAME", "NAMESPACE", "REVISION", "UPDATED", "STATUS", "CHART", "APP VERSION")
	for _, r := range r.releases {
		table.AddRow(r.Name, r.Namespace, r.Revision, r.Updated, r.Status, r.Chart, r.AppVersion)
	}
	return output.EncodeTable(out, table)
}

func (r *releaseListWriter) WriteJSON(out io.Writer) error {
	return output.EncodeJSON(out, r.releases)
}

func (r *releaseListWriter) WriteYAML(out io.Writer) error {
	return output.EncodeYAML(out, r.releases)
}
