package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
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
	projectDir    string
	contentDir    string
	templateDir   string
	buildDir      string
	indexDir      string
	initDir       string
	drafts        bool
	port          string
	launch        bool
	builtin       string
	verbosity     int
	rootConf      config
	confs         configs
	docs          documentsLookup
	idxs          indexes
	htmlTemplates htmlTemplates
	textTemplates textTemplates
}

func newProject() project {
	return project{}
}

// message strips leading project directory from path names to make the message
// more readable.
func (proj *project) message(msg string) string {
	return strings.Replace(msg, proj.projectDir+string(filepath.Separator), "", -1)
}

// logconsole prints a message if `-v` option verbosity is equal to or greater than
// verbosity.
func (proj *project) logconsole(verbosity int, msg string) {
	if proj.verbosity >= verbosity {
		fmt.Println(proj.message(msg))
	}
}

// logerror prints a message to stderr.
func (proj *project) logerror(msg string) {
	fmt.Fprintln(os.Stderr, "error: "+proj.message(msg))
}

// println unconditionally prints a message.
func (proj *project) println(msg string) {
	proj.logconsole(0, msg)
}

// verbose prints a message if `-v` verbose option was specified.
func (proj *project) verbose(msg string) {
	proj.logconsole(1, msg)
}

// verbose2 prints a message if the `-v` verbose option was specified more than
// once.
func (proj *project) verbose2(msg string) {
	proj.logconsole(2, msg)
}

func (proj *project) die(msg string) {
	if msg != "" {
		proj.logerror(msg)
	}
	os.Exit(1)
}

// parseArgs parses the hindsite command-line arguments.
func (proj *project) parseArgs(args []string) error {
	proj.port = "1212"
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
			if len(args) == 2 && opt != "help" {
				return fmt.Errorf("project directory not specifed")
			}
			proj.command = opt
		case i == 2:
			if !dirExists(opt) && strings.HasPrefix(opt, "-") {
				return fmt.Errorf("project directory not specifed")
			}
			proj.projectDir = opt
		case opt == "-drafts":
			proj.drafts = true
		case opt == "-launch":
			proj.launch = true
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
				proj.port = arg
			default:
				panic("illegal arugment: " + opt)
			}
			skip = true
		default:
			return fmt.Errorf("illegal option: %s", opt)
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
	proj.templateDir, err = getPath(proj.templateDir, filepath.Join(proj.projectDir, "template"))
	if err != nil {
		return err
	}
	proj.verbose2("template directory: " + proj.templateDir)
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
	return nil
}

func isCommand(name string) bool {
	return stringlist{"build", "help", "init", "serve"}.Contains(name)
}

func (proj *project) execute() error {
	// Execute command.
	var err error
	switch proj.command {
	case "build":
		err = proj.build()
	case "help":
		proj.help()
	case "init":
		err = proj.init()
	case "serve":
		err = proj.serve()
	default:
		panic("illegal command: " + proj.command)
	}
	return err
}

// help implements the help command.
func (proj *project) help() {
	fmt.Println(`Hindsite is a static website generator.

Usage:

    hindsite init  PROJECT_DIR [OPTIONS]
    hindsite build PROJECT_DIR [OPTIONS]
    hindsite serve PROJECT_DIR [OPTIONS]
    hindsite help

The commands are:

    init    initialize a new project
    build   build the website
    serve   start development webserver
    help    display usage summary

The options are:

    -content  CONTENT_DIR
    -template TEMPLATE_DIR
    -build    BUILD_DIR
    -port     PORT
    -builtin  NAME
    -drafts
    -launch
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
	for i, conf := range proj.confs {
		if pathIsInDir(dir, conf.origin) {
			result.merge(conf)
		}
		if i == 0 {
			// Assign global root configuration values.
			result.exclude = conf.exclude
			result.include = conf.include
			result.homepage = conf.homepage
			result.urlprefix = conf.urlprefix
		}
	}
	return result
}
