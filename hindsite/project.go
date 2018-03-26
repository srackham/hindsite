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

type project struct {
	command     string
	executable  string
	projectDir  string
	contentDir  string
	templateDir string
	buildDir    string
	indexDir    string
	drafts      bool
	topic       string
	port        string
	clean       bool
	builtin     string
	verbose     bool
	rootConf    config
	confs       configs
	tmpls       templates
}

func newProject() project {
	return project{}
}

// printlin prints a message if `-v` verbose option set.
func (proj *project) println(message string) {
	if proj.verbose {
		fmt.Println(message)
	}
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
			proj.command = opt
		case opt == "-drafts":
			proj.drafts = true
		case opt == "-clean":
			proj.clean = true
		case opt == "-v":
			proj.verbose = true
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
		case i == 2:
			if opt[0] == '-' {
				return fmt.Errorf("illegal option: %s", opt)
			}
			if proj.command == "help" {
				proj.topic = opt
			} else {
				proj.projectDir = opt
			}
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
	proj.templateDir, err = getPath(proj.templateDir, filepath.Join(proj.projectDir, "template"))
	if err != nil {
		return err
	}
	proj.buildDir, err = getPath(proj.buildDir, filepath.Join(proj.projectDir, "build"))
	if err != nil {
		return err
	}
	proj.indexDir, err = getPath(proj.indexDir, filepath.Join(proj.buildDir, "indexes"))
	if err != nil {
		return err
	}
	// Content, template and build directories cannot be nested.
	checkOverlap := func(name1, dir1, name2, dir2 string) error {
		if len(strings.TrimPrefix(dir1, dir2)) < len(dir1) {
			return fmt.Errorf("%s directory cannot reside inside %s directory", name1, name2)
		}
		if len(strings.TrimPrefix(dir2, dir1)) < len(dir2) {
			return fmt.Errorf("%s directory cannot reside inside %s directory", name2, name1)
		}
		return nil
	}
	if err := checkOverlap("content", proj.contentDir, "template", proj.templateDir); err != nil {
		return err
	}
	if err := checkOverlap("build", proj.buildDir, "content", proj.contentDir); err != nil {
		return err
	}
	if err := checkOverlap("build", proj.buildDir, "template", proj.templateDir); err != nil {
		return err
	}
	if !(pathIsInDir(proj.indexDir, proj.buildDir) || proj.indexDir == proj.buildDir) {
		return fmt.Errorf("index directory must reside in build directory: %s", proj.buildDir)
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

// init implements the init command.
func (proj *project) init() error {
	if dirExists(proj.contentDir) {
		files, err := ioutil.ReadDir(proj.contentDir)
		if err != nil {
			return err
		}
		if len(files) > 0 {
			return fmt.Errorf("non-empty content directory: " + proj.contentDir)
		}
	}
	if proj.builtin != "" {
		// Load template directory from the built-in project.
		if dirExists(proj.templateDir) {
			files, err := ioutil.ReadDir(proj.templateDir)
			if err != nil {
				return err
			}
			if len(files) > 0 {
				return fmt.Errorf("non-empty template directory: " + proj.templateDir)
			}
		}
		proj.println("installing builtin template: " + proj.builtin)
		if err := RestoreAssets(proj.templateDir, proj.builtin+"/template"); err != nil {
			return err
		}
		// Hoist the restored template files up into the root of the project template directory.
		files, _ := filepath.Glob(filepath.Join(proj.templateDir, proj.builtin, "template", "*"))
		for _, f := range files {
			if err := os.Rename(f, filepath.Join(proj.templateDir, filepath.Base(f))); err != nil {
				return err
			}
		}
		// Remove empty restored path.
		if err := os.RemoveAll(filepath.Join(proj.templateDir, proj.builtin)); err != nil {
			return err
		}
	} else {
		if !dirExists(proj.templateDir) {
			return fmt.Errorf("missing template directory: " + proj.templateDir)
		}
	}
	// Initialize content from template directory.
	if err := mkMissingDir(proj.contentDir); err != nil {
		return err
	}
	err := filepath.Walk(proj.templateDir, func(f string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if f == proj.templateDir {
			return nil
		}
		dst, err := pathTranslate(f, proj.templateDir, proj.contentDir)
		if err != nil {
			return err
		}
		if info.IsDir() {
			proj.println("make directory:   " + dst)
			err = mkMissingDir(dst)
		} else {
			// Copy example documents to content directory.
			switch filepath.Ext(f) {
			case ".md", ".rmu":
				proj.println("copy example: " + f)
				err = copyFile(f, dst)
				if err != nil {
					return err
				}
				proj.println("write:        " + dst)
			}
		}
		return err
	})
	return err
}

// help implements the help command.
func (proj *project) help() {
	println(`Hindsite is a static website generator.

Usage:

    hindsite init  [PROJECT_DIR] [OPTIONS]
    hindsite build [PROJECT_DIR] [OPTIONS]
    hindsite serve [PROJECT_DIR] [OPTIONS]
    hindsite help  [TOPIC]

The commands are:

    init    create a new project
    build   generate the website
    serve   start development webserver
    help    display documentation

The options are:

    -content  CONTENT_DIR
    -template TEMPLATE_DIR
    -build    BUILD_DIR
    -port     PORT
    -builtin  NAME
    -clean
    -drafts
    -v
`)
}

// build implements the build command.
func (proj *project) build() error {
	if err := proj.parseConfigs(); err != nil {
		return err
	}
	if !dirExists(proj.buildDir) {
		if err := os.Mkdir(proj.buildDir, 0775); err != nil {
			return err
		}
	}
	if proj.clean {
		// Delete everything in the build directory.
		files, _ := filepath.Glob(filepath.Join(proj.buildDir, "*"))
		for _, f := range files {
			if err := os.RemoveAll(f); err != nil {
				return err
			}
		}
	}
	proj.tmpls = newTemplates(proj.templateDir)
	var confMod time.Time // The most recent date a change was made to a configuration file or a template file.
	// Copy static files from template directory to build directory and parse all template files.
	err := filepath.Walk(proj.templateDir, func(f string, info os.FileInfo, err error) error {
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
				proj.println("parse template: " + f)
				err = proj.tmpls.add(f)
			default:
				err = proj.copyStaticFile(f, proj.templateDir, proj.buildDir)
			}
		}
		return err
	})
	if err != nil {
		return err
	}
	// Parse content documents and copy static files to the build directory.
	docs := documents{}
	err = filepath.Walk(proj.contentDir, func(f string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if proj.exclude(info) {
			proj.println("exclude: " + f)
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !info.IsDir() {
			switch filepath.Ext(f) {
			case ".md", ".rmu":
				// Parse document.
				doc, err := newDocument(f, proj)
				if err != nil {
					return err
				}
				if doc.draft && !proj.drafts {
					proj.println("skip draft: " + f)
					return nil
				}
				docs = append(docs, &doc)
			case ".toml", ".yaml":
				// Skip configuration file.
				if isOlder(confMod, info.ModTime()) {
					confMod = info.ModTime()
				}
			case ".html":
				err = proj.copyStaticFile(f, proj.contentDir, proj.buildDir)
			default:
				err = proj.copyStaticFile(f, proj.contentDir, proj.buildDir)
			}
		}
		return err
	})
	if err != nil {
		return err
	}
	idxs, err := newIndexes(proj)
	if err != nil {
		return err
	}
	for _, doc := range docs {
		idxs.addDocument(doc)
	}
	err = idxs.build(proj, confMod)
	if err != nil {
		return err
	}
	// Render documents.
	for _, doc := range docs {
		if !rebuild(doc.buildpath, confMod, doc) {
			continue
		}
		proj.println("render: " + doc.contentpath)
		data := templateData{}
		data.merge(doc.frontMatter())
		data.merge(doc.prevNext())
		data.merge(proj.data())
		data["body"] = template.HTML(doc.render(doc.content))
		err = proj.tmpls.render(doc.layout, data, doc.buildpath)
		if err != nil {
			return err
		}
		proj.println("write:  " + doc.buildpath)
		proj.println(doc.String())
	}
	if proj.rootConf.homepage != "" {
		// Install home page.
		src := proj.rootConf.homepage
		dst := filepath.Join(proj.buildDir, "index.html")
		if !fileExists(src) {
			return fmt.Errorf("homepage file missing: %s", src)
		}
		if !upToDate(dst, src) {
			proj.println("copy homepage: " + src)
			if err := copyFile(src, dst); err != nil {
				return err
			}
			proj.println("write:         " + dst)
		}
	}
	return nil
}

// rebuild returns true if the target file does not exist or is newer than
// modified time or newer than any document.
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

// upToDate returns false target file is newer than the prerequisite file or if
// target does not exist.
func upToDate(target, prerequisite string) bool {
	result, err := fileIsOlder(prerequisite, target)
	if err != nil {
		return false
	}
	return result
}

// copyStaticFile copies srcFile to corresponding path in dstRoot.
// Skips if the destination file is up to date.
// Creates missing destination directories.
func (proj *project) copyStaticFile(srcFile, srcRoot, dstRoot string) error {
	dstFile, err := pathTranslate(srcFile, srcRoot, dstRoot)
	if err != nil {
		return err
	}
	if upToDate(dstFile, srcFile) {
		return nil
	}
	proj.println("copy static: " + srcFile)
	err = mkMissingDir(filepath.Dir(dstFile))
	if err != nil {
		return err
	}
	err = copyFile(srcFile, dstFile)
	if err != nil {
		return err
	}
	proj.println("write:       " + dstFile)
	return nil
}

// exclude returns true if a file should be excluded from processing.
func (proj *project) exclude(info os.FileInfo) bool {
	for _, pat := range proj.rootConf.exclude {
		if info.IsDir() && strings.HasSuffix(pat, "/") {
			pat = strings.TrimSuffix(pat, "/")
		}
		result, _ := filepath.Match(pat, info.Name())
		if result {
			return true
		}
	}
	return false
}

// data returns project global template variables.
func (proj *project) data() templateData {
	return templateData{"site": templateData{
		"urlprefix": proj.rootConf.urlprefix,
	}}
}

// serve implements the serve comand.
func (proj *project) serve() error {
	if err := proj.parseConfigs(); err != nil {
		return err
	}
	if !dirExists(proj.buildDir) {
		return fmt.Errorf("missing build directory: " + proj.buildDir)
	}
	// Tweaked http.StripPrefix() handler
	// (https://golang.org/pkg/net/http/#StripPrefix). If URL does not start
	// with prefix serve unmodified URL.
	stripPrefix := func(prefix string, h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			proj.println("request: " + r.URL.Path)
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
	http.Handle("/", stripPrefix(proj.rootConf.urlprefix, http.FileServer(http.Dir(proj.buildDir))))
	fmt.Printf("\nServing build directory %s on http://localhost:%s/\nPress Ctrl+C to stop\n", proj.buildDir, proj.port)
	return http.ListenAndServe(":"+proj.port, nil)
}
