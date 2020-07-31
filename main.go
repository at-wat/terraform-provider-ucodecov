package main

import (
	"github.com/at-wat/terraform-provider-ucodecov/codecov"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: codecov.Provider,
	})
}
