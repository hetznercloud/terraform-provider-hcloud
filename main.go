package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6/tf6server"

	"github.com/hetznercloud/terraform-provider-hcloud/hcloud"
)

func main() {
	ctx := context.Background()

	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	providerFactory, err := hcloud.GetMuxedProvider(ctx)
	if err != nil {
		log.Fatal(err)
	}

	var serveOpts []tf6server.ServeOpt

	if debug {
		serveOpts = append(serveOpts, tf6server.WithManagedDebug())
	}

	err = tf6server.Serve(
		"registry.terraform.io/hetznercloud/hcloud",
		providerFactory,
		serveOpts...,
	)

	if err != nil {
		log.Fatal(err)
	}
}
