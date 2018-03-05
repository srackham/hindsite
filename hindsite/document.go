package main

import (
	"bufio"
	"bytes"
	"fmt"
	"html/template"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	blackfriday "gopkg.in/russross/blackfriday.v2"
	yaml "gopkg.in/yaml.v2"
)

type TemplateData map[string]interface{}

// Document TODO
type Document struct {
	contentpath string // Content directory file path.
	buildpath   string // Build directory file path.
	urlpath     string // URL path relatative to server root with leading slash.
	content     string // Markup text (without front matter header).
	html        string // Rendered content.
	// Front matter.
	title    string
	date     time.Time
	synopsis string
	addendum string
	tags     []string
	draft    bool
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
	doc.contentpath = name
	// Synthesis default front matter from file name.
	doc.title = fileName(name)
	if doc.title[0] == '~' {
		doc.draft = true
		doc.title = doc.title[1:]
	}
	p, err := filepath.Rel(Cmd.contentDir, doc.contentpath)
	if err != nil {
		return err
	}
	p = filepath.Dir(p)
	p = filepath.Join(p, doc.title+".html")
	doc.buildpath = path.Join(Cmd.buildDir, p)
	doc.urlpath = "/" + filepath.ToSlash(p)
	if regexp.MustCompile(`^\d\d\d\d-\d\d-\d\d-.+`).MatchString(doc.title) {
		d, err := parseDate(doc.title[0:10], nil)
		if err != nil {
			return err
		}
		doc.date = d
		doc.title = doc.title[11:]
	}
	doc.title = strings.Title(strings.Replace(doc.title, "-", " ", -1))
	// Parse embedded front matter.
	doc.content, err = readFile(doc.contentpath)
	if err != nil {
		return err
	}
	err = doc.extractFrontMatter()
	if err != nil {
		return err
	}
	// Render document.
	doc.html = string(blackfriday.Run([]byte(doc.content)))
	return nil
}

// Extract and parse front matter from the start of the document.
func (doc *Document) extractFrontMatter() error {
	scanner := bufio.NewScanner(strings.NewReader(doc.content))
	if !scanner.Scan() {
		return scanner.Err()
	}
	var end, format string
	switch scanner.Text() {
	case "<!--":
		format = "yaml"
		end = "-->"
	case "---":
		format = "yaml"
		end = "---"
	case "+++":
		format = "toml"
		end = "+++"
	default:
		return nil
	}
	fmText := ""
	for scanner.Scan() {
		if scanner.Text() == end {
			break
		}
		fmText += scanner.Text() + "\n"
	}
	content := ""
	for scanner.Scan() {
		content += scanner.Text() + "\n"
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	doc.content = content
	fm := struct {
		Title       string
		Date        string
		Synopsis    string
		Description string
		Addendum    string
		Tags        string
		Categories  []string
		Draft       bool
	}{}
	switch format {
	case "toml":
		_, err := toml.Decode(fmText, &fm)
		if err != nil {
			return err
		}
	case "yaml":
		err := yaml.Unmarshal([]byte(fmText), &fm)
		if err != nil {
			return err
		}
	}
	// Merge parsed front matter.
	if fm.Title != "" {
		doc.title = fm.Title
	}
	if fm.Date != "" {
		// TODO parse
		d, err := parseDate(fm.Date, nil)
		if err != nil {
			return err
		}
		doc.date = d
	}
	if fm.Synopsis != "" {
		doc.synopsis = fm.Synopsis
	}
	if fm.Description != "" {
		doc.synopsis = fm.Description
	}
	if fm.Addendum != "" {
		doc.addendum = fm.Addendum
	}
	if fm.Tags != "" {
		doc.tags = strings.Split(fm.Tags, ",")
		for i, v := range doc.tags {
			doc.tags[i] = strings.TrimSpace(v)
		}
	}
	if len(fm.Categories) > 0 {
		doc.tags = fm.Categories
	}
	if !doc.draft { // File name tilda flag overrides embedded draft flag.
		doc.draft = fm.Draft
	}
	return nil
}

func (doc *Document) mergeToTemplateData(data TemplateData) {
	data["body"] = template.HTML(doc.html)
	data["title"] = doc.title
	data["date"] = doc.date.Format("02-Jan-2006")
	data["tags"] = strings.Join(doc.tags, ", ")
	data["synopsis"] = doc.synopsis
	data["addendum"] = doc.addendum
}

func (doc *Document) renderWebpage(tmpl *template.Template, data TemplateData) string {
	doc.mergeToTemplateData(data)
	buf := bytes.NewBufferString("")
	tmpl.Execute(buf, data)
	return buf.String()
}
