package runscope

import (
	"log"
)

// Config contains runscope provider settings
type Config struct {
	AccessToken string
	ApiUrl      string
}

func (c *Config) Client() (*Client, error) {
	client := NewClient(c.ApiUrl, c.AccessToken)

	log.Printf("[INFO] runscope client configured for server %s", c.ApiUrl)

	return client, nil
}
