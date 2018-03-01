package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"

	blackfriday "gopkg.in/russross/blackfriday.v2"
)

func main() {
	markup := readFile(os.Args[1] + "/index.md")
	tmpl, _ := template.ParseFiles(os.Args[1] + "/layout.html")
	output := renderWebpage(markup, tmpl)
	writeFile(os.Args[1]+"/index.html", output)
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
