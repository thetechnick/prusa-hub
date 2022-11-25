package linkclient

import "strings"

type ClientOptions struct {
	Endpoint string
	APIKey   string
}

type clientOption interface {
	ApplyToClientOptions(o *ClientOptions)
}

type WithEndpoint string

func (ep WithEndpoint) ApplyToClientOptions(o *ClientOptions) {
	// ensure there is always a single trailing "/"
	o.Endpoint = strings.TrimRight(string(ep), "/") + "/"
}

type WithAPIKey string

func (key WithAPIKey) ApplyToClientOptions(o *ClientOptions) {
	o.APIKey = string(key)
}
