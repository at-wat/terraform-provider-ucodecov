package codecov

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider returns a Provider.
func Provider() *schema.Provider {
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"token_v2": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CODECOV_API_V2_TOKEN", nil),
				Sensitive:   true,
			},
			"endpoint_base": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "https://codecov.io",
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			"ucodecov_settings": dataSourceCodecovConfig(),
		},
		ConfigureContextFunc: providerConfigure,
	}
	return provider
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return &providerConfig{
		TokenV2:      d.Get("token_v2").(string),
		EndpointBase: d.Get("endpoint_base").(string),
	}, diag.Diagnostics{}
}

type providerConfig struct {
	TokenV2      string
	EndpointBase string
}
