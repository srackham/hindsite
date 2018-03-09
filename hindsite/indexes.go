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
	templateDir   string            // The template directory that contains the index templates.
	templateFiles []string          // List of template file names.
	buildDir      string            // The build directory that the index pages are written.
	url           string            // URL of index directory.
	docs          []*document       // Parsed documents belonging to index.
	tagURLs       map[string]string // Maps tags to tag index URL.
}

type indexes []index

// Search templates directory for indexes and add them to indexes.
func (idxs *indexes) init(templateDir string) error {
	err := filepath.Walk(templateDir, func(f string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			files, err := filepath.Glob(filepath.Join(f, "*"))
			if err != nil {
				return err
			}
			tmplFiles := []string{}
			for _, fn := range files {
				if templateFileNames.Contains(fn) {
					tmplFiles = append(tmplFiles, fn)
				}
			}
			if len(tmplFiles) > 0 {
				idx := index{}
				idx.templateDir = f
				idx.templateFiles = tmplFiles
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
	render := func(tmplfile string, datafunc func(idx index) templateData) error {
		if !templateFileNames.Contains(tmplfile) {
			return nil
		}
		tmpl, err := template.ParseFiles(filepath.Join(idx.templateDir, tmplfile))
		if err != nil {
			return err
		}
		data := datafunc(idx)
		buf := bytes.NewBufferString("")
		tmpl.Execute(buf, data)
		html := buf.String()
		return writeFile(filepath.Join(idx.buildDir, tmplfile), html)
	}
	if err := render("all.html", index.allData); err != nil {
		return err
	}
	if err := render("recent.html", index.recentData); err != nil {
		return err
	}
	return nil
}

// Template data sorted by date descending.
// n >= 0 limits the maxiumum number returned.
func (idx index) dataByDate(n int) templateData {
	// Sort index documents by decending date.
	sort.Slice(idx.docs, func(i, j int) bool {
		return !idx.docs[i].date.Before(idx.docs[j].date)
	})
	// Build list of document template data.
	docs := []templateData{}
	for i, doc := range idx.docs {
		if n >= 0 && i < n {
			break
		}
		docs = append(docs, doc.frontMatter())
	}
	return templateData{"docs": docs}
}

func (idx index) allData() templateData {
	return idx.dataByDate(-1)
}

func (idx index) recentData() templateData {
	return idx.dataByDate(5)
}
