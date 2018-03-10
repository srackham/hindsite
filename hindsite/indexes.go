package main

import (
	"bytes"
	"html/template"
	"os"
	"path/filepath"
	"sort"
)

var templateFileNames = stringlist{"all.html", "recent.html", "tags.html", "tag.html"}

type index struct {
	templateDir string                 // The template directory that contains the index templates.
	buildDir    string                 // The build directory that the index pages are written.
	url         string                 // URL of index directory.
	docs        []*document            // Parsed documents belonging to index.
	tagdocs     map[string][]*document // Groups indexed documents by tag.
}

type indexes []index

// Search templates directory for indexes and add them to indexes.
func (idxs *indexes) init(templateDir string) error {
	err := filepath.Walk(templateDir, func(f string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			files, err := filepath.Glob(filepath.Join(f, "*.html"))
			if err != nil {
				return err
			}
			found := false
			for _, fn := range files {
				if templateFileNames.Contains(fn) {
					found = true
					break
				}
			}
			if found {
				idx := index{}
				idx.templateDir = f
				p, err := filepath.Rel(templateDir, f)
				if err != nil {
					return err
				}
				idx.buildDir = filepath.Join(Cmd.buildDir, p)
				idx.url = Config.urlprefix + "/" + filepath.ToSlash(p)
				*idxs = append(*idxs, idx)
			}
		}
		return nil
	})
	return err
}

// Add document to all indexes that it belongs to.
func (idxs indexes) addDocument(doc *document) {
	for i, idx := range idxs {
		if pathIsInDir(doc.templatepath, idx.templateDir) {
			idxs[i].docs = append(idx.docs, doc)
		}
	}
}

// Build all indexes.
func (idxs indexes) build() error {
	for _, idx := range idxs {
		if err := idx.build(); err != nil {
			return err
		}
	}
	return nil
}

func (idx index) build() error {
	render := func(tmplfile string, data templateData, outfile string) error {
		tmpl, err := template.ParseFiles(tmplfile)
		if err != nil {
			return err
		}
		buf := bytes.NewBufferString("")
		if err := tmpl.Execute(buf, data); err != nil {
			return err
		}
		html := buf.String()
		if err := mkMissingDir(filepath.Dir(outfile)); err != nil {
			return err
		}
		return writeFile(outfile, html)
	}
	tmplfile := filepath.Join(idx.templateDir, "all.html")
	var outfile string
	if fileExists(tmplfile) {
		outfile = filepath.Join(idx.buildDir, "all.html")
		err := render(tmplfile, docsByDate(idx.docs, -1), outfile)
		if err != nil {
			return err
		}
	}
	tmplfile = filepath.Join(idx.templateDir, "recent.html")
	if fileExists(tmplfile) {
		outfile = filepath.Join(idx.buildDir, "recent.html")
		err := render(tmplfile, docsByDate(idx.docs, 5), outfile)
		if err != nil {
			return err
		}
	}
	tagsTemplate := filepath.Join(idx.templateDir, "tags.html")
	tagTemplate := filepath.Join(idx.templateDir, "tag.html")
	if fileExists(tagsTemplate) || fileExists(tagTemplate) {
		// Build idx.tagdocs[].
		for _, doc := range idx.docs {
			for _, tag := range doc.tags {
				idx.tagdocs[tag] = append(idx.tagdocs[tag], doc)
			}
		}
		if fileExists(tagsTemplate) {
			outfile = filepath.Join(idx.buildDir, "tags.html")
			err := render(tagsTemplate, idx.tagsData(), outfile)
			if err != nil {
				return err
			}
		}
		if fileExists(tagTemplate) {
			for tag := range idx.tagdocs {
				err := render(tagsTemplate, docsByDate(idx.tagdocs[tag], -1),
					filepath.Join(idx.buildDir, "tags", slugify(tag, nil)+".html"))
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// Dcouments template data sorted by date descending.
// n >= 0 limits the maxiumum number returned.
func docsByDate(docs []*document, n int) templateData {
	// Sort index documents by decending date.
	sort.Slice(docs, func(i, j int) bool {
		return !docs[i].date.Before(docs[j].date)
	})
	// Build list of document template data.
	data := []templateData{}
	for i, doc := range docs {
		if n >= 0 && i < n {
			break
		}
		data = append(data, doc.frontMatter())
	}
	return templateData{"docs": data}
}

type tagData struct {
	tag string
	url string
}

func (idx index) tagsData() templateData {
	tags := []tagData{}
	for tag := range idx.tagdocs {
		data := tagData{
			tag,
			idx.url + "/tags/" + slugify(tag, nil) + ".html",
		}
		tags = append(tags, data)
	}
	sort.Slice(tags, func(i, j int) bool {
		return tags[i].tag < tags[j].tag
	})
	return templateData{"tags": tags}
}
