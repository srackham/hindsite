package main

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
	"time"

	blackfriday "gopkg.in/russross/blackfriday.v2"
)

type TemplateData map[string]interface{}

// Document TODO
type Document struct {
	Title    string
	Date     time.Time
	Synopsis string
	Addendum string
	Tags     []string
	Draft    bool
	path     string // File path.
	content  string // Markup text (without front matter header).
	html     string // Rendered content.
}

/*
// NewDocument TODO
func NewDocument(docfile string) *Document {
	// TODO
	result := new(Document)
	return result
}
*/

// Parse document content and front matter.
func (doc *Document) parseFile(name string) error {
	if !fileExists(name) {
		return fmt.Errorf("missing document: %s", name)
	}
	doc.path = name
	doc.content = readFile(name)
	doc.html = string(blackfriday.Run([]byte(doc.content)))
	doc.Title = strings.Title(strings.Replace(fileName(name), "-", " ", -1))
	return nil
}

func (doc *Document) mergeData(data TemplateData) {
	data["title"] = doc.Title
	data["body"] = template.HTML(doc.html)
}

func (doc *Document) renderWebpage(tmpl *template.Template, data TemplateData) string {
	doc.mergeData(data)
	buf := bytes.NewBufferString("")
	tmpl.Execute(buf, data)
	return buf.String()
}
