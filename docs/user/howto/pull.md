# Pulling charts to the local machine

Sometimes there is a need to modify charts from a repo or just have a look at what the chart is actually doing.

For that you can use `hypper pull`, which will download the chart to your local machine.


## Pulling a chart

```shell
$ hypper pull hypper/fleet
$ ls fleet*
fleet-0.3.500.tgz
```

## Pulling a chart and extracting it

Usually, pull will download the tar.gz chart package, but you can use `--untar` so instead the package is extracted automatically.

```shell
$ hypper pull hypper/fleet --untar
$ ls fleet/     
charts  Chart.yaml  templates  values.yaml
```

## Pulling a specific chart version

Pull will automatically download the latest chart version, but you can use `--version` to get the specified version instead.

```shell
$ hypper pull hypper/fleet --version 0.3.400
$ ls fleet*
fleet-0.3.400.tgz
```

## Changing the output dir

By default, pull will download the chart into the current dir. Use the `-d` flag to set the output dir.

Note: The output dir needs to exist beforehand.

```shell
$ hypper pull hypper/fleet -d chartslocal
$ ls chartslocal 
fleet-0.3.500.tgz
```

## Pulling development versions

You can use the `--devel` flag to pull devel versions. By default, hypper won't download, search or list devel version unless specified by the flag.

```shell
$ hypper pull hypper/fleet --version 0.3.600-rc1 --devel
$ ls fleet*
fleet-0.3.600-rc1.tgz
```
