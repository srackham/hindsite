package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path"
	"path/filepath"
)

type Command struct {
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
}

// Cmd is singleton command.
var Cmd = Command{}

func (cmd *Command) Parse(args []string) error {
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
		case stringlist{"-project", "-content", "-template", "-build", "-port"}.Contains(v):
			if i+1 >= len(args) {
				return fmt.Errorf("missing %s argument value", v)
			}
			switch v {
			case "-project":
				cmd.projectDir = args[i+1]
			case "-content":
				cmd.contentDir = args[i+1]
			case "-template":
				cmd.templateDir = args[i+1]
			case "-build":
				cmd.buildDir = args[i+1]
			case "-port":
				cmd.port = args[i+1]
			default:
				panic("illegal arugment: " + v)
			}
			skip = true
		default:
			return fmt.Errorf("illegal argument: %s", v)
		}
	}
	if !path.IsAbs(cmd.contentDir) {
		cmd.contentDir = path.Join(cmd.projectDir, cmd.contentDir)
	}
	if !path.IsAbs(cmd.templateDir) {
		cmd.templateDir = path.Join(cmd.projectDir, cmd.templateDir)
	}
	if !path.IsAbs(cmd.buildDir) {
		cmd.buildDir = path.Join(cmd.projectDir, cmd.buildDir)
	}
	return nil
}

func isCommand(name string) bool {
	return stringlist{"build", "help", "init", "serve"}.Contains(name)
}

func (cmd *Command) Execute() error {
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

func (cmd *Command) help() {
	println("Usage: hindsite command [arguments]")
}

func (cmd *Command) build() error {
	if !dirExists(cmd.contentDir) {
		return fmt.Errorf("content directory does not exist: " + cmd.contentDir)
	}
	if !dirExists(cmd.templateDir) {
		return fmt.Errorf("template directory does not exist: " + cmd.templateDir)
	}
	if !dirExists(cmd.buildDir) {
		os.Mkdir(cmd.buildDir, 0775)
	}
	// Delete everything in the build directory.
	files, _ := filepath.Glob(path.Join(cmd.buildDir, "*"))
	for _, f := range files {
		os.RemoveAll(f)
	}
	// Process all content documents in the content directory.
	err := filepath.Walk(cmd.contentDir, func(f string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		// TODO: Skip build directory.
		println("infile: " + f)
		switch filepath.Ext(f) {
		case ".toml", ".yaml", ".html":
			// Skip configuration and template files.
			return nil
		case ".md":
			doc := Document{}
			err = doc.parseFile(f)
			if err != nil {
				return err
			}
			if doc.draft && !cmd.drafts {
				return nil
			}
			tmpl, err := template.ParseFiles(path.Join(cmd.templateDir, "layout.html"))
			if err != nil {
				return err
			}
			data := TemplateData{}
			html := doc.renderWebpage(tmpl, data)
			err = mkFileDir(doc.buildpath)
			if err != nil {
				return err
			}
			err = writeFile(doc.buildpath, html)
			if err != nil {
				return err
			}
			println("outfile: " + doc.buildpath)
			println("URL: " + doc.urlpath)
		default:
			// Copy static files verbatim.
			outfile, err := filepath.Rel(cmd.contentDir, f)
			if err != nil {
				return err
			}
			outfile = path.Join(cmd.buildDir, outfile)
			err = mkFileDir(outfile)
			if err != nil {
				return err
			}
			err = copyFile(f, outfile)
			if err != nil {
				return err
			}
			println("outfile: " + outfile)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (cmd *Command) serve() error {
	if !dirExists(cmd.buildDir) {
		return fmt.Errorf("build directory does not exist: " + cmd.buildDir)
	}
	http.Handle("/", http.FileServer(http.Dir(cmd.buildDir)))
	fmt.Printf("\nServing build directory %s on http://localhost:%s/\nPress Ctrl+C to stop\n", cmd.buildDir, cmd.port)
	return http.ListenAndServe(":"+cmd.port, nil)
}

func (cmd *Command) init() error {
	// TODO
	// Use bindata RestoreAssets() to write the builtin example to the target template directory recursively.
	return nil
}
