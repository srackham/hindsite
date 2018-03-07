package main

import (
	"fmt"
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
	drafts      bool
	slugify     bool
	topic       string
	port        string
	update      bool
	verbose     bool
}

// Cmd is global singleton.
var Cmd = command{}

func (cmd *command) Parse(args []string) error {
	cmd.projectDir = "."
	cmd.contentDir = "content"
	cmd.templateDir = "template"
	cmd.buildDir = "build"
	cmd.port = "1212"
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
		case v == "-update":
			cmd.update = true
		case v == "-v":
			cmd.verbose = true
		case stringlist{"-project", "-content", "-template", "-build", "-port", "-set"}.Contains(v):
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
			case "-port":
				cmd.port = arg
			case "-set":
				m := regexp.MustCompile(`^(\w+?)=(.*)$`).FindStringSubmatch(arg)
				if m == nil {
					return fmt.Errorf("illegal -set name=value argument: %s", arg)
				}
				if err := Config.set(m[1], m[2]); err != nil {
					return err
				}
			default:
				panic("illegal arugment: " + v)
			}
			skip = true
		default:
			return fmt.Errorf("illegal argument: %s", v)
		}
	}
	// Clean and convert directories to absolute paths.
	var err error
	cmd.projectDir, err = filepath.Abs(cmd.projectDir)
	if err != nil {
		return err
	}
	if !filepath.IsAbs(cmd.contentDir) {
		cmd.contentDir = filepath.Join(cmd.projectDir, cmd.contentDir)
	}
	if !filepath.IsAbs(cmd.templateDir) {
		cmd.templateDir = filepath.Join(cmd.projectDir, cmd.templateDir)
	}
	if !filepath.IsAbs(cmd.buildDir) {
		cmd.buildDir = filepath.Join(cmd.projectDir, cmd.buildDir)
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
	if Cmd.contentDir != Cmd.templateDir {
		if err := checkOverlap("content", Cmd.contentDir, "template", cmd.templateDir); err != nil {
			return err
		}
	}
	if filepath.Dir(Cmd.buildDir) != Cmd.contentDir {
		if err := checkOverlap("build", Cmd.buildDir, "content", cmd.contentDir); err != nil {
			return err
		}
		if err := checkOverlap("build", Cmd.buildDir, "template", cmd.templateDir); err != nil {
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
	if !cmd.update {
		// Delete everything in the build directory.
		files, _ := filepath.Glob(filepath.Join(cmd.buildDir, "*"))
		for _, f := range files {
			if err := os.RemoveAll(f); err != nil {
				return err
			}
		}
	}
	// Process all content documents in the content directory.
	err := filepath.Walk(cmd.contentDir, func(f string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if f == Cmd.buildDir {
				return filepath.SkipDir
			}
			return nil
		}
		switch filepath.Ext(f) {
		case ".toml", ".yaml", ".html":
			// Skip configuration and template files.
			verbose("skipping: " + f)
			return nil
		case ".md", ".rmu":
			doc := document{}
			err = doc.parseFile(f)
			if err != nil {
			}
			if cmd.upToDate(f, doc.buildpath) {
				return nil
			}
			if doc.draft && !cmd.drafts {
				verbose("skipping: " + f)
				return nil
			}
			verbose("render:   " + f)
			html, err := doc.render()
			if err != nil {
				return err
			}
			err = mkMissingDir(filepath.Dir(doc.buildpath))
			if err != nil {
				return err
			}
			err = writeFile(doc.buildpath, html)
			if err != nil {
				return err
			}
			verbose("outfile:  " + doc.buildpath)
			verbose(doc.String())
		default:
			// Copy static files verbatim.
			outfile, err := filepath.Rel(cmd.contentDir, f)
			if err != nil {
				return err
			}
			outfile = filepath.Join(cmd.buildDir, outfile)
			if cmd.upToDate(f, outfile) {
				return nil
			}
			verbose("copying:  " + f)
			err = mkMissingDir(filepath.Dir(outfile))
			if err != nil {
				return err
			}
			err = copyFile(f, outfile)
			if err != nil {
				return err
			}
			verbose("outfile:  " + outfile)
		}
		return nil
	})
	if err != nil {
		return err
	}
	// Build indexes.
	for _, idx := range Indexes {
		if err := idx.build(); err != nil {
			return err
		}
	}
	return nil
}

func (cmd *command) upToDate(infile, outfile string) (result bool) {
	// Return true if the -update option is set and the infile is older than the
	// outfile.
	if !cmd.update || !fileExists(outfile) {
		return false
	}
	result, err := fileIsOlder(infile, outfile)
	if err != nil {
		return false
	}
	return result
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
