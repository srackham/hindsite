package main

import (
	"bytes"
	"fmt"
	"html/template"
	"regexp"
	"strings"
	"time"

	blackfriday "gopkg.in/russross/blackfriday.v2"
)

type TemplateData map[string]interface{}

// Document TODO
type Document struct {
	title    string
	date     time.Time
	synopsis string
	addendum string
	tags     []string
	draft    bool
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
	doc.title = fileName(name)
	if regexp.MustCompile(`^\d\d\d\d-\d\d-\d\d-.+`).MatchString(doc.title) {
		loc, _ := time.LoadLocation("Local")
		t, err := time.ParseInLocation(time.RFC3339, doc.title[0:10]+"T00:00:00+00:00", loc)
		if err != nil {
			return err
		}
		doc.date = t
		doc.title = doc.title[11:]
	}
	doc.title = strings.Title(strings.Replace(doc.title, "-", " ", -1))
	return nil
}

func (doc *Document) mergeData(data TemplateData) {
	data["title"] = doc.title
	data["date"] = doc.date.Format("Jan-02-2006")
	data["body"] = template.HTML(doc.html)
}

func (doc *Document) renderWebpage(tmpl *template.Template, data TemplateData) string {
	doc.mergeData(data)
	buf := bytes.NewBufferString("")
	tmpl.Execute(buf, data)
	return buf.String()
}
