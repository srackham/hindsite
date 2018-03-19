package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	yaml "gopkg.in/yaml.v2"
)

// config defines global configuration parameters.
type config struct {
	origin    string // Configuration file directory.
	author    string // Default document author.
	homepage  string // Use this file (relative to the build directory) for /index.html.
	recent    int    // Maximum number of recent index entries.
	urlprefix string // For document and index page URLs.
}

type configs []config

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

// Parse all config files from project content and templates directory into
// project confs.
func (proj *project) parseConfigs() error {
	for _, d := range []string{proj.contentDir, proj.templateDir} {
		if proj.contentDir == proj.templateDir && d == proj.templateDir {
			break
		}
		err := filepath.Walk(d, func(f string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				return nil
			}
			conf := config{}
			conf.origin = f
			found := false
			for _, v := range []string{"config.toml", "config.yaml"} {
				cf := filepath.Join(f, v)
				if fileExists(cf) {
					found = true
					proj.println("read config: " + cf)
					if err := conf.parseFile(proj, cf); err != nil {
						return err
					}
				}
			}
			if found {
				proj.confs = append(proj.confs, conf)
				proj.println(conf.String())
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	// Sort configurations by ascending origin directory to ensure deeper
	// configurations have precedence.
	sort.Slice(proj.confs, func(i, j int) bool {
		return proj.confs[i].origin < proj.confs[j].origin
	})
	return nil
}

// Merge non-"zero" configuration fields into configuration.
func (conf *config) merge(from config) {
	if from.origin != "" {
		conf.origin = from.origin
	}
	if from.author != "" {
		conf.author = from.author
	}
	if from.homepage != "" {
		conf.homepage = from.homepage
	}
	if from.recent != 0 {
		conf.recent = from.recent
	}
	if from.urlprefix != "" {
		conf.urlprefix = from.urlprefix
	}
}

// Return merged configuration that will be applied to files in contentDir and
// templateDir locations. Content directory configuration files take precedence
// over template directory configuration files. proj.confs are sorted by
// project.confs.origin ascending which ensures configuration directory
// heirarchy precedence.
func (proj *project) configFor(contentDir, templateDir string) config {
	result := config{recent: 5, urlprefix: "/"} // Set default configuration.
	for _, conf := range proj.confs {
		if templateDir == conf.origin || pathIsInDir(templateDir, conf.origin) {
			result.merge(conf)
		}
	}
	for _, conf := range proj.confs {
		if contentDir == conf.origin || pathIsInDir(contentDir, conf.origin) {
			result.merge(conf)
		}
	}
	return result
}
