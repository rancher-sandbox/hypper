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

package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type tristate int

const (
	Unknown tristate = iota
	Present
	Absent
)

// Pkg is the minimum object the solver reasons about.
// Note that each package is unique. The same chart, with the same default
// release name & ns, but different version, is a different package. E.g:
// prometheus-1.2.0 and prometheus-1.3.0 are different packages.
type Pkg struct {
	// unique key:
	ReleaseName string // Release name, or default chart release-name
	Version     string // sem ver (without a range)
	Namespace   string // Installed ns, or default chart namespace
	ChartName   string // chart name

	DependsRel         []*PkgRel // list of dependencies' fingerprints
	DependsOptionalRel []*PkgRel // list of optional dependencies' fingerprints
	Repository         string    // repository where to find ChartName
	ParentChartPath    string    // Absolute path of parent chart, if existing
	CurrentState       tristate  // current state of the package
	DesiredState       tristate  // desired state of the package
	PinnedVer          tristate  // if we have a pinnedVer or not in pkg.Version
}

// PkgRel codifies a shared dependency relation to another package
type PkgRel struct {
	// unique key for base fingerprint:
	ReleaseName string
	Namespace   string
	SemverRange string // e.g: 1.0.0, ~1.0.0, ^1.0.0
	ChartName   string
}

// NewPkg creates a new Pkg struct. It does not give value to DependsRel,
// DependsOptionalRel.
func NewPkg(relName, chartName, version, namespace string,
	currentState, desiredState, pinnedVer tristate,
	repo, parentChartPath string) *Pkg {

	p := &Pkg{
		ReleaseName:        relName,
		Version:            version,
		Namespace:          namespace,
		ChartName:          chartName,
		DependsRel:         []*PkgRel{},
		DependsOptionalRel: []*PkgRel{},
		Repository:         repo,
		ParentChartPath:    parentChartPath,
		CurrentState:       currentState,
		DesiredState:       desiredState,
		PinnedVer:          pinnedVer,
	}

	return p
}

// NewPkgMock creates a new package.
// Useful for testing.
func NewPkgMock(name, version, namespace string,
	depends, dependsOptional []*PkgRel,
	currentState, desiredState tristate) *Pkg {

	p := NewPkg(name, name, version, namespace,
		currentState, desiredState, Unknown, "ourrepo", "")

	p.DependsRel = depends
	p.DependsOptionalRel = dependsOptional

	return p
}

// JSON serializes package p into JSON, returning a []byte
func (p *Pkg) JSON() ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(p)
	return buffer.Bytes(), err
}

// GetFingerPrint returns a unique id of the package.
func (p *Pkg) GetFingerPrint() string {
	return fmt.Sprintf("%s_%s_%s_%s", p.ReleaseName, p.Version, p.Namespace, p.ChartName)
}

// CreateFingerPrint creates a fingerprint from the specific strings.
func CreateFingerPrint(releaseName, version, ns, chartName string) string {
	return fmt.Sprintf("%s_%s_%s_%s", releaseName, version, ns, chartName)
}

// GetBaseFingerPrint returns a unique id of the package minus version.
// This helps when filtering packages to find those that are similar and differ
// only in the version.
func (p *Pkg) GetBaseFingerPrint() string {
	return fmt.Sprintf("%s_%s_%s", p.ReleaseName, p.Namespace, p.ChartName)
}

// CreateBaseFingerPrint returns a base fingerprint
func CreateBaseFingerPrint(releaseName, ns, chartName string) string {
	return fmt.Sprintf("%s_%s_%s", releaseName, ns, chartName)
}

// Encode encodes the package to string in a JSON.
func (p *Pkg) Encode() (string, error) {

	encodedPackage, err := p.JSON()
	if err != nil {
		return "", err
	}

	return string(encodedPackage), nil
}

// String returns a string of some of the Pkg fields.
// Useful for debug output.
func (p *Pkg) String() (retString string) {
	retString = retString +
		fmt.Sprintf("fp: %s Currentstate: %v DesiredState: %v PinnedVersion: %v ChartName: %s\n",
			p.GetFingerPrint(),
			p.CurrentState, p.DesiredState, p.PinnedVer,
			p.ChartName,
		)
	for _, rel := range p.DependsRel {
		retString = retString + fmt.Sprintf("          DepRel: %v\n", rel)
	}
	for _, rel := range p.DependsOptionalRel {
		retString = retString + fmt.Sprintf("          DepOptionalRel: %v\n", rel)
	}
	return
}
