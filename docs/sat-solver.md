## Using a SAT solver for solving dependencies

The problem space is defined by a set of packages that depend on each other in
various ways. A **package** is comprised of:
- An implicit **chart**.
- **URI** (repo + name)
- **semver ranges** (e.g: `~`, `^`). They get collapsed into a specific semver
  when deployed.
- Kubernetes **namespace**. Either where to install it, or where are they
  installed, if already installed.
- Defined relations to other packages (**Depends**, **Optional-Depends**,
  **Provides**, **Conflicts**).

Also, not considered now but could be considered in the future:
- **Information about charts containing only CRDs**. Given how CRD objects
  behave on remove and update in the cluster, this may mean to never uninstall
  them in normal operations, and pinning versions for them to not get updated
  automatically.
- **Values.yaml** could be added to the package definition. Changing the values
  set and redeploying the chart means having a package that behaves differently
  (in the same way that ABI changes of usual Linux OS package makes them
  compatible or not with other packages).
- **Pinned** installed packages. Those pinned packages may not be
  uninstalled/changed in version. 
- **Automatically installed** vs manually installed. This allows for example to
  auto-remove packages once nothing depends on them.

Given a set of installed packages in the system, and a set of packages that we
want to be installed, we want to know if there's a set of solutions to it, and
which one is better. Or, we want to know if there is no solution, and why. This
is a satisfiability problem that is NP-hard. 

Some of these package features are a non-issue for resolving the dependency
problem: namespaces are just part of what defines a package, pinned and manually
installed packages just create a new hard constraint to be added, etc.

## Pure SAT, MaxSAT, Pseudo-Boolean SAT problem?

In the past, dependency problems have been solved either by custom heuristics or
by using a SAT solver and pruning results with ad-hoc heuristics after the
resolution. Nowadays, more and more dependency problems are solved either by
reducing the problem space (eg: go modules) or by using a SAT solver, with or
without heuristics to help reduce the problem space.

In our case, we have several strategies we want to apply to our problem, some of
which may be selected at run time. Given the amount of strategies and the
constraints arising from our package definition, we will define our problem as a
Pseudo-Boolean SAT one.
This allows specifying weights in our statements to implement our strategies,
and enables multicriteria optimization.

Also, since the expected number of charts installed in a cluster is going to be
"low" (in the order of 10^1 instead of 10^3 or more that is normally seen for
these problems), we may try to obtain more optimized solutions in a reasonable
time.
In addition, some solvers have the ability to return sub-optimal solutions as
soon as they find them, without waiting for the optimal solution calculation.

#### How does an upgrade operation work in a pure SAT dependency problem?

On traditional Linux OS dependency problems, the satisfiability of the upgrades
is checked after the fact that packages have entered the repositories.

Usually this is accomplished by running tools against the repositories. These
can be SAT solvers, specific scheduled jobs (e.g:
[debcheck](https://qa.debian.org/dose/debcheck.html) &
[paper](https://upsilon.cc/~zack/research/publications/debconf8-mancoosi.pdf),
[debgraph](http://debian.semistable.com/debgraph.out.html) in Debian), CI jobs
for default installations, CI jobs for upgrades between different distribution
releases, etc, in addition to normal QA work.

This is also complemented by gating the package set through several
repositories. E.g:
- OpenSUSE Factory -> Tumbleweed -> Leap, or
- Debian Unstable -> Testing -> Stable

On the distribution receiving new packages (OpenSUSE Factory, Debian Unstable),
transient uninstallability problems are expected. These problems are
increasingly exposed when the same package cannot be installed in several
architectures. On the more stable distributions (OpenSUSE Leap, Debian Stable),
it is expected that all or most of the installability problems have been found.

In Hypper's case, we want the possibility to depend on packages (Helm charts)
outside of our curated repositories. Therefore, we would benefit from knowing if
a system can be upgraded or not, regardless of our ability on curating the
repositories. Hence, we use a SAT solver directly when performing those
upgrades.

## Strategies

- **Minimize package removal**.
- **Minimize number of package charts**.
- **Maximize freshness of packages**. By assigning a distance between package
  versions, we can optimize for those that have the smaller distance. Doing a
  full upgrade of the system means selecting to maximize freshness of all
  installed packages, without asking for installing new packages by default (but
  allowing for packages to be pulled into the installed list).
- **Minimize install of optional packages**.
- **Maximize install of optional packages**.

Not considered right now:
- **Minimize changes to installed packages** (values.yaml, CRD charts).

## Encoding the Pseudo-Boolean Optimization problem

<!-- Showing latex formulas in GH markdown: -->
<!-- https://gist.github.com/a-rodin/fef3f543412d6e1ec5b6cf55bf197d7b -->
<!-- <img src="https://render.githubusercontent.com/render/math?math=foo"> -->
<!-- Replace plus signs with %2B. -->

## Pseudo-Boolean Optimization

We will encode the problem as a linear Pseudo-Boolean Optimization one. This
kind of problem is an optimization problem in
<img src="https://render.githubusercontent.com/render/math?math=f: \{0,1\}^n \mapsto \Re">,
Over _n_ Boolean variables
<img src="https://render.githubusercontent.com/render/math?math=x_{1}, \dots ,x_{n}">.
In our case, whether package _x_ is present (installed) in the final solution.
The optimization problem form is in the following function to minimize:

<img src="https://render.githubusercontent.com/render/math?math=min \sum_{j \in N} c_{j}*x_{j}">

subject to:

<img src="https://render.githubusercontent.com/render/math?math=\sum_{j \in N} a_{ij}*l_{j} \geq b_{i}">

<img src="https://render.githubusercontent.com/render/math?math=x_{j} \in \{0,1\}, a_{ij},b_{i}, c_{j} \in N_{0}^{+}, i \in M">

<img src="https://render.githubusercontent.com/render/math?math=M = 1, \dots, m">

Where:

- Each <img src="https://render.githubusercontent.com/render/math?math=c_{j}">
  is a non-negative integer cost, associated with variable
  <img src="https://render.githubusercontent.com/render/math?math=x_{j}, j \in N">.
- <img src="https://render.githubusercontent.com/render/math?math=a_{ij}">
  denotes the coefficients of the literals
  <img src="https://render.githubusercontent.com/render/math?math=l_{j}">
  in the set m of linear constraints (a literal being a Boolean variable).

In our case, a package _x_ belongs to the repository _P_, where _P_ is the set
of tuples
<img src="https://render.githubusercontent.com/render/math?math=P = \{ (repoUrl, name, semverRange, ns), \dots \}">.

## Constraints definitions

We will define Pseudo-Boolean functions that will be optimized by selecting
their input values. This Pseudo-Boolean functions can then be implemented in
Hypper easily.

### Installation

<img src="https://render.githubusercontent.com/render/math?math=x_{1}"> is installed:

<img src="https://render.githubusercontent.com/render/math?math=x_{1} \geq 1">,
with <img src="https://render.githubusercontent.com/render/math?math=x_{1} \in P">

For simplicity, the tuple of package
<img src="https://render.githubusercontent.com/render/math?math=x = (repoUrl, name, semverRange = 1.0, ns)">
is represented as
<img src="https://render.githubusercontent.com/render/math?math=x_{1}">.


### Depends

<img src="https://render.githubusercontent.com/render/math?math=p_{1}"> depends on
<img src="https://render.githubusercontent.com/render/math?math=x_{1}">:

<img src="https://render.githubusercontent.com/render/math?math=x_{1} - p_{1} \geq 0">

Installing <img src="https://render.githubusercontent.com/render/math?math=p_{1}"> implies
installing <img src="https://render.githubusercontent.com/render/math?math=x_{1}">.
Although <img src="https://render.githubusercontent.com/render/math?math=x_{1}"> may be
installed without <img src="https://render.githubusercontent.com/render/math?math=p_{1}">.


### Depends on multiple versions

If <img src="https://render.githubusercontent.com/render/math?math=p_{1}">
depends on multiple versions of _x_, for example,
<img src="https://render.githubusercontent.com/render/math?math=x_{1}"> and
<img src="https://render.githubusercontent.com/render/math?math=x_{2}">:

<img src="https://render.githubusercontent.com/render/math?math=x_{1} %2B x_{2} - p_{1} \geq 0">


### Conflicts

If <img src="https://render.githubusercontent.com/render/math?math=y_{2}"> conflicts with
<img src="https://render.githubusercontent.com/render/math?math=x_{1}">:

<img src="https://render.githubusercontent.com/render/math?math=x_{1} %2B y_{3} \leq 1">


### Provides

If <img src="https://render.githubusercontent.com/render/math?math=x_{1}">
provides
<img src="https://render.githubusercontent.com/render/math?math=b_{1}">:

<img src="https://render.githubusercontent.com/render/math?math=x_{1} \geq 1">,
<img src="https://render.githubusercontent.com/render/math?math=b_{1} \geq 1">

### Optional-Depends

If <img src="https://render.githubusercontent.com/render/math?math=x_{1}">
optional-depends on
<img src="https://render.githubusercontent.com/render/math?math=a_{1}, b_{1}">: 

<img src="https://render.githubusercontent.com/render/math?math=a_{1} %2B b_{1} \geq 0">, 
together with
<img src="https://render.githubusercontent.com/render/math?math=x_{1} \geq 1">

## Objective function

The quality of the solution depends on the definition of the objective function.
We can use several criteria when optimizing the objective function.

### Minimize package removal

Given  <img src="https://render.githubusercontent.com/render/math?math=PI'_{1} \dots PI'_{n}">
as packages already installed:

<img src="https://render.githubusercontent.com/render/math?math=f_{1} = min( (1 - PI'_{1}) %2B \dots %2B (1 - PI'_{n}))">

The solver will try to set <img src="https://render.githubusercontent.com/render/math?math=PI_{i}">
to 1, which implies not removing the package.

### Minimize the number of installed packages

Given <img src="https://render.githubusercontent.com/render/math?math=P_{1} \dots P_{n}">
as the new packages targetted to be installed (either existent or new), the
objective function will be:

<img src="https://render.githubusercontent.com/render/math?math=f_{2}(P) = min(P_{1} %2B \dots %2B P_{n})">

### Maximizing the freshness of packages

When the objective is to install the most recent versions, independently of what
is already installed.

Consider
<img src="https://render.githubusercontent.com/render/math?math=P1_{1} \dots P1_{k1}">
to be different versions of package
<img src="https://render.githubusercontent.com/render/math?math=P1">.

Consider  <img src="https://render.githubusercontent.com/render/math?math=d(P1_{1})">
to be the normalized precalculated distance between
<img src="https://render.githubusercontent.com/render/math?math=P1_{1}">
version and the newest version
<img src="https://render.githubusercontent.com/render/math?math=P1_{newest}">
available in the repositories. This distance is a
constant for the purposes of the PBO problem. In our case, we can define it as:

<img src="https://render.githubusercontent.com/render/math?math=d(P1_{1}) = P1_{newest} * \overline{semver_{newest}} - P1_{1} * \overline{semver_{1}}">

Where 
<img src="https://render.githubusercontent.com/render/math?math=\overline{semver_{k}}">
with version _k_ in the form of Helm's [semantic version
ranges](https://devhints.io/semver).

The normalized semantic version is calculated by constructing a fractional
number such that:

- `major version -> integer part`
- `minor.patch version -> fractional part`

E.g:

- `1.2.1 -> 1,21`
- `0.46.3 -> 0,463`
- `7.10.1 -> 7,101`


The objective function will be:

<img src="https://render.githubusercontent.com/render/math?math=f_{3} = min( (P1_{1} * d(P1_{1}) %2B \dots %2B  P1_{k1} * d(P1_{k1}) ) %2B  (PZ_{1} * d(PZ_{1}) %2B \dots %2B  PZ_{kn} * d(PZ_{kn}) ))">

where the value of 
<img src="https://render.githubusercontent.com/render/math?math=d(Pi_{ki})">
is zero if the package is the newest in the repo.


### Multicriteria optimization

Usually, installing a package follows multiple of the already mentioned
criteria. Even if one is more important than the others, not following several
of them can lead to undesired solutions (e.g: minimizing the amount of
installed packages, yet wanting all optional dependencies to be installed).

Given that each criteria has a different utility for the user, we can specify a
weight
<img src="https://render.githubusercontent.com/render/math?math=W_{i})">
for each criteria _i_ (e.g: weight given to removals of packages, weight given
to the present of packages, weight given to having an older version in the
solution when a newer exists).

Then, the objective function is defined as:

<img src="https://render.githubusercontent.com/render/math?math=f_{4} = min( (W_{1} * f_{1}(P) %2B (W_{2} * f_{2}(P)) %2B (W_{3} * f_{3}(P))">

# Existing solver libraries in Golang that we considered

Ideally, we want a MAXSAT/Pseudo-Boolean SAT solver library with ample feature
support, good docs, good API, a thriving community, verbose output, in Golang,
and with a compatible license that can be statically linked. We want a solver
that is used in the field, updated continuously with latest SAT improvements
from SAT competitions. We want a solver that handles optimization problems
(hence, a Pseudo-Boolean SAT solver).

## [Gini](https://github.com/IRIFrance/gini)

- 1 big known consumer: github.com/operator-framework/operator-lifecycle-manager.
- Only pure SAT CNF problems.

## [Luet](https://github.com/mudler/luet)

- GPLv3, which would force us to move to GPLv3 given that Golang is usually
  statically compiled.
- It defines a pure SAT problem (with constraints and literals inspired by OPIUM)
(see [here](https://github.com/mudler/luet/blob/56e9c6f82ecb1aeffe5f7d0f152c6210cd6fe4ed/pkg/solver/solver.go#L37-L54)).
- It uses the pure SAT solving of Gophersat, and, if there's some failure, it uses
Q-Learning to recover (see [here](https://github.com/mudler/luet/blob/master/pkg/solver/resolver.go)).
See also their [docs](https://luet-lab.github.io/docs/docs/concepts/overview/constraints/).

## [Go-sat](https://github.com/mitchellh/go-sat)

- Pure SAT solver.
- Last developed 4 years ago.
- No known consumers.

## [Saturday](https://github.com/cespare/saturday)

- Pure SAT solver.

## [DPLL](https://github.com/bmatsuo/dpll)

- Pure SAT solver.

## [Go dep](https://github.com/golang/dep/tree/master/gps)

- Deprecated, too tightly coupled with go modules.

## [Vgo](https://github.com/golang/go)

- Uses non-SAT solver heuristics:
  * https://research.swtch.com/vgo-principles
  * https://research.swtch.com/vgo-mvs
- Non exported API, all under `src/cmd/go/internal/{mod*,mvs}`.

# Selected implementation

## [Gophersat](https://github.com/crillab/gophersat)

- No big known consumers, but several medium consumers (see
  [here](https://github.com/search?q=gophersat&type=code).
- Solves Pseudo-Boolean, MAXSAT problems (e.g: efficiency problems).
- Provides sub-optimal solutions to PB/MaxSAT problems as soon as it finds them,
  without needing to wait for the optimal solution. We can then decide how much we
  want to wait for a solution.
- The one with most active development at the time of evaluation.
- [Announcement email (2017)](https://groups.google.com/g/golang-nuts/c/SE1dNC8N46o/m/0drzp9jdAQAJ).

# literature

PBO problem:
- https://arxiv.org/pdf/1007.1022.pdf (CC BY-ND 3.0, yet contains minor errors)
- https://ths.rwth-aachen.de/wp-content/uploads/sites/4/teaching/theses/grobelna_bachelor.pdf (PB equations optimization)

General literature:
- https://research.swtch.com/version-sat
- https://arxiv.org/pdf/2011.07851.pdf
- http://www.lifl.fr/LION9/presentation_orator/lion9.pdf
- https://www.cril.univ-artois.fr/documents/nanjing-leberre.pdf
- http://citeseerx.ist.psu.edu/viewdoc/download;jsessionid=1BF7B2469673BA27F3F17649A156991E?doi=10.1.1.22.8880&rep=rep1&type=pdf
- https://en.opensuse.org/openSUSE:Libzypp_satsolver_basics
- https://en.opensuse.org/openSUSE:Libzypp_satsolver_internals
- https://github.com/operator-framework/enhancements/blob/master/enhancements/operator-dependency-resolution.md
- https://ranjitjhala.github.io/static/opium.pdf
- https://resources.mpi-inf.mpg.de/departments/rg1/conferences/vtsa09/slides/leberre1.pdf
- https://arxiv.org/pdf/1007.1022.pdf?origin=publicationDetail
- https://github.com/operator-framework/enhancements/pull/7/files
- https://hal.inria.fr/hal-00697463/document
