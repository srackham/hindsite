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
	number int    // 1...
	file   string // File name.
	url    string
	next   *page
	prev   *page
	docs   documents
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
	// NOTE:
	// - Document prev/next corresponds to the primary index.
	// - Index document ordering ensures subsequent derived document tag indexes
	//   are also ordered.
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
	// Render document index pages.
	idx.paginate()
	tmpl := tmpls.name(idx.templateDir, "all.html")
	for _, pg := range idx.pages {
		if rebuild(pg.file, modified, pg.docs...) {
			fm := pg.docs.frontMatter()
			fm["page"] = pg.frontMatter()
			err := tmpls.render(tmpl, fm, pg.file)
			proj.println("write index: " + pg.file)
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

// Synthesize index pages.
func (idx *index) paginate() {
	pgs := []page{}
	pagesize := idx.conf.paginate
	var pagecount int
	if pagesize <= 0 {
		pagecount = 1
	} else {
		pagecount = (len(idx.docs)-1)/pagesize + 1 // Total number of pages.
	}
	for pageno := 1; pageno <= pagecount; pageno++ {
		pg := page{number: pageno}
		i := (pageno - 1) * pagesize
		if pageno == pagecount {
			pg.docs = idx.docs[i:]
		} else {
			pg.docs = idx.docs[i : i+pagesize]
		}
		f := fmt.Sprintf("all-%d.html", pg.number)
		pg.file = filepath.Join(idx.indexDir, f)
		pg.url = path.Join(idx.url, f)
		pgs = append(pgs, pg)
	}
	for i := range pgs {
		if i != 0 {
			pgs[i].prev = &pgs[i-1]
		}
		if i < len(pgs)-1 {
			pgs[i].next = &pgs[i+1]
		}
	}
	idx.pages = pgs
}

func (pg page) frontMatter() (data templateData) {
	data = templateData{}
	data["number"] = pg.number
	data["url"] = pg.url
	prev := templateData{}
	if pg.prev != nil {
		prev["number"] = pg.prev.number
		prev["url"] = pg.prev.url
	}
	data["prev"] = prev
	next := templateData{}
	if pg.next != nil {
		next["number"] = pg.next.number
		next["url"] = pg.next.url
	}
	data["next"] = next
	return data
}
