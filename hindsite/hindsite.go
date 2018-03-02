package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path"

	blackfriday "gopkg.in/russross/blackfriday.v2"
)

func main() {
	if len(os.Args) == 1 {
		helpCommand()
		os.Exit(0)
	}
	cmd := os.Args[1]
	switch cmd {
	case "help":
		helpCommand()
	case "build":
		Config.InitDirs(os.Args[3], "", "", "")
		buildCommand()
	case "run":
		runCommand()
	default:
		die("illegal command:" + cmd)
	}
}

func helpCommand() {
	println("Usage: hindsite build -project PROJECT_DIR")
}

func buildCommand() {
	markup := readFile(path.Join(Config.contentDir, "index.md"))
	tmpl, _ := template.ParseFiles(path.Join(Config.templateDir, "layout.html"))
	output := renderWebpage(markup, tmpl)
	writeFile(Config.buildDir+"/index.html", output)
}

func runCommand() {
	// TODO
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

// Helpers.
func die(message string) {
	if message != "" {
		fmt.Fprintln(os.Stderr, message)
	}
	os.Exit(1)
}

func fileExists(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}

func readFile(filename string) string {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		die(err.Error())
	}
	return string(bytes)
}

func writeFile(filename string, text string) {
	err := ioutil.WriteFile(filename, []byte(text), 0644)
	if err != nil {
		die(err.Error())
	}
}
