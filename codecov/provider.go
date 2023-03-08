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
			"token": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CODECOV_API_V2_TOKEN", nil),
				Sensitive:   true,
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
	return d.Get("token").(string), diag.Diagnostics{}
}
