package main

// ConfigParams defines global configuration parameters.
type ConfigParams struct {
	urlprefix string
}

// Config contains global configuration parameters.
var Config ConfigParams

func init() {
	Config.urlprefix = "/blog"
}
