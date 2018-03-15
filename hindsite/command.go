package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type command struct {
	name        string
	executable  string
	projectDir  string
	contentDir  string
	templateDir string
	buildDir    string
	indexDir    string
	drafts      bool
	slugify     bool
	topic       string
	port        string
	clean       bool
	builtin     bool
	verbose     bool
}

// Cmd is global singleton.
var Cmd command

func newCommand() command {
	return command{}
}

func (cmd *command) Parse(args []string) error {
	cmd.port = "1212"
	skip := false
	for i, opt := range args {
		if skip {
			skip = false
			continue
		}
		switch {
		case i == 0:
			cmd.executable = opt
			if len(args) == 1 {
				cmd.name = "help"
			}
		case i == 1:
			if opt == "-h" || opt == "--help" {
				opt = "help"
			}
			if !isCommand(opt) {
				return fmt.Errorf("illegal command: %s", opt)
			}
			cmd.name = opt
		case i == 2 && cmd.name == "help":
			if !isCommand(opt) {
				return fmt.Errorf("illegal help topic: %s", opt)
			}
			cmd.topic = opt
		case opt == "-drafts":
			cmd.drafts = true
		case opt == "-slugify":
			cmd.slugify = true
		case opt == "-clean":
			cmd.clean = true
		case opt == "-builtin":
			cmd.builtin = true
		case opt == "-v":
			cmd.verbose = true
		case stringlist{"-project", "-content", "-template", "-build", "-index", "-port"}.Contains(opt):
			if i+1 >= len(args) {
				return fmt.Errorf("missing %s argument value", opt)
			}
			arg := args[i+1]
			switch opt {
			case "-project":
				cmd.projectDir = arg
			case "-content":
				cmd.contentDir = arg
			case "-template":
				cmd.templateDir = arg
			case "-build":
				cmd.buildDir = arg
			case "-index":
				cmd.indexDir = arg
			case "-port":
				cmd.port = arg
			default:
				panic("illegal arugment: " + opt)
			}
			skip = true
		default:
			return fmt.Errorf("illegal argument: %s", opt)
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
	cmd.projectDir, err = getPath(cmd.projectDir, ".")
	if err != nil {
		return err
	}
	cmd.contentDir, err = getPath(cmd.contentDir, filepath.Join(cmd.projectDir, "content"))
	if err != nil {
		return err
	}
	cmd.templateDir, err = getPath(cmd.templateDir, filepath.Join(cmd.projectDir, "template"))
	if err != nil {
		return err
	}
	cmd.buildDir, err = getPath(cmd.buildDir, filepath.Join(cmd.projectDir, "build"))
	if err != nil {
		return err
	}
	cmd.indexDir, err = getPath(cmd.indexDir, filepath.Join(cmd.buildDir, "indexes"))
	if err != nil {
		return err
	}
	// Content and build directories can be the same. The build directory is
	// allowed at root of content directory. In all other cases content,
	// template and build directories cannot be nested.
	checkOverlap := func(name1, dir1, name2, dir2 string) error {
		if len(strings.TrimPrefix(dir1, dir2)) < len(dir1) {
			return fmt.Errorf("%s directory cannot reside inside %s directory", name1, name2)
		}
		if len(strings.TrimPrefix(dir2, dir1)) < len(dir2) {
			return fmt.Errorf("%s directory cannot reside inside %s directory", name2, name1)
		}
		return nil
	}
	if cmd.contentDir != cmd.templateDir {
		if err := checkOverlap("content", cmd.contentDir, "template", cmd.templateDir); err != nil {
			return err
		}
	}
	if filepath.Dir(cmd.buildDir) != cmd.contentDir {
		if err := checkOverlap("build", cmd.buildDir, "content", cmd.contentDir); err != nil {
			return err
		}
		if err := checkOverlap("build", cmd.buildDir, "template", cmd.templateDir); err != nil {
			return err
		}
	}
	if !(pathIsInDir(cmd.indexDir, cmd.buildDir) || cmd.indexDir == cmd.buildDir) {
		return fmt.Errorf("index directory must reside in build directory: %s", cmd.buildDir)
	}
	return nil
}

func isCommand(name string) bool {
	return stringlist{"build", "help", "init", "serve"}.Contains(name)
}

func (cmd *command) Execute() error {
	var err error
	// Parse configuration files from template and content directories (content directory config has precedence).
	for _, dir := range []string{cmd.templateDir, cmd.contentDir} {
		for _, conf := range []string{"config.toml", "config.yaml"} {
			f := filepath.Join(dir, conf)
			if fileExists(f) {
				verbose("read config: " + f)
				if err := Config.parseFile(f); err != nil {
					return err
				}
			}
		}
	}
	verbose("config: \n" + Config.String())
	// Execute command.
	switch cmd.name {
	case "build":
		err = cmd.build()
	case "help":
		cmd.help()
	case "init":
		err = cmd.init()
	case "serve":
		err = cmd.serve()
	default:
		panic("illegal command: " + cmd.name)
	}
	return err
}

func (cmd *command) init() error {
	if dirExists(cmd.contentDir) {
		files, err := ioutil.ReadDir(cmd.contentDir)
		if err != nil {
			return err
		}
		if len(files) > 0 {
			return fmt.Errorf("non-empty content directory: " + cmd.contentDir)
		}
	}
	if cmd.builtin {
		// Load template directory from the built-in project.
		if dirExists(cmd.templateDir) {
			files, err := ioutil.ReadDir(cmd.templateDir)
			if err != nil {
				return err
			}
			if len(files) > 0 {
				return fmt.Errorf("non-empty template directory: " + cmd.templateDir)
			}
		}
		verbose("installing builtin template")
		if err := RestoreAssets(cmd.templateDir, ""); err != nil {
			return err
		}
	}
	// Initialize content from template directory.
	if err := mkMissingDir(cmd.contentDir); err != nil {
		return err
	}
	err := filepath.Walk(cmd.templateDir, func(f string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if f == cmd.templateDir {
			return nil
		}
		dst, err := pathTranslate(f, cmd.templateDir, cmd.contentDir)
		if err != nil {
			return err
		}
		if info.IsDir() {
			verbose("make directory:   " + dst)
			err = mkMissingDir(dst)
		} else {
			// Copy example documents to content directory.
			switch filepath.Ext(f) {
			case ".md", ".rmu":
				verbose("copy example: " + f)
				err = copyFile(f, dst)
				if err != nil {
					return err
				}
				verbose("write:        " + dst)
			}
		}
		return err
	})
	return err
}

func (cmd *command) help() {
	println("Usage: hindsite command [arguments]")
}

func (cmd *command) build() error {
	if !dirExists(cmd.contentDir) {
		return fmt.Errorf("missing content directory: " + cmd.contentDir)
	}
	if !dirExists(cmd.templateDir) {
		return fmt.Errorf("missing template directory: " + cmd.templateDir)
	}
	if err := cmd.slugifyDir(cmd.contentDir); err != nil {
		return err
	}
	if cmd.contentDir != cmd.templateDir {
		if err := cmd.slugifyDir(cmd.templateDir); err != nil {
			return err
		}
	}
	if !dirExists(cmd.buildDir) {
		if err := os.Mkdir(cmd.buildDir, 0775); err != nil {
			return err
		}
	}
	if cmd.clean {
		// Delete everything in the build directory.
		files, _ := filepath.Glob(filepath.Join(cmd.buildDir, "*"))
		for _, f := range files {
			if err := os.RemoveAll(f); err != nil {
				return err
			}
		}
	}
	tmpls := newTemplates(cmd.templateDir)
	var confMod time.Time // The most recent date a change was made to a configuration file or a template file.
	// Copy static files from template directory to build directory and parse all template files.
	err := filepath.Walk(cmd.templateDir, func(f string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			switch filepath.Ext(f) {
			case ".toml", ".yaml":
				// Skip configuration file.
				if isOlder(confMod, info.ModTime()) {
					confMod = info.ModTime()
				}
			case ".md", ".rmu":
				// Skip example.
			case ".html":
				// Compile template.
				if isOlder(confMod, info.ModTime()) {
					confMod = info.ModTime()
				}
				tmpl := tmpls.name(f)
				verbose("parse template: " + tmpl)
				err = tmpls.add(tmpl)
			default:
				if cmd.contentDir != cmd.templateDir {
					err = cmd.copyStaticFile(f, cmd.templateDir, cmd.buildDir)
				}
			}
		}
		return err
	})
	if err != nil {
		return err
	}
	// Parse content documents and copy static files to the build directory.
	docs := documents{}
	err = filepath.Walk(cmd.contentDir, func(f string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if f == cmd.buildDir {
				// Do not process the build directory.
				return filepath.SkipDir
			}
			return nil
		}
		switch filepath.Ext(f) {
		case ".md", ".rmu":
			doc := document{}
			err = doc.parseFile(f, tmpls)
			if err != nil {
			}
			if doc.draft && !cmd.drafts {
				verbose("skip draft: " + f)
				return nil
			}
			docs = append(docs, &doc)
		case ".toml", ".yaml":
			if isOlder(confMod, info.ModTime()) {
				confMod = info.ModTime()
			}
		case ".html":
			if cmd.contentDir != cmd.templateDir {
				err = cmd.copyStaticFile(f, cmd.contentDir, cmd.buildDir)
			}
		default:
			err = cmd.copyStaticFile(f, cmd.contentDir, cmd.buildDir)
		}
		return err
	})
	if err != nil {
		return err
	}
	// Build indexes.
	idxs := indexes{}
	idxs.init(cmd.templateDir, cmd.buildDir, cmd.indexDir)
	for _, doc := range docs {
		idxs.addDocument(doc)
	}
	err = idxs.build(tmpls, confMod)
	if err != nil {
		return err
	}
	// Render documents.
	for _, doc := range docs {
		if !rebuild(doc.buildpath, confMod, doc) {
			continue
		}
		verbose("render: " + doc.contentpath)
		data := templateData{}
		data.add(doc.frontMatter())
		data["body"] = template.HTML(doc.render())
		err = tmpls.render(doc.layout, data, doc.buildpath)
		if err != nil {
			return err
		}
		verbose("write:  " + doc.buildpath)
		verbose(doc.String())
	}
	if Config.homepage != "" {
		// Install home page.
		src := Config.homepage
		dst := filepath.Join(cmd.buildDir, "index.html")
		if !fileExists(src) {
			return fmt.Errorf("homepage file missing: %s", src)
		}
		if !upToDate(dst, src) {
			verbose("copy homepage: " + src)
			if err := copyFile(src, dst); err != nil {
				return err
			}
			verbose("write:         " + dst)
		}
	}
	return nil
}

// Return true if the target file is newer than modified time or newer than any
// document.
func rebuild(target string, modified time.Time, docs ...*document) bool {
	info, err := os.Stat(target)
	if err != nil {
		return true
	}
	targetMod := info.ModTime()
	if isOlder(targetMod, modified) {
		return true
	}
	for _, doc := range docs {
		if isOlder(targetMod, doc.modified) {
			return true
		}
	}
	return false
}

// Return false target file is newer than the prerequisite file or if target
// does not exist.
func upToDate(target, prerequisite string) bool {
	result, err := fileIsOlder(prerequisite, target)
	if err != nil {
		return false
	}
	return result
}

// Copy srcFile to corresponding path in dstRoot.
// Skip if the destination file is up to date.
// Creates any missing destination directories.
func (cmd *command) copyStaticFile(srcFile, srcRoot, dstRoot string) error {
	dstFile, err := pathTranslate(srcFile, srcRoot, dstRoot)
	if err != nil {
		return err
	}
	if upToDate(dstFile, srcFile) {
		return nil
	}
	verbose("copy static: " + srcFile)
	err = mkMissingDir(filepath.Dir(dstFile))
	if err != nil {
		return err
	}
	err = copyFile(srcFile, dstFile)
	if err != nil {
		return err
	}
	verbose("write:       " + dstFile)
	return nil
}

// Recursively and slugify directory and file names.
func (cmd *command) slugifyDir(dir string) error {
	// TODO
	return nil
}

func (cmd *command) serve() error {
	if !dirExists(cmd.contentDir) {
		fmt.Fprintln(os.Stderr, "warning: missing content directory: "+cmd.contentDir)
	}
	if !dirExists(cmd.templateDir) {
		fmt.Fprintln(os.Stderr, "warning: missing template directory: "+cmd.templateDir)
	}
	if !dirExists(cmd.buildDir) {
		return fmt.Errorf("missing build directory: " + cmd.buildDir)
	}
	// Tweaked http.StripPrefix() handler
	// (https://golang.org/pkg/net/http/#StripPrefix). If URL does not start
	// with prefix serve unmodified URL.
	stripPrefix := func(prefix string, h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			verbose("request: " + r.URL.Path)
			if p := strings.TrimPrefix(r.URL.Path, prefix); len(p) < len(r.URL.Path) {
				r2 := new(http.Request)
				*r2 = *r
				r2.URL = new(url.URL)
				*r2.URL = *r.URL
				r2.URL.Path = p
				h.ServeHTTP(w, r2)
			} else {
				h.ServeHTTP(w, r)
			}
		})
	}
	http.Handle("/", stripPrefix(Config.urlprefix, http.FileServer(http.Dir(cmd.buildDir))))
	fmt.Printf("\nServing build directory %s on http://localhost:%s/\nPress Ctrl+C to stop\n", cmd.buildDir, cmd.port)
	return http.ListenAndServe(":"+cmd.port, nil)
}
