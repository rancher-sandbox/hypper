## How to lint a hypper chart

We just created or modified our chart or added annotations to take advantage of Hypper features but...how do we know those annotations are correct?


It's very simple, just run `hypper lint` against your chart, and it will verify that the chart is well-formed.

`hypper lint` runs both Helm checks and Hypper checks against the chart.

If the linter encounters things that will cause the chart to fail installation, it will emit [ERROR] messages. If it encounters issues that break with convention or recommendation, it will emit [WARNING] messages.


For example, running it against one of our test charts with Hypper annotations should produce a warning due to a missing icon:

```shell
$ hypper lint cmd/hypper/testdata/testcharts/hypper-annot
==> Linting cmd/hypper/testdata/testcharts/hypper-annot
[INFO] Chart.yaml: icon is recommended

1 chart(s) linted, 0 chart(s) failed
```


While running it against a vanilla Helm chart with no extra Hypper annotations will emit several [WARNING], recommending to set certain values that Hypper supports:


```shell
$ hypper lint cmd/hypper/testdata/testcharts/vanilla-helm 
==> Linting cmd/hypper/testdata/testcharts/vanilla-helm
[INFO] Chart.yaml: icon is recommended
[INFO] Chart.yaml: Setting hypper.cattle.io/release-name in annotations is recommended
[INFO] Chart.yaml: Setting hypper.cattle.io/namespace in annotations is recommended
[INFO] Chart.yaml: Setting hypper.cattle.io/shared-dependencies in annotations is recommended

1 chart(s) linted, 0 chart(s) failed
```

Running against a shared-dependencies annotation that is malformed will emit an [ERROR]:

```shell
$ hypper lint cmd/hypper/testdata/testcharts/hypper-annot
==> Linting cmd/hypper/testdata/testcharts/hypper-annot
[INFO] Chart.yaml: icon is recommended
[ERROR] Chart.yaml: Shared dependencies list is broken, please check the correct format

Error: 1 chart(s) linted, 1 chart(s) failed
```


Now if you were to set some kind of automated CI in place to check for linting and are required to have Hypper annotations as mandatory, you can run `hypper lint` with the `--strict` flag so all warnings are marked as errors.

```shell
$ hypper lint cmd/hypper/testdata/testcharts/fallback-annot --strict
==> Linting cmd/hypper/testdata/testcharts/fallback-annot
[INFO] Chart.yaml: icon is recommended
[WARNING] Chart.yaml: Setting hypper.cattle.io/release-name in annotations is recommended
[WARNING] Chart.yaml: Setting hypper.cattle.io/namespace in annotations is recommended
[WARNING] Chart.yaml: Setting hypper.cattle.io/shared-dependencies in annotations is recommended

Error: 1 chart(s) linted, 1 chart(s) failed
```