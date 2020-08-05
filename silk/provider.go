package silk

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider -
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"server": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("SILK_SDP_SERVER", nil),
				Description: "The IP Address of a Silk server.",
			},
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("SILK_SDP_USERNAME", nil),
				Description: "The username used to authenticate against the Silk server.",
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("SILK_SDP_PASSWORD", nil),
				Description: "The password used to authenticate against the Silk Sever.",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"silk_volume":       resourceSilkVolume(),
			"silk_volume_group": resourceSilkVolumeGroup(),
			"silk_host":         resourceSilkHost(),
			"silk_host_pwwn":    resourceSilkHostPWWN(),
			"silk_host_group":   resourceSilkHostGroup(),
		},
		DataSourcesMap: map[string]*schema.Resource{},

		ConfigureFunc: providerConfigure,
	}
}

// providerConfigure creates an interface{} that is stored and passed into subsequent resources as the
// meta parameter. This return value is used to pass along the configured Silk Go SDK API client
func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	config := Config{
		Server:   d.Get("server").(string),
		Username: d.Get("username").(string),
		Password: d.Get("password").(string),
	}

	return config.Client()
}
