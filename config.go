package main

import (
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	yaml "gopkg.in/yaml.v2"
)

type config struct {
	origin string // Configuration file directory.
	// Configuration parameters.
	author    *string           // Default document author (nil if undefined).
	templates *string           // Comma separated list of content file name extensions to undergo text template expansion (nil if undefined).
	homepage  string            // Use this built file for /index.html.
	paginate  int               // Number of documents per index page. No pagination if zero or less.
	urlprefix string            // Prefix for synthesized document and index page URLs.
	permalink string            // URL template.
	id        string            // Front matter id behavior: "optional",  "mandatory" or "urlpath".
	exclude   []string          // List of excluded content patterns.
	include   []string          // List of included content patterns.
	timezone  *time.Location    // Time zone for site generation.
	user      map[string]string // User defined configuration key/values.
	// Date formats for template variables: date, shortdate, mediumdate, longdate.
	shortdate  string
	mediumdate string
	longdate   string
}

type configs []config

// checkTemplates returns an error if the templates configuration value is illegal.
func checkTemplates(templates string) error {
	re := regexp.MustCompile("^(\\.\\w+\\|)*(\\.\\w+)+$")
	if templates != "" && !re.MatchString(templates) {
		return fmt.Errorf("illegal templates: %s", templates)
	}
	return nil
}

// Return default configuration.
func newConfig() config {
	conf := config{
		exclude:    []string{".*"},
		id:         "optional",
		paginate:   5,
		shortdate:  "2006-01-02",
		mediumdate: "2-Jan-2006",
		longdate:   "Mon Jan 2, 2006",
		user:       map[string]string{},
	}
	conf.timezone, _ = time.LoadLocation("Local")
	return conf
}

// parseFile parses a configuration file.
func (conf *config) parseFile(proj *project, f string) error {
	text, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}
	cf := struct {
		Author     *string
		Templates  *string
		Exclude    *string
		Include    *string
		Homepage   string
		ID         string
		Permalink  string
		URLPrefix  string
		Paginate   int
		Timezone   string
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
		panic("illegal configuration file extension: " + f)
	}
	// Validate and merge parsed configuration.
	if cf.Author != nil {
		conf.author = cf.Author
	}
	if cf.Templates != nil {
		if err := checkTemplates(*cf.Templates); err != nil {
			return err
		}
		conf.templates = cf.Templates
	}
	if cf.Permalink != "" {
		conf.permalink = cf.Permalink
	}
	if cf.Homepage != "" {
		home := cf.Homepage
		home = filepath.FromSlash(home)
		if !filepath.IsAbs(home) {
			home = filepath.Join(proj.buildDir, home)
		} else {
			return fmt.Errorf("homepage must be relative to the build directory: %s", proj.buildDir)
		}
		if !pathIsInDir(home, proj.buildDir) {
			return fmt.Errorf("homepage must reside in build directory: %s", proj.buildDir)
		}
		if dirExists(home) {
			return fmt.Errorf("homepage cannot be a directory: %s", home)
		}
		conf.homepage = home
	}
	if cf.ID != "" {
		switch cf.ID {
		case "optional", "mandatory", "urlpath":
			conf.id = cf.ID
		default:
			return fmt.Errorf("illegal id configuration parameter value: %s", cf.ID)
		}
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
	if cf.Exclude != nil {
		conf.exclude = append([]string{".*"}, strings.Split(filepath.ToSlash(*cf.Exclude), "|")...)
		for _, pat := range conf.exclude {
			if pat == "" {
				return fmt.Errorf("exclude pattern cannot be blank: %s", *cf.Exclude)
			}
		}
	}
	if cf.Include != nil {
		conf.include = strings.Split(filepath.ToSlash(*cf.Include), "|")
		for _, pat := range conf.include {
			if pat == "" {
				return fmt.Errorf("include pattern cannot be blank: %s", *cf.Include)
			}
		}
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

// Return configuration as YAML formatted string.
func (conf *config) data() templateData {
	data := templateData{}
	data["author"] = nz(conf.author)
	data["templates"] = nz(conf.templates)
	data["id"] = conf.id
	data["permalink"] = conf.permalink
	data["homepage"] = conf.homepage
	data["paginate"] = conf.paginate
	data["urlprefix"] = conf.urlprefix
	data["exclude"] = strings.Join(conf.exclude, "|")
	data["include"] = strings.Join(conf.include, "|")
	data["timezone"] = conf.timezone.String()
	data["shortdate"] = conf.shortdate
	data["mediumdate"] = conf.mediumdate
	data["longdate"] = conf.longdate
	data["user"] = conf.user
	return data
}

// Return configuration as YAML formatted string.
func (conf *config) String() (result string) {
	d, _ := yaml.Marshal(conf.data())
	return string(d)
}

// merge merges "non-zero" configuration parameters into configuration.
func (conf *config) merge(from config) {
	if from.origin != "" {
		conf.origin = from.origin
	}
	if from.author != nil {
		conf.author = from.author
	}
	if from.id != "" {
		conf.id = from.id
	}
	if from.templates != nil {
		conf.templates = from.templates
	}
	if from.permalink != "" {
		conf.permalink = from.permalink
	}
	if from.paginate != 0 {
		conf.paginate = from.paginate
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
	if from.homepage != "" {
		conf.homepage = from.homepage
	}
	if from.urlprefix != "" {
		conf.urlprefix = from.urlprefix
	}
	if from.exclude != nil {
		conf.exclude = from.exclude
	}
	if from.include != nil {
		conf.include = from.include
	}
	for k, v := range from.user {
		conf.user[k] = v
	}
}

// joinPrefix joins path elements and prefixes them with the urlprefix. The
// urlprefix cannot be processed by path.Join because it would replace '//' with
// '/' in an absolute urlprefix (e.g. http://example.com),
func (conf *config) joinPrefix(elem ...string) string {
	if strings.HasSuffix(conf.urlprefix, "/") {
		panic("urlprefix has '/' suffix: " + conf.urlprefix)
	}
	if len(elem[0]) == 0 {
		panic("joinPrefix: missing argument")
	}
	if strings.HasPrefix(elem[0], "/") {
		panic("relative URL has '/' prefix: " + elem[0])
	}
	return conf.urlprefix + "/" + path.Join(elem...)
}
