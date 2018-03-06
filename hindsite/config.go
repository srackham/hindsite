package main

import (
	"fmt"
	"regexp"
	"strings"
)

// config defines global configuration parameters.
type config struct {
	urlprefix string // For document and index page URLs.
}

// Config is global singleton.
var Config config

func (conf *config) set(name, value string) error {
	switch name {
	case "urlprefix":
		re := regexp.MustCompile(`^((http|/)\S+|)$`)
		if !re.MatchString(value) {
			return fmt.Errorf("illegal urlprefix value: %s", value)
		}
		conf.urlprefix = strings.TrimSuffix(value, "/")
	default:
		return fmt.Errorf("illegal configuration parameter name: %s", name)
	}
	return nil
}

// TODO ???
func (conf *config) toURL(rootdir, filename string) error {
	return nil
}
