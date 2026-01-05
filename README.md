# Terraform Provider for the Hetzner Cloud

[![GitHub release](https://img.shields.io/github/tag/hetznercloud/terraform-provider-hcloud.svg?label=release)](https://github.com/hetznercloud/terraform-provider-hcloud/releases/latest) [![Actions Status](https://github.com/hetznercloud/terraform-provider-hcloud/workflows/test/badge.svg)](https://github.com/hetznercloud/terraform-provider-hcloud/actions)[![Actions Status](https://github.com/hetznercloud/terraform-provider-hcloud/workflows/release/badge.svg)](https://github.com/hetznercloud/terraform-provider-hcloud/actions)
[![Codecov](https://codecov.io/gh/hetznercloud/terraform-provider-hcloud/graph/badge.svg?token=og7OhpoV5W)](https://codecov.io/gh/hetznercloud/terraform-provider-hcloud/tree/main)

- Documentation: https://registry.terraform.io/providers/hetznercloud/hcloud/latest/docs

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/install) or [OpenTofu](https://opentofu.org/docs/intro/install/)
  - Our provider tests run with Terraform or OpenTofu releases that are supported upstream.
  - Our provider should work with any tool that supports the [terraform plugin protocol version 6](https://developer.hashicorp.com/terraform/plugin/terraform-plugin-protocol#protocol-version-6).
- [Go](https://go.dev/doc/install) (to build the provider plugin)

## Development

### API Stability

This Go module implements a Terraform Provider for Hetzner Cloud
Services. We thus guarantee backwards compatibility only for use through
Terraform HCL. The actual _Go code_ in this repository _may change
without a major version increase_.

Currently, the code is mostly located in the `hcloud` package. In the
long term we want to move most of the `hcloud` package into individual
sub-packages located in the `internal` directory. The goal is a
structure similar to HashiCorp's [Terraform Provider
Scaffolding](https://github.com/hashicorp/terraform-provider-scaffolding)

### Using the provider

If you are building the provider, follow the instructions to [install it as a plugin](https://www.terraform.io/docs/plugins/basics.html#installing-a-plugin). After placing it into your plugins directory, run `terraform init` to initialize it.

### Building the provider

Clone repository to: `$GOPATH/src/github.com/hetznercloud/terraform-provider-hcloud`

```sh
$ mkdir -p $GOPATH/src/github.com/hetznercloud; cd $GOPATH/src/github.com/hetznercloud
$ git clone https://github.com/hetznercloud/terraform-provider-hcloud.git
```

Enter the provider directory and build the provider

```sh
$ cd $GOPATH/src/github.com/hetznercloud/terraform-provider-hcloud
$ make build
```

### Developing the provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (version 1.14+ is _required_). You'll also need to correctly setup a [GOPATH](http://golang.org/doc/code.html#GOPATH), as well as adding `$GOPATH/bin` to your `$PATH`.

To compile the provider, run `make build`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

```sh
$ make build
...
$ ./bin/terraform-provider-hcloud
...
```

In order to test the provider, you can simply run `make test`.

```sh
$ make test
```

In order to run the full suite of Acceptance tests run `make testacc`.

_Note:_ Acceptance tests create real resources, and often cost money to run.

```
$ make testacc
```

You may save your acceptance tests environment variables in the `.env` file, for example:

```sh
$ cat .env
HCLOUD_TOKEN=YOUR_API_TEST_TOKEN
#TF_ACC=1
TF_LOG=DEBUG
TF_LOG_PATH_MASK=test-%s.log

$ go test -v -timeout=30m -parallel=8 ./internal/server
=== RUN   TestAccHcloudDataSourceServerTest
# ...
```

### Running a local build

Choose a terraform cli config file path:

```sh
export TF_CLI_CONFIG_FILE="terraform.tfrc"
```

In the terraform cli config file, override the lookup path for the `hetznercloud/hcloud` provider to use the local build:

```sh
cat <<EOF >"$TF_CLI_CONFIG_FILE"
provider_installation {
  dev_overrides {
    "hetznercloud/hcloud" = "$PWD"
  }

  direct {}
}
EOF
```

Build the provider, resulting in a `terraform-provider-hcloud` binary:

```sh
make build
```

Finally, run your terraform plan to see if it works:

```sh
tofu plan
```

You should see the following warning:

```
╷
│ Warning: Provider development overrides are in effect
│
│ The following provider development overrides are set in the CLI configuration:
│  - hetznercloud/hcloud in /home/user/code/github.com/hetznercloud/terraform-provider-hcloud
│
│ The behavior may therefore not match any released version of the provider and applying changes may cause the state to become incompatible with published releases.
╵
```

### Releasing experimental features

To publish experimental features as part of regular releases:

- an announcement, including a link to a changelog entry, must be added to the release notes.

- an `Experimental` notice, must be added to the experimental resource, datasource, and functions descriptions:

  ```go
  func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema.MarkdownDescription = `
  Manage a Hetzner Cloud Product.

  See https://docs.hetzner.cloud/reference/cloud#product for more details.
  `
    experimental.Product.AppendNotice(&resp.Schema.MarkdownDescription)
  }
  ```

- a `Experimental` warning must be logged when experimental resource, datasource, or functions are being used:

  ```go
  func (r *Resource) Configure(_ context.Context, _ resource.ConfigureRequest, resp *resource.ConfigureResponse) {
    experimental.Product.AppendDiagnostic(&resp.Diagnostics)
  }
  ```

### Deprecating attributes

When deprecating an attributes:

1. Mark the attribute in the schema as deprecated (for SDKv2, the docs template may also need updating), in the message explain the deprecation and link to changelog, for example:

   ```go
   func Resource() *schema.Resource {
       return &schema.Resource{
           // ...
           Schema: map[string]*schema.Schema{
               "datacenter": {
                   // ...
                   Deprecated: "The datacenter attribute is deprecated and will be removed after 1 July 2026. Please use the location attribute instead. See https://docs.hetzner.cloud/changelog#2025-12-16-phasing-out-datacenters.",
                   // ...
               },
           },
       }
   }
   ```

2. - For inputs: Implement backwards-compatible behaviour if possible
   - For output: Keep writing the attribute for as long as it is returned from the API. Once it is no longer
     returned the code should return the user config/previous state if available, or `""` (SDKv2) / `null` (Plugin Framework).

3. In the resource and datasource docs, add a `## Deprecations` section with a subsection for this field, explaining the behaviour and urging users to upgrade to a compatible version, for example:

   ```md
   ### `datacenter` attribute

   The `datacenter` attribute is deprecated, use the `ABC` attribute instead.

   See our the [API changelog](https://docs.hetzner.cloud/changelog#2025-12-16-phasing-out-datacenters) for more details.

   -> Please upgrade to `v1.W.0+` of the provider to avoid issues once the Hetzner Cloud API no longer returns the `XYZ` attribute.
   ```

4. Highlight the deprecation in the Release Notes.
