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
	"bufio"
	"os"
	"strings"

	"github.com/pkg/errors"

	"github.com/Masterminds/log-go"
	logio "github.com/Masterminds/log-go/io"
	"github.com/Masterminds/semver/v3"
	"github.com/jinzhu/copier"

	pkg "github.com/rancher-sandbox/hypper/internal/package"
	"github.com/rancher-sandbox/hypper/internal/solver"
	"github.com/rancher-sandbox/hypper/pkg/cli"
	"github.com/rancher-sandbox/hypper/pkg/eyecandy"

	"github.com/rancher-sandbox/hypper/internal/third-party/helm/resolver"
	"github.com/rancher-sandbox/hypper/pkg/repo"
	"helm.sh/helm/v3/pkg/action"
	helmChart "helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/release"
)

// OptionalDepsStrategy defines a strategy for determining wether to use optional deps
type optionalDepsStrategy int

const (
	// OptionalDepsAll will use all the optional deps
	OptionalDepsAll optionalDepsStrategy = iota
	// OptionalDepsAsk will interactively prompt on each optional dep
	OptionalDepsAsk
	// OptionalDepsNone with skip all the optional deps
	OptionalDepsNone
)

// Install is a composite type of Helm's Install type
type Install struct {
	*action.Install

	// Hypper specific:
	NoSharedDeps      bool
	OptionalDeps      optionalDepsStrategy
	NoCreateNamespace bool

	// Config stores the actionconfig so it can be retrieved and used again
	Config *Configuration
}

// NewInstall creates a new Install object with the given configuration,
// by wrapping action.NewInstall
func NewInstall(cfg *Configuration) *Install {
	return &Install{
		Install: action.NewInstall(cfg.Configuration),
		Config:  cfg,
	}
}

// CheckDependencies checks the dependencies for a chart.
// by wrapping action.CheckDependencies
func CheckDependencies(ch *helmChart.Chart, reqs []*helmChart.Dependency) error {
	return action.CheckDependencies(ch, reqs)
}

// Run executes the installation
//
// If DryRun is set to true, this will prepare the release, but not install it.
// It returns a slice of releases deployed to the cluster.
//
// It will create a DB of packages from all known charts in repos, releases and
// desired ones. Then, it will solve with the SAT solver, and if relevant,
// install the wanted chart and its dependencies. If dependencies are already
// satisfied, they will be silently skipped.
func (i *Install) Run(strategy solver.SolverStrategy, wantedChrt *helmChart.Chart, vals map[string]interface{},
	settings *cli.EnvSettings, logger log.Logger) ([]*release.Release, error) {

	// TODO obtain lock
	// defer release lock

	// get all releases
	clientInstallForGetRels := NewInstall(i.Config)
	// do a deep copy, in case install struct changes in the future:
	if err := copier.Copy(&clientInstallForGetRels, &i); err != nil {
		return nil, err
	}
	rels, err := clientInstallForGetRels.GetAllReleases()
	if err != nil {
		return nil, err
	}

	// honour settings.NamespaceFromFlag:
	SetNamespace(i, wantedChrt, settings.Namespace(), settings.NamespaceFromFlag)

	if i.ReleaseName == "" {
		// no release provided, obtain it from annotations or chart name
		var err error
		i.ReleaseName, err = GetName(wantedChrt, i.NameTemplate)
		if err != nil {
			return nil, err
		}
	}

	// create pkg with chart to be installed:
	version := i.ChartPathOptions.Version
	pinnedVer := pkg.Unknown
	if i.Version == "" {
		// no pinned ver, take the chart as a filler for fp:
		version = wantedChrt.Metadata.Version
	} else {
		pinnedVer = pkg.Present
	}

	wantedPkg := pkg.NewPkg(i.ReleaseName, wantedChrt.Metadata.Name, version, i.Namespace,
		pkg.Unknown, pkg.Present, pinnedVer, i.ChartPathOptions.RepoURL)

	// get all repo entries, continue if there's none:
	rf, err := repo.LoadFile(settings.EnvSettings.RegistryConfig)
	if err != nil {
		if !os.IsNotExist(errors.Cause(err)) {
			return nil, err
		}
		logger.Debug("No repository present, continuing…")
	}

	s := solver.New(strategy, logger)

	err = i.BuildWorld(s.PkgDB, rf.Repositories, rels, wantedPkg, wantedChrt, settings, logger)
	if err != nil {
		return nil, err
	}

	s.PkgDB.DebugPrintDB(logger)

	// Promote optional deps to normal deps, depending on the strategy selected:
	// TODO use wantedPkg instead of wantedPkgInDB once wantedPkg from local chart gets depRel correctly built
	wantedPkgInDB := solver.PkgDBInstance.GetPackageByFingerprint(wantedPkg.GetFingerPrint())
	switch i.OptionalDeps {
	case OptionalDepsAll:
		logger.Debugf("Promoting all optional deps of package %s to normal deps\n", wantedPkgInDB.GetFingerPrint())
		// promote all optional deps of wanted package to normal deps:
		wantedPkgInDB.DependsRel = append(wantedPkgInDB.DependsRel, wantedPkgInDB.DependsOptionalRel...)
	case OptionalDepsNone:
		logger.Debugf("Disregarding all optional deps of package %s\n", wantedPkgInDB.GetFingerPrint())
	case OptionalDepsAsk:
		logger.Debugf("Asking for each optional deps of package %s if they should be promoted\n", wantedPkgInDB.GetFingerPrint())
		for _, rel := range wantedPkgInDB.DependsOptionalRel {
			reader := bufio.NewReader(os.Stdin)
			question := eyecandy.ESPrintf(settings.NoEmojis,
				":red_question_mark:Install optional shared dependency \"%s\" of chart \"%s\"?",
				rel.ReleaseName,
				wantedPkgInDB.ChartName,
			)
			if promptBool(question, reader, logger) {
				wantedPkgInDB.DependsRel = append(wantedPkgInDB.DependsRel, rel)
			}
		}
	}

	// s.PkgDB.DebugPrintDB(logger)

	s.Solve(wantedPkgInDB)

	if s.IsSAT() {
		installedRels, err := i.postOrderInstall(s.PkgResultSet.ToInstall, wantedPkgInDB, wantedChrt, vals, settings, logger)
		if err != nil {
			return installedRels, err
		}
		return installedRels, nil
	} else {
		// UNSAT, error with inconsistencies
		incons := ""
		for _, incon := range s.PkgResultSet.Inconsistencies {
			incons = incons + incon
		}
		return make([]*release.Release, 0), errors.New(incons)
	}
}

// postOrderInstall traverses the dependency tree in post-order, and calls for
// installation of packages. This will install the dependencies of a chart
// before the chart itself.
func (i *Install) postOrderInstall(tr *solver.PkgTree,
	wantedPkg *pkg.Pkg, wantedChrt *helmChart.Chart, vals map[string]interface{},
	settings *cli.EnvSettings, logger log.Logger) (installedRels []*release.Release, err error) {

	// recursively install dependencies of node:
	for _, depTR := range tr.Relations {
		// if first dep and we have dep, print info:
		// for all deps, call recursively:
		installedDeps, err := i.postOrderInstall(depTR, wantedPkg, wantedChrt, vals, settings, logger)
		if err != nil {
			return installedDeps, err
		}
		installedRels = append(installedRels, installedDeps...)
	}

	// install node:
	if i.NoSharedDeps && tr.Node.GetFingerPrint() != wantedPkg.GetFingerPrint() {
		// skip if node is a dependency and not wantedPkg:
		logger.Infof(eyecandy.ESPrintf(settings.NoEmojis, ":next_track_button: Skipping dependency \"%s\", flag `no-shared-deps` has been set",
			tr.Node.ChartName))
	} else {
		rel, err := i.InstallPkg(tr.Node, wantedPkg, wantedChrt, vals, 0, settings, logger)
		if err != nil {
			return installedRels, err
		}
		installedRels = append(installedRels, rel)
	}

	return installedRels, err
}

// Chart returns the chart that should be used.
//
// This will read the flags and skip args if necessary.
func (i *Install) Chart(args []string) (string, error) {
	if len(args) > 2 {
		return args[1], errors.Errorf("expected at most two arguments, unexpected arguments: %v", strings.Join(args[2:], ", "))
	}

	if len(args) == 2 {
		return args[1], nil
	}

	return args[0], nil
}

// NameAndChart overloads Helm's NameAndChart. It always fails.
//
// On Hypper, we need to read the chart annotations to know the correct release name.
// Therefore, it cannot happen in this function.
func (i *Install) NameAndChart(args []string) (string, string, error) {
	return "", "", errors.New("NameAndChart() cannot be used")
}

// checkIfInstallable validates if a chart can be installed
//
// Application chart type is only installable
func CheckIfInstallable(ch *helmChart.Chart) error {
	switch ch.Metadata.Type {
	case "", "application":
		return nil
	}
	return errors.Errorf("%s charts are not installable", ch.Metadata.Type)
}

// GetAllReleases obtains the releases in all namespaces that we have access to.
func (i *Install) GetAllReleases() (releases []*release.Release, err error) {
	// use empty namespace, to target all namespaces the kubeclient has access to:
	i.Config.SetNamespace("")
	return i.GetReleases()
}

// GetReleases obtains the releases in the specific namespace that we are searching into.
func (i *Install) GetReleases() (releases []*release.Release, err error) {
	clientList := NewList(i.Config)
	clientList.SetStateMask()
	releases, err = clientList.Run()
	if err != nil {
		return nil, err
	}
	return releases, nil
}

func (i *Install) LoadChart(chartName, repo, version string,
	settings *cli.EnvSettings, logger log.Logger) (*helmChart.Chart, error) {

	if strings.HasPrefix(repo, "file://") {
		logger.Debugf("Repository from local path: %s\n", repo)

		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		localPath, err := resolver.GetLocalPath(repo, wd)
		if err != nil {
			return nil, err
		}
		chartRequested, err := loader.LoadDir(localPath)
		if err != nil {
			return nil, err
		}

		constraint, err := semver.NewConstraint(version)
		if err != nil {
			return nil, errors.Wrapf(err, "dependency \"%s\" has an invalid version/constraint format", chartName)
		}

		v, err := semver.NewVersion(chartRequested.Metadata.Version)
		if err != nil {
			return nil, err
		}

		if constraint.Check(v) {
			return chartRequested, err
		}

		return nil, errors.Errorf("can't get a valid version for dependency %s", chartName)
	}

	i.ChartPathOptions.RepoURL = repo
	i.ChartPathOptions.Version = version
	cp, err := i.ChartPathOptions.LocateChart(chartName, settings.EnvSettings)
	if err != nil {
		return nil, err
	}

	logger.Debugf("CHART PATH: %s\n", cp)

	chartRequested, err := loader.Load(cp)
	if err != nil {
		return nil, err
	}
	return chartRequested, nil
}

// InstallPkg installs the passed package by pulling its related chart. It takes
// care of using the desired namespace for it.
func (i *Install) InstallPkg(p *pkg.Pkg, wantedPkg *pkg.Pkg, wantedChart *helmChart.Chart,
	vals map[string]interface{}, lvl int,
	settings *cli.EnvSettings, logger log.Logger) (*release.Release, error) {
	// FIXME don't pass wantedPkg and wantedChart and skip things more cleanly

	logger.Debug("Installing package: " + p.String())

	clientInstall := NewInstall(i.Config)
	// we need to automatically satisfy all install options (i.CreateNamespace,
	// i.DryRun, etc) when we are installing the dep using clientInstall. Doing
	// a shallow copy sounds like asking for trouble when the install struct
	// changes, so let's do a deep copy instead:
	if err := copier.Copy(&clientInstall, &i); err != nil {
		return nil, err
	}

	var chartRequested *helmChart.Chart
	var chartpath string
	if p.GetFingerPrint() == wantedPkg.GetFingerPrint() { // wantedPkg
		// don't load chart, we already have it
		chartRequested = wantedChart
	} else { // dependency
		// we don't have a chart, load it
		var err error
		chartRequested, err = clientInstall.LoadChart(p.ChartName, p.Repository, p.Version,
			settings, logger)
		if err != nil {
			return nil, err
		}
		// default to empty vals:
		vals = make(map[string]interface{})
	}

	getter := getter.All(settings.EnvSettings)

	logger.Debugf("Original chart version: %q", chartRequested.Metadata.Version)
	if clientInstall.Devel {
		logger.Debug("setting version to >0.0.0-0")
		clientInstall.Version = ">0.0.0-0"
	}

	// Set Namespace, Releasename for the install client without reevaluating them
	// from the dependent:
	SetNamespace(clientInstall, chartRequested, p.Namespace, settings.NamespaceFromFlag)
	clientInstall.ReleaseName = p.ReleaseName

	if err := CheckIfInstallable(chartRequested); err != nil {
		return nil, err
	}

	if chartRequested.Metadata.Deprecated {
		logger.Warnf("Chart \"$s\" is deprecated", chartRequested.Name())
	}

	wInfo := logio.NewWriter(logger, log.InfoLevel)

	// Check chart dependencies to make sure all are present in /charts
	if req := chartRequested.Metadata.Dependencies; req != nil {
		// If CheckDependencies returns an error, we have unfulfilled dependencies.
		// As of Helm 2.4.0, this is treated as a stopping condition:
		// https://github.com/helm/helm/issues/2209
		if err := action.CheckDependencies(chartRequested, req); err != nil {
			if clientInstall.DependencyUpdate {
				man := &downloader.Manager{
					Out:              wInfo,
					ChartPath:        chartpath,
					Keyring:          clientInstall.ChartPathOptions.Keyring,
					SkipUpdate:       false,
					Getters:          getter,
					RepositoryConfig: settings.RepositoryConfig,
					RepositoryCache:  settings.RepositoryCache,
					Debug:            settings.Debug,
				}
				if err := man.Update(); err != nil {
					return nil, err
				}
				// Reload the chart with the updated Chart.lock file.
				if chartRequested, err = loader.Load(chartpath); err != nil {
					return nil, errors.Wrap(err, "failed reloading chart after repo update")
				}
			} else {
				return nil, err
			}
		}
	}

	// pretty print:
	logger.Infof(eyecandy.ESPrintf(settings.NoEmojis, ":cruise_ship: %sInstalling chart \"%s\" as \"%s\" in namespace \"%s\"…",
		strings.Repeat("  ", lvl), chartRequested.Name(), clientInstall.ReleaseName, clientInstall.Namespace))

	// perform instal:
	helmInstall := clientInstall.Install
	i.Config.SetNamespace(clientInstall.Namespace)
	rel, err := helmInstall.Run(chartRequested, vals) // wrap Helm's i.Run for now
	if err != nil {
		return rel, err
	}
	return rel, nil
}

func promptBool(question string, reader *bufio.Reader, logger log.Logger) bool {
	for {
		log.Infof("%s [Y/n]:", question)

		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		response = strings.ToLower(strings.TrimSpace(response))

		if response == "y" || response == "yes" || response == "" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
	}
}
