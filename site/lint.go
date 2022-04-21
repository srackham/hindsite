package site

import (
	urlpkg "net/url"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/srackham/hindsite/fsx"
	"github.com/srackham/hindsite/set"
	"github.com/srackham/hindsite/slice"
)

// documentLink represents an intra-site document link.
type documentLink struct {
	target string // Target build fle path
	anchor string // URL fragment
}

// parseLink computes the URL's target build file path and URL anchor fragment from
// `url` and returns them in `link`.
// `offsite` is returned set to `true` if `url` refers to an external resource.
func (doc *document) parseLink(url string) (link documentLink, offsite bool) {
	// Split into URL and anchor.
	s := strings.Split(url, "#")
	url = s[0]
	if len(s) > 1 {
		link.anchor = s[1]
	}
	re := regexp.MustCompile(`(?i)^(?:` + regexp.QuoteMeta(doc.conf.urlprefix) + `)?/(.+)$`)
	matches := re.FindStringSubmatch(url)
	if matches != nil { // Matches absolute or root-relative URL.
		link.target = filepath.Join(doc.site.buildDir, matches[1])
		if fsx.DirExists(link.target) {
			link.target = filepath.Join(link.target, "index.html")
		}
	} else { // Matches relative URLs.
		re := regexp.MustCompile(`(?i)^([\w][\w./-]*)$`)
		matches := re.FindStringSubmatch(url)
		if matches != nil {
			link.target = filepath.Join(filepath.Dir(doc.buildPath), matches[1])
		} else {
			offsite = true
		}
	}
	return
}

// parseHTML scans the document's `html` saving id attributes to `doc.ids` and href and src
// URL attributes to `doc.urls` which are used for subsequent site link validation.
func (doc *document) parseHTML(html string) {
	// Scan HTML element id attributes.
	doc.ids = slice.Slice[string]{}
	pat := regexp.MustCompile(`(?i)id="(.+?)"`)
	matches := pat.FindAllStringSubmatch(html, -1)
	for _, match := range matches {
		doc.ids = append(doc.ids, match[1])
	}
	// Scan HTML for href and src URL attributes.
	doc.urls = slice.Slice[string]{}
	pat = regexp.MustCompile(`(?i)(?:href|src)="(.+?)"`)
	matches = pat.FindAllStringSubmatch(html, -1)
	for _, match := range matches {
		doc.urls = append(doc.urls, match[1])
	}

}

// lintChecks checks that all document intra-site URLs point to valid target
// files and valid HTML id attributes.
func (site *site) lintChecks() (errCount int) {
	for _, doc := range site.docs.byContentPath {
		site.verbose("lint document: %s", doc.contentPath)
		// Check for llicit or duplicate ids.
		ids := set.New(doc.ids...)
		for id, _ := range ids {
			re := regexp.MustCompile(`^[A-Za-z][\w:.-]*$`) // https://www.w3.org/TR/html4/types.html
			if !re.MatchString(id) {
				doc.site.logerror("%s: contains illicit element id: \"%s\"", doc.contentPath, id)
				errCount++
			}
			if ids.Count(id) > 1 {
				doc.site.logerror("%s: contains duplicate element id: \"%s\"", doc.contentPath, id)
				errCount++
			}
		}
		// Iterate the document's href/src attribute URLs.
		for _, url := range doc.urls {
			_, err := urlpkg.Parse(url)
			if err != nil {
				doc.site.logerror("%s: contains illicit URL: \"%s\"", doc.contentPath, url)
				errCount++
				continue
			}
			if strings.HasPrefix(url, "#") { // Intra-document URL fragment.
				if !doc.ids.Has(url[1:]) {
					doc.site.logerror("%s: contains link to missing anchor: \"%s\"", doc.contentPath, url)
					errCount++
					continue
				}
			} else {
				link, offsite := doc.parseLink(url)
				if offsite {
					site.verbose2("lint: %s: skipped offsite link: \"%s\"", doc.contentPath, url)
					continue
				}
				// Check the target URL file exists.
				if !fsx.FileExists(link.target) {
					doc.site.logerror("%s: contains link to missing file: \"%s\"", doc.contentPath, strings.TrimPrefix(link.target, site.buildDir+string(filepath.Separator)))
					errCount++
					continue
				}
				// Check the URL anchor has a matching HTML id attribute in the target document.
				if link.anchor != "" {
					target, ok := site.docs.byBuildPath[link.target]
					if !ok || !target.ids.Has(link.anchor) {
						doc.site.logerror("%s: contains link to missing anchor: \"%s\"", doc.contentPath, strings.TrimPrefix(url, site.confs[0].urlprefix+string(filepath.Separator)))
						errCount++
						continue
					}
				}
				site.verbose2("lint: %s: validated link: \"%s\"", doc.contentPath, url)
			}
		}
	}
	return
}
