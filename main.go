package main

import (
	"flag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/hetznercloud/terraform-provider-hcloud/hcloud"
)

func main() {
	var debugMode bool

	flag.BoolVar(&debugMode, "debuggable", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	if debugMode {
		plugin.Serve(&plugin.ServeOpts{
			ProviderAddr: "registry.terraform.io/hetznercloud/hcloud",
			ProviderFunc: hcloud.Provider,
		})
	} else {
		plugin.Serve(&plugin.ServeOpts{
			ProviderFunc: hcloud.Provider})
	}
}
