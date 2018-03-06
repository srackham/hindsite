package main

import (
	"fmt"
	"regexp"
)

// ConfigParams defines global configuration parameters.
type ConfigParams struct {
	urlprefix string
}

// Config contains global configuration parameters.
var Config ConfigParams

func (conf *ConfigParams) set(name, value string) error {
	switch name {
	case "urlprefix":
		re := regexp.MustCompile(`^(/\w+|)$`)
		if !re.MatchString(value) {
			return fmt.Errorf("illegal urlprefix value: %s", value)
		}
		conf.urlprefix = value
	default:
		return fmt.Errorf("illegal configuration parameter name: %s", name)
	}
	return nil
}
