package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/jaceq/terraform-provider-pagerduty/pagerduty"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: pagerduty.Provider})
}
