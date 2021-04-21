## Searching for charts in repos

Search reads through all the repositories configured on the system, and looks for matches. Searching on these repositories uses the metadata stored on the system.

It will display the latest stable versions of the charts found.

```console
$ hypper search repo fleet
NAME                            CHART VERSION   APP VERSION     DESCRIPTION
hypper-charts/fleet             0.3.500         0.3.5           Fleet Manager - GitOps at Scale
hypper-charts/fleet-agent       0.3.500         0.3.5           Fleet Manager Agent - GitOps at Scale
hypper-charts/fleet-crd         0.3.500         0.3.5           Fleet Manager CustomResourceDefinitions
```

By default, it will only show the latest version that match the keyword used. In order to see all versions you can use the `-l` flag.

```console
$ hypper search repo fleet-agent -l
NAME                            CHART VERSION   APP VERSION     DESCRIPTION
hypper-charts/fleet-agent       0.3.500         0.3.5           Fleet Manager Agent - GitOps at Scale
```

You can also search for a specific chart version by using the `--version VERSION` flag..

Note that VERSION needs to be a valid SemVer version.

```console
$ hypper search repo fleet-agent --version 0.3.500
NAME              	CHART VERSION	APP VERSION	DESCRIPTION                          
hypper/fleet-agent	0.3.500      	0.3.5      	Fleet Manager Agent - GitOps at Scale
```

It's also possible to pass the `--regexp` flag to use regexp in the search.

```console
$ hypper search repo "fleet-" --regexp
NAME                            CHART VERSION   APP VERSION     DESCRIPTION
hypper-charts/fleet-agent       0.3.500         0.3.5           Fleet Manager Agent - GitOps at Scale
hypper-charts/fleet-crd         0.3.500         0.3.5           Fleet Manager CustomResourceDefinitions
```

If you want to search for development charts (only stable charts are shown by default), use the `--devel` flag to show those development charts.


```console
$ hypper search repo fleet-agent --devel
NAME                            CHART VERSION   APP VERSION     DESCRIPTION
hypper-charts/fleet-agent       0.3.500         0.3.5           Fleet Manager Agent - GitOps at Scale
```
