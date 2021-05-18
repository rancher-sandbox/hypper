## How to lint a hypper chart

We just created or modified our chart or added annotations to take advantage of Hypper features but...how do we know those annotations are correct?


It's very simple, just run `hypper lint` against your chart, and it will verify that the chart is well-formed.

`hypper lint` runs both Helm checks and Hypper checks against the chart.

If the linter encounters things that will cause the chart to fail installation,
it will emit `[ERROR]` messages. If it encounters issues that break with
convention or recommendation, it will emit `[WARNING]` messages. For optional
features, it will emit `[INFO]` messages.

First, lets add the hypper-charts repo, to obtain some chart examples:

```console
$ hypper repo add hypper-charts 'https://rancher-sandbox.github.io/hypper-charts/repo'
"hypper-charts" has been added to your repositories
$ hypper repo update
Hang tight while we grab the latest from your chart repositories...
...Successfully got an update from the "hypper-charts" chart repository
🛳  Update Complete.
$ hypper pull hypper-charts/our-app --untar

```

Now, let's lint. For example, linting against one of our test charts with
Hypper annotations should produce a warning due to a missing icon:

```console
$ hypper lint ./our-app
==> Linting our-app/
[INFO] Chart.yaml: icon is recommended

1 chart(s) linted, 0 chart(s) failed
```


While running it against a vanilla Helm chart with no extra Hypper annotations
will emit several `[INFO]`, recommending to set certain values that Hypper
supports:


```console
$ hypper repo add bitnami https://charts.bitnami.com/bitnami
"bitnami" has been added to your repositories

$ hypper repo update
Hang tight while we grab the latest from your chart repositories...
...Successfully got an update from the "hypper-charts" chart repository
...Successfully got an update from the "bitnami" chart repository
🛳  Update Complete.

$ hypper pull bitnami/mariadb --untar

$ hypper lint ./mariadb
==> Linting ./mariadb
[INFO] Chart.yaml: Setting hypper.cattle.io/release-name in annotations is recommended
[INFO] Chart.yaml: Setting hypper.cattle.io/namespace in annotations is recommended
[INFO] Chart.yaml: Setting hypper.cattle.io/shared-dependencies in annotations is optional
[INFO] Chart.yaml: Setting hypper.cattle.io/optional-dependencies in annotations is optional

1 chart(s) linted, 0 chart(s) failed
```

Running against a shared-dependencies annotation that is malformed will emit an `[ERROR]`. Let's edit the MariaDB chart, so it has a wrong `shared-dependencies` stanza, for example. Make the ./mariadb/Chart.yaml look like this:

```diff
# this is a diff of ./mariadb/Cahrt.yaml

annotations:
  category: Database
+  hypper.cattle.io/shared-dependencies: |
+    - name: fleet
+      version: "this-is-incorrect-0.3.500"
+      repository: "https://rancher-sandbox.github.io/hypper-charts/repo"
apiVersion: v2
appVersion: 10.5.10
dependencies:
- name: common
  repository: https://charts.bitnami.com/bitnami
  tags:
  - bitnami-common
  version: 1.x.x
description: Fast, reliable, scalable, and easy to use open-source relational database
  system. MariaDB Server is intended for mission-critical, heavy-load production systems
  as well as for embedding into mass-deployed software. Highly available MariaDB cluster.
home: https://github.com/bitnami/charts/tree/master/bitnami/mariadb
icon: https://bitnami.com/assets/stacks/mariadb/img/mariadb-stack-220x234.png
keywords:
- mariadb
- mysql
- database
- sql
- prometheus
maintainers:
- email: containers@bitnami.com
  name: Bitnami
name: mariadb
sources:
- https://github.com/bitnami/bitnami-docker-mariadb
- https://github.com/prometheus/mysqld_exporter
- https://mariadb.org
version: 9.3.11

```


```console
$ hypper lint ./mariadb
==> Linting ./mariadb
[INFO] Chart.yaml: Setting hypper.cattle.io/release-name in annotations is recommended
[INFO] Chart.yaml: Setting hypper.cattle.io/namespace in annotations is recommended
[INFO] Chart.yaml: Setting hypper.cattle.io/optional-dependencies in annotations is optional
[ERROR] Chart.yaml: Shared dependency version is broken: improper constraint: this-is-incorrect-0.3.500

Error: 1 chart(s) linted, 1 chart(s) failed
```


Now if you were to set some kind of automated CI in place to check for linting
and are required to have Hypper annotations as mandatory, you can run `hypper
lint` with the `--strict` flag so all warnings are marked as errors. For now, we
don't have warning-level messages.
