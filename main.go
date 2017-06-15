package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/shelmangroup/terraform-provider-coredns/coredns"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: coredns.Provider,
	})
}
