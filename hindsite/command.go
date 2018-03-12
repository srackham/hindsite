package main

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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
	verbose     bool
	set         map[string]string // -set option name/values map.
}

// Cmd is global singleton.
var Cmd = command{}

func (cmd *command) Parse(args []string) error {
	cmd.port = "1212"
	cmd.set = map[string]string{}
	skip := false
	for i, v := range args {
		if skip {
			skip = false
			continue
		}
		switch {
		case i == 0:
			cmd.executable = v
			if len(args) == 1 {
				cmd.name = "help"
			}
		case i == 1:
			if v == "-h" || v == "--help" {
				v = "help"
			}
			if !isCommand(v) {
				return fmt.Errorf("illegal command: %s", v)
			}
			cmd.name = v
		case i == 2 && cmd.name == "help":
			if !isCommand(v) {
				return fmt.Errorf("illegal help topic: %s", v)
			}
			cmd.topic = v
		case v == "-drafts":
			cmd.drafts = true
		case v == "-slugify":
			cmd.slugify = true
		case v == "-clean":
			cmd.clean = true
		case v == "-v":
			cmd.verbose = true
		case stringlist{"-project", "-content", "-template", "-build", "-index", "-port", "-set"}.Contains(v):
			if i+1 >= len(args) {
				return fmt.Errorf("missing %s argument value", v)
			}
			arg := args[i+1]
			switch v {
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
			case "-set":
				m := regexp.MustCompile(`^(\w+?)=(.*)$`).FindStringSubmatch(arg)
				if m == nil {
					return fmt.Errorf("illegal -set name=value argument: %s", arg)
				}
				cmd.set[m[1]] = m[2]
			default:
				panic("illegal arugment: " + v)
			}
			skip = true
		default:
			return fmt.Errorf("illegal argument: %s", v)
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
	// Set configuration values.
	for k, v := range cmd.set {
		if err := Config.set(k, v); err != nil {
			return err
		}
	}
	return nil
}

func isCommand(name string) bool {
	return stringlist{"build", "help", "init", "serve"}.Contains(name)
}

func (cmd *command) Execute() error {
	var err error
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

func (cmd *command) help() {
	println("Usage: hindsite command [arguments]")
}

func (cmd *command) build() error {
	if !dirExists(cmd.contentDir) {
		return fmt.Errorf("content directory does not exist: " + cmd.contentDir)
	}
	if !dirExists(cmd.templateDir) {
		return fmt.Errorf("template directory does not exist: " + cmd.templateDir)
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
	templates := template.New("")
	if cmd.contentDir != cmd.templateDir {
		// Copy static files from template directory to build directory and compile templates.
		err := filepath.Walk(cmd.templateDir, func(f string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				switch filepath.Ext(f) {
				case ".toml", ".yaml":
					// Skip configuration file.
					return nil
				case ".html":
					// Compile template.
					text, err := readFile(f)
					if err != nil {
						return err
					}
					name, err := filepath.Rel(cmd.templateDir, f)
					if err != nil {
						return err
					}
					_, err = templates.New(name).Parse(text)
					if err != nil {
						return err
					}
				default:
					return cmd.copyStaticFile(f)
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	// Parse content documents and copy static files to the build directory.
	docs := []*document{}
	err := filepath.Walk(cmd.contentDir, func(f string, info os.FileInfo, err error) error {
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
			err = doc.parseFile(f)
			if err != nil {
			}
			if doc.draft && !cmd.drafts {
				verbose("skip draft: " + f)
				return nil
			}
			docs = append(docs, &doc)
		case ".toml", ".yaml":
			verbose("skip configuration: " + f)
			return nil
		case ".html":
			if cmd.contentDir == cmd.templateDir {
				verbose("skip template: " + f)
			} else {
				if err := cmd.copyStaticFile(f); err != nil {
					return err
				}
			}
		default:
			if err := cmd.copyStaticFile(f); err != nil {
				return err
			}
		}
		return nil
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
	err = idxs.build(templates)
	if err != nil {
		return err
	}
	// Render documents.
	for _, doc := range docs {
		if cmd.upToDate(doc.contentpath, doc.buildpath) && cmd.upToDate(doc.layoutpath, doc.buildpath) {
			continue
		}
		verbose("render: " + doc.contentpath)
		data := templateData{}
		data.add(doc.frontMatter())
		data["body"] = template.HTML(doc.render())
		tmpl, err := filepath.Rel(cmd.templateDir, doc.layoutpath)
		if err != nil {
			return err
		}
		err = renderTemplate(templates, tmpl, data, doc.buildpath)
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
		if !cmd.upToDate(src, dst) {
			verbose("copy homepage: " + src)
			if err := copyFile(src, dst); err != nil {
				return err
			}
			verbose("write:         " + dst)
		}
	}
	return nil
}

func (cmd *command) upToDate(infile, outfile string) (result bool) {
	// Return true if the -clean option is not set and the infile is older than the
	// outfile.
	if cmd.clean || !fileExists(outfile) {
		return false
	}
	result, err := fileIsOlder(infile, outfile)
	if err != nil {
		return false
	}
	return result
}

// Copy file from content or template directory to corresponding location in build directory.
// Creates any missing build directories.
func (cmd *command) copyStaticFile(f string) error {
	// Copy static files verbatim.
	var inDir string
	switch {
	case pathIsInDir(f, cmd.contentDir):
		inDir = cmd.contentDir
	case pathIsInDir(f, cmd.templateDir):
		inDir = cmd.templateDir
	default:
		return fmt.Errorf("file is not in content or template directory: %s", f)
	}
	outfile, err := filepath.Rel(inDir, f)
	if err != nil {
		return err
	}
	outfile = filepath.Join(cmd.buildDir, outfile)
	if cmd.upToDate(f, outfile) {
		return nil
	}
	verbose("copy static: " + f)
	err = mkMissingDir(filepath.Dir(outfile))
	if err != nil {
		return err
	}
	err = copyFile(f, outfile)
	if err != nil {
		return err
	}
	verbose("write:       " + outfile)
	return nil
}

// Recursively and slugify directory and file names.
func (cmd *command) slugifyDir(dir string) error {
	// TODO
	return nil
}

func (cmd *command) serve() error {
	if !dirExists(cmd.buildDir) {
		return fmt.Errorf("build directory does not exist: " + cmd.buildDir)
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

func (cmd *command) init() error {
	// TODO
	// Use bindata RestoreAssets() to write the builtin example to the target template directory recursively.
	return nil
}
