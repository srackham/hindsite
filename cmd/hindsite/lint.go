package main

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// documentLink represents an intra-site document link.
type documentLink struct {
	target string // Target build path
	anchor string // URL fragment
}

// parseLink computes the URL's target build path and URL anchor fragment from
// `url` and returns them in `link`. `offsite` returns `true` if the URL targets
// an external resource.
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
		if dirExists(link.target) {
			link.target = filepath.Join(link.target, "index.html")
		}
	} else { // Matches relative URLs.
		re := regexp.MustCompile(`(?i)^([\w][\w./-]*)$`)
		matches := re.FindStringSubmatch(url)
		if matches == nil {
			return link, true
		}
		link.target = filepath.Join(filepath.Dir(doc.buildPath), matches[1])
	}
	return link, false
}

// parseHTML scans the document's `html` saving id attributes to `doc.ids` and href and src
// URL attributes to `doc.urls` which are used for subsequent site link validation.
func (doc *document) parseHTML(html string) {
	// Scan HTML element id attributes.
	doc.ids = stringlist{}
	pat := regexp.MustCompile(`(?i)id="(.+?)"`)
	matches := pat.FindAllStringSubmatch(html, -1)
	for _, match := range matches {
		doc.ids = append(doc.ids, match[1])
	}
	// Scan HTML for href and src URL attributes.
	doc.urls = stringlist{}
	pat = regexp.MustCompile(`(?i)(?:href|src)="(.+?)"`)
	matches = pat.FindAllStringSubmatch(html, -1)
	for _, match := range matches {
		doc.urls = append(doc.urls, match[1])
	}

}

// lintLinks checks that all document intra-site URLs point to valid target
// files and valid HTML id attributes.
func (site *site) lintLinks() (errCount int) {
	for _, doc := range site.docs.byContentPath {
		// Iterate the document's href/src attribute URLs.
		for _, url := range doc.urls {
			if strings.HasPrefix(url, "#") { // Intra-document URL fragment.
				if !doc.ids.Contains(url[1:]) {
					err := fmt.Errorf("%s: contains link to missing anchor: %s", doc.contentPath, url)
					errCount++
					doc.site.logerror(err.Error())
					continue
				}
			} else {
				link, offsite := doc.parseLink(url)
				if offsite {
					site.verbose2("lint: %s: skipped offsite link: %s", doc.contentPath, url)
					continue
				}
				// Check the target URL file exists.
				if !fileExists(link.target) {
					err := fmt.Errorf("%s: contains link to missing file: %s", doc.contentPath, strings.TrimPrefix(link.target, site.buildDir+string(filepath.Separator)))
					errCount++
					site.logerror(err.Error())
					continue
				}
				// Check the URL anchor has a matching HTML id attribute in the target document.
				if link.anchor != "" {
					target, ok := site.docs.byBuildPath[link.target]
					if !ok || !target.ids.Contains(link.anchor) {
						err := fmt.Errorf("%s: contains link to missing anchor: %s", doc.contentPath, strings.TrimPrefix(url, site.rootConf.urlprefix+string(filepath.Separator)))
						errCount++
						site.logerror(err.Error())
						continue
					}
				}
				site.verbose2("lint: %s: validated link: %s", doc.contentPath, url)
			}
		}
	}
	return
}
