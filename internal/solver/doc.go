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

A package is an object comprised of a chart, its version, and its
characteristics when installed (release name, namespace, etc). See the pkg
package.

To perform a package operation, for example, install packageA, we:

 1. Build a database of all packages in the world, which contains:
 - packages deployed in cluster (releases).
 - packages in the known repositories.
 - requested changes to packages (to install, to remove, to upgrade).
 The gophersat/solver Pseudo-Boolean solver operates over IDs that are
 integer. Hence, the database maps packages with consecutive unique IDs.
 The database contains information on a package current state (unknown,
 installed, removed) and desired state (unknown, installed, removed).
 The database can also be queried to obtain a list of packages that differ
 only in the version.
 The solver library operates with those integer IDs, and one needs to know
 all possible IDs before generating constraints, as for example one can be
 adding a "depends" constraint of a package that hasn't had a constraint
 added yet.
 For simplifying the implementation of the solver and constraint creation,
 adding packages to the database can happen in any order (e.g: first toModify,
 later releases, and at last repos). This means that db.Add() will intelligently
 merge new information into the package in the db, if the package is already
 present.

 2. Iterate through the package database and create pseudo-boolean
 constraints for the package ID:
 - If package ID needs to be installed or not
 - If it depends on another package(s) ID(s)
 - If it conflicts with other similar packages that differ with it only in
   version).
 - If we want to minimize or maximize the distance between present version
  and wanted version (upgrade to major versions, never upgrade, etc)

 3. Find a solution to the SAT dependency problem if exists, or the
 contradiction if there's no solution.
 The result is a list of tuple of:
 - IDs (each corresponding with a package), and
 - Resulting state of the package (if the package should be present in the
   system or not).

 4. We then iterate through the result list, separating the packages into
 different sets by checking their current, desired, and resulting state:
 unchanged packages, packages to install, packages to remove.

 TODO how to detect upgrades efficiently, instead of iterating through all
 releases, given that the sat solver informs of an upgrade by providing a
 package to be installed, and a package to be removed.
*/
package solver
