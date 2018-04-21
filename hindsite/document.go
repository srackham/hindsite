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
	contentPath  string
	buildPath    string
	templatePath string    // Virtual path used to find document related templates.
	header       string    // Front matter document header text.
	content      string    // Markup text (without front matter header).
	modified     time.Time // Document source file modified timestamp.
	primaryIndex *index    // Top-level document index (nil if document is not indexed).
	prev         *document // Previous document in primary index.
	next         *document // Next document in primary index.
	// Front matter.
	title       string
	date        time.Time
	author      *string
	templates   *string
	description string
	addendum    string
	url         string // Synthesised document URL.
	tags        []string
	draft       bool
	slug        string
	layout      string            // Document template name.
	user        map[string]string // User defined configuration key/values.
}

type documents []*document

// Parse document content and front matter.
func newDocument(contentfile string, proj *project) (document, error) {
	if !pathIsInDir(contentfile, proj.contentDir) {
		panic("newDocument: document is outside content directory: " + contentfile)
	}
	if !fileExists(contentfile) {
		panic("newDocument: missing document: " + contentfile)
	}
	doc := document{}
	doc.proj = proj
	info, err := os.Stat(contentfile)
	if err != nil {
		return doc, err
	}
	doc.modified = info.ModTime()
	doc.contentPath = contentfile
	p, _ := filepath.Rel(proj.contentDir, doc.contentPath)
	switch filepath.Ext(p) {
	case ".md", ".rmu":
		p = replaceExt(p, ".html")
	}
	doc.buildPath = filepath.Join(proj.buildDir, p)
	doc.templatePath = filepath.Join(proj.templateDir, p)
	doc.conf = proj.configFor(doc.contentPath)
	doc.url = path.Join("/", doc.conf.urlprefix, filepath.ToSlash(p))
	// Extract title and date from file name.
	doc.title = fileName(contentfile)
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
	doc.author = doc.conf.author       // Default author.
	doc.templates = doc.conf.templates // Default templates.
	doc.content, err = readFile(doc.contentPath)
	if err != nil {
		return doc, err
	}
	if err := doc.extractFrontMatter(); err != nil {
		return doc, err
	}
	if doc.slug != "" {
		// Change output file names to match document slug variable.
		f := doc.slug + filepath.Ext(doc.buildPath)
		doc.buildPath = filepath.Join(filepath.Dir(doc.buildPath), f)
		doc.url = path.Join(path.Dir(doc.url), f)
	}
	if doc.layout == "" {
		// Find nearest document layout template file.
		layout := ""
		for _, l := range proj.htmlTemplates.layouts {
			if len(l) > len(layout) && pathIsInDir(doc.templatePath, filepath.Dir(l)) {
				layout = l
			}
		}
		if layout == "" {
			return doc, fmt.Errorf("missing layout.html template for: %s", doc.contentPath)
		}
		doc.layout = proj.htmlTemplates.name(layout)
	}
	return doc, nil
}

// extractFrontMatter extracts and parses front matter and description from the
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
	case "---":
		format = "yaml"
		end = "---"
	case "+++":
		format = "toml"
		end = "+++"
	case "<!--":
		format = "yaml"
		end = "-->"
	case "/***":
		format = "yaml"
		end = "***/"
	default:
		return nil
	}
	header, eof, err := readTo(end, scanner)
	if err != nil {
		return err
	}
	if eof {
		return fmt.Errorf("missing front matter delimiter: %s: %s", end, doc.contentPath)
	}
	doc.header = header
	description, eof, err := readTo("<!--more-->", scanner)
	if err != nil {
		return err
	}
	if !eof {
		doc.description = description
		content, _, err := readTo("", scanner)
		if err != nil {
			return err
		}
		doc.content = description + content
	} else {
		doc.content = description
	}
	fm := struct {
		Title       string
		Date        string
		Description string
		Addendum    string
		Author      *string
		Templates   *string
		Tags        []string
		Draft       bool
		Slug        string
		Layout      string
		User        map[string]string
	}{}
	switch format {
	case "toml":
		_, err := toml.Decode(header, &fm)
		if err != nil {
			return err
		}
	case "yaml":
		err := yaml.Unmarshal([]byte(header), &fm)
		if err != nil {
			return err
		}
	}
	// Merge parsed front matter.
	if fm.Title != "" {
		doc.title = fm.Title
	}
	if fm.Date != "" {
		d, err := parseDate(fm.Date, nil)
		if err != nil {
			return err
		}
		doc.date = d
	}
	if fm.Author != nil {
		doc.author = fm.Author
	}
	if fm.Templates != nil {
		doc.templates = fm.Templates
	}
	if fm.Description != "" {
		doc.description = fm.Description
	}
	if fm.Addendum != "" {
		doc.addendum = fm.Addendum
	}
	if fm.Tags != nil {
		doc.tags = fm.Tags
	}
	if !doc.draft {
		doc.draft = fm.Draft
	}
	if fm.Slug != "" {
		doc.slug = fm.Slug
	}
	if fm.Layout != "" {
		doc.layout = fm.Layout
	}
	if fm.User != nil {
		doc.user = fm.User
	}
	return nil
}

// frontMatter returns docment template data including merged configuration variables.
func (doc *document) frontMatter() templateData {
	data := templateData{}
	data["title"] = doc.title
	data["author"] = nz(doc.author)
	data["templates"] = nz(doc.templates)
	data["shortdate"] = doc.date.In(doc.conf.timezone).Format(doc.conf.shortdate)
	data["mediumdate"] = doc.date.In(doc.conf.timezone).Format(doc.conf.mediumdate)
	data["longdate"] = doc.date.In(doc.conf.timezone).Format(doc.conf.longdate)
	data["date"] = data["mediumdate"] // Alias.
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
	data["urlprefix"] = doc.conf.urlprefix
	user := doc.conf.user
	for k, v := range doc.user {
		user[k] = v
	}
	data["user"] = user
	// Process addendum and description as a text templates before rendering to HTML.
	addendum := doc.addendum
	description := doc.description
	if isTemplate(doc.contentPath, doc.templates) {
		addendum, _ = doc.proj.textTemplates.renderText("documentAddendum", addendum, data)
		description, _ = doc.proj.textTemplates.renderText("documentDescription", description, data)
	}
	data["addendum"] = doc.render(addendum)
	data["description"] = doc.render(description)
	return data
}

// Return front matter as YAML formatted string.
func (doc *document) String() (result string) {
	d, _ := yaml.Marshal(doc.frontMatter())
	return string(d)
}

// Render document markup to HTML.
func (doc *document) render(text string) template.HTML {
	// Render document.
	var html string
	switch filepath.Ext(doc.contentPath) {
	case ".md":
		html = string(blackfriday.Run([]byte(text)))
	case ".rmu":
		conf, err := readFile(filepath.Join(doc.proj.contentDir, "config.rmu"))
		if err == nil {
			text = conf + "\n\n" + text
		}
		html = rimu.Render(text, rimu.RenderOptions{Reset: true})
	}
	return template.HTML(html)
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

// updateFrom copies fields set by newDocument from src document.
func (doc *document) updateFrom(src document) {
	doc.proj = src.proj
	doc.conf = src.conf
	doc.contentPath = src.contentPath
	doc.buildPath = src.buildPath
	doc.templatePath = src.templatePath
	doc.header = src.header
	doc.content = src.content
	doc.modified = src.modified
	// doc.primaryIndex = src.primaryIndex
	// doc.prev = src.prev
	// doc.next = src.next
	doc.title = src.title
	doc.date = src.date
	doc.author = src.author
	doc.templates = src.templates
	doc.description = src.description
	doc.addendum = src.addendum
	doc.url = src.url
	doc.tags = src.tags
	doc.draft = src.draft
	doc.slug = src.slug
	doc.layout = src.layout
	doc.user = src.user
}
