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

type configs []struct {
	dir  string // Directory contain a configuration file.
	conf config // Combined YAML + TOML config.
}

func newConfig() config {
	conf := config{}
	conf.urlprefix = "/"
	conf.recent = 5
	return conf
}

func (conf *config) set(proj *project, name, value string) error {
	switch name {
	case "author":
		conf.author = value
	case "homepage":
		if !filepath.IsAbs(value) {
			value = filepath.Join(proj.buildDir, value)
		} else if !pathIsInDir(value, proj.buildDir) {
			return fmt.Errorf("homepage must reside in build directory: %s", proj.buildDir)
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
func (conf *config) parseFile(proj *project, f string) error {
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
		if err := conf.set(proj, "author", cf.Author); err != nil {
			return err
		}
	}
	if cf.Homepage != "" {
		if err := conf.set(proj, "homepage", cf.Homepage); err != nil {
			return err
		}
	}
	if cf.Recent != "" {
		if err := conf.set(proj, "recent", cf.Recent); err != nil {
			return err
		}
	}
	if cf.URLPrefix != "" {
		if err := conf.set(proj, "urlprefix", cf.URLPrefix); err != nil {
			return err
		}
	}
	return nil
}

// Parse and merge YAML and TOML config files in directory dir.
func (conf *config) parseConfigFiles(proj *project, dir string) error {
	for _, cf := range []string{"config.toml", "config.yaml"} {
		f := filepath.Join(dir, cf)
		if fileExists(f) {
			proj.println("read config: " + f)
			if err := conf.parseFile(proj, f); err != nil {
				return err
			}
		}
	}
	return nil
}

// Return configuration as YAML formatted string.
func (conf *config) data() (data templateData) {
	data = templateData{}
	data["author"] = conf.author
	data["homepage"] = conf.homepage
	data["recent"] = strconv.Itoa(conf.recent)
	data["urlprefix"] = conf.urlprefix
	return data
}

// Return configuration as YAML formatted string.
func (conf *config) String() (result string) {
	d, _ := yaml.Marshal(conf.data())
	return string(d)
}

// Return merged configurations for contentDir and templateDir locations.
// TODO: This routine will be called many times with the same arguments
// and the same results, caching the results would speed it up.
func (confs configs) configFor(contentDir, templateDir string) config {
	result := newConfig()
	return result
}
