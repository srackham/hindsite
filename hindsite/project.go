package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type project struct {
	command       string
	executable    string
	projectDir    string
	contentDir    string
	templateDir   string
	buildDir      string
	indexDir      string
	drafts        bool
	topic         string
	port          string
	incremental   bool
	builtin       string
	verbosity     int
	rootConf      config
	confs         configs
	htmlTemplates htmlTemplates
	textTemplates textTemplates
}

func newProject() project {
	return project{}
}

// verbose prints a message if `-v` option verbosity is equal to or greater than
// verbosity.
func (proj *project) println(verbosity int, message string) {
	if proj.verbosity >= verbosity {
		message = strings.Replace(message, proj.projectDir+string(filepath.Separator), "", -1)
		fmt.Println(message)
	}
}

// verbose prints a message if `-v` verbose option was specified.
func (proj *project) verbose(message string) {
	proj.println(1, message)
}

// verbose2 prints a message if the `-v` verbose option was specified more than
// once.
func (proj *project) verbose2(message string) {
	proj.println(2, message)
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
			if proj.command == "help" {
				proj.topic = opt
			} else {
				if !dirExists(opt) && strings.HasPrefix(opt, "-") {
					return fmt.Errorf("project directory not specifed")
				}
				proj.projectDir = opt
			}
		case opt == "-drafts":
			proj.drafts = true
		case opt == "-incremental":
			proj.incremental = true
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
				panic("parseArgs: illegal arugment: " + opt)
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
	proj.indexDir, err = getPath(proj.indexDir, filepath.Join(proj.buildDir, "indexes"))
	if err != nil {
		return err
	}
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
		// It's OK to build the init directory.
		if !(proj.command == "build" && proj.contentDir == filepath.Join(proj.templateDir, "init")) {
			return err
		}
	}
	if err := checkOverlap("build", proj.buildDir, "content", proj.contentDir); err != nil {
		return err
	}
	if err := checkOverlap("build", proj.buildDir, "template", proj.templateDir); err != nil {
		return err
	}
	if !pathIsInDir(proj.indexDir, proj.buildDir) {
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
		panic("execute: illegal command: " + proj.command)
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
		proj.verbose("installing builtin template: " + proj.builtin)
		if err := RestoreAssets(proj.templateDir, proj.builtin+"/template"); err != nil {
			return err
		}
		// Hoist the restored template files from the root of the restored
		// builtin directery up one level into the root of the project template
		// directory.
		files, _ := filepath.Glob(filepath.Join(proj.templateDir, proj.builtin, "template", "*"))
		for _, f := range files {
			if err := os.Rename(f, filepath.Join(proj.templateDir, filepath.Base(f))); err != nil {
				return err
			}
		}
		// Remove the now empty restored path.
		if err := os.RemoveAll(filepath.Join(proj.templateDir, proj.builtin)); err != nil {
			return err
		}
	} else {
		if !dirExists(proj.templateDir) {
			return fmt.Errorf("missing template directory: " + proj.templateDir)
		}
	}
	// Create the template directory structure in the content directory.
	initDir := filepath.Join(proj.templateDir, "init")
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
		if info.IsDir() && f == initDir {
			return filepath.SkipDir
		}
		if info.IsDir() {
			dst := pathTranslate(f, proj.templateDir, proj.contentDir)
			proj.verbose("make directory: " + dst)
			err = mkMissingDir(dst)
		}
		return err
	})
	if err != nil {
		return err
	}
	// Copy the contents of the optional template init directory to the content directory.
	if dirExists(initDir) {
		err = filepath.Walk(initDir, func(f string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if f == initDir {
				return nil
			}
			dst := pathTranslate(f, initDir, proj.contentDir)
			if info.IsDir() {
				if !dirExists(dst) {
					proj.verbose("make directory: " + dst)
					err = mkMissingDir(dst)
				}
			} else {
				proj.verbose2("copy init: " + f)
				proj.verbose("write init: " + dst)
				err = copyFile(f, dst)
			}
			return err
		})
	}
	return err
}

// help implements the help command.
func (proj *project) help() {
	println(`Hindsite is a static website generator.

Usage:

    hindsite init  PROJECT_DIR [OPTIONS]
    hindsite build PROJECT_DIR [OPTIONS]
    hindsite serve PROJECT_DIR [OPTIONS]
    hindsite help  [TOPIC]

The commands are:

    init    initialize a new project
    build   generate the website
    serve   start development webserver
    help    display documentation

The options are:

    -content  CONTENT_DIR
    -template TEMPLATE_DIR
    -build    BUILD_DIR
    -port     PORT
    -builtin  NAME
    -incremental
    -drafts
    -v
`)
}

// build implements the build command.
func (proj *project) build() error {
	startTime := time.Now()
	if err := proj.parseConfigs(); err != nil {
		return err
	}
	if !dirExists(proj.buildDir) {
		if err := os.Mkdir(proj.buildDir, 0775); err != nil {
			return err
		}
	}
	if !proj.incremental {
		// Delete everything in the build directory forcing a complete site rebuild.
		files, _ := filepath.Glob(filepath.Join(proj.buildDir, "*"))
		for _, f := range files {
			if err := os.RemoveAll(f); err != nil {
				return err
			}
		}
	}
	// confMod records the most recent date a change was made to a configuration file or a template file.
	var confMod time.Time
	updateConfMod := func(info os.FileInfo) {
		if isOlder(confMod, info.ModTime()) {
			confMod = info.ModTime()
		}
	}
	// Parse all template files.
	proj.htmlTemplates = newHtmlTemplates(proj.templateDir)
	proj.textTemplates = newTextTemplates(proj.templateDir)
	err := filepath.Walk(proj.templateDir, func(f string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && f == filepath.Join(proj.templateDir, "init") {
			return filepath.SkipDir
		}
		if !info.IsDir() {
			switch filepath.Ext(f) {
			case ".toml", ".yaml":
				// Skip configuration file.
				updateConfMod(info)
			case ".html":
				// Compile HTML template.
				updateConfMod(info)
				proj.verbose("parse template: " + f)
				err = proj.htmlTemplates.add(f)
			case ".txt":
				// Compile text template.
				updateConfMod(info)
				proj.verbose("parse template: " + f)
				err = proj.textTemplates.add(f)
			}
		}
		return err
	})
	if err != nil {
		return err
	}
	// Parse content directory documents and copy/render static files to the build directory.
	draftsCount := 0
	docsCount := 0
	docs := documents{}
	err = filepath.Walk(proj.contentDir, func(f string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(proj.contentDir, f)
		if proj.exclude(rel) {
			proj.verbose("exclude: " + f)
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !info.IsDir() {
			switch filepath.Ext(f) {
			case ".md", ".rmu":
				docsCount++
				// Parse document.
				doc, err := newDocument(f, proj)
				if err != nil {
					return err
				}
				if doc.draft && !proj.drafts {
					draftsCount++
					proj.verbose("skip draft: " + f)
					return nil
				}
				docs = append(docs, &doc)
			default:
				conf := proj.configFor(f)
				if isTemplate(f, conf.templates) {
					err = proj.renderStaticFile(f, confMod)
				} else {
					err = proj.copyStaticFile(f)
				}
			}
		}
		return err
	})
	if err != nil {
		return err
	}
	// Create indexes.
	idxs, err := newIndexes(proj)
	if err != nil {
		return err
	}
	for _, doc := range docs {
		idxs.addDocument(doc)
	}
	// Sort index documents then assign document prev/next according to the
	// primary index ordering. Index document ordering ensures subsequent
	// derived document tag indexes are also ordered.
	for _, idx := range idxs {
		idx.docs.sortByDate()
		if idx.primary {
			idx.docs.setPrevNext()
		}
	}
	// Build index pages.
	err = idxs.build(confMod)
	if err != nil {
		return err
	}
	// Render documents. Documents are written before writing indexes so that
	// they are available as soon as possible.
	for _, doc := range docs {
		if !rebuild(doc.buildPath, confMod, doc) {
			continue
		}
		data := doc.frontMatter()
		markup := doc.content
		// Render document markup as a text template.
		if isTemplate(doc.contentPath, doc.templates) {
			proj.verbose2("render template: " + doc.contentPath)
			markup, err = proj.textTemplates.renderText("documentMarkup", markup, data)
			if err != nil {
				return err
			}
		}
		proj.verbose2("render document: " + doc.contentPath)
		// Convert markup to HTML then render document layout to build directory.
		data["body"] = doc.render(markup)
		err = proj.htmlTemplates.render(doc.layout, data, doc.buildPath)
		if err != nil {
			return err
		}
		proj.verbose("write document: " + doc.buildPath)
		proj.verbose2(doc.String())
	}
	fmt.Printf("documents: %d\n", docsCount)
	fmt.Printf("drafts: %d\n", draftsCount)
	// Install home page.
	if proj.rootConf.homepage != "" {
		src := proj.rootConf.homepage
		if !fileExists(src) {
			return fmt.Errorf("homepage file missing: %s", src)
		}
		dst := filepath.Join(proj.buildDir, "index.html")
		if !fileExists(dst) || upToDate(src, dst) {
			proj.verbose2("copy homepage: " + src)
			proj.verbose("write homepage: " + dst)
			if err := copyFile(src, dst); err != nil {
				return err
			}
		}
	}
	fmt.Printf("time: %.2fs\n", time.Now().Sub(startTime).Seconds())
	return nil
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
			proj.verbose("request: " + r.URL.Path)
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
	proj.println(0, fmt.Sprintf("\nServing build directory %s on http://localhost:%s/\nPress Ctrl+C to stop\n", proj.buildDir, proj.port))
	return http.ListenAndServe(":"+proj.port, nil)
}

// isTemplate returns true if the file path f is in the templates configuration value.
func isTemplate(f string, templates *string) bool {
	return strings.Contains(","+nz(templates)+",", ","+filepath.Ext(f)+",")
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

// copyStaticFile copies the content directory srcFile to corresponding build
// directory. Skips if the destination file is up to date. Creates missing
// destination directories.
func (proj *project) copyStaticFile(srcFile string) error {
	if !pathIsInDir(srcFile, proj.contentDir) {
		panic("copyStaticFile: static file is outside content directory: " + srcFile)
	}
	dstFile := pathTranslate(srcFile, proj.contentDir, proj.buildDir)
	if upToDate(dstFile, srcFile) {
		return nil
	}
	proj.verbose("copy static:  " + srcFile)
	err := mkMissingDir(filepath.Dir(dstFile))
	if err != nil {
		return err
	}
	err = copyFile(srcFile, dstFile)
	if err != nil {
		return err
	}
	proj.verbose2("write static: " + dstFile)
	return nil
}

// renderStaticFile renders file f from the content directory as a text template
// and writes it to the corresponding build directory. Skips if the destination
// file is newer than f and is newer than the modified time. Creates missing
// destination directories.
func (proj *project) renderStaticFile(f string, modified time.Time) error {
	// Parse document.
	doc, err := newDocument(f, proj)
	if err != nil {
		return err
	}
	if !rebuild(doc.buildPath, modified, &doc) {
		return nil
	}
	// Render document markup as a text template.
	proj.verbose2("render static: " + doc.contentPath)
	proj.verbose2(doc.String())
	markup := doc.content
	if isTemplate(doc.contentPath, doc.templates) {
		data := doc.frontMatter()
		markup, err = proj.textTemplates.renderText("staticFile", markup, data)
		if err != nil {
			return err
		}
	}
	proj.verbose("write static: " + doc.buildPath)
	err = mkMissingDir(filepath.Dir(doc.buildPath))
	if err != nil {
		return err
	}
	return writeFile(doc.buildPath, markup)
}

// exclude returns true if path name f matches a configuration exclude path.
func (proj *project) exclude(f string) bool {
	for _, pat := range proj.rootConf.exclude {
		if matched, _ := filepath.Match(pat, f); matched {
			return true
		}
	}
	return false
}

// configFor returns the merged configuration for content directory path p.
// Configuration files that are in the corresponding template directory path are
// merged working from top (lowest precedence) to bottom.
//
// For example, if th path is `template/posts/james` then directories are
// searched in the following order: `template`, `template/posts`,
// `template/posts/james` with configuration entries from `template` having
// lowest precedence.
//
// The `proj.confs` have been sorted by configuration `origin` in ascending
// order to ensure the directory precedence.
func (proj *project) configFor(p string) config {
	if !pathIsInDir(p, proj.contentDir) {
		panic("configFor: path outside content directory: " + p)
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
			result.homepage = conf.homepage
			result.urlprefix = conf.urlprefix
		}
	}
	return result
}
