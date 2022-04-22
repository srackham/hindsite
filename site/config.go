package site

import (
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/srackham/hindsite/fsx"

	"github.com/BurntSushi/toml"
	yaml "gopkg.in/yaml.v3"
)

type config struct {
	origin string // Configuration file directory.
	// Configuration variables.
	author    *string           // Default document author (nil if undefined).
	templates []string          // List of included content templates.
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

// Unvalidated configuration variable values.
// Undefined configuration variables have a nil pointer value.
type rawConfig struct {
	// Configuration variables
	Author     *string
	Templates  *string
	Exclude    *string
	Include    *string
	Homepage   *string
	ID         *string
	Permalink  *string
	URLPrefix  *string
	Paginate   *string
	Timezone   *string
	ShortDate  *string
	MediumDate *string
	LongDate   *string
	// Command specific variables
	Template *string // new document template
	User     map[string]string
}

// parseVar parses the `NAME=VALUE` var argument `arg` into `vars`.
func parseVar(vars *rawConfig, arg string) error {
	s := strings.SplitN(arg, "=", 2)
	if len(s) != 2 {
		return fmt.Errorf("illegal -var syntax: %s", arg)
	}
	name := s[0]
	val := s[1]
	if strings.HasPrefix(name, "user.") {
		name = strings.TrimPrefix(name, "user.")
		vars.User[name] = val
	} else {
		switch name {
		case "author":
			vars.Author = &val
		case "templates":
			vars.Templates = &val
		case "exclude":
			vars.Exclude = &val
		case "include":
			vars.Include = &val
		case "id":
			vars.ID = &val
		case "permalink":
			vars.Permalink = &val
		case "urlprefix":
			vars.URLPrefix = &val
		case "paginate":
			vars.Paginate = &val
		case "timezone":
			vars.Timezone = &val
		case "shortdate":
			vars.ShortDate = &val
		case "mediumdate":
			vars.MediumDate = &val
		case "longdate":
			vars.LongDate = &val
		case "template":
			vars.Template = &val
		default:
			return fmt.Errorf("illegal -var name: %s", name)
		}
	}
	return nil
}

func (conf *config) parseFile(site *site, f string) error {
	text, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}
	raw := rawConfig{}
	switch filepath.Ext(f) {
	case ".toml":
		_, err := toml.Decode(string(text), &raw)
		if err != nil {
			return err
		}
	case ".yaml":
		err := yaml.Unmarshal(text, &raw)
		if err != nil {
			return err
		}
	default:
		panic("illegal configuration file extension: " + f)
	}
	return conf.mergeRaw(site, raw)
}

// mergeRaw validates raw configuration values and merges them into `conf`.
func (conf *config) mergeRaw(site *site, raw rawConfig) error {
	// Validate and merge parsed configuration.
	if raw.Author != nil {
		conf.author = raw.Author
	}
	if raw.Templates != nil {
		conf.templates = splitPatterns(*raw.Templates)
	}
	if raw.Permalink != nil {
		conf.permalink = *raw.Permalink
	}
	if raw.Homepage != nil {
		home := *raw.Homepage
		home = filepath.FromSlash(home)
		if !filepath.IsAbs(home) {
			home = filepath.Join(site.buildDir, home)
		} else {
			return fmt.Errorf("homepage must be relative to the build directory: %s", site.buildDir)
		}
		if !fsx.PathIsInDir(home, site.buildDir) {
			return fmt.Errorf("homepage must reside in build directory: %s", site.buildDir)
		}
		if fsx.DirExists(home) {
			return fmt.Errorf("homepage cannot be a directory: %s", home)
		}
		conf.homepage = home
	}
	if raw.ID != nil {
		switch *raw.ID {
		case "optional", "mandatory", "urlpath":
			conf.id = *raw.ID
		default:
			return fmt.Errorf("illegal id: %s", *raw.ID)
		}
	}
	if raw.Paginate != nil {
		if n, err := strconv.Atoi(*raw.Paginate); err != nil {
			return fmt.Errorf("illegal paginate value: %s", *raw.Paginate)
		} else {
			conf.paginate = n
		}
	}
	if raw.URLPrefix != nil {
		value := *raw.URLPrefix
		re := regexp.MustCompile(`^((http|/)\S+|)$`)
		if !re.MatchString(value) {
			return fmt.Errorf("illegal urlprefix: %s", value)
		}
		conf.urlprefix = strings.TrimSuffix(value, "/")
	}
	if raw.Exclude != nil {
		conf.exclude = append([]string{".*"}, splitPatterns(*raw.Exclude)...)
	}
	if raw.Include != nil {
		conf.include = splitPatterns(*raw.Include)
	}
	if raw.Timezone != nil {
		tz, err := time.LoadLocation(*raw.Timezone)
		if err != nil {
			return err
		}
		conf.timezone = tz
	}
	if raw.ShortDate != nil {
		conf.shortdate = *raw.ShortDate
	}
	if raw.MediumDate != nil {
		conf.mediumdate = *raw.MediumDate
	}
	if raw.LongDate != nil {
		conf.longdate = *raw.LongDate
	}
	if raw.User != nil {
		conf.user = raw.User
	}
	return nil
}

// Return configuration as a map keyed by parameter name.
func (conf *config) data() templateData {
	data := templateData{}
	data["author"] = nz(conf.author)
	data["templates"] = strings.Join(conf.templates, "|")
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

// String returns configuration as YAML formatted string.
func (conf *config) String() (result string) {
	d, _ := yaml.Marshal(conf.data())
	return string(d)
}

// merge merges "non-zero" configuration variables into configuration.
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

// splitPatterns splits `|` separated file patterns
func splitPatterns(patterns string) []string {
	return strings.Split(filepath.ToSlash(patterns), "|")
}
