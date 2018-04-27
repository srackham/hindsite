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
	content      string    // Markup text (without front matter header).
	modtime      time.Time // Document source file modified timestamp.
	// TODO: Is primaryIndex field necessary, it's only used on one place? Factor it out?
	primaryIndex *index    // Top-level document index (nil if document is not indexed).
	prev         *document // Previous document in primary index.
	next         *document // Next document in primary index.
	// Front matter.
	title       string
	date        time.Time
	author      *string
	templates   *string
	description string
	url         string // Synthesised document URL.
	tags        []string
	draft       bool
	permalink   string // URL template.
	slug        string
	layout      string            // Document template name.
	user        map[string]string // User defined configuration key/values.
}

// Parse document content and front matter.
func newDocument(contentfile string, proj *project) (document, error) {
	if !pathIsInDir(contentfile, proj.contentDir) {
		panic("document is outside content directory: " + contentfile)
	}
	if !fileExists(contentfile) {
		panic("missing document: " + contentfile)
	}
	doc := document{}
	doc.proj = proj
	info, err := os.Stat(contentfile)
	if err != nil {
		return doc, err
	}
	doc.modtime = info.ModTime()
	doc.contentPath = contentfile
	doc.conf = proj.configFor(doc.contentPath)
	// Extract title and date from file name.
	doc.title = fileName(contentfile)
	if regexp.MustCompile(`^\d\d\d\d-\d\d-\d\d-.+`).MatchString(doc.title) {
		d, err := parseDate(doc.title[0:10], doc.conf.timezone)
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
	doc.permalink = doc.conf.permalink // Default permalink.
	doc.content, err = readFile(doc.contentPath)
	if err != nil {
		return doc, err
	}
	if err := doc.extractFrontMatter(); err != nil {
		return doc, err
	}
	// Synthesize build path and URL according to content path, permalink and slug values.
	rel, _ := filepath.Rel(proj.contentDir, doc.contentPath)
	doc.templatePath = filepath.Join(proj.templateDir, rel)
	f := filepath.Base(rel)
	switch filepath.Ext(f) {
	case ".md", ".rmu":
		f = replaceExt(f, ".html")
	}
	if doc.slug != "" {
		f = doc.slug + filepath.Ext(f)
	}
	if doc.permalink != "" {
		link := doc.permalink
		link = strings.Replace(link, "%y", doc.date.Format("2006"), -1)
		link = strings.Replace(link, "%m", doc.date.Format("01"), -1)
		link = strings.Replace(link, "%d", doc.date.Format("02"), -1)
		link = strings.Replace(link, "%f", f, -1)
		link = strings.Replace(link, "%p", fileName(f), -1)
		link = strings.TrimPrefix(link, "/")
		if strings.HasSuffix(link, "/") {
			// "Pretty" URLs.
			doc.buildPath = filepath.Join(proj.buildDir, filepath.FromSlash(link), "index.html")
			doc.url = path.Join("/", doc.conf.urlprefix, link) + "/"
		} else {
			doc.buildPath = filepath.Join(proj.buildDir, filepath.FromSlash(link))
			doc.url = path.Join("/", doc.conf.urlprefix, link)
		}
	} else {
		doc.buildPath = filepath.Join(proj.buildDir, filepath.Dir(rel), f)
		doc.url = path.Join("/", doc.conf.urlprefix, path.Dir(filepath.ToSlash(rel)), f)
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
		Author      *string
		Templates   *string
		Tags        []string
		Draft       bool
		Permalink   string
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
		d, err := parseDate(fm.Date, doc.conf.timezone)
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
	if fm.Permalink != "" {
		doc.permalink = fm.Permalink
	}
	if fm.Description != "" {
		doc.description = fm.Description
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

// frontMatter returns document template data including merged configuration variables.
func (doc *document) frontMatter() templateData {
	data := templateData{}
	data["title"] = doc.title
	data["author"] = nz(doc.author)
	data["templates"] = nz(doc.templates) // TODO: Necessary in front matter?
	data["permalink"] = doc.permalink     // TODO: Necessary in front matter?
	data["shortdate"] = doc.date.In(doc.conf.timezone).Format(doc.conf.shortdate)
	data["mediumdate"] = doc.date.In(doc.conf.timezone).Format(doc.conf.mediumdate)
	data["longdate"] = doc.date.In(doc.conf.timezone).Format(doc.conf.longdate)
	data["date"] = doc.date
	data["modtime"] = doc.modtime
	data["slug"] = doc.slug
	data["url"] = doc.url
	tags := []map[string]string{}
	for _, tag := range doc.tags {
		url := ""
		if doc.primaryIndex != nil {
			// TODO: Remove temporary diagnostics (added after mysterious loss of tag URLS 23-Apr-2018).
			if doc.primaryIndex.slugs[tag] == "" {
				panic(fmt.Sprintf("%#v", doc))
			}
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
	// Process description as a text template before rendering to HTML.
	description := doc.description
	if isTemplate(doc.contentPath, doc.templates) {
		description, _ = doc.proj.textTemplates.renderText("documentDescription", description, data)
	}
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

// updateFrom copies fields set by newDocument from src document.
func (doc *document) updateFrom(src document) {
	doc.proj = src.proj
	doc.conf = src.conf
	doc.contentPath = src.contentPath
	doc.buildPath = src.buildPath
	doc.templatePath = src.templatePath
	doc.content = src.content
	doc.modtime = src.modtime
	doc.title = src.title
	doc.date = src.date
	doc.author = src.author
	doc.templates = src.templates
	doc.permalink = src.permalink
	doc.description = src.description
	doc.url = src.url
	doc.tags = src.tags
	doc.draft = src.draft
	doc.slug = src.slug
	doc.layout = src.layout
	doc.user = src.user
}

// isDraft returns true if document is a draft and the drafts option is not true.
func (doc *document) isDraft() bool {
	return doc.draft && !doc.proj.drafts
}

/*
	documentsList stores documentsList in an ordered list.
*/
type documentsList []*document

// Assign previous and next according to the current sort order.
func (docs documentsList) setPrevNext() {
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
func (docs documentsList) sortByDate() {
	// Sort documents by decending date.
	sort.Slice(docs, func(i, j int) bool {
		return !docs[i].date.Before(docs[j].date)
	})
}

// Return slice of first n documents.
func (docs documentsList) first(n int) documentsList {
	result := documentsList{}
	for i, doc := range docs {
		if n >= 0 && i >= n {
			break
		}
		result = append(result, doc)
	}
	return result
}

// delete deletes document from docs and returns resulting slice. Panics if
// document not in slice.
func (docs documentsList) delete(doc *document) documentsList {
	for i, d := range docs {
		if d == doc {
			return append(docs[:i], docs[i+1:]...)
		}
	}
	panic("missing document: " + doc.contentPath)
}

// contains returns true if doc is in docs.
func (docs documentsList) contains(doc *document) bool {
	for _, d := range docs {
		if d == doc {
			return true
		}
	}
	return false
}

// Return documents front matter template data.
func (docs documentsList) frontMatter() templateData {
	data := []templateData{}
	for _, doc := range docs {
		data = append(data, doc.frontMatter())
	}
	return templateData{"docs": data}
}

/*
	documentsLookup implement storage and retrieval of documents by contentPath and
	buildPath
*/
type documentsLookup struct {
	byContentPath map[string]*document // Documents keyed by contentPath.
	byBuildPath   map[string]*document // Documents keyed by buildPath.
}

func newDocumentsLookup() documentsLookup {
	return documentsLookup{map[string]*document{}, map[string]*document{}}
}

func (lookup *documentsLookup) add(doc *document) error {
	d := lookup.byContentPath[doc.contentPath]
	if d != nil {
		panic("duplicate document: " + d.contentPath)
	}
	d = lookup.byBuildPath[doc.buildPath]
	if d != nil {
		return fmt.Errorf("documents have same build path: " + d.contentPath + ": " + doc.contentPath)
	}
	lookup.byBuildPath[doc.buildPath] = doc
	lookup.byContentPath[doc.contentPath] = doc
	return nil
}

func (lookup *documentsLookup) delete(doc *document) {
	lookup.byBuildPath[doc.buildPath] = nil
	lookup.byContentPath[doc.contentPath] = nil
}
