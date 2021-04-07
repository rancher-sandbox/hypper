/*
Copyright The Helm Authors.

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
	"github.com/Masterminds/log-go"
	"github.com/rancher-sandbox/hypper/pkg/search"
	"github.com/spf13/cobra"
)

const searchDesc = `
Search provides the ability to search for Hypper charts in the various places
they can be stored including repositories you have added.
Use search subcommands to search different locations for charts.
`

const searchRepoDesc = `
Search reads through all of the repositories configured on the system, and
looks for matches. Search of these repositories uses the metadata stored on
the system.

It will display the latest stable versions of the charts found. If you
specify the --devel flag, the output will include pre-release versions.
If you want to search using a version constraint, use --version.

Examples:

    # Search for stable release versions matching the keyword "nginx"
    $ hypper search repo nginx

    # Search for release versions matching the keyword "nginx", including pre-release versions
    $ hypper search repo nginx --devel

    # Search for the latest stable release for nginx-ingress with a major version of 1
    $ hypper search repo nginx-ingress --version ^1.0.0

Repositories are managed with 'hypper repo' commands.
`

// newSearchCmd is the generic search command that can contain subcommands
func newSearchCmd(logger log.Logger) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "search [keyword]",
		Short: "search for a keyword in charts",
		Long:  searchDesc,
	}

	cmd.AddCommand(newSearchRepoCmd(logger))

	return cmd
}

// newSearchRepoCmd search on repos
func newSearchRepoCmd(logger log.Logger) *cobra.Command {
	o := &search.RepoOptions{}

	cmd := &cobra.Command{
		Use:   "repo [keyword]",
		Short: "search repositories for a keyword in charts",
		Long:  searchRepoDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			o.RepoFile = settings.RepositoryConfig
			o.RepoCacheDir = settings.RepositoryCache
			return o.Run(logger, args)
		},
	}

	f := cmd.Flags()
	f.BoolVarP(&o.Regexp, "regexp", "r", false, "use regular expressions for searching repositories you have added")
	f.BoolVarP(&o.Versions, "versions", "l", false, "show the long listing, with each version of each chart on its own line, for repositories you have added")
	f.BoolVar(&o.Devel, "devel", false, "use development versions (alpha, beta, and release candidate releases), too. Equivalent to version '>0.0.0-0'. If --version is set, this is ignored")
	f.StringVar(&o.Version, "version", "", "search using semantic versioning constraints on repositories you have added")
	f.UintVar(&o.MaxColWidth, "max-col-width", 50, "maximum column width for output table")
	bindOutputFlag(cmd, &o.OutputFormat)

	return cmd
}
