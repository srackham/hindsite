package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type index struct {
	conf        config               // Merged configuration for this index.
	contentDir  string               // The directory that contains the indexed documents.
	templateDir string               // The directory that contains the index templates.
	indexDir    string               // The build directory that the index pages are written to.
	url         string               // URL of index directory.
	docs        documents            // Parsed documents belonging to index.
	tagdocs     map[string]documents // Partitions index documents by tag.
	tagfiles    map[string]string    // Slugified tag file names.
	pages       []page               // Paginated docs.
}

type indexes []index

type page struct {
	url  string
	next string // URL.
	prev string // URL.
	docs documents
}

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
func (idxs *indexes) init(proj *project) error {
	err := filepath.Walk(proj.templateDir, func(f string, info os.FileInfo, err error) error {
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
				p, err := filepath.Rel(proj.templateDir, f)
				if err != nil {
					return err
				}
				idx.contentDir = filepath.Join(proj.contentDir, p)
				if !dirExists(idx.contentDir) {
					return fmt.Errorf("missing indexed content directory: %s", idx.contentDir)
				}
				idx.indexDir = filepath.Join(proj.indexDir, p)
				p, err = filepath.Rel(proj.buildDir, idx.indexDir)
				if err != nil {
					return err
				}
				idx.conf = proj.configFor(idx.contentDir, idx.templateDir)
				idx.url = path.Join(idx.conf.urlprefix, filepath.ToSlash(p))
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

// Build all indexes. modified is the date of the most recently modified
// configuration or template file.
func (idxs indexes) build(proj *project, modified time.Time) error {
	// Sort index documents then assign previous and next documents according to
	// the primary index ordering.
	// NOTE: Document prev/next corresponds to the primary index.
	for _, idx1 := range idxs {
		primary := true
		for _, idx2 := range idxs {
			if pathIsInDir(idx1.templateDir, idx2.templateDir) {
				primary = false
			}
		}
		idx1.docs.sortByDate()
		if primary {
			idx1.docs.setPrevNext()
		}
	}
	// Build all indexes.
	for _, idx := range idxs {
		if err := idx.build(proj, modified); err != nil {
			return err
		}
	}
	return nil
}

func (idx index) build(proj *project, modified time.Time) error {
	tmpls := &proj.tmpls // Lexical shortcut.
	tagsTemplate := tmpls.name(idx.templateDir, "tags.html")
	tagTemplate := tmpls.name(idx.templateDir, "tag.html")
	if tmpls.contains(tagsTemplate) || tmpls.contains(tagTemplate) {
		if !tmpls.contains(tagsTemplate) {
			return fmt.Errorf("missing tags template: %s", tagsTemplate)
		}
		if !tmpls.contains(tagTemplate) {
			return fmt.Errorf("missing tag template: %s", tagTemplate)
		}
		outfile := filepath.Join(idx.indexDir, "tags.html")
		if rebuild(outfile, modified, idx.docs...) {
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
			// Render tags index.
			err := tmpls.render(tagsTemplate, idx.tagsData(), outfile)
			proj.println("write index: " + outfile)
			if err != nil {
				return err
			}
			// Render per-tag indexes.
			for tag := range idx.tagdocs {
				outfile := filepath.Join(idx.indexDir, "tags", idx.tagfiles[tag])
				if rebuild(outfile, modified, idx.tagdocs[tag]...) {
					data := idx.tagdocs[tag].frontMatter()
					data["tag"] = tag
					err := tmpls.render(tagTemplate, data, outfile)
					proj.println("write index: " + outfile)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	// Render all index.
	tmpl := tmpls.name(idx.templateDir, "all.html")
	var outfile string
	if tmpls.contains(tmpl) {
		outfile = filepath.Join(idx.indexDir, "all.html")
		docs := idx.docs
		if rebuild(outfile, modified, docs...) {
			err := tmpls.render(tmpl, docs.frontMatter(), outfile)
			proj.println("write index: " + outfile)
			if err != nil {
				return err
			}
		}
	}
	// Render recent index.
	tmpl = tmpls.name(idx.templateDir, "recent.html")
	if tmpls.contains(tmpl) {
		outfile = filepath.Join(idx.indexDir, "recent.html")
		docs := idx.docs.first(idx.conf.recent)
		if rebuild(outfile, modified, docs...) {
			err := tmpls.render(tmpl, docs.frontMatter(), outfile)
			proj.println("write index: " + outfile)
			if err != nil {
				return err
			}
		}
	}
	return nil
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
