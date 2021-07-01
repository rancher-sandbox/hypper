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

// Pkg is the minimum object the solver reasons about. It is comprised of a
// chart, its version, and its characteristics when installed (release name,
// namespace, etc).
// Note that each package is unique. The same chart, with the same default
// release name & ns, but different version, is a different package. E.g:
// prometheus-1.2.0 and prometheus-1.3.0 are different packages.
type Pkg struct {
	ID                 int  `json:"-" yaml:"-"`     // ID, position on the solver model
	ReleaseName        string    // Release name, or default chart release-name
	ChartName          string    // chart name
	Version            string    // sem ver (without a range)
	Namespace          string    // Installed ns, or default chart namespace
	DependsRel         []*PkgRel // list of dependencies' fingerprints
	DependsOptionalRel []*PkgRel // list of optional dependencies' fingerprints
	CurrentState       tristate  // current state of the package
	DesiredState       tristate  // desired state of the package
	Repository         string
}

type PkgRel struct {
	BaseFingerprint string // base fingerprint of dependency with releasename, namespace
	SemverRange     string // e.g: 1.0.0, ~1.0.0
}

func NewPkg(relName, chartName, version, namespace string,
	currentState, desiredState tristate, repo string) *Pkg {

	p := &Pkg{
		ID:                 -1,
		ReleaseName:        relName,
		ChartName:          chartName,
		Version:            version,
		Namespace:          namespace,
		DependsRel:         []*PkgRel{},
		DependsOptionalRel: []*PkgRel{},
		CurrentState:       currentState,
		DesiredState:       desiredState,
		Repository:         repo,
	}

	return p
}

// NewPkgMock creates a new package, with a digest based in the package name,
// and a nil chart pointer.
// Useful for testing.
func NewPkgMock(name, version, namespace string,
	depends, dependsOptional []*PkgRel,
	currentState, desiredState tristate) *Pkg {

	p := NewPkg(name, name, version, namespace, currentState, desiredState, "ourrepo")

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
	return fmt.Sprintf("%s-%s-%s", p.ReleaseName, p.Version, p.Namespace)
}

func CreateFingerPrint(releaseName, version, ns string) string {
	return fmt.Sprintf("%s-%s-%s", releaseName, version, ns)
}

// GetBaseFingerPrint returns a unique id of the package minus version.
// This helps when filtering packages to find those that are similar and differ
// only in the version.
func (p *Pkg) GetBaseFingerPrint() string {
	return fmt.Sprintf("%s-%s", p.ReleaseName, p.Namespace)
}

// CreateBaseFingerPrint returns a base fingerprint (name-ns)
func CreateBaseFingerPrint(name, ns string) string {
	return fmt.Sprintf("%s-%s", name, ns)
}

// Encode encodes the package to string.
// It returns an ID which can be used to retrieve the package later on.
func (p *Pkg) Encode() (string, error) {

	encodedPackage, err := p.JSON()
	if err != nil {
		return "", err
	}

	return string(encodedPackage), nil
}
