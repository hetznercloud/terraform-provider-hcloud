# Terraform Provider for the Hetzner Cloud

[![GitHub release](https://img.shields.io/github/tag/hetznercloud/terraform-provider-hcloud.svg?label=release)](https://github.com/hetznercloud/terraform-provider-hcloud/releases/latest) [![Actions Status](https://github.com/hetznercloud/terraform-provider-hcloud/workflows/test/badge.svg)](https://github.com/hetznercloud/terraform-provider-hcloud/actions)[![Actions Status](https://github.com/hetznercloud/terraform-provider-hcloud/workflows/release/badge.svg)](https://github.com/hetznercloud/terraform-provider-hcloud/actions)
[![Codecov](https://codecov.io/gh/hetznercloud/terraform-provider-hcloud/graph/badge.svg?token=og7OhpoV5W)](https://codecov.io/gh/hetznercloud/terraform-provider-hcloud/tree/main)

- Documentation: https://registry.terraform.io/providers/hetznercloud/hcloud/latest/docs

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads)
  - HashiCorp recommends to use the two latest terraform releases (1.6.x, 1.7.x). Our test suite validates that our provider works with these versions.
  - This provider uses the [terraform plugin protocol version 6](https://developer.hashicorp.com/terraform/plugin/terraform-plugin-protocol#protocol-version-6), and should work with all tools (ie. Terraform & OpenTofu) that supports it.
- [Go](https://go.dev/doc/install) 1.21.x (to build the provider plugin)

## API Stability

This Go module implements a Terraform Provider for Hetzner Cloud
Services. We thus guarantee backwards compatibility only for use through
Terraform HCL. The actual _Go code_ in this repository _may change
without a major version increase_.

Currently the code is mostly located in the `hcloud` package. In the
long term we want to move most of the `hcloud` package into individual
sub-packages located in the `internal` directory. The goal is a
structure similar to HashiCorp's [Terraform Provider
Scaffolding](https://github.com/hashicorp/terraform-provider-scaffolding)

## Using the provider

If you are building the provider, follow the instructions to [install it as a plugin](https://www.terraform.io/docs/plugins/basics.html#installing-a-plugin). After placing it into your plugins directory, run `terraform init` to initialize it.

## Building the provider

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

## Developing the provider

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
