package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type index struct {
	proj        *project             // Context.
	conf        config               // Merged configuration for this index.
	contentDir  string               // The directory that contains the indexed documents.
	templateDir string               // The directory that contains the index templates.
	indexDir    string               // The build directory that the index pages are written to.
	url         string               // Synthesised absolute or root-relative index directory URL.
	docs        documents            // Parsed documents belonging to index.
	tagdocs     map[string]documents // Partitions indexed documents by tag.
	slugs       map[string]string    // Slugified tags.
	primary     bool                 // True if this is a primary index.
}

type indexes []index

// page represents a document index page.
type page struct {
	number int    // 1...
	file   string // File name.
	url    string
	next   *page
	prev   *page
	first  *page
	last   *page
	docs   documents
}

func newIndex(proj *project) index {
	idx := index{}
	idx.proj = proj
	idx.tagdocs = map[string]documents{}
	idx.slugs = map[string]string{}
	return idx
}

// Search templateDir directory for indexed directories and add them to indexes.
func newIndexes(proj *project) (indexes, error) {
	idxs := indexes{}
	err := filepath.Walk(proj.templateDir, func(f string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && fileExists(filepath.Join(f, "docs.html")) {
			idx := newIndex(proj)
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
			idx.conf = proj.configFor(idx.templateDir)
			idx.url = path.Join("/", idx.conf.urlprefix, filepath.ToSlash(p))
			idxs = append(idxs, idx)
		}
		return nil
	})
	// Assign primary indexes.
	for i, idx1 := range idxs {
		idxs[i].primary = true
		for _, idx2 := range idxs {
			if pathIsInDir(idx1.templateDir, idx2.templateDir) {
				idxs[i].primary = false
			}
		}
	}
	return idxs, err
}

// addDocument add a document to all indexes that it belongs to, it also assigns
// the document its primary index.
func (idxs indexes) addDocument(doc *document) {
	for i, idx := range idxs {
		if pathIsInDir(doc.templatepath, idx.templateDir) {
			idxs[i].docs = append(idx.docs, doc)
			if idx.primary {
				doc.primaryIndex = &idxs[i]
			}
		}
	}
}

// build builds all indexes. modified is the date of the most recently modified
// configuration or template file. If any document in the index has been
// modified since the index was last built then the index must be completely
// rebuild.
func (idxs indexes) build(modified time.Time) error {
	for _, idx := range idxs {
		target := filepath.Join(idx.indexDir, "docs-1.html")
		if rebuild(target, modified, idx.docs...) {
			if err := idx.build(); err != nil {
				return err
			}
		}
	}
	return nil
}

// build builds document and tag index pages.
func (idx index) build() error {
	tmpls := &idx.proj.tmpls // Lexical shortcut.
	// renderPages renders paginated document pages with named template.
	// Additional template data is included.
	renderPages := func(pgs []page, tmpl string, data templateData) error {
		count := 0
		for _, pg := range pgs {
			count += len(pg.docs)
		}
		for _, pg := range pgs {
			fm := pg.docs.frontMatter()
			fm["count"] = strconv.Itoa(count)
			fm["page"] = pg.frontMatter()
			fm.merge(data)
			// Merge applicable (lower precedence) configuration variables.
			data["urlprefix"] = idx.conf.urlprefix
			data["user"] = idx.conf.user
			err := tmpls.render(tmpl, fm, pg.file)
			idx.proj.verbose("write index: " + pg.file)
			if err != nil {
				return err
			}
		}
		return nil
	}
	docsTemplate := tmpls.name(idx.templateDir, "docs.html")
	tagsTemplate := tmpls.name(idx.templateDir, "tags.html")
	if tmpls.contains(tagsTemplate) {
		// Build idx.tagdocs[].
		for _, doc := range idx.docs {
			for _, tag := range doc.tags {
				idx.tagdocs[tag] = append(idx.tagdocs[tag], doc)
			}
		}
		// Build index tag slugs.
		slugs := []string{}
		for tag := range idx.tagdocs {
			slug := slugify(tag, slugs)
			slugs = append(slugs, slug)
			idx.slugs[tag] = slug
		}
		// Render tags index.
		data := idx.tagsData()
		// Merge applicable configuration variables.
		data["urlprefix"] = idx.conf.urlprefix
		data["user"] = idx.conf.user
		outfile := filepath.Join(idx.indexDir, "tags.html")
		err := tmpls.render(tagsTemplate, data, outfile)
		idx.proj.verbose("write index: " + outfile)
		if err != nil {
			return err
		}
		// Render per-tag document index pages.
		for tag := range idx.tagdocs {
			pgs := idx.paginate(idx.tagdocs[tag], filepath.Join("tags", idx.slugs[tag]+"-%d.html"))
			if err := renderPages(pgs, docsTemplate, templateData{"tag": tag}); err != nil {
				return err
			}
		}
	}
	// Render document index pages.
	pgs := idx.paginate(idx.docs, "docs-%d.html")
	return renderPages(pgs, docsTemplate, templateData{})
}

func (idx index) tagsData() templateData {
	tags := []map[string]string{} // An array of "tag", "url" key value maps.
	for tag, docs := range idx.tagdocs {
		data := map[string]string{
			"tag":   tag,
			"url":   path.Join(idx.url, "tags", idx.slugs[tag]+"-1.html"),
			"count": strconv.Itoa(len(docs)),
		}
		tags = append(tags, data)
	}
	sort.Slice(tags, func(i, j int) bool {
		return strings.ToLower(tags[i]["tag"]) < strings.ToLower(tags[j]["tag"])
	})
	return templateData{"tags": tags}
}

// Synthesize index pages.
func (idx *index) paginate(docs documents, filename string) []page {
	pgs := []page{}
	pagesize := idx.conf.paginate
	var pagecount int
	if pagesize <= 0 {
		pagecount = 1
	} else {
		pagecount = (len(docs)-1)/pagesize + 1 // Total number of pages.
	}
	for pageno := 1; pageno <= pagecount; pageno++ {
		pg := page{number: pageno}
		i := (pageno - 1) * pagesize
		if pageno == pagecount {
			pg.docs = docs[i:]
		} else {
			pg.docs = docs[i : i+pagesize]
		}
		f := fmt.Sprintf(filename, pg.number)
		pg.file = filepath.Join(idx.indexDir, f)
		pg.url = path.Join(idx.url, filepath.ToSlash(f))
		pgs = append(pgs, pg)
	}
	for i := range pgs {
		if i != 0 {
			pgs[i].prev = &pgs[i-1]
		}
		if i < len(pgs)-1 {
			pgs[i].next = &pgs[i+1]
		}
		pgs[i].first = &pgs[0]
		pgs[i].last = &pgs[len(pgs)-1]
	}
	return pgs
}

func (pg page) frontMatter() templateData {
	dataFor := func(pg *page) templateData {
		data := templateData{}
		if pg != nil {
			data["number"] = strconv.Itoa(pg.number)
			data["url"] = pg.url
		}
		return data
	}
	data := dataFor(&pg)
	data["prev"] = dataFor(pg.prev)
	data["next"] = dataFor(pg.next)
	data["first"] = dataFor(pg.first)
	data["last"] = dataFor(pg.last)
	return data
}
