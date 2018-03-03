package main

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path"
	"strings"

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

func (cmd *Command) Parse(args []string) bool {
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
				fmt.Printf("illegal command: %s\n", v)
				return false
			}
			cmd.name = v
		case i == 2 && cmd.name == "help":
			if !isCommand(v) {
				fmt.Printf("illegal help topic: %s\n", v)
				return false
			}
			cmd.topic = v
		case v == "-drafts":
			cmd.drafts = true
		case v == "-slugify":
			cmd.slugify = true
		case strings.Contains("-project|-content|-template|-build", v):
			if i >= len(args) {
				fmt.Printf("missing argument value: %s\n", v)
				return false
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
			fmt.Printf("illegal argument: %s\n", v)
			return false
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
	return true
}

func isCommand(name string) bool {
	return strings.Contains("build|help|init|run", name)
}

func (cmd *Command) Execute() {
	switch cmd.name {
	case "build":
		cmd.build()
	case "help":
		cmd.help()
	case "init":
		cmd.init()
	case "run":
		cmd.run()
	default:
		panic("illegal command: " + cmd.name)
	}
}

func (cmd *Command) help() {
	println("Usage: hindsite command [arguments]")
}

func (cmd *Command) build() {
	if !dirExists(cmd.contentDir) {
		die("content directory does not exist: " + cmd.contentDir)
	}
	if !dirExists(cmd.templateDir) {
		die("template directory does not exist: " + cmd.templateDir)
	}
	if !dirExists(cmd.buildDir) {
		os.Mkdir(cmd.buildDir, 0775)
	}
	markup := readFile(path.Join(cmd.contentDir, "index.md"))
	tmpl, _ := template.ParseFiles(path.Join(cmd.templateDir, "layout.html"))
	output := renderWebpage(markup, tmpl)
	writeFile(path.Join(cmd.buildDir, "index.html"), output)
}

func (cmd *Command) run() {
	// TODO
}

func (cmd *Command) init() {
	// TODO
	// Use bindata RestoreAssets() to write the builtin example to the target template directory recursively.
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
