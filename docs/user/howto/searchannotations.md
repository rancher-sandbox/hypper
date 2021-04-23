## Searching for charts in repos using annotations

Apart from having the normal [search](docs/user/howto/search.md) in hypper, we support searching for specific annotations in a search.

Using the `-a` flag in `hypper search repo` allows to filter the returned charts based on the KEY=VALUE values in annotations.

```shell
$ hypper search repo -a 'hypper.cattle.io/release-name=fleet'       
NAME        	CHART VERSION	APP VERSION	DESCRIPTION                    
hypper/fleet	0.3.500      	0.3.5      	Fleet Manager - GitOps at Scale

```


We also support the `--regexp` flag when using annotations, so you can search for the VALUE with a regexp

```shell
$ hypper search repo -a 'hypper.cattle.io/release-name=.*crd.*' --regexp
NAME                            	CHART VERSION	APP VERSION	DESCRIPTION                                 
hypper/fleet-crd                	0.3.500      	0.3.5      	Fleet Manager CustomResourceDefinitions     
hypper/longhorn-crd             	1.1.001      	           	Installs the CRDs for longhorn.             
hypper/rancher-backup-crd       	1.0.400      	1.0.4      	Installs the CRDs for rancher-backup.       
hypper/rancher-cis-benchmark-crd	1.0.301      	           	Installs the CRDs for rancher-cis-benchmark.
hypper/rancher-gatekeeper-crd   	3.3.001      	           	Installs the CRDs for rancher-gatekeeper.   
hypper/rancher-logging-crd      	3.9.002      	           	Installs the CRDs for rancher-logging.      
hypper/rancher-monitoring-crd   	9.4.204      	           	Installs the CRDs for rancher-monitoring.   
hypper/rancher-operator-crd     	0.1.300      	0.1.3      	Rancher Operator CustomResourceDefinitions  

```

Of course the usual search flags also work in conjunction with annotations, so you can search for annotations *and* a specific version with he `--version` flag for example. 


Note: Currently you can pass more than one annotation to search for. This is currently an OR filter, so if any of the given annotation are found in a chart, it will match.