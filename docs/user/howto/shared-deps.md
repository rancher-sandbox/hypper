# Working with shared dependencies

Hypper charts support the concept of shared dependency charts.

A chart declared as a shared dependency is a chart that more than one chart may
depend on; once deployed, it can be reused by multiple other deployments. Charts
deployed as shared dependencies are the analogous of system libraries in an OS:
dependencies that are used by several consumers.

## Creating a Hypper chart with shared dependencies

Let's create the simplest chart possible, and add charts as shared dependencies
to it.

First, we create a simple empty chart:

```console
$ helm create our-app
```

Now, we can edit `./our-app/Chart.yaml`, and add some shared dependencies to it:

```diff
apiVersion: v2
name: our-app
description: A Helm chart for Kubernetes
type: application
version: 0.1.0
appVersion: 1.16.0
annotations:
+ hypper.cattle.io/shared-dependencies: |
+   - name: fleet
+     version: "^0.3.500"
+     repository: "https://rancher-sandbox.github.io/hypper-charts/repo"
+ hypper.cattle.io/optional-dependencies: |
+   - name: rancher-tracing
+     version: "^1.20.002"
+     repository: "https://rancher-sandbox.github.io/hypper-charts/repo"
```

Shared dependencies are just normal Helm
[Dependencies](https://helm.sh/docs/topics/charts/#chart-dependencies). As
such, they are defined with:

- The name field is the name of the chart you want.
- The version field is the version of the chart you want.
- The repository field is the full URL to the chart repository. Note that you
  must also use `hypper repo add` to add that repo locally. You might use the
  name of the repo instead of URL.

We can also add a default release name and namespace, where our-app and its
shared-dependencies will get installed:

```diff
apiVersion: v2
name: our-app
description: A Helm chart for Kubernetes
type: application
version: 0.1.0
appVersion: 1.16.0
annotations:
+ hypper.cattle.io/namespace: hypper
+ hypper.cattle.io/release-name: our-app-name
  hypper.cattle.io/shared-dependencies: |
    - name: fleet
      version: "^0.3.500"
      repository: "https://rancher-sandbox.github.io/hypper-charts/repo"
  hypper.cattle.io/optional-dependencies: |
    - name: rancher-tracing
      version: "^1.20.002"
      repository: "https://rancher-sandbox.github.io/hypper-charts/repo"
```

To verify that we did create the correctly, let's lint it:

```console
$ hypper lint ./our-app
==> Linting our-app
[INFO] Chart.yaml: icon is recommended

1 chart(s) linted, 0 chart(s) failed
```

## Listing shared dependencies

Hypper's `shared-dep list` command will list the shared dependencies, its status, and other information:

```console
$ hypper shared-deps list ./our-app
NAME            VERSION         REPOSITORY                                              STATUS          NAMESPACE       TYPE
fleet           ^0.3.500        https://rancher-sandbox.github.io/hypper-charts/repo    not-installed   fleet-system    shared
rancher-tracing ^1.20.002       https://rancher-sandbox.github.io/hypper-charts/repo    not-installed   istio-system    shared-optional
```


## Deploying shared dependencies

Now, let's pretend that we had `fleet` already installed, so let's install
it by hand.

First, add the repos of the `fleet` shared dependency, so they are found when
installing manually.

Note that, when hypper installs the shared dependency on its own, you don't need
to add the repos.

```console
$ hypper repo add hypper-charts 'https://rancher-sandbox.github.io/hypper-charts/repo'
"hypper-charts" has been added to your repositories
$ hypper repo update
Hang tight while we grab the latest from your chart repositories...
...Successfully got an update from the "hypper-charts" chart repository
🛳  Update Complete.
```

Now we install `fleet`:

```console
$ hypper install hypper-charts/fleet -n fleet-system
🛳  Installing chart "fleet" as "fleet" in namespace "fleet-system"…
👏 Done!
```

That satisfies one shared dependency of `our-app`:

```console
$ hypper shared-deps list ./our-app
NAME            VERSION         REPOSITORY                                              STATUS          NAMESPACE       TYPE
fleet           ^0.3.500        https://rancher-sandbox.github.io/hypper-charts/repo    deployed        fleet-system    shared
rancher-tracing ^1.20.002       https://rancher-sandbox.github.io/hypper-charts/repo    not-installed   istio-system    shared-optional
```

Then, we can install `our-app`, and any of its missing shared dependencies.
By default, Hypper asks for each optional shared dependency if we want to
install it or not:

```console
$ hypper install ./our-app
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

What has happened?

1. Hypper has made sure that all declared shared dependencies of `our-app` are
   satisfied, installing those that are missing, and skipping those present.
   Since the chart of the shared dependency didn't have annotations for default
   namespace, it will install them on the dependent namespace.
2. Since we haven't specified the release name or namespace in the command,
   Hypper has installed `our-app` in the default release-name (`our-app-name`)
   and namespace (`hypper`) we specified in the Hypper annotations.

Let's see:

```console
$ ./bin/hypper shared-deps list ./our-app
NAME            VERSION         REPOSITORY                                              STATUS          NAMESPACE       TYPE
fleet           ^0.3.500        https://rancher-sandbox.github.io/hypper-charts/repo    deployed        fleet-system    shared
rancher-tracing ^1.20.002       https://rancher-sandbox.github.io/hypper-charts/repo    deployed        istio-system    shared-optional

$ hypper list -A
NAME            NAMESPACE       REVISION        UPDATED                                         STATUS          CHART                           APP VERSION
fleet           fleet-system    1               2021-05-17 17:32:45.889083671 +0200 CEST        deployed        fleet-0.3.500                   0.3.5
our-app-name    hypper          1               2021-05-17 17:35:13.517422915 +0200 CEST        deployed        our-app-0.0.2                   0.0.1
rancher-tracing istio-system    1               2021-05-17 17:35:13.311231418 +0200 CEST        deployed        rancher-tracing-1.20.002        1.20.0
```

Yay! they are all installed.

If you want, you can always install `our-app` without the shared-dependencies,
by passing the flag `--no-shared-deps`. And you can either install all optional
dependencies by default with `--optional-deps=all`, or skip them with
`--optional-deps=none`.
