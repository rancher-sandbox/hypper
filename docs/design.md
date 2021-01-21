# Hypper Design

Hypper is a package manager built on Helm and inspired by other package managers
like zypper.

## How Hypper Differs From Helm

Why build a package manager on a package manager? Helm is a package manager that
tries to make few assumptions. For example, Custom Resource Definitions (CRD)s
are a Kubernetes resource that extend the Kubernetes API. They can be used in
various ways and there are pitfalls with them. Helm makes very few assumptions
when it comes to them. Hypper is going to make some different assumptions which
will result in different features.

As differences between Hypper and Helm arise they will be documented along with
the reasons for the difference. This is in an effort to provide clarity.

When features from Hypper can be upstreamed into Helm they will be. In that way,
Hypper can act as a testing ground for Helm. We do not expect that every feature
will fit in upstream Helm.

A second element in the difference is that Helm, as a package manager, does not
handle environment deployment operations. Those are left to higher level tools
like Helmfile (a push model like Ansible) and the Flux Helm Operator (a pull
model like Chef). Helm expects higher level tools will use Helm and charts for
operating in different environments.

Hypper will move will make some assumptions that move it more towards the
environment management. This is in a different way from Helmfile or the Flux
Helm Operator but further than Helm will go.

## Directory Structure

The directory structure is inspired by Helm with the following layout.

```
.
├── cmd
│   └── hypper
├── docs
└── pkg
    └── action 
    └── ... 
```

- The _cmd/hypper_ directory contains the client application. This CLI for mac,
  linux, and windows is both a functional application and reference application
  of the library.
- _docs_ contains documentation on the design and usage of hypper.
- The _pkg_ directory contains a library meant to be imported by _cmd/hypper_ and
  other applications. The functionality needed is implemented here.
- _pkg/actions_ contains the main functionality used by _cmd/hypper_.

This is the same layout as Helm itself.

## Design Intent

The following outline the initial major intentions of hypper.

### Modern Terminal UI

Colors and icons are often present in modern terminals. Sometimes referred to
as eye candy, these features can improve user experience by drawing attention to
important things.

### Optional Dependencies

In Helm there are two ways to handle dependencies.

1. Explicitly declare other charts as dependencies. This will cause those charts
   to be installed. When the templates are rendered they happen in the same
   sandboxed environment.
2. Install charts separately and use values to connect them to each other. For
   example, install PostreSQL with one chart and then pass connection information
   into another chart using a Secret or values that generate a Secret.

Hypper intends to add a 3rd form of dependency via an optional dependency. For
example, you are developing something that requires a prometheus CRD. You can
depend on a chart that provides that CRD but does not provide prometheus. The
CRD chart or your chart can declare an optional dependency on a chart that
installs Prometheus. When using hypper to install your chart you will be asked
about installing the optional dependencies. In development you can choose not
to install the dependencies while in production you can choose to install them.

The optional dependency handling will extend to all commands.

Optional dependencies will be specified using Helm chart annotations.

### Shared Dependencies

Sometimes you want to have a shared system dependency. For example, you may want
to have just a single Prometheus or Istio in a cluster. When you install an
application that depends on this shared dependency you want to install it if not
present and leverage the existing one if present.

With Helm there is a separation between those who provide charts and those who
consume them. Charts from various repositories can be installed in the same
cluster but are not guaranteed to work together. That is left as an exercise for
the chart authors.

Hypper will make it easier to have a repository of charts that work together but
can be installed separately.

The additional metadata will be stored as Helm chart annotations.

### OCI Storage

OCI artifacts are a new way to store things. Helm supports storing Helm charts
in OCI distributions as an experimental feature. Many container registries
support storing non-images in them. The list includes Azure, AWS, Google Cloud,
GitHub Container Registry, and Harbor. The one notable exception is Docker Hub.

Kubernetes already needs to have a container registry. In air gapped, on-premise,
and custom environments you need a registry to store your images. If that can be
used to store charts instead of a separate system you have fewer system
dependencies.

Hypper will make OCI registries a primary element in the system.

One notable disadvantage to using an OCI registry for storing charts is a lack of
an `index.yaml` file. This file has metadata enabling some functionality that
can't be duplicated with an OCI registry due to APIs not existing (e.g., search).
To restore these features we are looking to put an index into an OCI registry.

### Name As Chart Identifier

Zypper, the package manager used for openSUSE and other Linux distributions,
uses the package name as the identifier for it. You can change from one vendor
to another for a package with some differences but still the same package.

Helm does not make this form of assumption. It could be true. For example, you
could install a chart from one location and then update from a fork in another
location. Or two charts with the same name could be completely different. It
is the responsibility of the end user to make this decision.

Hypper is going to make the assumption that charts of the same name are the
same chart from different locations.

### Performance

When working with many files over a network there are opportunities to use
caching, features of protocols, and design choices to improve performance. The
following examples illustrate this:

- JSON is faster to parse and uses less memory than YAML does. An `index.json`
  file would faster and less memory intensive for Helm to work with than
  `index.yaml`. A change like this could be done in a backwards compatible
  manner.
- `index.yaml` files are regularly pulled down to get updates and this is a
  manual step. Using HTTP headers we can check if the content changed and only
  pull it in those cases. By doing that some of the updates can be automated,
  as well. A header that could be used, for example, is `etag`.
- Charts can be stored locally in a cache and used from there rather than being
  downloaded each time. This cache could be stored in places like a CI system
  (e.g., like the way packages are stored for programming languages) or on a
  local development computer.
