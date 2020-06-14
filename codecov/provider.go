package codecov

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

// Provider returns a terraform.ResourceProvider.
func Provider() terraform.ResourceProvider {
	provider := &schema.Provider{
		DataSourcesMap: map[string]*schema.Resource{
			"ucodecov_settings": dataSourceCodecovSettings(),
		},
	}
	return provider
}
