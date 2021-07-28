# Introduction

Hypper provides Kubernetes package management for cluster admins. It is a
package manager built on Helm and inspired by other package managers like
zypper.

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

### Hypper Features Not In Helm

#### Installing With Release Name and Namespace Repeatedly

When dealing with system level charts you may want to have them be installed in
the same namespace or with the same release name everywhere. For example, when
you install a cluster wide logging service. With Hypper, you can capture this
information as an annotation and only specify a different one if you want to
override the default. Then when you tell Hypper to install your service it will
be repeatedly installed to the same location.

#### Shared Dependencies

Sometimes you want to have a shared system dependency. For example, you may want
to have just a single Prometheus or Istio in a cluster. When you install an
application that depends on this shared dependency you want to install it if not
present and leverage the existing one if present.

The additional metadata will be stored as Helm chart annotations.

#### Optional Shared Dependencies

There are occasions where you may want a shared dependency to be optional.
Instead of being checked and installed all the time you want Hypper to prompt
you about using it. Or, you can tell it what to do using flags.

Hypper provides this ability for the direct chart you want to install and the
additional metadata is stored in annotations.

## Client Application and Software Development Kit (SDK)

Hypper provides both a client application you can use in a terminal and a SDK
in Go that you can use for the development of your applications. The
documentation provided here primarily focuses on chart customizations for
Hypper and the client application.

The SDK documentation can be found at
[pkg.go.dev/github.com/rancher-sandbox/hypper](https://pkg.go.dev/github.com/rancher-sandbox/hypper).
The client source provides an example of using the SDK and is in the `cmd`
sub-directory.
