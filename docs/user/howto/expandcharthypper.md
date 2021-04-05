# How to expand a chart with Hypper annotations

Hypper charts build on top of Helm charts so any chart out there will work out of the box.

But we want to go further, we want to take advantage of Hypper features like release and namespace from annotations.

This is a simple as modifying our `Chart.yaml` to add to its annotations.


For example using the [Mysql chart from Bitnami](https://github.com/bitnami/charts/tree/master/bitnami/mysql) as example, these are the contents of `Chart.yaml`:

```yaml
annotations:
  category: Database
apiVersion: v2
appVersion: 8.0.23
dependencies:
  - name: common
    repository: https://charts.bitnami.com/bitnami
    tags:
      - bitnami-common
    version: 1.x.x
description: Chart to create a Highly available MySQL cluster
home: https://github.com/bitnami/charts/tree/master/bitnami/mysql
icon: https://bitnami.com/assets/stacks/mysql/img/mysql-stack-220x234.png
keywords:
  - mysql
  - database
  - sql
  - cluster
  - high availability
maintainers:
  - email: containers@bitnami.com
    name: Bitnami
name: mysql
sources:
  - https://github.com/bitnami/bitnami-docker-mysql
  - https://mysql.com
version: 8.5.1
```

But we want to install it always on the same namespace called `databases`

So we will expand the `Chart.yaml` to look like this:

```diff
annotations:
  category: Database
+ hypper.cattle.io/namespace: databases
apiVersion: v2
appVersion: 8.0.23
dependencies:
  - name: common
    repository: https://charts.bitnami.com/bitnami
    tags:
      - bitnami-common
    version: 1.x.x
description: Chart to create a Highly available MySQL cluster
home: https://github.com/bitnami/charts/tree/master/bitnami/mysql
icon: https://bitnami.com/assets/stacks/mysql/img/mysql-stack-220x234.png
keywords:
  - mysql
  - database
  - sql
  - cluster
  - high availability
maintainers:
  - email: containers@bitnami.com
    name: Bitnami
name: mysql
sources:
  - https://github.com/bitnami/bitnami-docker-mysql
  - https://mysql.com
version: 8.5.1
```

We added the `hypper.cattle.io/namespace: databases` to the annotation, so now when we install this chart with Hypper it will install into the proper namespace without any need to passing the value on the CLI:

```shell
$ hypper install mysql/                 
Installing chart "mysql" in namespace "databases"…
Done! 👏 

```

See also how we didn't to specify any name for the release? Hypper is smart enough to try to obtain the name from the annotations (like the namespace!) and if it doesn't find it, it uses the name value on the `Chart.yaml`

If we wanted to specify the release name in the annotations as well we just need to add `hypper.cattle.io/release-name` to the annotations as we did above with the namespace and hypper will take care of setting it! 