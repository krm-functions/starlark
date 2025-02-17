# Starlark function

The `starlark` KRM function provides a [Starlark](https://starlark-lang.org/) interpreter that allows for resource modification, including adding or removing resources.

Starlark programs should be provided as a `StarlarkRun` function-config:

```yaml
apiVersion: fn.kpt.dev/v1alpha1
kind: StarlarkRun
metadata:
  name: set-annotations
params:
  toAdd:
    foo: bar
    baz: olo
source: |
  def main():
    toAdd = ctx.resource_list["functionConfig"]["params"]["toAdd"]
    for resource in ctx.resource_list["items"]:
      for key in toAdd:
        resource["metadata"]["annotations"][key] = toAdd[key]
  main()
```

Alternatively a `ConfigMap` can be used. The Starlark source must provided in `source`
and additional fields can be referenced in the code:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: set-annotations
data:
  foo: bar
  baz: olo
  source: |
    def main():
      fncfg = ctx.resource_list["functionConfig"]["data"]
      for resource in ctx.resource_list["items"]:
        for key in fncfg: 
          resource["metadata"]["annotations"][key] = fncfg[key]
    main()
```

Example:

```shell
kpt fn source examples | kpt fn eval --results-dir _results - --image ghcr.io/krm-functions/starlark --fn-config example-function-config/set-annotation.yaml
```

which will produce:

```
[RUNNING] "ghcr.io/krm-functions/starlark"
[PASS] "ghcr.io/krm-functions/starlark" in 3s

...

apiVersion: config.kubernetes.io/v1
kind: ResourceList
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    labels:
      app: test
    name: test
    annotations:
      baz: olo
      foo: bar
```

note how the annotations have been added by the Starlark program.
