package main

import (
	"bufio"
	"fmt"
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
	conf         config // Merged configuration for this index.
	contentpath  string
	buildpath    string
	templatepath string    // Virtual path used to find document related templates.
	content      string    // Markup text (without front matter header).
	rootIndex    *index    // Top-level document index (nil if document is not indexed).
	modified     time.Time // Document source file modified timestamp.
	prev         *document // Previous document in primary index.
	next         *document // Next document in primary index.
	// Front matter.
	title    string
	date     time.Time
	author   string
	synopsis string
	addendum string
	url      string // Absolute or root-relative URL.
	tags     []string
	draft    bool
	slug     string
	layout   string // Document template name.
}

type documents []*document

// Parse document content and front matter.
func (doc *document) parseFile(contentfile string, proj *project) error {
	if !fileExists(contentfile) {
		return fmt.Errorf("missing document: %s", contentfile)
	}
	info, err := os.Stat(contentfile)
	if err != nil {
		return err
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
		return err
	}
	p = filepath.Dir(p)
	p = filepath.Join(p, doc.title+".html")
	doc.buildpath = filepath.Join(proj.buildDir, p)
	doc.templatepath = filepath.Join(proj.templateDir, p)
	doc.conf = proj.configFor(filepath.Dir(doc.contentpath), filepath.Dir(doc.templatepath))
	doc.url = path.Join(doc.conf.urlprefix, filepath.ToSlash(p))
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
	doc.author = doc.conf.author // Default author.
	doc.content, err = readFile(doc.contentpath)
	if err != nil {
		return err
	}
	err = doc.extractFrontMatter()
	if err != nil {
		return err
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
			return fmt.Errorf("missing layout.html template for: %s", doc.contentpath)
		}
		doc.layout = proj.tmpls.name(layout)
	}
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

func (doc *document) frontMatter() (data templateData) {
	data = templateData{}
	data["title"] = doc.title
	data["date"] = doc.date.Format("02-Jan-2006")
	data["author"] = doc.author
	data["synopsis"] = doc.synopsis
	data["addendum"] = doc.addendum
	data["slug"] = doc.slug
	data["url"] = doc.url
	prev := templateData{}
	if doc.prev != nil {
		prev["title"] = doc.prev.title
		prev["url"] = doc.prev.url
	}
	data["prev"] = prev
	next := templateData{}
	if doc.next != nil {
		next["title"] = doc.next.title
		next["url"] = doc.next.url
	}
	data["next"] = next
	tags := []map[string]string{}
	for _, tag := range doc.tags {
		url := ""
		if doc.rootIndex != nil {
			url = path.Join(doc.rootIndex.url, "tags", doc.rootIndex.tagfiles[tag])
		}
		tags = append(tags, map[string]string{
			"tag": tag,
			"url": url,
		})
	}
	data["tags"] = tags
	return data
}

// Return front matter as YAML formatted string.
func (doc *document) String() (result string) {
	d, _ := yaml.Marshal(doc.frontMatter())
	return string(d)
}

// Render document markup to HTML.
func (doc *document) render() (html string) {
	// Render document.
	switch filepath.Ext(doc.contentpath) {
	case ".md":
		html = string(blackfriday.Run([]byte(doc.content)))
	case ".rmu":
		html = rimu.Render(doc.content, rimu.RenderOptions{})
	}
	return html
}

// Assign previous and next according to the current sort order.
func (docs documents) setPrevNext() documents {
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
	return docs
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
