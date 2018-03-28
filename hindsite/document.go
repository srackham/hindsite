package main

import (
	"bufio"
	"fmt"
	"html/template"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/srackham/go-rimu/rimu"
	blackfriday "gopkg.in/russross/blackfriday.v2"
	yaml "gopkg.in/yaml.v2"
)

// document TODO
type document struct {
	proj         *project // Context.
	conf         config   // Merged configuration for this document.
	contentpath  string
	buildpath    string
	templatepath string    // Virtual path used to find document related templates.
	content      string    // Markup text (without front matter header).
	primaryIndex *index    // Top-level document index (nil if document is not indexed).
	modified     time.Time // Document source file modified timestamp.
	prev         *document // Previous document in primary index.
	next         *document // Next document in primary index.
	// Front matter.
	title    string
	date     time.Time
	author   string
	synopsis string
	addendum string
	url      string // Synthesised absolute or root-relative document URL.
	tags     []string
	draft    bool
	slug     string
	layout   string // Document template name.
}

type documents []*document

// Parse document content and front matter.
func newDocument(contentfile string, proj *project) (document, error) {
	doc := document{}
	doc.proj = proj
	if !fileExists(contentfile) {
		return doc, fmt.Errorf("missing document: %s", contentfile)
	}
	info, err := os.Stat(contentfile)
	if err != nil {
		return doc, err
	}
	doc.modified = info.ModTime()
	doc.contentpath = contentfile
	// Synthesis title, url, draft front matter from content document file name.
	doc.title = fileName(contentfile)
	if doc.title[0] == '~' {
		doc.draft = true
		doc.title = doc.title[1:]
	}
	p, err := filepath.Rel(proj.contentDir, doc.contentpath)
	if err != nil {
		return doc, err
	}
	p = filepath.Dir(p)
	p = filepath.Join(p, doc.title+".html")
	doc.buildpath = filepath.Join(proj.buildDir, p)
	doc.templatepath = filepath.Join(proj.templateDir, p)
	doc.conf = proj.configFor(filepath.Dir(doc.contentpath), filepath.Dir(doc.templatepath))
	doc.url = path.Join("/", doc.conf.urlprefix, filepath.ToSlash(p))
	if regexp.MustCompile(`^\d\d\d\d-\d\d-\d\d-.+`).MatchString(doc.title) {
		d, err := parseDate(doc.title[0:10], nil)
		if err != nil {
			return doc, err
		}
		doc.date = d
		doc.title = doc.title[11:]
	}
	doc.title = strings.Title(strings.Replace(doc.title, "-", " ", -1))
	// Parse embedded front matter.
	doc.author = doc.conf.author // Default author.
	doc.content, err = readFile(doc.contentpath)
	if err != nil {
		return doc, err
	}
	err = doc.extractFrontMatter()
	if err != nil {
		return doc, err
	}
	// If necessary change output file names to match document slug variable.
	if doc.slug != "" {
		doc.buildpath = filepath.Join(filepath.Dir(doc.buildpath), doc.slug+".html")
		doc.url = path.Join(path.Dir(doc.url), doc.slug+".html")
	}
	if doc.layout == "" {
		// Find nearest document layout template.
		layout := ""
		for _, tmpl := range proj.tmpls.layouts {
			if len(tmpl) > len(layout) && pathIsInDir(doc.templatepath, filepath.Dir(tmpl)) {
				layout = tmpl
			}
		}
		if layout == "" {
			return doc, fmt.Errorf("missing layout.html template for: %s", doc.contentpath)
		}
		doc.layout = proj.tmpls.name(layout)
	}
	return doc, nil
}

// extractFrontMatter extracts and parses front matter and synopsis from the
// start of the document. The front matter is stripped from the content.
func (doc *document) extractFrontMatter() error {
	readTo := func(end string, scanner *bufio.Scanner) (text string, eof bool, err error) {
		for scanner.Scan() {
			if end != "" && scanner.Text() == end {
				return text, false, nil
			}
			text += scanner.Text() + "\n"
		}
		if err := scanner.Err(); err != nil {
			return "", false, err
		}
		return text, true, nil
	}
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
	fmText, eof, err := readTo(end, scanner)
	if err != nil {
		return err
	}
	if eof {
		return fmt.Errorf("missing front matter delimiter: %s: %s", end, doc.contentpath)
	}
	synopsis, eof, err := readTo("<!--more-->", scanner)
	if err != nil {
		return err
	}
	if !eof {
		doc.synopsis = synopsis
		content, _, err := readTo("", scanner)
		if err != nil {
			return err
		}
		doc.content = synopsis + content
	} else {
		doc.content = synopsis
	}
	fm := struct {
		Title       string
		Date        string
		Synopsis    string
		Author      string
		Description string
		Addendum    string
		Tags        string   // Comma-separated tags.
		Categories  []string // Tags slice.
		Draft       bool
		Slug        string
		Layout      string
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
	if fm.Author != "" {
		doc.author = fm.Author
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
	if fm.Slug != "" {
		doc.slug = fm.Slug
	}
	if fm.Layout != "" {
		doc.layout = fm.Layout
	}
	return nil
}

func (doc *document) frontMatter() templateData {
	data := templateData{}
	data["title"] = doc.title
	data["shortdate"] = doc.date.In(doc.conf.timezone).Format(doc.conf.shortdate)
	data["mediumdate"] = doc.date.In(doc.conf.timezone).Format(doc.conf.mediumdate)
	data["longdate"] = doc.date.In(doc.conf.timezone).Format(doc.conf.longdate)
	data["date"] = data["mediumdate"] // Alias.
	data["author"] = doc.author
	data["synopsis"] = template.HTML(doc.render(doc.synopsis))
	data["addendum"] = template.HTML(doc.render(doc.addendum))
	data["slug"] = doc.slug
	data["url"] = doc.url
	tags := []map[string]string{}
	for _, tag := range doc.tags {
		url := ""
		if doc.primaryIndex != nil {
			url = path.Join(doc.primaryIndex.url, "tags", doc.primaryIndex.slugs[tag]+"-1.html")
		}
		tags = append(tags, map[string]string{
			"tag": tag,
			"url": url,
		})
	}
	data["tags"] = tags
	// prev/next were assigned when the indexes were built.
	if doc.prev != nil {
		data["prev"] = templateData{"url": doc.prev.url}
	}
	if doc.next != nil {
		data["next"] = templateData{"url": doc.next.url}
	}
	return data
}

// Return front matter as YAML formatted string.
func (doc *document) String() (result string) {
	d, _ := yaml.Marshal(doc.frontMatter())
	return string(d)
}

// Render document markup to HTML.
func (doc *document) render(text string) (html string) {
	// Render document.
	switch filepath.Ext(doc.contentpath) {
	case ".md":
		html = string(blackfriday.Run([]byte(text)))
	case ".rmu":
		conf, err := readFile(filepath.Join(doc.proj.contentDir, "config.rmu"))
		if err == nil {
			text = conf + "\n\n" + text
		}
		html = rimu.Render(text, rimu.RenderOptions{})
	}
	return html
}

// Assign previous and next according to the current sort order.
func (docs documents) setPrevNext() {
	for i, doc := range docs {
		if i == 0 {
			doc.prev = nil
		} else {
			doc.prev = docs[i-1]
		}
		if i >= len(docs)-1 {
			doc.next = nil
		} else {
			doc.next = docs[i+1]
		}
	}
}

// Return documents slice sorted by date descending.
func (docs documents) sortByDate() {
	// Sort documents by decending date.
	sort.Slice(docs, func(i, j int) bool {
		return !docs[i].date.Before(docs[j].date)
	})
}

// Return slice of first n documents.
func (docs documents) first(n int) documents {
	result := documents{}
	for i, doc := range docs {
		if n >= 0 && i >= n {
			break
		}
		result = append(result, doc)
	}
	return result
}

// Return documents front matter template data.
func (docs documents) frontMatter() templateData {
	data := []templateData{}
	for _, doc := range docs {
		data = append(data, doc.frontMatter())
	}
	return templateData{"docs": data}
}
