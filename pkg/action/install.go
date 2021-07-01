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
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"

	"github.com/Masterminds/log-go"
	logio "github.com/Masterminds/log-go/io"
	"github.com/jinzhu/copier"

	pkg "github.com/rancher-sandbox/hypper/internal/package"
	"github.com/rancher-sandbox/hypper/internal/solver"
	"github.com/rancher-sandbox/hypper/pkg/cli"
	"github.com/rancher-sandbox/hypper/pkg/eyecandy"

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

	// hypper specific
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
// If DryRun is set to true, this will prepare the release, but not install it
// lvl is used for printing nested stagered output on recursion. Starts at 0.
func (i *Install) Run(chrt *helmChart.Chart, vals map[string]interface{}, settings *cli.EnvSettings, logger log.Logger, lvl int) ([]*release.Release, error) {

	// TODO obtain lock
	// defer release lock

	// create pkg with chart to be installed:
	// TODO chrt.Metadata.Version & repo, are incorrect and irrelevant
	// we are just updating an already existing package from the repos with
	// DesiredState: Present
	wantedPkg := pkg.NewPkg(i.ReleaseName, chrt.Metadata.Name, chrt.Metadata.Version, i.Namespace, pkg.Unknown, pkg.Present, "")

	// FIXME deprels are not taken if wantedPkg is local and not in repos
	// FIXME wantedPkg from repos is not taken correctly: failed to download hypper/our-app

	// get all releases
	rels, err := i.GetAllReleases(settings)
	if err != nil {
		return nil, err
	}

	// get all repo entries
	rf, err := repo.LoadFile(settings.RepositoryConfig)
	if err != nil {
		return nil, err
	}

	s := solver.New()

	err = BuildWorld(s.PkgDB, rf.Repositories, rels, []*pkg.Pkg{wantedPkg}, settings, logger)
	if err != nil {
		return nil, err
	}

	s.PkgDB.DebugPrintDB(logger)

	// Promote optional deps to normal deps, depending on the strategy selected:
	switch i.OptionalDeps {
	case OptionalDepsAll:
		logger.Debugf("Promoting all optional deps of package %s to normal deps\n", wantedPkg.GetFingerPrint())
		// promote all optional deps of wanted package to normal deps:
		for _, rel := range wantedPkg.DependsOptionalRel {
			wantedPkg.DependsRel = append(wantedPkg.DependsRel, rel)
		}
	case OptionalDepsNone:
		logger.Debugf("Disregarding all optional deps of package %s\n", wantedPkg.GetFingerPrint())
	case OptionalDepsAsk:
		logger.Debugf("Asking for each optional deps of package %s if they should be promoted\n", wantedPkg.GetFingerPrint())
		for _, rel := range wantedPkg.DependsOptionalRel {
			reader := bufio.NewReader(os.Stdin)
			question := eyecandy.ESPrintf(settings.NoEmojis,
				":red_question_mark:Install optional shared dependency \"%s\" of chart \"%s\"?",
				rel.BaseFingerprint,
				wantedPkg.ChartName,
			)
			if promptBool(question, reader, logger) {
				wantedPkg.DependsRel = append(wantedPkg.DependsRel, rel)
			}
		}
	}

	s.Solve()

	fmt.Println(s.FormatOutput(solver.Table))

	installedRels := make([]*release.Release, 0)
	if s.IsSAT() {
		for _, p := range s.PkgResultSet.ToInstall {

			// TODO
			// if i.NoSharedDeps && package.isDependency {
			// 	  // skip this package and not install it
			//    continue
			// }

			// install package:
			rel, err := i.InstallPkg(p, settings, logger)
			if err != nil {
				return installedRels, err
			}
			installedRels = append(installedRels, rel)
		}
	}

	return installedRels, nil
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

	// len(args) == 1
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

func (i *Install) GetReleases() (releases []*release.Release, err error) {
	// obtain the releases for the specific ns that we are searching into
	clientList := NewList(i.Config)
	clientList.SetStateMask()
	releases, err = clientList.Run()
	if err != nil {
		return nil, err
	}
	return releases, nil
}

func (i *Install) GetAllReleases(settings *cli.EnvSettings) (releases []*release.Release, err error) {

	if err := i.Config.Init(settings.RESTClientGetter(), "", os.Getenv("HELM_DRIVER"), i.Config.Log); err != nil {
		return nil, err
	}
	return i.GetReleases()
}

func (i *Install) LoadChartFromPkg(p *pkg.Pkg, settings *cli.EnvSettings, logger log.Logger) (*helmChart.Chart, error) {
	i.ChartPathOptions.RepoURL = p.Repository
	i.ChartPathOptions.Version = p.Version
	cp, err := i.ChartPathOptions.LocateChart(p.ChartName, settings.EnvSettings)
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

func (i *Install) InstallPkg(p *pkg.Pkg, settings *cli.EnvSettings, logger log.Logger) (*release.Release, error) {

	logger.Debugf("Installing package: %v\n", p)

	clientInstall := NewInstall(i.Config)
	// we need to automatically satisfy all install options (i.CreateNamespace,
	// i.DryRun, etc) when we are installing the dep using clientInstall. Doing
	// a shallow copy sounds like asking for trouble when the install struct
	// changes, so let's do a deep copy instead:
	if err := copier.Copy(&clientInstall, &i); err != nil {
		return nil, err
	}

	chartRequested, err := clientInstall.LoadChartFromPkg(p, settings, logger)
	if err != nil {
		return nil, err
	}

	getter := getter.All(settings.EnvSettings)
	vals := make(map[string]interface{}) // TODO calculate vals instead of {}

	logger.Debugf("Original chart version: %q", chartRequested.Metadata.Version)
	if clientInstall.Devel {
		logger.Debug("setting version to >0.0.0-0")
		clientInstall.Version = ">0.0.0-0"
	}

	// Set Namespace, Releasename for the install client without reevaluating them
	// from the dependent:
	SetNamespace(clientInstall, chartRequested, p.Namespace, settings.NamespaceFromFlag)
	clientInstall.ReleaseName, err = GetName(chartRequested, clientInstall.NameTemplate, p.ReleaseName)
	if err != nil {
		return nil, err
	}

	if err := CheckIfInstallable(chartRequested); err != nil {
		return nil, err
	}

	if chartRequested.Metadata.Deprecated {
		logger.Warnf("Chart \"$s\" is deprecated", chartRequested.Name())
	}

	// re-obtain the chartpath again, for Metadata.Dependencies. FIXME deduplicate.
	clientInstall.ChartPathOptions.RepoURL = p.Repository
	chartpath, err := clientInstall.ChartPathOptions.LocateChart(p.ChartName, settings.EnvSettings)
	if err != nil {
		return nil, err
	}

	logger.Debugf("CHART PATH: %s\n", chartpath)

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

	logger.Infof(eyecandy.ESPrintf(settings.NoEmojis, ":cruise_ship: Installing chart \"%s\" as \"%s\" in namespace \"%s\"…",
		chartRequested.Name(), clientInstall.ReleaseName, clientInstall.Namespace))
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
