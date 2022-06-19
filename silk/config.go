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

// find is a helper function that is used to determine if val is in the slice
// This is mainly used to find the PWWN that need be added or removed from the host.
func find(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

func unique(stringSlice []string) []string {
    keys := make(map[string]bool)
    list := []string{}	
    for _, entry := range stringSlice {
        if _, value := keys[entry]; !value {
            keys[entry] = true
            list = append(list, entry)
        }
    }    
    return list
}