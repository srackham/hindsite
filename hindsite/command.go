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
	slugify     bool
	topic       string
	port        string
	clean       bool
	builtin     bool
	verbose     bool
	// config      config // Root configuration.
}

func newProject() project {
	return project{}
}

// Print message if `-v` verbose option set.
func (proj *project) println(message string) {
	if proj.verbose {
		fmt.Println(message)
	}
}

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
		case i == 2 && proj.command == "help":
			if !isCommand(opt) {
				return fmt.Errorf("illegal help topic: %s", opt)
			}
			proj.topic = opt
		case opt == "-drafts":
			proj.drafts = true
		case opt == "-slugify":
			proj.slugify = true
		case opt == "-clean":
			proj.clean = true
		case opt == "-builtin":
			proj.builtin = true
		case opt == "-v":
			proj.verbose = true
		case stringlist{"-project", "-content", "-template", "-build", "-index", "-port"}.Contains(opt):
			if i+1 >= len(args) {
				return fmt.Errorf("missing %s argument value", opt)
			}
			arg := args[i+1]
			switch opt {
			case "-project":
				proj.projectDir = arg
			case "-content":
				proj.contentDir = arg
			case "-template":
				proj.templateDir = arg
			case "-build":
				proj.buildDir = arg
			case "-index":
				proj.indexDir = arg
			case "-port":
				proj.port = arg
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
	proj.projectDir, err = getPath(proj.projectDir, ".")
	if err != nil {
		return err
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
	if proj.contentDir != proj.templateDir {
		if err := checkOverlap("content", proj.contentDir, "template", proj.templateDir); err != nil {
			return err
		}
	}
	if filepath.Dir(proj.buildDir) != proj.contentDir {
		if err := checkOverlap("build", proj.buildDir, "content", proj.contentDir); err != nil {
			return err
		}
		if err := checkOverlap("build", proj.buildDir, "template", proj.templateDir); err != nil {
			return err
		}
	}
	if !(pathIsInDir(proj.indexDir, proj.buildDir) || proj.indexDir == proj.buildDir) {
		return fmt.Errorf("index directory must reside in build directory: %s", proj.buildDir)
	}
	if !dirExists(proj.projectDir) {
		return fmt.Errorf("missing project directory: " + proj.projectDir)
	}
	return nil
}

func isCommand(name string) bool {
	return stringlist{"build", "help", "init", "serve"}.Contains(name)
}

func (proj *project) execute() error {
	var err error
	// Parse configuration files from template and content directories (content directory has precedence).
	for _, dir := range []string{proj.templateDir, proj.contentDir} {
		for _, conf := range []string{"config.toml", "config.yaml"} {
			f := filepath.Join(dir, conf)
			if fileExists(f) {
				proj.println("read config: " + f)
				if err := Config.parseFile(f, proj); err != nil {
					return err
				}
			}
		}
	}
	proj.println("config: \n" + Config.String())
	// Execute command.
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
	if proj.builtin {
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
		proj.println("installing builtin template")
		if err := RestoreAssets(proj.templateDir, ""); err != nil {
			return err
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

func (proj *project) help() {
	println("Usage: hindsite command [arguments]")
}

func (proj *project) build() error {
	if !dirExists(proj.contentDir) {
		return fmt.Errorf("missing content directory: " + proj.contentDir)
	}
	if !dirExists(proj.templateDir) {
		return fmt.Errorf("missing template directory: " + proj.templateDir)
	}
	if err := proj.slugifyDir(proj.contentDir); err != nil {
		return err
	}
	if proj.contentDir != proj.templateDir {
		if err := proj.slugifyDir(proj.templateDir); err != nil {
			return err
		}
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
	tmpls := newTemplates(proj.templateDir)
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
				tmpl := tmpls.name(f)
				proj.println("parse template: " + tmpl)
				err = tmpls.add(tmpl)
			default:
				if proj.contentDir != proj.templateDir {
					err = proj.copyStaticFile(f, proj.templateDir, proj.buildDir)
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
	err = filepath.Walk(proj.contentDir, func(f string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if f == proj.buildDir {
				// Do not process the build directory.
				return filepath.SkipDir
			}
			return nil
		}
		switch filepath.Ext(f) {
		case ".md", ".rmu":
			doc := document{}
			err = doc.parseFile(f, proj, tmpls)
			if err != nil {
			}
			if doc.draft && !proj.drafts {
				proj.println("skip draft: " + f)
				return nil
			}
			docs = append(docs, &doc)
		case ".toml", ".yaml":
			if isOlder(confMod, info.ModTime()) {
				confMod = info.ModTime()
			}
		case ".html":
			if proj.contentDir != proj.templateDir {
				err = proj.copyStaticFile(f, proj.contentDir, proj.buildDir)
			}
		default:
			err = proj.copyStaticFile(f, proj.contentDir, proj.buildDir)
		}
		return err
	})
	if err != nil {
		return err
	}
	// Build indexes.
	idxs := indexes{}
	idxs.init(proj)
	for _, doc := range docs {
		idxs.addDocument(doc)
	}
	err = idxs.build(proj, tmpls, confMod)
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
		data.add(doc.frontMatter())
		data["body"] = template.HTML(doc.render())
		err = tmpls.render(doc.layout, data, doc.buildpath)
		if err != nil {
			return err
		}
		proj.println("write:  " + doc.buildpath)
		proj.println(doc.String())
	}
	if Config.homepage != "" {
		// Install home page.
		src := Config.homepage
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

// Recursively and slugify directory and file names.
func (proj *project) slugifyDir(dir string) error {
	// TODO
	return nil
}

func (proj *project) serve() error {
	if !dirExists(proj.buildDir) {
		return fmt.Errorf("missing build directory: " + proj.buildDir)
	}
	if !dirExists(proj.contentDir) {
		fmt.Fprintln(os.Stderr, "warning: missing content directory: "+proj.contentDir)
	}
	if !dirExists(proj.templateDir) {
		fmt.Fprintln(os.Stderr, "warning: missing template directory: "+proj.templateDir)
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
	http.Handle("/", stripPrefix(Config.urlprefix, http.FileServer(http.Dir(proj.buildDir))))
	fmt.Printf("\nServing build directory %s on http://localhost:%s/\nPress Ctrl+C to stop\n", proj.buildDir, proj.port)
	return http.ListenAndServe(":"+proj.port, nil)
}
