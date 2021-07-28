# Quickstart

This guide covers how you can quickly get started using Hypper.

If you already know Helm, this will be a breeze, as Hypper follows Helm's workflow
and extends its functionalities to install Helm charts that have been extended
to work with Hypper.

## Prerequisites

The following prerequisites are required:

1. A Kubernetes cluster.
1. A Chart repository to install charts from.

## Install Hypper

See the [installation guide](./installing.md).

## Configuring Hypper by initializing repositories

Hypper enables you to deploy Helm and Hypper charts into your Kubernetes
cluster. To simplify that, the best way is to add a chart repository, for
example one with Helm charts:

```console
$ hypper repo add bitnami https://charts.bitnami.com/bitnami
```

You can add several repositories. Let's add a repository containing charts
with some Hypper functionality:

```console
$ hypper repo add hypper-charts https://rancher-sandbox.github.io/hypper-charts/repo
```

Now, you can list the repositories that Hypper can install charts from:

```console
$ hypper repo list
NAME            URL
bitnami         https://charts.bitnami.com/bitnami
hypper-charts   https://rancher-sandbox.github.io/hypper-charts/repo
```

## Install an example Helm chart

To install a chart, you can run the `hypper install` command. Hypper has several
ways to find and install a chart, but the easiest is to use a repository, in
this case the Bitnami chart repository.

```console
$ hypper install mariadb bitnami/mariadb
🛳  Installing chart "mariadb" as "mariadb" in namespace "default"…
👏 Done!
```

Whenever you install a chart, a new release is created. So one Helm chart can be
installed multiple times into the same cluster. And each can be independently
managed and upgraded.

## Install an example Hypper chart

Hypper charts are supersets of Helm charts with more functionality, such as
specifying default release name, namespace of installation, or installing
shared-dependency charts automatically.
They are thought to be installed system-wide in the cluster: 1 Hypper chart for
all users of the cluster. Think of them as typical system OS libraries/services.

The commands are the same as you have already used:

```console
$ hypper install hypper-charts/our-app
❓ Install optional shared dependency "rancher-tracing" of chart "demo"? [Y/n]:
y
The following charts are going to be installed:
our-app v0.1.0
 ├─ fleet v0.3.500
 └─ rancher-tracing v1.20.002

🛳  Installing chart "fleet" as "fleet" in namespace "fleet-system"…
🛳  Installing chart "rancher-tracing" as "rancher-tracing" in namespace "istio-system"…
🛳  Installing chart "our-app" as "our-app-name" in namespace "hypper"…
👏 Done!
```

This time, the chart got installed with a default name `our-app-name`, and into
a default namespace  `hypper`. Hypper to creates the namespace if it doesn't
exist (you can pass `--no-create-namespace` if you don't want that). It also
installed its defined shared and optional shared dependencies.

## Learn about releases

It's easy to see what has been released with hypper:

```console
$ hypper ls --all-namespaces
NAME            NAMESPACE       REVISION        UPDATED                                         STATUS          CHART                           APP VERSION
fleet           fleet-system    1               2021-05-18 15:44:16.11805509 +0200 CEST         deployed        fleet-0.3.500                   0.3.5
mariadb         default         1               2021-05-18 15:44:43.106328879 +0200 CEST        deployed        mariadb-9.3.11                  10.5.10
our-app-name    hypper          1               2021-05-18 15:44:18.687033582 +0200 CEST        deployed        our-app-0.0.2                   0.0.1
rancher-tracing istio-system    1               2021-05-18 15:44:18.592656807 +0200 CEST        deployed        rancher-tracing-1.20.002        1.20.0
```

## Uninstall a release

To uninstall a release, use the hypper uninstall command:

```console
$ hypper uninstall fleet -n fleet-system
🔥  uninstalling fleet
✅  release "fleet" uninstalled
```

This will uninstall fleet from Kubernetes, which will remove all resources
associated with the release as well as the release history.

If the flag `--keep-history` is provided, release history will be kept. You will
be able to request information about that release:

```console
$ hypper status fleet -n fleet-system
NAME: fleet
LAST DEPLOYED: Fri Mar 12 12:14:28 2021
NAMESPACE: fleet-system
STATUS: uninstalled
REVISION: 1
TEST SUITE: None
```

## Read the help text

To learn more about available Hypper commands, use `hypper help`, or type a
command followed by the `-h` flag: `hypper install -h`.
