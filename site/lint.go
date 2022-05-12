package site

import (
	urlpkg "net/url"
	"path/filepath"
	"regexp"
	"sort"
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
	re := regexp.MustCompile(`(?i)^(?:` + regexp.QuoteMeta(doc.site.urlprefix()) + `)?/([^/].*)$`) // Extracts root-relative URLs.
	matches := re.FindStringSubmatch(url)
	if matches != nil {
		link.target = filepath.Join(doc.site.buildDir, unescapeURL(matches[1]))
	} else {
		re := regexp.MustCompile(`(?i)^([\w][\w./-]*)$`) // Extracts page-relative URLs.
		matches := re.FindStringSubmatch(url)
		if matches != nil {
			link.target = filepath.Join(filepath.Dir(doc.buildPath), unescapeURL(matches[1]))
		} else {
			offsite = true
		}
	}
	if !offsite && fsx.DirExists(link.target) {
		link.target = filepath.Join(link.target, "index.html")
	}
	return
}

// parseHTML scans the document's `html` saving id attributes to `doc.ids` and href and src
// URL attributes to `doc.urls` which are used for subsequent site link validation.
func (doc *document) parseHTML(html string) {
	// Scan HTML element id attributes.
	doc.ids = slice.Slice[string]{}
	re := regexp.MustCompile(`(?i)id="(.+?)"`)
	matches := re.FindAllStringSubmatch(html, -1)
	for _, match := range matches {
		doc.ids = append(doc.ids, match[1])
	}
	// Scan HTML for href and src URL attributes.
	doc.urls = slice.Slice[string]{}
	re = regexp.MustCompile(`(?i)(?:href|src)="(.+?)"`)
	matches = re.FindAllStringSubmatch(html, -1)
	for _, match := range matches {
		doc.urls = append(doc.urls, match[1])
	}

}

// lintChecks checks that all document intra-site URLs point to valid target
// files and valid HTML id attributes.
func (site *site) lintChecks() (errCount int, warnCount int) {
	for _, k := range sortedKeys(site.docs.byContentPath) {
		doc := site.docs.byContentPath[k]
		site.verbose("lint document: %s", doc.contentPath)
		// Check for illicit or duplicate ids.
		ids := set.New(doc.ids...)
		sortedIds := ids.Values()
		sort.Strings(sortedIds)
		for _, id := range sortedIds {
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
						doc.site.logerror("%s: contains link to missing anchor: \"%s\"", doc.contentPath, strings.TrimPrefix(url, site.urlprefix()+"/"))
						errCount++
						continue
					}
				}
				site.verbose2("lint: %s: validated link: \"%s\"", doc.contentPath, url)
			}
		}
	}
	// TODO this will be unnecessary, convert it to isValidPath(docPath string) helper
	// Check document URL path names are lowercase alphanumeric with hyphens.
	// site.verbose("lint URL paths")
	// for p := range site.docs.byBuildPath {
	// 	p, _ = filepath.Rel(site.buildDir, p)
	// 	p = filepath.ToSlash(p)
	// 	if !isCleanURLPath(p) {
	// 		site.warning("dubious URL path name: \"%s\"", p)
	// 		warnCount++
	// 	}
	// }
	return
}
