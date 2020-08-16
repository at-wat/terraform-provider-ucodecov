package codecov

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider returns a Provider.
func Provider() *schema.Provider {
	provider := &schema.Provider{
		DataSourcesMap: map[string]*schema.Resource{
			"ucodecov_settings": dataSourceCodecovSettings(),
		},
	}
	return provider
}
