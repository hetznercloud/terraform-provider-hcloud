Terraform Provider for the Hetzner Cloud
==================
[![GitHub release](https://img.shields.io/github/release/hetznercloud/terraform-provider-hcloud.svg)](https://github.com/hetznercloud/terraform-provider-hcloud/releases/latest) [![Build Status](https://travis-ci.org/hetznercloud/terraform-provider-hcloud.svg?branch=master)](https://travis-ci.org/hetznercloud/terraform-provider-hcloud)

- Website: https://www.terraform.io
<img src="https://cdn.rawgit.com/hashicorp/terraform-website/master/content/source/assets/images/logo-hashicorp.svg" width="600px">

Maintainers
-----------

This provider plugin is maintained by:

* The Hetzner Cloud Team

Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) 0.11.x
-	[Go](https://golang.org/doc/install) 1.9 (to build the provider plugin)

Installing the provider
---------------------
To install the Hetzner Cloud Terraform provider use the binary distributions from the Releases page. The packages are available for the same OS/ARCH combinations as Terraform itself:

- Mac OS X
`64-bit`
- FreeBSD
`32-bit` `64-bit` `Arm`
- Linux
`32-bit` `64-bit` `Arm`
- OpenBSD
`32-bit` `64-bit`
- Solaris
`64-bit`
- Windows
`32-bit` `64-bit`

Download and uncompress the latest release for your OS. This example uses the linux binary for amd64.

```sh
$ wget https://github.com/hetznercloud/terraform-provider-hcloud/releases/download/v1.1.0/terraform-provider-hcloud_v1.1.0_linux_amd64.zip
$ unzip terraform-provider-hcloud_v1.1.0_linux_amd64.zip
```

Now copy the binary into the Terraform plugins folder.

```sh
$ mkdir -p ~/.terraform.d/plugins/
$ mv terraform-provider-hcloud ~/.terraform.d/plugins/
```

Building the provider
---------------------

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

Using the provider
----------------------

See the [Hetzner Cloud Provider documentation](docs/readme.md) to get started using the Hetzner Cloud provider.


Developing the provider
---------------------------

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (version 1.8+ is *required*). You'll also need to correctly setup a [GOPATH](http://golang.org/doc/code.html#GOPATH), as well as adding `$GOPATH/bin` to your `$PATH`.

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
