package codecov

import (
	"context"
	"time"

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
			"api_interval": {
				Type:     schema.TypeFloat,
				Optional: true,
				Default:  1.0,
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
	apiCallTick := time.NewTicker(time.Duration(
		float64(time.Second) * d.Get("api_interval").(float64),
	))
	return &providerConfig{
		TokenV2:      d.Get("token_v2").(string),
		EndpointBase: d.Get("endpoint_base").(string),
		APICallTick:  apiCallTick.C,
	}, diag.Diagnostics{}
}

type providerConfig struct {
	TokenV2      string
	EndpointBase string
	APICallTick  <-chan time.Time
}
