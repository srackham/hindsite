package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	yaml "gopkg.in/yaml.v2"
)

// config defines global configuration parameters.
type config struct {
	author    string // Default document author.
	homepage  string // Use this file (relative to the build directory) for /index.html.
	recent    int    // Maximum number of recent index entries.
	urlprefix string // For document and index page URLs.
}

// Config global singleton.
var Config config

func newConfig() config {
	conf := config{}
	Config.urlprefix = "/"
	Config.recent = 5
	return conf
}

func (conf *config) set(name, value string) error {
	switch name {
	case "author":
		conf.author = value
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

// Parse config file.
func (conf *config) parseFile(f string) error {
	text, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}
	cf := struct {
		Author    string
		Homepage  string
		URLPrefix string
		Recent    string
	}{}
	switch filepath.Ext(f) {
	case ".toml":
		_, err := toml.Decode(string(text), &cf)
		if err != nil {
			return err
		}
	case ".yaml":
		err := yaml.Unmarshal(text, &cf)
		if err != nil {
			return err
		}
	default:
		panic("illegal configuration file extension")
	}
	// Merge parsed configuration.
	if cf.Author != "" {
		if err := conf.set("author", cf.Author); err != nil {
			return err
		}
	}
	if cf.Homepage != "" {
		if err := conf.set("homepage", cf.Homepage); err != nil {
			return err
		}
	}
	if cf.Recent != "" {
		if err := conf.set("recent", cf.Recent); err != nil {
			return err
		}
	}
	if cf.URLPrefix != "" {
		if err := conf.set("urlprefix", cf.URLPrefix); err != nil {
			return err
		}
	}
	return nil
}
