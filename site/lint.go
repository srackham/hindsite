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

// linkTarget computes the URL's target build file path.
// Returns blank string if the URL is off-site.
func (doc *document) linkTarget(u *urlpkg.URL) (target string) {
	url := u.String()
	url = strings.TrimSuffix(url, "#"+u.Fragment)
	re := regexp.MustCompile(`(?i)^(?:` + regexp.QuoteMeta(doc.site.urlprefix()) + `)?/([^/].*)$`) // Extracts root-relative URLs.
	matches := re.FindStringSubmatch(url)
	if matches != nil {
		target = filepath.Join(doc.site.buildDir, decodeURL(matches[1]))
	} else {
		re := regexp.MustCompile(`(?i)^([\w][\w./-]*)$`) // Extracts page-relative URLs.
		matches := re.FindStringSubmatch(url)
		if matches != nil {
			target = filepath.Join(filepath.Dir(doc.buildPath), decodeURL(matches[1]))
		}
	}
	if target != "" && fsx.DirExists(target) {
		target = filepath.Join(target, "index.html")
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
func (site *site) lintChecks() {
	for _, k := range sortedKeys(site.docs.byContentPath) {
		doc := site.docs.byContentPath[k]
		site.logVerbose("lint document: \"%s\"", doc.contentPath)
		// Check for illicit or duplicate ids.
		ids := set.New(doc.ids...)
		sortedIds := ids.Values()
		sort.Strings(sortedIds)
		for _, id := range sortedIds {
			re := regexp.MustCompile(`^[A-Za-z][\w:.-]*$`) // https://www.w3.org/TR/html4/types.html
			if !re.MatchString(id) {
				doc.site.logError("\"%s\": contains illicit element id: \"%s\"", doc.contentPath, id)
			}
			if ids.Count(id) > 1 {
				doc.site.logError("\"%s\": contains duplicate element id: \"%s\"", doc.contentPath, id)
			}
		}
		// Iterate the document's href/src attribute URLs.
		for _, url := range doc.urls {
			u, err := urlpkg.Parse(url)
			if err != nil {
				doc.site.logError("\"%s\": contains illicit URL: \"%s\"", doc.contentPath, url)
				continue
			}
			if strings.HasPrefix(url, "#") { // Intra-document URL fragment.
				if !doc.ids.Has(url[1:]) {
					doc.site.logError("\"%s\": contains link to missing anchor: \"%s\"", doc.contentPath, url)
					continue
				}
			} else {
				target := doc.linkTarget(u)
				if target == "" { // Off-site URL.
					site.logVerbose2("lint: \"%s\": skipped offsite link: \"%s\"", doc.contentPath, url)
					continue
				}
				// Check the target URL file exists.
				if !fsx.FileExists(target) {
					doc.site.logError("\"%s\": contains link to missing file: \"%s\"", doc.contentPath, strings.TrimPrefix(target, site.buildDir+string(filepath.Separator)))
					continue
				}
				// Check the URL anchor has a matching HTML id attribute in the target document.
				if u.Fragment != "" {
					targetDoc, ok := site.docs.byBuildPath[target]
					if !ok || !targetDoc.ids.Has(u.Fragment) {
						doc.site.logError("\"%s\": contains link to missing anchor: \"%s\"", doc.contentPath, strings.TrimPrefix(url, site.urlprefix()+"/"))
						continue
					}
				}
				site.logVerbose2("lint: \"%s\": validated link: \"%s\"", doc.contentPath, url)
			}
		}
	}
	return
}
