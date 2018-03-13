package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

type index struct {
	templateDir string               // The directory that contains the index templates.
	indexDir    string               // The build directory that the index pages are written to.
	url         string               // URL of index directory.
	docs        documents            // Parsed documents belonging to index.
	tagdocs     map[string]documents // Partitions index documents by tag.
	tagfiles    map[string]string    // Slugified tag file names.
}

type indexes []index

func newIndex() index {
	idx := index{}
	idx.tagdocs = map[string]documents{}
	idx.tagfiles = map[string]string{}
	return idx
}

func isIndexFile(filename string) bool {
	return stringlist{
		"all.html",
		"recent.html",
		"tags.html",
		"tag.html",
	}.Contains(filepath.Base(filename))
}

// Search templateDir directory for indexed directories and add them to indexes.
// indexDir is the directory in the build directory that contains built indexes.
func (idxs *indexes) init(templateDir, buildDir, indexDir string) error {
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
				if isIndexFile(fn) {
					found = true
					break
				}
			}
			if found {
				idx := newIndex()
				idx.templateDir = f
				p, err := filepath.Rel(templateDir, f)
				if err != nil {
					return err
				}
				idx.indexDir = filepath.Join(indexDir, p)
				p, err = filepath.Rel(buildDir, idx.indexDir)
				if err != nil {
					return err
				}
				idx.url = path.Join(Config.urlprefix, filepath.ToSlash(p))
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
			if doc.rootIndex == nil {
				doc.rootIndex = &idxs[i]
			}
		}
	}
}

// Build all indexes.
func (idxs indexes) build(tmpls templates) error {
	for _, idx := range idxs {
		if err := idx.build(tmpls); err != nil {
			return err
		}
	}
	return nil
}

func (idx index) build(tmpls templates) error {
	tagsTemplate := tmpls.name(idx.templateDir, "tags.html")
	tagTemplate := tmpls.name(idx.templateDir, "tag.html")
	if tmpls.contains(tagsTemplate) || tmpls.contains(tagTemplate) {
		if !tmpls.contains(tagsTemplate) {
			return fmt.Errorf("missing tags template: %s", tagsTemplate)
		}
		if !tmpls.contains(tagTemplate) {
			return fmt.Errorf("missing tag template: %s", tagTemplate)
		}
		// TODO: Break if all idx.docs are older than tags.html outfile.
		// Build idx.tagdocs[].
		for _, doc := range idx.docs {
			for _, tag := range doc.tags {
				idx.tagdocs[tag] = append(idx.tagdocs[tag], doc)
			}
		}
		// Build idx.tagfiles[].
		slugs := []string{}
		for tag := range idx.tagdocs {
			slug := slugify(tag, slugs)
			slugs = append(slugs, slug)
			idx.tagfiles[tag] = slug + ".html"
		}
		outfile := filepath.Join(idx.indexDir, "tags.html")
		err := tmpls.render(tagsTemplate, idx.tagsData(), outfile)
		verbose("write index: " + outfile)
		if err != nil {
			return err
		}
		for tag := range idx.tagdocs {
			// TODO: Break if all idx.tagdocs[tag] document s are older than outfile.
			data := docsByDate(idx.tagdocs[tag], -1)
			data["tag"] = tag
			outfile = filepath.Join(idx.indexDir, "tags", idx.tagfiles[tag])
			err := tmpls.render(tagTemplate, data, outfile)
			verbose("write index: " + outfile)
			if err != nil {
				return err
			}
		}
	}
	tmpl := tmpls.name(idx.templateDir, "all.html")
	var outfile string
	if tmpls.contains(tmpl) {
		outfile = filepath.Join(idx.indexDir, "all.html")
		err := tmpls.render(tmpl, docsByDate(idx.docs, -1), outfile)
		verbose("write index: " + outfile)
		if err != nil {
			return err
		}
	}
	tmpl = tmpls.name(idx.templateDir, "recent.html")
	if tmpls.contains(tmpl) {
		outfile = filepath.Join(idx.indexDir, "recent.html")
		err := tmpls.render(tmpl, docsByDate(idx.docs, 5), outfile)
		verbose("write index: " + outfile)
		if err != nil {
			return err
		}
	}
	return nil
}

// Dcouments template data sorted by date descending.
// n >= 0 limits the maxiumum number returned.
func docsByDate(docs documents, n int) templateData {
	// Sort index documents by decending date.
	sort.Slice(docs, func(i, j int) bool {
		return !docs[i].date.Before(docs[j].date)
	})
	// Build list of document template data.
	data := []templateData{}
	for i, doc := range docs {
		if n >= 0 && i >= n {
			break
		}
		data = append(data, doc.frontMatter())
	}
	return templateData{"docs": data}
}

func (idx index) tagsData() templateData {
	tags := []map[string]string{} // An array of "tag", "url" key value maps.
	for tag := range idx.tagdocs {
		data := map[string]string{
			"tag": tag,
			"url": path.Join(idx.url, "tags", idx.tagfiles[tag]),
		}
		tags = append(tags, data)
	}
	sort.Slice(tags, func(i, j int) bool {
		return strings.ToLower(tags[i]["tag"]) < strings.ToLower(tags[j]["tag"])
	})
	return templateData{"tags": tags}
}
