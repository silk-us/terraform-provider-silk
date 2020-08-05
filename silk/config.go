package silk

import "github.com/silk-us/silk-sdp-go-sdk/silksdp"

// Config is per-provider, specifies where to connect to Rubrik CDM
type Config struct {
	Server   string
	Username string
	Password string
}

// Client returns a *rubrik.Client to interact with the configured Rubrik CDM instance
func (c *Config) Client() (*silksdp.Credentials, error) {

	return silksdp.Connect(c.Server, c.Username, c.Password), nil
}
