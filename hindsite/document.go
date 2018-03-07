package main

import (
	"bufio"
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	blackfriday "gopkg.in/russross/blackfriday.v2"
	yaml "gopkg.in/yaml.v2"
)

type templateData map[string]interface{}

// document TODO
type document struct {
	contentpath string // Content directory file path.
	buildpath   string // Build directory file path.
	content     string // Markup text (without front matter header).
	// Front matter.
	title    string
	date     time.Time
	synopsis string
	addendum string
	url      string // URL path relatative to server root with leading slash.
	tags     []string
	draft    bool
}

// Parse document content and front matter.
func (doc *document) parseFile(contentfile string) error {
	if !fileExists(contentfile) {
		return fmt.Errorf("missing document: %s", contentfile)
	}
	doc.contentpath = contentfile
	// Synthesis title, url, draft front matter from content document file name.
	doc.title = fileName(contentfile)
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
	doc.buildpath = filepath.Join(Cmd.buildDir, p)
	doc.url = Config.urlprefix + "/" + filepath.ToSlash(p)
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
	Indexes.add(doc)
	return nil
}

// Extract and parse front matter from the start of the document.
func (doc *document) extractFrontMatter() error {
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

func (doc *document) mergeToTemplateData(data templateData) {
	data["title"] = doc.title
	data["date"] = doc.date.Format("02-Jan-2006")
	data["tags"] = strings.Join(doc.tags, ", ")
	data["synopsis"] = doc.synopsis
	data["addendum"] = doc.addendum
	data["url"] = doc.url
}

// func (doc *document) toString() (result string) {
// 	result += "---\n"
// 	for k, v := range doc.frontMatter() {
// 		result += fmt.Sprintf("%-8s: %s\n", k, v)
// 	}
// 	result += "---"
// 	return result
// }

// Render document markup and document variables with the document layout template.
// Return rendered HTML.
func (doc *document) render() (string, error) {
	// Render document.
	var body string
	switch filepath.Ext(doc.contentpath) {
	case ".md":
		body = string(blackfriday.Run([]byte(doc.content)))
	}
	// TODO: Look up layout.
	layout := filepath.Join(Cmd.templateDir, "layout.html")
	tmpl, err := template.ParseFiles(layout)
	if err != nil {
		return "", err
	}
	data := templateData{}
	data["body"] = template.HTML(body)
	// data.add(doc.frontMatter())
	doc.mergeToTemplateData(data)
	buf := bytes.NewBufferString("")
	tmpl.Execute(buf, data)
	return buf.String(), nil
}
