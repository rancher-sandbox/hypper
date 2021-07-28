# hypper

[![CI](https://github.com/rancher-sandbox/hypper/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/rancher-sandbox/hypper/actions/workflows/ci.yml)

pronounced hip-er.

A Kubernetes package manager for cluster admins that's built on top of Helm and charts.

## Get in touch

Want to contribute or talk with use about hypper? Leave us an issue on [github](https://github.com/rancher-sandbox/hypper/issues) or join us on the [rancher-users slack](https://slack.com/app_redirect?channel=C01V9NSD308).

## Hypper vs Helm

[Helm](https://helm.sh) is a package manager that targets application operation. It is designed with application operators in mind rather than cluster operators. Applications operators may have restrictions on the namespaces they can install workloads into and see. The clusters used by application operators may be multi-tenant and some application operators may install the same application multiple times into a namespace. Helm handles these forms of cases well and makes assumptions that these happen. [Cluster operator use cases are generally out of scope for Helm](https://github.com/helm/community/blob/main/user-profiles.md).

Hypper builds upon Helm and charts to handle package management with cluster operators in mind. Think of it like installing packages onto a single Linux server where different packages can typically know where each other are. Hypper is designed to work with shared dependencies. That is, dependencies installed once on a cluster that more than one different chart can depend on. This is in addition to the dependencies specified the Helm way that are tied to an individual chart. Hypper works best with repositories of charts designed to work together.

## Annotations

The additional information Hypper needs is specified via annotations in the `Chart.yaml` file. For example, a chart can specify a namespace and release name so that it can be installed into a well known location. This enables releases of other charts to know where to find it. Those annotations are:

```yaml
annotations:
  hypper.cattle.io/release-name: example-name
  hypper.cattle.io/namespace: example-namespace
```

Shared dependencies also happen via annotation.

```yaml
annotations:
  hypper.cattle.io/shared-dependencies: |
    - name: fleet
      version: "^0.3.500"
      repository: "https://rancher-sandbox.github.io/hypper-charts/repo"
    - name: rancher-tracing
      version: "^1.20.002"
      repository: "https://rancher-sandbox.github.io/hypper-charts/repo"
```

Notice that the value is a multi-line string (noted by the `|` after the `:`). Hypper parses the value as YAML. The other information looks similar to Helms existing dependencies, which Hypper supports. Shared dependencies are handled differently. Instead of being installed along with the charts resources they are installed as separate charts before this chart is installed. These charts are installed as their own releases with their own lifecycles.

## Roadmap

- [x] Install charts to a chart specified release name and namespace (via chart annotations)
- [x] Install share dependencies (those that more than one application may rely on)
- [x] Install optional/suggests dependencies
- [ ] List outdated dependencies
- [ ] Ensure all versions resolve correctly across dependencies

## License

Copyright (c) 2020-2021 [SUSE, LLC](http://suse.com)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
