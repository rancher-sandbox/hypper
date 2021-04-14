# Quickstart

This guide covers how you can quickly get started using Hypper.

If you already know Helm, this will be a breeze, as Hypper follows Helm's workflow
and extends its functionalities to install Helm and charts that have been extended
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

```terminal
$ hypper repo add bitnami https://charts.bitnami.com/bitnami
```

You can add several repositories. Let's add a repository containing charts
with some Hypper functionality:

```terminal
$ hypper repo add rancher-charts https://charts.rancher.io
```

Now, you can list the repositories that Hypper can install charts from:

```terminal
$ hypper repo list
NAME            URL
bitnami         https://charts.bitnami.com/bitnami
rancher-charts  https://charts.rancher.io
```

## Install an example Helm chart

To install a chart, you can run the `hypper install` command. Hypper has several
ways to find and install a chart, but the easiest is to use a repository, in
this case the Bitnami chart repository.

```terminal
$ hypper install mariadb bitnami/mariadb
Installing chart "mariadb" in namespace "default"…
Done! 👏
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

```terminal
$ hypper install rancher-charts/fleet --create-namespace
Installing chart "fleet" in namespace "fleet-system"…
Done! 👏
```

This time, the chart got installed with a default name `fleet`, and into a
default namespace  `fleet-system`. We passed the flag `--create-namespace` to
tell Hypper to create the namespace if it doesn't exist.

## Learn about releases

It's easy to see what has been released with hypper:

```terminal
$ hypper ls --all-namespaces
NAME       NAMESPACE       REVISION        UPDATED                                 STATUS          CHART                 APP VERSION
fleet      fleet-system    1               2021-03-12 12:06:35.951012048 +0100 CET deployed        fleet-0.3.400         0.3.4
mariadb    default         1               2021-03-12 12:09:46.670491535 +0100 CET deployed        mariadb-9.3.5         10.5.9
```

## Uninstall a release

To uninstall a release, use the hypper uninstall command:

```terminal
$ hypper uninstall fleet -n fleet-system
🔥  uninstalling fleet
🔥  release "fleet" uninstalled
```

This will uninstall fleet from Kubernetes, which will remove all resources
associated with the release as well as the release history.

If the flag `--keep-history` is provided, release history will be kept. You will
be able to request information about that release:

```terminal
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
