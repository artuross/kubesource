# `kubesource`: vendor Kubernetes manifests from upstream sources

`kubesource` is a simple tool designed to render Kubernetes manifests and vendor them directly into your repository. It uses `kustomize` under the hood to handle the rendering process.

`kubesource` can also filter out unwanted resources based on some metadata.

## installation

```sh
go install github.com/artuross/kubesource/cmd/kubesource@latest
```

## usage

```sh
kubesource [flags]
```

`kubesource` will scan the current directory and all subdirectories for `kubesource.yaml` files, then process each file it finds.

## why

I created `kubesource` to solve 2 problems:

1. Vendor Kubernetes manifests

	I do not want to rely on external servers to serve the manifests my cluster depends on. I want to have a copy of those manifests in my repository, so I can easily review and version them.

	This is particularly difficult, because Kubernetes is heavily fragmented. Some projects use Helm charts, some other expose raw manifests.

2. Filter any unwanted resources

	While working with Kubernetes, I found that quality of 3rd party manifests vary a lot.

	Examples:

	- I prefer to use a custom namespace for each component, but found that some Helm charts **always** create a namespace. I prefer to manage namespaces myself (outside of the provisioned app to avoid finalizer hell), so I want to filter out those resources.
	- Some Helm charts create CRDs with no way to disable that. I like to manage CRDs separately.
	- I would like to exclude any default Secrets, such as generated certificates.

While both objectives can be achieved with `kustomize` alone, I found that the process is a bit difficult. Thus, `kubesource` aims to simplify the process while leaving further customization to `kustomize`.

## examples

Consider an example `kubesource.yaml` file:

```yaml
apiVersion: kubesource.rcwz.pl/v1alpha2
kind: Config
sources:
  - sourceDir: ./source
    targets:
      - directory: ./app
        filter:
          exclude:
            - kind: CustomResourceDefinition
            - kind: Secret
      - directory: ./crds
        filter:
          include:
            - kind: CustomResourceDefinition
```

When executed, `kubesource` will attempt to find all directories containing `kubesource.yaml` file. For each file, it will:

1. For each entry in `sources`:
   1. Render the manifests in `sourceDir` with `kustomize`.
   2. For each directory in `targets`:
      1. Filter the rendered manifests based on `filter`.
      2. Split the manifests into multiple files, one per resource.
      3. Write all matching manifests to `directory`.

In the example above, `kubesource` will render manifests from `./source`, then:

- write all resources except `CustomResourceDefinition` and `Secret` to `./app`;
- write all `CustomResourceDefinition` to `./crds`.

### multiple sources

You can specify multiple sources in a single configuration file:

```yaml
apiVersion: kubesource.rcwz.pl/v1alpha2
kind: Config
sources:
  - sourceDir: _source-internal
    targets:
      - directory: app-internal/base
  - sourceDir: _source-external
    targets:
      - directory: app-external/base
```

This will render each source directory separately and write the results to the specified target directories.

### valid filters

Example below includes all supported filters.

```yaml
apiVersion: kubesource.rcwz.pl/v1alpha2
kind: Config
sources:
  - sourceDir: ./source
    targets:
      - directory: ./app
        filter:
          include:
            - apiVersion: v1
              kind: ConfigMap
              metadata:
                name: my-config
                namespace: my-namespace
                labels:
                  app: my-app
                  version: v1.0.0
          exclude:
            - apiVersion: v1
              kind: Secret
              metadata:
                name: my-secret
                namespace: my-namespace
                labels:
                  app: my-app
                  version: v1.0.0
```

In filters, ALL fields are optional. If a property is not specified, it will not be used for filtering.

For example, in the example above a `ConfigMap` will only be included if it matches all specified fields (`apiVersion`, `kind`, `metadata.name`, `metadata.namespace` and both labels), but it could have additional labels which will be ignored.

#### `include` filters

`include` filters are additive. If multiple `include` filters are specified, a resource will be included if it matches **any** of the filters.

If no `include` filters are specified, all resources are included (unless excluded by `exclude` filter).

#### `exclude` filters

`exclude` filters are subtractive. If multiple `exclude` filters are specified, a resource will be excluded if it matches **any** of the filters.

If no `exclude` filters are specified, no resources are excluded.

`exclude` filters are applied after `include` filters. Thus, it is possible to include all `ConfigMap` resources and exclude a specific one.
