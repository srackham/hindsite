package site

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/srackham/hindsite/fsx"
	"github.com/srackham/hindsite/slice"

	"github.com/fatih/color"
)

// Build ldflags.
var (
	// VERS is the latest hindsite version tag. Set by linker -ldflags "-X main.VERS=..."
	VERS = "-"
	// OS is the target operating system and architecture. Set by linker -ldflags "-X main.OS=..."
	OS = "-"
	// BUILD is the date the executable was built.
	BUILT = "-"
	// COMMIT is the Git commit hash.
	COMMIT = "-"
)

type site struct {
	command       string
	cmdargs       []string
	executable    string
	in            chan string
	out           chan string
	confs         []config
	docs          documentsLookup
	idxs          indexes
	htmlTemplates htmlTemplates
	textTemplates textTemplates
	// Command options
	siteDir     string
	contentDir  string
	templateDir string
	buildDir    string
	indexDir    string
	initDir     string
	from        string
	drafts      bool
	lint        bool
	launch      bool
	httpport    uint16
	lrport      uint16
	livereload  bool
	navigate    bool
	keep        bool
	verbosity   int
	vars        rawConfig
}

// New creates a new site.
func New() site {
	return site{
		httpport:   1212,
		lrport:     35729,
		livereload: true,
	}
}

// Execute runs a hindsite command specified by CLI args.
func (site *site) Execute(args []string) error {
	var err error
	err = site.parseArgs(args)
	if err == nil {
		switch site.command {
		case "build":
			err = site.build()
		case "help":
			err = site.help()
		case "init":
			err = site.init()
		case "new":
			err = site.new()
		case "serve":
			svr := newServer(site)
			err = svr.serve()
		case "nop":
			// Do nothing, used by tests
		default:
			panic("illegal command: " + site.command)
		}
	}
	if err != nil && err != ErrNonFatal {
		site.logerror(err.Error())
	}
	return err
}

// parseArgs parses and validates hindsite command-line arguments.
func (site *site) parseArgs(args []string) error {
	skip := false
	for i, opt := range args {
		if skip {
			skip = false
			continue
		}
		switch {
		case i == 0:
			site.executable = opt
			if len(args) == 1 {
				site.command = "help"
			}
		case i == 1:
			if opt == "-h" || opt == "--help" {
				opt = "help"
			}
			if !isCommand(opt) {
				return fmt.Errorf("illegal command: %s", opt)
			}
			site.command = opt
		case opt == "-drafts":
			site.drafts = true
		case opt == "-lint":
			site.lint = true
		case opt == "-launch":
			site.launch = true
		case opt == "-navigate":
			site.navigate = true
		case opt == "-keep":
			site.keep = true
		case opt == "-v":
			site.verbosity++
		case opt == "-vv":
			site.verbosity += 2
		case slice.New("-site", "-content", "-template", "-build", "-from", "-port", "-var", "-config").Has(opt):
			// Process option argument.
			if i+1 >= len(args) {
				return fmt.Errorf("missing %s argument value", opt)
			}
			arg := args[i+1]
			switch opt {
			case "-site":
				site.siteDir = arg
			case "-content":
				site.contentDir = arg
			case "-template":
				site.templateDir = arg
			case "-build":
				site.buildDir = arg
			case "-from":
				site.from = arg
			case "-port":
				ports := strings.SplitN(arg, ":", 2)
				if len(ports) > 0 && ports[0] != "" {
					i, err := strconv.ParseUint(ports[0], 10, 16)
					if err != nil {
						return fmt.Errorf("illegal -port: %s", arg)
					}
					site.httpport = uint16(i)
				}
				if len(ports) > 1 && ports[1] != "" {
					if ports[1] == "-1" {
						site.livereload = false
					} else {
						i, err := strconv.ParseUint(ports[1], 10, 16)
						if err != nil {
							return fmt.Errorf("illegal -port: %s", arg)
						}
						site.lrport = uint16(i)
					}
				}
			case "-var":
				if err := site.vars.parseVar(arg); err != nil {
					return err
				}
			case "-config":
				if err := site.vars.parseConfigFile(arg); err != nil {
					return err
				}
			default:
				panic("unexpected option: " + opt)
			}
			skip = true
		case strings.HasPrefix(opt, "-"):
			return fmt.Errorf("illegal option: %s", opt)
		default:
			site.cmdargs = append(site.cmdargs, opt)
		}
	}
	if site.command == "help" {
		return nil
	}
	// Validate site, content, template and build directories.
	getPath := func(path, defaultPath string) (string, error) {
		if path == "" {
			path = defaultPath
		}
		return filepath.Abs(path)
	}
	var err error
	site.siteDir, err = getPath(site.siteDir, ".")
	if err != nil {
		return err
	}
	if !fsx.DirExists(site.siteDir) {
		return fmt.Errorf("missing site directory: " + site.siteDir)
	}
	site.contentDir, err = getPath(site.contentDir, filepath.Join(site.siteDir, "content"))
	if err != nil {
		return err
	}
	site.verbose2("content directory: " + site.contentDir)
	if site.command != "init" && !fsx.DirExists(site.contentDir) {
		return fmt.Errorf("missing content directory: " + site.contentDir)
	}
	site.templateDir, err = getPath(site.templateDir, filepath.Join(site.siteDir, "template"))
	if err != nil {
		return err
	}
	site.verbose2("template directory: " + site.templateDir)
	if site.command != "init" && !fsx.DirExists(site.templateDir) {
		return fmt.Errorf("missing template directory: " + site.templateDir)
	}
	site.buildDir, err = getPath(site.buildDir, filepath.Join(site.siteDir, "build"))
	if err != nil {
		return err
	}
	site.verbose2("build directory: " + site.buildDir)
	// init and indexes directories are hardwired.
	site.indexDir = filepath.Join(site.buildDir, "indexes")
	site.initDir = filepath.Join(site.templateDir, "init")
	// Content, template and build directories cannot be nested.
	checkOverlap := func(name1, dir1, name2, dir2 string) error {
		if dir1 == dir2 {
			return fmt.Errorf("%s directory cannot be the same as %s directory", name1, name2)
		}
		if fsx.PathIsInDir(dir1, dir2) {
			return fmt.Errorf("%s directory cannot reside inside %s directory", name1, name2)
		}
		if fsx.PathIsInDir(dir2, dir1) {
			return fmt.Errorf("%s directory cannot reside inside %s directory", name2, name1)
		}
		return nil
	}
	if err := checkOverlap("content", site.contentDir, "template", site.templateDir); err != nil {
		// It's OK for the content directory to be the the template init directory.
		if site.contentDir != site.initDir {
			return err
		}
	}
	if err := checkOverlap("build", site.buildDir, "content", site.contentDir); err != nil {
		return err
	}
	if err := checkOverlap("build", site.buildDir, "template", site.templateDir); err != nil {
		return err
	}
	return nil
}

// output prints a line to `out` if `-v` option verbosity is equal to or
// greater than verbosity.
func (site *site) output(out io.Writer, verbosity int, format string, v ...interface{}) {
	if site.verbosity >= verbosity {
		msg := fmt.Sprintf(format, v...)
		// Strip leading site directory from path names to make message more readable.
		if filepath.IsAbs(site.siteDir) {
			msg = strings.Replace(msg, " "+site.siteDir+string(filepath.Separator), " ", -1)
			msg = strings.TrimPrefix(msg, site.siteDir+string(filepath.Separator))
		}
		if site.out == nil {
			fmt.Fprintln(out, msg)
		} else {
			site.out <- msg
		}
	}
}

// logconsole prints a line to logout.
func (site *site) logconsole(format string, v ...interface{}) {
	site.output(os.Stdout, 0, format, v...)
}

// verbose prints a line to logout if `-v` verbose option was specified.
func (site *site) verbose(format string, v ...interface{}) {
	site.output(os.Stdout, 1, format, v...)
}

// verbose2 prints a a line to logout the `-v` verbose option was specified more
// than once.
func (site *site) verbose2(format string, v ...interface{}) {
	site.output(os.Stdout, 2, format, v...)
}

// logerror prints a line to stderr.
func (site *site) logerror(format string, v ...interface{}) {
	color.Set(color.FgRed, color.Bold)
	site.output(os.Stderr, 0, format, v...)
	color.Unset()
}

// warning prints a line to stdout.
func (site *site) warning(format string, v ...interface{}) {
	color.Set(color.FgRed)
	site.output(os.Stdout, 0, "warning: "+format, v...)
	color.Unset()
}

func isCommand(name string) bool {
	return slice.New("build", "help", "init", "nop", "new", "serve").Has(name)
}

// help implements the help command.
func (site *site) help() (err error) {
	github := "https://github.com/srackham/hindsite"
	docsite := "https://srackham.github.io/hindsite"
	summary := `Hindsite is a static website generator.

Usage:

    hindsite init   [OPTION]...
    hindsite build  [OPTION]...
    hindsite serve  [OPTION]...
    hindsite new    [OPTION]... DOCUMENT
    hindsite help   [COMMAND]

Commands:

    init    initialize a new site
    build   build the website
    serve   start development webserver
    new     create a new content document
    help    display documentation

Options:

    -site     SITE_DIR
    -content  CONTENT_DIR
    -template TEMPLATE_DIR
    -build    BUILD_DIR
    -var      NAME=VALUE
    -port     [HTTP_PORT][:LR_PORT]
    -from     SOURCE
    -drafts
    -lint
    -launch
    -navigate
    -keep
    -v

Version:    ` + VERS + " (" + OS + ")" + `
Git commit: ` + COMMIT + `
Built:      ` + BUILT + `
Github:     ` + github + `
Docs:       ` + docsite + ``

	switch {
	case len(site.cmdargs) > 1:
		err = fmt.Errorf("to many command arguments")
	case len(site.cmdargs) == 1:
		cmd := site.cmdargs[0]
		if !isCommand(cmd) {
			err = fmt.Errorf("illegal command: %s", cmd)
		} else {
			url := fmt.Sprintf("%s#%s-command", docsite, cmd)
			err = launchBrowser(url)
			if err != nil {
				err = fmt.Errorf("failed to open '%s' in web browser: %s", url, err.Error())
			}
		}
	default:
		site.logconsole(summary)
	}
	return err
}

func (site *site) isDocument(f string) bool {
	ext := filepath.Ext(f)
	return (ext == ".md" || ext == ".rmu") && fsx.PathIsInDir(f, site.contentDir)
}

// homepage returns the site-wide root configuration `homepage` variable.
func (site *site) homepage() string {
	return site.confs[0].homepage
}

// urlprefix returns the site-wide root configuration `urlprefix` variable.
func (site *site) urlprefix() string {
	return site.confs[0].urlprefix
}

// match returns true if content file `f` matches one of the `patterns`.
// A blank pattern matches nothing.
// NOTE: Used for matching configuration `exclude`, `include`, `templates`
// configuration variables which use the `/` path separator.
func (site *site) match(f string, patterns []string) bool {
	if !fsx.PathIsInDir(f, site.contentDir) {
		panic("matched path must reside in content directory: " + f)
	}
	f, _ = filepath.Rel(site.contentDir, f)
	f = filepath.ToSlash(f)
	matched := false
	for _, pat := range patterns {
		pat = strings.TrimSpace(pat)
		if pat == "" {
			continue
		}
		if strings.HasSuffix(pat, "/") {
			if fsx.PathIsInDir(f, strings.TrimSuffix(pat, "/")) {
				matched = true
			}
		} else {
			if !strings.Contains(pat, "/") {
				f = path.Base(f)
			}
			if pat[0] == '/' {
				pat = pat[1:]
			}
			matched, _ = path.Match(pat, f)
		}
		if matched {
			return true
		}
	}
	return false
}

// exclude returns true if content file `f` is skipped by the `build` command.
func (site *site) exclude(f string) bool {
	return site.match(f, site.confs[0].exclude) && !site.match(f, site.confs[0].include)
}

// configFor returns the merged configuration for content directory path p.
// Configuration files that are in the corresponding template directory path are
// merged working from top (lowest precedence) to bottom.
//
// For example, if the path is `template/posts/james` then directories are
// searched in the following order: `template`, `template/posts`,
// `template/posts/james` with configuration entries from `template` having
// lowest precedence.
//
// The `site.confs` have been sorted by configuration `origin` in ascending
// order to ensure the directory precedence.
func (site *site) configFor(p string) config {
	if !fsx.PathIsInDir(p, site.contentDir) {
		panic("path outside content directory: " + p)
	}
	dir := fsx.PathTranslate(p, site.contentDir, site.templateDir)
	if fsx.FileExists(p) {
		dir = filepath.Dir(dir)
	}
	result := site.confs[0]
	result.user = copyMap(site.confs[0].user)
	for _, conf := range site.confs[1:] {
		if fsx.PathIsInDir(dir, conf.origin) {
			result.merge(conf)
		}
	}
	return result
}

// parseConfigFiles parses all configuration files from the site template
// directory to `site.confs`. The root configuration (site.confs[0]) is
// synthesised by merging configuration defaults with the root configuration
// file and -var options.
func (site *site) parseConfigFiles() error {
	site.confs = []config{}
	// Assign default root config.
	site.confs = append(site.confs, config{
		exclude:    []string{".*"},
		id:         "optional",
		paginate:   5,
		shortdate:  "2006-01-02",
		mediumdate: "2-Jan-2006",
		longdate:   "Mon Jan 2, 2006",
		user:       map[string]string{},
	})
	site.confs[0].timezone, _ = time.LoadLocation("Local")
	site.confs[0].origin = site.templateDir
	// Parse all config files.
	err := filepath.Walk(site.templateDir, func(f string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && f == site.initDir {
			return filepath.SkipDir
		}
		if !info.IsDir() {
			return nil
		}
		var conf config
		var found bool
		if f == site.templateDir {
			conf = site.confs[0]
		} else {
			conf = config{}
			conf.origin = f
		}
		for _, v := range []string{"config.toml", "config.yaml"} {
			cf := filepath.Join(f, v)
			if fsx.FileExists(cf) {
				found = true
				site.verbose("read config: " + cf)
				raw := rawConfig{}
				if err := raw.parseConfigFile(cf); err != nil {
					return fmt.Errorf("config file: \"%s\": %s", cf, err.Error())
				}
				if err := conf.mergeRaw(raw); err != nil {
					return fmt.Errorf("config file: \"%s\": %s", cf, err.Error())
				}
				if f != site.templateDir {
					msg := "root config variable \"%s\" in non-root config file: \"%s\""
					if conf.homepage != "" {
						site.warning(msg, "homepage", cf)
					}
					if conf.urlprefix != "" {
						site.warning(msg, "urlprefix", cf)
					}
					if conf.exclude != nil {
						site.warning(msg, "exclude", cf)
					}
					if conf.include != nil {
						site.warning(msg, "include", cf)
					}
				}
				break
			}
		}
		if found {
			if f == site.templateDir {
				site.confs[0] = conf // Update root config.
			} else {
				site.confs = append(site.confs, conf)
			}
			site.verbose2(conf.String())
		}
		return nil
	})
	if err != nil {
		return err
	}
	// Sort configurations by ascending origin directory to ensure deeper
	// configurations have precedence.
	sort.Slice(site.confs, func(i, j int) bool {
		return site.confs[i].origin < site.confs[j].origin
	})
	// Merge -var options into root config.
	if err := site.confs[0].mergeRaw(site.vars); err != nil {
		return fmt.Errorf("config variable: %s", err.Error())
	}
	site.verbose("root config: \n" + site.confs[0].String())
	// Sanity checks.
	if site.confs[0].origin != site.templateDir {
		panic("site.conf[0].origin != site.templateDir")
	}
	return nil
}
