package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	yaml "gopkg.in/yaml.v2"
)

type config struct {
	origin string // Configuration file directory.
	// Configuration parameters.
	author    string            // Default document author.
	homepage  string            // Use this file (relative to the build directory) for /index.html.
	paginate  int               // Number of documents per index page. No pagination if zero or less.
	urlprefix string            // Prefix for synthesised document and index page URLs.
	exclude   []string          // List of excluded content directory paths.
	timezone  *time.Location    // Time zone for site generation.
	user      map[string]string // User defined configuration key/values.
	// Date formats for template variables: date, shortdate, mediumdate, longdate.
	shortdate  string
	mediumdate string
	longdate   string
}

type configs []config

// Return default configuration.
func newConfig() config {
	conf := config{
		paginate:   5,
		shortdate:  "2006-01-02",
		mediumdate: "2-Jan-2006",
		longdate:   "Mon Jan 2, 2006",
		user:       map[string]string{},
	}
	conf.timezone, _ = time.LoadLocation("Local")
	return conf
}

// Parse config file.
func (conf *config) parseFile(proj *project, f string) error {
	text, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}
	cf := struct {
		Author     string
		Homepage   string
		URLPrefix  string
		Paginate   int
		Timezone   string
		Exclude    string
		ShortDate  string
		MediumDate string
		LongDate   string
		User       map[string]string
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
	// Validate and merge parsed configuration.
	if cf.Author != "" {
		conf.author = cf.Author
	}
	if cf.Homepage != "" {
		value := cf.Homepage
		if !filepath.IsAbs(value) {
			value = filepath.Join(proj.buildDir, value)
		} else if !pathIsInDir(value, proj.buildDir) {
			return fmt.Errorf("homepage must reside in build directory: %s", proj.buildDir)
		}
		conf.homepage = value
	}
	if cf.Paginate != 0 {
		conf.paginate = cf.Paginate
	}
	if cf.URLPrefix != "" {
		value := cf.URLPrefix
		re := regexp.MustCompile(`^((http|/)\S+|)$`)
		if !re.MatchString(value) {
			return fmt.Errorf("illegal urlprefix value: %s", value)
		}
		conf.urlprefix = strings.TrimSuffix(value, "/")
	}
	if cf.Exclude != "" {
		conf.exclude = strings.Split(cf.Exclude, "|")
	}
	if cf.Timezone != "" {
		tz, err := time.LoadLocation(cf.Timezone)
		if err != nil {
			return err
		}
		conf.timezone = tz
	}
	if cf.ShortDate != "" {
		conf.shortdate = cf.ShortDate
	}
	if cf.MediumDate != "" {
		conf.mediumdate = cf.MediumDate
	}
	if cf.LongDate != "" {
		conf.longdate = cf.LongDate
	}
	if cf.User != nil {
		conf.user = cf.User
	}
	return nil
}

// String returns configuration as YAML formatted string.
func (conf *config) String() (result string) {
	data := templateData{}
	data["author"] = conf.author
	data["homepage"] = conf.homepage
	data["paginate"] = conf.paginate
	data["urlprefix"] = conf.urlprefix
	data["exclude"] = strings.Join(conf.exclude, "|")
	data["timezone"] = conf.timezone.String()
	data["shortdate"] = conf.shortdate
	data["mediumdate"] = conf.mediumdate
	data["longdate"] = conf.longdate
	data["user"] = conf.user
	d, _ := yaml.Marshal(data)
	return string(d)
}

// parseConfig parses all configuration files from the project template
// directory to project `confs`.
func (proj *project) parseConfigs() error {
	if !dirExists(proj.templateDir) {
		return fmt.Errorf("missing template directory: " + proj.templateDir)
	}
	err := filepath.Walk(proj.templateDir, func(f string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && f == filepath.Join(proj.templateDir, "init") {
			return filepath.SkipDir
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
				proj.verbose("read config: " + cf)
				if err := conf.parseFile(proj, cf); err != nil {
					return err
				}
			}
		}
		if found {
			proj.confs = append(proj.confs, conf)
			proj.verbose2(conf.String())
		}
		return nil
	})
	if err != nil {
		return err
	}
	// Sort configurations by ascending origin directory to ensure deeper
	// configurations have precedence.
	sort.Slice(proj.confs, func(i, j int) bool {
		return proj.confs[i].origin < proj.confs[j].origin
	})
	proj.rootConf = proj.configFor(proj.templateDir)
	proj.verbose2("root config: \n" + proj.rootConf.String())
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
	if from.paginate != 0 {
		conf.paginate = from.paginate
	}
	if from.urlprefix != "" {
		conf.urlprefix = from.urlprefix
	}
	if len(from.exclude) != 0 {
		conf.exclude = from.exclude
	}
	if from.timezone != nil {
		conf.timezone = from.timezone
	}
	if from.shortdate != "" {
		conf.shortdate = from.shortdate
	}
	if from.mediumdate != "" {
		conf.mediumdate = from.mediumdate
	}
	if from.longdate != "" {
		conf.longdate = from.longdate
	}
	for k, v := range from.user {
		conf.user[k] = v
	}
}

// configFor returns the merged configuration for the dir directory.
// Configuration files that are in the dir path are merged working from top
// (lowest precedence) to bottom.
//
// For example, if dir is `template/posts/james` then directories are searched
// in the following order: `template`, `template/posts`, `template/posts/james`
// with configuration entries from `template` having lowest precedence.
//
// The `proj.confs` have been sorted by configuration `origin` in ascending
// order to ensure the directory precedence.
func (proj *project) configFor(dir string) config {
	result := newConfig()
	for _, conf := range proj.confs {
		if dir == conf.origin || pathIsInDir(dir, conf.origin) {
			result.merge(conf)
		}
	}
	return result
}
