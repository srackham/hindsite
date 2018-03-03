package main

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path"
	"path/filepath"

	blackfriday "gopkg.in/russross/blackfriday.v2"
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
}

func (cmd *Command) Parse(args []string) error {
	cmd.projectDir = "."
	cmd.contentDir = "content"
	cmd.templateDir = "template"
	cmd.buildDir = "build"
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
		case stringlist{"-project", "-content", "-template", "-build"}.Contains(v):
			// Consume the argument value and skip next iteration.
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
	return stringlist{"build", "help", "init", "run"}.Contains(name)
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
	case "run":
		err = cmd.run()
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
	files, _ := filepath.Glob(path.Join(cmd.buildDir, "*"))
	for _, f := range files {
		os.RemoveAll(f)
	}
	files, _ = filepath.Glob(path.Join(cmd.contentDir, "*.md"))
	for _, f := range files {
		markup := readFile(f)
		tmpl, _ := template.ParseFiles(path.Join(cmd.templateDir, "layout.html"))
		output := renderWebpage(markup, tmpl)
		outfile := path.Join(cmd.buildDir, path.Base(replaceExt(f, ".html")))
		writeFile(outfile, output)
	}
	return nil
}

func (cmd *Command) run() error {
	// TODO
	return nil
}

func (cmd *Command) init() error {
	// TODO
	// Use bindata RestoreAssets() to write the builtin example to the target template directory recursively.
	return nil
}

func renderWebpage(markup string, tmpl *template.Template) (result string) {
	html := blackfriday.Run([]byte(markup))
	data := struct {
		Title string
		Body  template.HTML
	}{"foobar", template.HTML(html)}
	buf := bytes.NewBufferString("")
	tmpl.Execute(buf, data)
	return buf.String()
}
