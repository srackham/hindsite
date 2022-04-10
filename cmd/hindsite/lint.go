package main

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

type documentLink struct {
	buildPath string // Target document build path
	anchor    string // URL fragment
}

// parseLink computes the URL's document build path and URL anchor fragment from
// `url` and returns them in `link`. If the URL is external to the site then `ok`
// is returned `false`.
func (doc *document) parseLink(url string) (link documentLink, ok bool) {
	// Split into URL and anchor.
	s := strings.Split(url, "#")
	url = s[0]
	if len(s) > 1 {
		link.anchor = s[1]
	}
	// Match URLs with a urlprefix or root-relative URLs.
	re := regexp.MustCompile(`(?i)^(?:` + regexp.QuoteMeta(doc.conf.urlprefix) + `)?/(.+)$`)
	matches := re.FindStringSubmatch(url)
	if matches != nil {
		link.buildPath = filepath.Join(doc.site.buildDir, matches[1])
		if dirExists(link.buildPath) {
			link.buildPath = filepath.Join(link.buildPath, "index.html")
		}
	} else {
		// Match relative URLs.
		re := regexp.MustCompile(`(?i)^([\w][\w./-]*)$`)
		matches := re.FindStringSubmatch(url)
		if matches == nil {
			return link, false // External link.
		}
		link.buildPath = filepath.Join(filepath.Dir(doc.buildPath), matches[1])
	}
	return link, true
}

// parseHTML scans the document's `html` saving id attributes to `doc.ids` and href and src
// URL attributes to `doc.urls`.
func (doc *document) parseHTML(html string) {
	// Scan HTML for intra-document anchor URLs and element ids.
	doc.ids = stringlist{}
	pat := regexp.MustCompile(`(?i)id="(.+?)"`)
	matches := pat.FindAllStringSubmatch(html, -1)
	for _, match := range matches {
		doc.ids = append(doc.ids, match[1])
	}
	doc.urls = stringlist{}
	pat = regexp.MustCompile(`(?i)(?:href|src)="(.+?)"`)
	matches = pat.FindAllStringSubmatch(html, -1)
	for _, match := range matches {
		doc.urls = append(doc.urls, match[1])
	}

}

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
			} else { // Inter-document URL
				link, ok := doc.parseLink(url) // TODO parse at render time???
				if !ok {
					site.verbose2("lint: %s: skipped external link: %s", doc.contentPath, url)
					continue
				}
				// Check the target URL file exists.
				var target *document
				if !fileExists(link.buildPath) {
					err := fmt.Errorf("%s: contains link to missing file: %s", doc.contentPath, strings.TrimPrefix(link.buildPath, site.buildDir+string(filepath.Separator)))
					errCount++
					site.logerror(err.Error())
					continue
				}
				// Check the href/src URL anchor has a matching HTML id attribute in
				// the target document.
				if link.anchor != "" {
					target, ok = site.docs.byBuildPath[link.buildPath]
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
