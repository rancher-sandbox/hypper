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

/*
Solver provides operations on packages: install, remove, upgrade,
check integrity and others.

A package is an object comprised of a unique key (tentative release name,
version, namespace, chart name), and digested information about the chart that
it relates to (dependency relations, chart and repo URL, current and desired
state..).

To perform a package operation, for example, "install packageA", we:

 1. Build a database of all packages in the world, which contains:
 - Packages deployed in cluster (releases).
 - Packages in the known repositories.
 - Requested changes to packages (to install, to remove, to upgrade).

 The gophersat/solver MAXSAT/Pseudo-Boolean solver operates over unique strings:
 in our case, the package string fingerprint, created from its unique key.

 The database contains information on a package current state (unknown,
 installed, removed) and desired state (unknown, installed, removed).
 The database can also be queried to obtain a list of packages that differ
 only in the version.

 Adding packages to the database can happen in any order (e.g: first toModify,
 later releases, and at last repos). This means that db.Add() will intelligently
 merge new information into the package in the db, if the package is already
 present.

 2. Iterate through the package database and create pseudo-boolean
 constraints for the package fingerprint:
 - If package needs to be installed or not
 - If it depends on another package(s)
 - If it conflicts with other similar packages that differ with it only in
   version).
 - If we want to minimize or maximize the distance between present version
  and wanted version (upgrade to major versions, never upgrade, etc)

 3. Find a solution to the SAT dependency problem if exists, or the
 contradiction if there's no solution.
 The result is a list of tuple of:
 - Fingerprints (each corresponding with a package), and
 - Resulting state of the package (if the package should be present in the
   system or not).

 4. We then iterate through the result list, separating the packages into
 different sets by checking their current, desired, and resulting state:
 unchanged packages, packages to install, packages to remove.

*/
package solver
