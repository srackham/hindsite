package main

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// config defines global configuration parameters.
type config struct {
	urlprefix string // For document and index page URLs.
	homepage  string // Use this file in the build directory for /index.html.
	recent    int    // Maximum number of recent index entries.
}

// Config global singleton.
var Config config

func init() {
	Config.urlprefix = "/"
	Config.recent = 5
}

func (conf *config) set(name, value string) error {
	switch name {
	case "homepage":
		if !filepath.IsAbs(value) {
			value = filepath.Join(Cmd.buildDir, value)
		} else if !pathIsInDir(value, Cmd.buildDir) {
			return fmt.Errorf("homepage must reside in build directory: %s", Cmd.buildDir)
		}
		conf.homepage = value
	case "recent":
		re := regexp.MustCompile(`^\d+$`)
		if !re.MatchString(value) {
			return fmt.Errorf("illegal recent value: %s", value)
		}
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		conf.recent = int(i)
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
