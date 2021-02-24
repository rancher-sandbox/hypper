package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/rancher-sandbox/hypper/pkg/hypperpath"
	"helm.sh/helm/v3/cmd/helm/require"
	"helm.sh/helm/v3/pkg/repo"
)

type repoRemoveOptions struct {
	names     []string
	repoFile  string
	repoCache string
}

func newRepoRemoveCmd(out io.Writer) *cobra.Command {
	o := &repoRemoveOptions{}

	cmd := &cobra.Command{
		Use:     "remove [REPO1 [REPO2 ...]]",
		Aliases: []string{"rm"},
		Short:   "remove one or more chart repositories",
		Args:    require.MinimumNArgs(1),
		//ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		//	return compListRepos(toComplete, args), cobra.ShellCompDirectiveNoFileComp
		//},
		RunE: func(cmd *cobra.Command, args []string) error {
			o.repoFile = settings.RepositoryConfig
			o.repoCache = settings.RepositoryCache
			o.names = args
			return o.run(out)
		},
	}
	return cmd
}

func (o *repoRemoveOptions) run(out io.Writer) error {
	r, err := repo.LoadFile(o.repoFile)
	if isNotExist(err) || len(r.Repositories) == 0 {
		return errors.New("no repositories configured")
	}

	for _, name := range o.names {
		if !r.Remove(name) {
			return errors.Errorf("no repo named %q found", name)
		}
		if err := r.WriteFile(o.repoFile, 0644); err != nil {
			return err
		}

		if err := removeRepoCache(o.repoCache, name); err != nil {
			return err
		}
		fmt.Fprintf(out, "%q has been removed from your repositories\n", name)
	}

	return nil
}

func removeRepoCache(root, name string) error {
	idx := filepath.Join(root, hypperpath.CacheChartsFile(name))
	if _, err := os.Stat(idx); err == nil {
		os.Remove(idx)
	}

	idx = filepath.Join(root, hypperpath.CacheIndexFile(name))
	if _, err := os.Stat(idx); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return errors.Wrapf(err, "can't remove index file %s", idx)
	}
	return os.Remove(idx)
}
