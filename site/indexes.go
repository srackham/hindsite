package site

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/srackham/hindsite/v2/fsx"
)

type index struct {
	site        *site                    // Context.
	conf        config                   // Merged configuration for this index.
	contentDir  string                   // The directory that contains the indexed documents.
	templateDir string                   // The directory that contains the index templates.
	indexDir    string                   // The build directory that the index pages are written to.
	url         string                   // Index directory relative URL (sans urlprefix).
	docs        documentsList            // Parsed documents belonging to index.
	tagDocs     map[string]documentsList // Partitions indexed documents by tag.
	slugs       map[string]string        // Slugified tags.
	isPrimary   bool                     // True if this is a primary index.
}

type indexes []*index

// page represents a document index page.
type page struct {
	number int    // 1...
	file   string // File name.
	url    string
	next   *page
	prev   *page
	first  *page
	last   *page
	docs   documentsList
}

func newIndex(site *site) index {
	idx := index{}
	idx.site = site
	return idx
}

// Search templateDir directory for indexed directories and add them to indexes.
func newIndexes(site *site) (indexes, error) {
	idxs := indexes{}
	err := filepath.Walk(site.templateDir, func(f string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && fsx.FileExists(filepath.Join(f, "docs.html")) {
			idx := newIndex(site)
			idx.templateDir = f
			p, err := filepath.Rel(site.templateDir, f)
			if err != nil {
				return err
			}
			idx.contentDir = filepath.Join(site.contentDir, p)
			if !fsx.DirExists(idx.contentDir) {
				return fmt.Errorf("missing indexed content directory: \"%s\"", idx.contentDir)
			}
			idx.indexDir = filepath.Join(site.indexDir, p)
			p, err = filepath.Rel(site.buildDir, idx.indexDir)
			if err != nil {
				return err
			}
			idx.conf = site.configFor(idx.contentDir)
			idx.url = filepath.ToSlash(p)
			idxs = append(idxs, &idx)
		}
		return nil
	})
	// Assign primary indexes.
	for i, idx1 := range idxs {
		idxs[i].isPrimary = true
		for _, idx2 := range idxs {
			if fsx.PathIsInDir(idx1.templateDir, idx2.templateDir) && idx1.templateDir != idx2.templateDir {
				// idx1 is child of idx2.
				idxs[i].isPrimary = false
			}
		}
	}
	return idxs, err
}

// addDocument add a document to all indexes that it belongs to, it also assigns
// the document its primary index.
func (idxs indexes) addDocument(doc *document) {
	for i, idx := range idxs {
		if fsx.PathIsInDir(doc.templatePath, idx.templateDir) {
			idxs[i].docs = append(idx.docs, doc)
			if idx.isPrimary {
				doc.primaryIndex = idxs[i]
			}
		}
	}
}

// build builds all indexes. modified is the date of the most recently modified
// configuration or template file. If any document in the index has been
// modified since the index was last built then the index must be completely
// rebuild.
func (idxs indexes) build() error {
	for _, idx := range idxs {
		if err := idx.build(nil); err != nil {
			return err
		}
	}
	return nil
}

// build builds document and tag index pages.
// If doc is nil then rebuild entire index.
// If doc is not nil then only those document index pages containing doc are rendered.
func (idx *index) build(doc *document) error {
	if doc == nil {
		idx.site.logVerbose("build index: \"%s\"", idx.indexDir)
	}
	tmpls := &idx.site.htmlTemplates // Lexical shortcut.
	// renderPages renders paginated document pages with named template.
	// Additional template data is included.
	renderPages := func(pgs []page, tmpl string, data templateData) error {
		count := 0
		for _, pg := range pgs {
			count += len(pg.docs)
		}
		for _, pg := range pgs {
			if doc != nil && !pg.docs.contains(doc) {
				continue
			}
			fm := pg.docs.frontMatter()
			fm["count"] = strconv.Itoa(count)
			fm["page"] = pg.frontMatter()
			mergeMap(fm, data)
			// Merge applicable configuration variables.
			fm["urlprefix"] = idx.conf.urlprefix
			fm["user"] = idx.conf.user
			if doc != nil {
				idx.site.logVerbose("write index: \"%s\"", pg.file)
			} else {
				idx.site.logVerbose2("write index: \"%s\"", pg.file)
			}
			html, err := tmpls.render(tmpl, fm)
			if err != nil {
				return err
			}
			html = idx.site.injectUrlprefix(html)
			if err = fsx.WritePath(pg.file, html); err != nil {
				return err
			}
		}
		return nil
	}
	if doc == nil {
		// Sort index documents then assign document prev/next according to the
		// primary index ordering. Index document ordering ensures subsequent
		// derived document tag indexes are also ordered.
		idx.docs.sortByDate()
		if idx.isPrimary {
			idx.docs.setPrevNext()
		}
	}
	docsTemplate := tmpls.name(idx.templateDir, "docs.html")
	tagsTemplate := tmpls.name(idx.templateDir, "tags.html")
	if tmpls.contains(tagsTemplate) {
		// Build idx.tagDocs[].
		idx.tagDocs = map[string]documentsList{}
		for _, doc := range idx.docs {
			for _, tag := range doc.tags {
				idx.tagDocs[tag] = append(idx.tagDocs[tag], doc)
			}
		}
		// Build index tag slugs.
		idx.slugs = map[string]string{}
		slugs := []string{}
		for _, tag := range sortedKeys(idx.tagDocs) {
			slug := slugify(tag, slugs)
			slugs = append(slugs, slug)
			idx.slugs[tag] = slug
		}
		if doc == nil {
			// Render tags index.
			data := idx.tagsData()
			// Merge applicable configuration variables.
			data["urlprefix"] = idx.conf.urlprefix
			data["user"] = idx.conf.user
			outfile := filepath.Join(idx.indexDir, "tags.html")
			html, err := tmpls.render(tagsTemplate, data)
			if err != nil {
				return err
			}
			html = idx.site.injectUrlprefix(html)
			idx.site.logVerbose2("write index: \"%s\"", outfile)
			if err = fsx.WritePath(outfile, html); err != nil {
				return err
			}
		}
		// Render per-tag document index pages.
		for _, tag := range sortedKeys(idx.tagDocs) {
			pgs := idx.paginate(idx.tagDocs[tag], filepath.Join("tags", idx.slugs[tag]+"-%d.html"))
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
	for tag, docs := range idx.tagDocs {
		data := map[string]string{
			"tag":   tag,
			"url":   rootRelURL(idx.url, "tags", idx.slugs[tag]+"-1.html"),
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
func (idx *index) paginate(docs documentsList, filename string) []page {
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
		pg.url = rootRelURL(idx.url, filepath.ToSlash(f))
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
