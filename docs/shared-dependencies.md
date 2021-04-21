# Proposal: Shared Dependencies

_Note, this is currently a proposal. Once the proposal has been turned into
an implementation this can be updated to document the implementation instead
of being a proposal._

Share dependencies are those that more than one chart may share. These are
useful for software running as system services.

## Use Case

As a cluster administrator, I would like to install a cluster level application
that has shared system dependencies. If one of the share dependencies is
already installed than the installed one should be used. Shared dependencies
should only be installed if not present already.

_Note: system level or cluster level used in this document are not a formal
Kubernetes or Rancher construct. It's an idea that cluster administrators need
to install some things at the cluster level where there will be just one
instance in the cluster._

### Prometheus Example

Consider the case where cluster operators want to have a single instance of
Prometheus running as a system service. There should only be one instance
running in the cluster and all (or most) applications that need prometheus
would use this single instance.

When someone installs an application (e.g., another system service) that depends
on Prometheus then Hypper should check to see if Prometheus is installed in the
expected location. If not, it should attempt to install it. If Prometheus is
already installed it would skip installing it and instead use that instance.

## Technical Details

This section contains the technical details on implementing the request. These
are suggest and open for discussion.

### Metadata

All metadata about declaring shared dependencies will be contained in the
`annotations` within the _Chart.yaml_ file of a chart. This uses an existing
mechanism that is commonly used in the Kubernetes and Helm space. This metadata
will also be contained in provenance files.

```yaml
annotations:
  hypper.cattle.io/shared: |
    - name: fleet
      version: "^0.3.500"
      repository: "https://rancher-sandbox.github.io/hypper-charts/repo"
```

Helm annotations are key/value pairs where the value is a string. In this case
the string is a multiline YAML string with YAML within it.

The `repository` contains the location of the dependent chart to install. This
is the same as with normal Helm dependencies.

`name` is the name of the chart in the repository.

The `version` property has the semantic version for the chart to be installed.
This can be a version range but there is no lock file so one version should be
installed in development and another in production if a range is used. If a
specific version is set that will be used.

### Name and Namespace

The metadata in the metadata section does not contain the release name or
namespace to install or look for the installed shared dependency in. This
metadata is contained in the dependent chart being installed.

If this metadata is not present an error should be given and it cannot be
installed. We may change this in a later revision to provide a way to pass this
in.

In order to know where to put or locate the shared dependency use the following
two annotations. These are already in use by the Rancher catalog system.

```yaml
annotations:
  catalog.cattle.io/namespace: fleet-system
  catalog.cattle.io/release-name: fleet-crd
```

This metadata can be fetched from 3 different locations:

1. The `index.yaml` file for the repository the chart is in
2. The chart in the local cache
3. From the chart in the remote repository

### Template Sandbox and Separate Releases

Dependencies in normal charts have their dependencies in the _charts_ directory
within a chart. Helm has it as a requirement that they are in there.

When Helm renders a chart and all of its dependencies they are rendered in a
shared template sandbox. For example, when a template function is created with
one name that name is global to the sandbox.

Shared dependencies will not have their own template sandbox. They will each be
in their own sandbox. This means we will be using the Helm library to install
each of the dependencies rather than combining them into one package to install.

Keeping the instances of the chart separate is important. When a shared
dependency is installed it will have its own Helm release record and be able to
be upgraded independently.

### Packaging Shared Dependencies

When someone runs `helm package` the dependencies declared in `dependencies`
need to be in the _charts_ directory and are packaged along with the Chart when
the archive is created.

Shared dependencies will not be doing that. To add shared dependencies to the
_charts_ directory would alter the chart in a way it was not intended to. These
charts should still be usable Helm charts that can be used by Helm and outside
of Hypper. Hypper will just make the experience better.

### Solving

Hypper will need to solve for the dependencies to make sure they are installed
in the proper order. This solver will only need to deal with the shared
dependencies, at this point. It should be a SAT solver (there are existing
Go SAT solvers that should be evaluated).
