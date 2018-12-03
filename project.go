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

type project struct {
	command       string
	executable    string
	in            chan string
	out           chan string
	projectDir    string
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

func newProject() project {
	proj := project{
		httpport:   1212,
		lrport:     35729,
		livereload: true,
	}
	return proj
}

// output prints a line to out writer if `-v` option verbosity is equal to or
// greater than verbosity.
func (proj *project) output(out io.Writer, verbosity int, format string, v ...interface{}) {
	if proj.verbosity >= verbosity {
		msg := fmt.Sprintf(format, v...)
		// Strip leading project directory from path names to make message more readable.
		if filepath.IsAbs(proj.projectDir) {
			msg = strings.Replace(msg, " "+proj.projectDir+string(filepath.Separator), " ", -1)
			msg = filepath.ToSlash(msg) // Normalize separators so automated tests pass.
		}
		if proj.out == nil {
			fmt.Fprintln(out, msg)
		} else {
			proj.out <- msg
		}
	}
}

// logconsole prints a line to logout.
func (proj *project) logconsole(format string, v ...interface{}) {
	proj.output(os.Stdout, 0, format, v...)
}

// verbose prints a line to logout if `-v` verbose option was specified.
func (proj *project) verbose(format string, v ...interface{}) {
	proj.output(os.Stdout, 1, format, v...)
}

// verbose2 prints a a line to logout the `-v` verbose option was specified more
// than once.
func (proj *project) verbose2(format string, v ...interface{}) {
	proj.output(os.Stdout, 2, format, v...)
}

// logerror prints a line to logerr.
func (proj *project) logerror(format string, v ...interface{}) {
	color.Set(color.FgRed, color.Bold)
	proj.output(os.Stderr, 0, "error: "+format, v...)
	color.Unset()
}

// parseArgs parses the hindsite command-line arguments.
func (proj *project) parseArgs(args []string) error {
	skip := false
	for i, opt := range args {
		if skip {
			skip = false
			continue
		}
		switch {
		case i == 0:
			proj.executable = opt
			if len(args) == 1 {
				proj.command = "help"
			}
		case i == 1:
			if opt == "-h" || opt == "--help" {
				opt = "help"
			}
			if !isCommand(opt) {
				return fmt.Errorf("illegal command: %s", opt)
			}
			proj.command = opt
		case i == 2 && proj.command == "new":
			if strings.HasPrefix(opt, "-") {
				return fmt.Errorf("illegal document file name: %s", opt)
			}
			proj.newFile = opt
		case i == 3 && proj.command == "new" && !strings.HasPrefix(opt, "-"):
			proj.projectDir = args[2]
			proj.newFile = opt
		case i == 2 && !strings.HasPrefix(opt, "-"):
			proj.projectDir = opt
		case opt == "-drafts":
			proj.drafts = true
		case opt == "-launch":
			proj.launch = true
		case opt == "-navigate":
			proj.navigate = true
		case opt == "-v":
			proj.verbosity++
		case opt == "-vv":
			proj.verbosity += 2
		case stringlist{"-content", "-template", "-build", "-builtin", "-port"}.Contains(opt):
			if i+1 >= len(args) {
				return fmt.Errorf("missing %s argument value", opt)
			}
			arg := args[i+1]
			switch opt {
			case "-content":
				proj.contentDir = arg
			case "-template":
				proj.templateDir = arg
			case "-build":
				proj.buildDir = arg
			case "-builtin":
				proj.builtin = arg
			case "-port":
				ports := strings.SplitN(arg, ":", 2)
				if len(ports) > 0 && ports[0] != "" {
					i, err := strconv.ParseUint(ports[0], 10, 16)
					if err != nil {
						return fmt.Errorf("illegal -port: %s", arg)
					}
					proj.httpport = uint16(i)
				}
				if len(ports) > 1 && ports[1] != "" {
					if ports[1] == "-1" {
						proj.livereload = false
					} else {
						i, err := strconv.ParseUint(ports[1], 10, 16)
						if err != nil {
							return fmt.Errorf("illegal -port: %s", arg)
						}
						proj.lrport = uint16(i)
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
	if proj.command == "help" {
		return nil
	}
	if proj.command == "new" {
		if proj.newFile == "" {
			return fmt.Errorf("document has not been specified")
		}
		if dirExists(proj.newFile) {
			return fmt.Errorf("document is a directory: %s", proj.newFile)
		}
		if d := filepath.Dir(proj.newFile); !dirExists(d) {
			return fmt.Errorf("missing document directory: %s", d)
		}
		if fileExists(proj.newFile) {
			return fmt.Errorf("document already exists: %s", proj.newFile)
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
	proj.projectDir, err = getPath(proj.projectDir, ".")
	if err != nil {
		return err
	}
	if !dirExists(proj.projectDir) {
		return fmt.Errorf("missing project directory: " + proj.projectDir)
	}
	proj.contentDir, err = getPath(proj.contentDir, filepath.Join(proj.projectDir, "content"))
	if err != nil {
		return err
	}
	proj.verbose2("content directory: " + proj.contentDir)
	if proj.command != "init" && !dirExists(proj.contentDir) {
		return fmt.Errorf("missing content directory: " + proj.contentDir)
	}
	proj.templateDir, err = getPath(proj.templateDir, filepath.Join(proj.projectDir, "template"))
	if err != nil {
		return err
	}
	proj.verbose2("template directory: " + proj.templateDir)
	if !(proj.command == "init" && proj.builtin != "") && !dirExists(proj.templateDir) {
		return fmt.Errorf("missing template directory: " + proj.templateDir)
	}
	proj.buildDir, err = getPath(proj.buildDir, filepath.Join(proj.projectDir, "build"))
	if err != nil {
		return err
	}
	proj.verbose2("build directory: " + proj.buildDir)
	// init and indexes directories are hardwired.
	proj.indexDir = filepath.Join(proj.buildDir, "indexes")
	proj.initDir = filepath.Join(proj.templateDir, "init")
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
	if err := checkOverlap("content", proj.contentDir, "template", proj.templateDir); err != nil {
		// It's OK for the content directory to be the the template init directory.
		if proj.contentDir != proj.initDir {
			return err
		}
	}
	if err := checkOverlap("build", proj.buildDir, "content", proj.contentDir); err != nil {
		return err
	}
	if err := checkOverlap("build", proj.buildDir, "template", proj.templateDir); err != nil {
		return err
	}
	if proj.command == "new" {
		proj.newFile, err = filepath.Abs(proj.newFile)
		if err != nil {
			return err
		}
		if !pathIsInDir(proj.newFile, proj.contentDir) {
			return fmt.Errorf("document must reside in %s directory", proj.contentDir)
		}
	}
	return nil
}

func isCommand(name string) bool {
	return stringlist{"build", "help", "init", "new", "serve"}.Contains(name)
}

// executeArgs runs a hindsite command specified by CLI args and returns a
// non-zero exit code if an error occurred.
func (proj *project) executeArgs(args []string) int {
	var err error
	err = proj.parseArgs(args)
	if err == nil {
		switch proj.command {
		case "build":
			err = proj.build()
		case "help":
			proj.help()
		case "init":
			err = proj.init()
		case "new":
			err = proj.new()
		case "serve":
			svr := newServer(proj)
			err = svr.serve()
		default:
			panic("illegal command: " + proj.command)
		}
	}
	if err != nil {
		proj.logerror(err.Error())
		return 1
	}
	return 0
}

// help implements the help command.
func (proj *project) help() {
	proj.logconsole(`Hindsite is a static website generator.

Usage:

    hindsite init  [PROJECT_DIR] [OPTIONS]
    hindsite build [PROJECT_DIR] [OPTIONS]
    hindsite serve [PROJECT_DIR] [OPTIONS]
    hindsite new   [PROJECT_DIR] DOCUMENT [OPTIONS]
    hindsite help

Commands:

    init    initialize a new project
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

func (proj *project) isDocument(f string) bool {
	ext := filepath.Ext(f)
	return (ext == ".md" || ext == ".rmu") && pathIsInDir(f, proj.contentDir)
}

// match returns true if path name f matches one of the patterns.
// The match is purely lexical.
func (proj *project) match(f string, patterns []string) bool {
	switch {
	case pathIsInDir(f, proj.contentDir):
		f, _ = filepath.Rel(proj.contentDir, f)
	case pathIsInDir(f, proj.templateDir):
		f, _ = filepath.Rel(proj.templateDir, f)
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
func (proj *project) exclude(f string) bool {
	return proj.match(f, proj.rootConf.exclude) && !proj.match(f, proj.rootConf.include)
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
// The `proj.confs` have been sorted by configuration `origin` in ascending
// order to ensure the directory precedence.
func (proj *project) configFor(p string) config {
	if !pathIsInDir(p, proj.contentDir) {
		panic("path outside content directory: " + p)
	}
	dir := pathTranslate(p, proj.contentDir, proj.templateDir)
	if fileExists(p) {
		dir = filepath.Dir(dir)
	}
	result := newConfig()
	for _, conf := range proj.confs {
		if pathIsInDir(dir, conf.origin) {
			result.merge(conf)
		}
	}
	// Global root configuration values.
	result.exclude = proj.rootConf.exclude
	result.include = proj.rootConf.include
	result.homepage = proj.rootConf.homepage
	result.urlprefix = proj.rootConf.urlprefix
	return result
}

// parseConfig parses all configuration files from the project template
// directory to project `confs`.
func (proj *project) parseConfigs() error {
	proj.confs = configs{}
	err := filepath.Walk(proj.templateDir, func(f string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && f == proj.initDir {
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
					return fmt.Errorf("config file: %s: %s", cf, err.Error())
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
	return nil
}
