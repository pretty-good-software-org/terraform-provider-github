package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"

	"github.com/integrations/terraform-provider-github/v6/github"
)

func main() {
	opts := &plugin.ServeOpts{
		ProviderAddr: "registry.opentofu.org/pretty-good-software-org/github",
		ProviderFunc: github.NewProvider(),
	}

	plugin.Serve(opts)
}
