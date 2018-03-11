package main

import (
	"bufio"
	"fmt"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/srackham/go-rimu/rimu"
	blackfriday "gopkg.in/russross/blackfriday.v2"
	yaml "gopkg.in/yaml.v2"
)

// document TODO
type document struct {
	contentpath  string
	buildpath    string
	layoutpath   string
	templatepath string // Virtual path used to find document related templates.
	content      string // Markup text (without front matter header).
	rootIndex    *index // Top-level document index (nil if document is not indexed).
	// Front matter.
	title    string
	date     time.Time
	synopsis string
	addendum string
	url      string // Absolute or root-relative URL.
	tags     []string
	draft    bool
	slug     string
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
	doc.templatepath = filepath.Join(Cmd.templateDir, p)
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
	// If necessary change output file names to match document slug variable.
	if doc.slug != "" {
		doc.buildpath = filepath.Join(filepath.Dir(doc.buildpath), doc.slug+".html")
		doc.url = path.Join(path.Dir(doc.url), doc.slug+".html")
	}
	// Find document layout template.
	layouts, err := filesInPath(filepath.Dir(doc.templatepath), Cmd.templateDir, []string{"layout.html"}, 1)
	if err != nil {
		return err
	}
	if len(layouts) == 0 {
		return fmt.Errorf("missing layout.html template for: %s", doc.contentpath)
	}
	doc.layoutpath = layouts[0]
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
		Slug        string
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
	if fm.Slug != "" {
		doc.slug = fm.Slug
	}
	return nil
}

func (doc *document) frontMatter() (data templateData) {
	data = templateData{}
	data["title"] = doc.title
	data["date"] = doc.date.Format("02-Jan-2006")
	data["synopsis"] = doc.synopsis
	data["addendum"] = doc.addendum
	data["slug"] = doc.slug
	data["url"] = doc.url
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
