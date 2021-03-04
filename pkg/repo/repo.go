/*
Copyright The Helm Authors & SUSE.

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

package repo

import (
	"io/ioutil"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	helmRepo "helm.sh/helm/v3/pkg/repo"
)

// File represents the repositories.yaml file
//
// File is a composite type of helm/pkg/repo.File
type File struct {
	helmRepo.File
}

// NewFile generates an empty repositories file.
//
// Generated and APIVersion are automatically set.
func NewFile() *File {
	helmFile := helmRepo.NewFile()
	return &File{
		*helmFile,
	}
}

// LoadFile takes a file at the given path and returns a File object
func LoadFile(path string) (*File, error) {
	r := new(File)
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return r, errors.Wrapf(err, "couldn't load repositories file (%s)", path)
	}

	err = yaml.Unmarshal(b, &r.File)
	return r, err
}
