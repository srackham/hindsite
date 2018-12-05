package main

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

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
	executable    string
	in            chan string
	out           chan string
	siteDir       string
	contentDir    string
	templateDir   string
	buildDir      string
	indexDir      string
	initDir       string
	newFile       string
	builtin       string
	drafts        bool
	launch        bool
	httpport      uint16
	lrport        uint16
	livereload    bool
	navigate      bool
	verbosity     int
	rootConf      config
	confs         configs
	docs          documentsLookup
	idxs          indexes
	htmlTemplates htmlTemplates
	textTemplates textTemplates
}

func newSite() site {
	site := site{
		httpport:   1212,
		lrport:     35729,
		livereload: true,
	}
	return site
}

// output prints a line to out writer if `-v` option verbosity is equal to or
// greater than verbosity.
func (site *site) output(out io.Writer, verbosity int, format string, v ...interface{}) {
	if site.verbosity >= verbosity {
		msg := fmt.Sprintf(format, v...)
		// Strip leading site directory from path names to make message more readable.
		if filepath.IsAbs(site.siteDir) {
			msg = strings.Replace(msg, " "+site.siteDir+string(filepath.Separator), " ", -1)
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

// logerror prints a line to logerr.
func (site *site) logerror(format string, v ...interface{}) {
	color.Set(color.FgRed, color.Bold)
	site.output(os.Stderr, 0, "error: "+format, v...)
	color.Unset()
}

// parseArgs parses the hindsite command-line arguments.
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
		case i == 2 && site.command == "new":
			if strings.HasPrefix(opt, "-") {
				return fmt.Errorf("illegal document file name: %s", opt)
			}
			site.newFile = opt
		case i == 3 && site.command == "new" && !strings.HasPrefix(opt, "-"):
			site.siteDir = args[2]
			site.newFile = opt
		case i == 2 && !strings.HasPrefix(opt, "-"):
			site.siteDir = opt
		case opt == "-drafts":
			site.drafts = true
		case opt == "-launch":
			site.launch = true
		case opt == "-navigate":
			site.navigate = true
		case opt == "-v":
			site.verbosity++
		case opt == "-vv":
			site.verbosity += 2
		case stringlist{"-content", "-template", "-build", "-builtin", "-port"}.Contains(opt):
			if i+1 >= len(args) {
				return fmt.Errorf("missing %s argument value", opt)
			}
			arg := args[i+1]
			switch opt {
			case "-content":
				site.contentDir = arg
			case "-template":
				site.templateDir = arg
			case "-build":
				site.buildDir = arg
			case "-builtin":
				site.builtin = arg
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
			default:
				panic("illegal argument: " + opt)
			}
			skip = true
		default:
			return fmt.Errorf("illegal option: %s", opt)
		}
	}
	if site.command == "help" {
		return nil
	}
	if site.command == "new" {
		if site.newFile == "" {
			return fmt.Errorf("document has not been specified")
		}
		if dirExists(site.newFile) {
			return fmt.Errorf("document is a directory: %s", site.newFile)
		}
		if d := filepath.Dir(site.newFile); !dirExists(d) {
			return fmt.Errorf("missing document directory: %s", d)
		}
		if fileExists(site.newFile) {
			return fmt.Errorf("document already exists: %s", site.newFile)
		}
	}
	// Clean and convert directories to absolute paths.
	// Internally all file paths are absolute.
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
	if !dirExists(site.siteDir) {
		return fmt.Errorf("missing site directory: " + site.siteDir)
	}
	site.contentDir, err = getPath(site.contentDir, filepath.Join(site.siteDir, "content"))
	if err != nil {
		return err
	}
	site.verbose2("content directory: " + site.contentDir)
	if site.command != "init" && !dirExists(site.contentDir) {
		return fmt.Errorf("missing content directory: " + site.contentDir)
	}
	site.templateDir, err = getPath(site.templateDir, filepath.Join(site.siteDir, "template"))
	if err != nil {
		return err
	}
	site.verbose2("template directory: " + site.templateDir)
	if !(site.command == "init" && site.builtin != "") && !dirExists(site.templateDir) {
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
		if pathIsInDir(dir1, dir2) {
			return fmt.Errorf("%s directory cannot reside inside %s directory", name1, name2)
		}
		if pathIsInDir(dir2, dir1) {
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
	if site.command == "new" {
		site.newFile, err = filepath.Abs(site.newFile)
		if err != nil {
			return err
		}
		if !pathIsInDir(site.newFile, site.contentDir) {
			return fmt.Errorf("document must reside in %s directory", site.contentDir)
		}
	}
	return nil
}

func isCommand(name string) bool {
	return stringlist{"build", "help", "init", "new", "serve"}.Contains(name)
}

// executeArgs runs a hindsite command specified by CLI args and returns a
// non-zero exit code if an error occurred.
func (site *site) executeArgs(args []string) int {
	var err error
	err = site.parseArgs(args)
	if err == nil {
		switch site.command {
		case "build":
			err = site.build()
		case "help":
			site.help()
		case "init":
			err = site.init()
		case "new":
			err = site.new()
		case "serve":
			svr := newServer(site)
			err = svr.serve()
		default:
			panic("illegal command: " + site.command)
		}
	}
	if err != nil {
		site.logerror(err.Error())
		return 1
	}
	return 0
}

// help implements the help command.
func (site *site) help() {
	site.logconsole(`Hindsite is a static website generator.

Usage:

    hindsite init  [SITE_DIR] [OPTIONS]
    hindsite build [SITE_DIR] [OPTIONS]
    hindsite serve [SITE_DIR] [OPTIONS]
    hindsite new   [SITE_DIR] DOCUMENT [OPTIONS]
    hindsite help

Commands:

    init    initialize a new site
    build   build the website
    serve   start development webserver
    new     create a new content document
    help    display usage summary

Options:

    -content  CONTENT_DIR
    -template TEMPLATE_DIR
    -build    BUILD_DIR
    -port     [HTTP_PORT][:LR_PORT]
    -builtin  NAME
    -drafts
    -launch
    -navigate
    -v

Version:    ` + VERS + " (" + OS + ")" + `
Built:      ` + BUILT + `
Git commit: ` + COMMIT + `
Github:     https://github.com/srackham/hindsite
Docs:       https://srackham.github.io/hindsite`)
}

// isTemplate returns true if the file path f is in the templates configuration value.
func isTemplate(f string, templates *string) bool {
	return strings.Contains("|"+nz(templates)+"|", "|"+filepath.Ext(f)+"|")
}

func (site *site) isDocument(f string) bool {
	ext := filepath.Ext(f)
	return (ext == ".md" || ext == ".rmu") && pathIsInDir(f, site.contentDir)
}

// match returns true if path name f matches one of the patterns.
// The match is purely lexical.
func (site *site) match(f string, patterns []string) bool {
	switch {
	case pathIsInDir(f, site.contentDir):
		f, _ = filepath.Rel(site.contentDir, f)
	case pathIsInDir(f, site.templateDir):
		f, _ = filepath.Rel(site.templateDir, f)
	default:
		panic("matched path must reside in content or template directories: " + f)
	}
	f = filepath.ToSlash(f)
	matched := false
	for _, pat := range patterns {
		if strings.HasSuffix(pat, "/") {
			if pathIsInDir(f, strings.TrimSuffix(pat, "/")) {
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

// exclude returns true if path name f is excluded.
func (site *site) exclude(f string) bool {
	return site.match(f, site.rootConf.exclude) && !site.match(f, site.rootConf.include)
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
	if !pathIsInDir(p, site.contentDir) {
		panic("path outside content directory: " + p)
	}
	dir := pathTranslate(p, site.contentDir, site.templateDir)
	if fileExists(p) {
		dir = filepath.Dir(dir)
	}
	result := newConfig()
	for _, conf := range site.confs {
		if pathIsInDir(dir, conf.origin) {
			result.merge(conf)
		}
	}
	// Global root configuration values.
	result.exclude = site.rootConf.exclude
	result.include = site.rootConf.include
	result.homepage = site.rootConf.homepage
	result.urlprefix = site.rootConf.urlprefix
	return result
}

// parseConfig parses all configuration files from the site template
// directory to site `confs`.
func (site *site) parseConfigs() error {
	site.confs = configs{}
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
		conf := config{}
		conf.origin = f
		found := false
		for _, v := range []string{"config.toml", "config.yaml"} {
			cf := filepath.Join(f, v)
			if fileExists(cf) {
				found = true
				site.verbose("read config: " + cf)
				if err := conf.parseFile(site, cf); err != nil {
					return fmt.Errorf("config file: %s: %s", cf, err.Error())
				}
			}
		}
		if found {
			site.confs = append(site.confs, conf)
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
	return nil
}
