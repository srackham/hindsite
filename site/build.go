package site

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/fatih/color"
	"github.com/srackham/hindsite/fsx"
)

// build implements the build command.
func (site *site) build() error {
	if len(site.cmdargs) > 0 {
		return fmt.Errorf("to many command arguments")
	}
	startTime := time.Now()
	if err := site.parseConfigFiles(); err != nil {
		return err
	}
	site.docs = newDocumentsLookup()
	// Parse all template files.
	site.htmlTemplates = newHTMLTemplates(site.templateDir)
	site.textTemplates = newTextTemplates(site.templateDir)
	err := filepath.Walk(site.templateDir, func(f string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if f == site.templateDir {
			return nil
		}
		if info.IsDir() && f == site.initDir {
			return filepath.SkipDir
		}
		if !info.IsDir() {
			switch filepath.Ext(f) {
			case ".toml", ".yaml":
				// Skip configuration file.
			case ".html":
				// Compile HTML template.
				site.verbose("parse template: " + f)
				err = site.htmlTemplates.add(f)
			case ".txt":
				// Compile text template.
				site.verbose("parse template: " + f)
				err = site.textTemplates.add(f)
			}
		}
		return err
	})
	if err != nil {
		return err
	}
	if !fsx.DirExists(site.buildDir) {
		if err := os.Mkdir(site.buildDir, 0775); err != nil {
			return err
		}
	}
	// Delete everything in the build directory forcing a complete site rebuild.
	if !site.keep {
		files, _ := filepath.Glob(filepath.Join(site.buildDir, "*"))
		for _, f := range files {
			if err := os.RemoveAll(f); err != nil {
				return err
			}
		}
	}
	// Parse content directory documents and copy/render static files to the build directory.
	docsCount := 0
	staticCount := 0
	errCount := 0
	err = filepath.Walk(site.contentDir, func(f string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if f == site.contentDir {
			return nil
		}
		if site.exclude(f) {
			site.verbose("exclude: " + f)
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !info.IsDir() {
			switch filepath.Ext(f) {
			case ".md", ".rmu":
				docsCount++
				// Parse document.
				doc, err := newDocument(f, site)
				if err != nil {
					errCount++
					site.logerror(err.Error())
					return nil
				}
				if doc.isDraft() {
					site.verbose("skip draft: " + f)
					return nil
				}
				if err := site.docs.add(&doc); err != nil {
					errCount++
					site.logerror(err.Error())
					return nil
				}
			default:
				staticCount++
				if err := site.buildStaticFile(f); err != nil {
					errCount++
					site.logerror(err.Error())
					return nil
				}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	// Create indexes.
	site.idxs, err = newIndexes(site)
	if err != nil {
		return err
	}
	for _, doc := range site.docs.byContentPath {
		site.idxs.addDocument(doc)
	}
	// Build index pages.
	err = site.idxs.build()
	if err != nil {
		return err
	}
	// Render documents.
	for _, doc := range site.docs.byContentPath {
		if err = site.renderDocument(doc); err != nil {
			return err
		}
	}
	// Install home page.
	if err := site.copyHomePage(); err != nil {
		return err
	}
	// Lint documents.
	if site.lint {
		errCount += site.lintChecks()
	}
	// Print summary.
	if errCount == 0 {
		color.Set(color.FgGreen, color.Bold)
	}
	site.logconsole("documents: %d", docsCount)
	site.logconsole("static: %d", staticCount)
	site.logconsole("time: %.2fs", time.Since(startTime).Seconds())
	color.Unset()
	if errCount > 0 {
		return fmt.Errorf("document errors: %d", errCount)
	}
	return nil
}

// copyHomePage copies the `homepage` config variable file to `/index.html` and
// adds it to the list of built documents.
func (site *site) copyHomePage() error {
	if site.confs[0].homepage != "" {
		homepage := site.confs[0].homepage
		homepage = filepath.FromSlash(homepage)
		if !filepath.IsAbs(homepage) {
			homepage = filepath.Join(site.buildDir, homepage)
		} else {
			return fmt.Errorf("homepage must be relative to the build directory: %s", site.buildDir)
		}
		if !fsx.PathIsInDir(homepage, site.buildDir) {
			return fmt.Errorf("homepage must reside in build directory: %s", site.buildDir)
		}
		if fsx.DirExists(homepage) {
			return fmt.Errorf("homepage cannot be a directory: %s", homepage)
		}
		if !fsx.FileExists(homepage) {
			return fmt.Errorf("homepage file missing: %s", homepage)
		}
		dst := filepath.Join(site.buildDir, "index.html")
		site.verbose2("copy homepage: " + homepage)
		site.verbose("write homepage: " + dst)
		if err := fsx.CopyFile(homepage, dst); err != nil {
			return err
		}
		site.docs.byBuildPath[dst] = site.docs.byBuildPath[homepage]
	}
	return nil
}

func (site *site) buildStaticFile(f string) error {
	conf := site.configFor(f)
	if site.match(f, conf.templates) {
		return site.renderStaticFile(f)
	}
	return site.copyStaticFile(f)
}

// copyStaticFile copies the content directory srcFile to corresponding build
// directory. Creates missing destination directories.
func (site *site) copyStaticFile(srcFile string) error {
	if !fsx.PathIsInDir(srcFile, site.contentDir) {
		panic("static file is outside content directory: " + srcFile)
	}
	dstFile := fsx.PathTranslate(srcFile, site.contentDir, site.buildDir)
	site.verbose("copy static:  " + srcFile)
	err := fsx.MkMissingDir(filepath.Dir(dstFile))
	if err != nil {
		return err
	}
	err = fsx.CopyFile(srcFile, dstFile)
	if err != nil {
		return err
	}
	site.verbose2("write static: " + dstFile)
	return nil
}

// renderStaticFile renders file f from the content directory as a text template
// and writes it to the corresponding build directory. Creates missing
// destination directories.
func (site *site) renderStaticFile(f string) error {
	// Parse document.
	doc, err := newDocument(f, site)
	if err != nil {
		return err
	}
	// Render file as a text template.
	site.verbose2("render static: " + doc.contentPath)
	site.verbose2(doc.String())
	content := doc.content
	if site.match(doc.contentPath, doc.templates) {
		data := doc.frontMatter()
		content, err = site.textTemplates.renderText("staticFile", content, data)
		if err != nil {
			return err
		}
	}
	site.verbose("write static: " + doc.buildPath)
	return fsx.WritePath(doc.buildPath, content)
}

func (site *site) renderDocument(doc *document) error {
	var err error
	data := doc.frontMatter()
	markup := doc.content
	// Render document markup as a text template.
	if site.match(doc.contentPath, doc.templates) {
		site.verbose2("render template: " + doc.contentPath)
		markup, err = site.textTemplates.renderText("documentMarkup", markup, data)
		if err != nil {
			return err
		}
	}
	// Convert markup to HTML then render document layout to build directory.
	site.verbose2("render document: " + doc.contentPath)
	data["body"] = doc.render(markup)
	html, err := site.htmlTemplates.render(doc.layout, data)
	if err != nil {
		return err
	}
	html = site.injectUrlprefix(html)
	if site.lint {
		doc.parseHTML(html)
		// doc.parseHTML(string(data["body"].(template.HTML)))
	}
	site.verbose("write document: " + doc.buildPath)
	if err = fsx.WritePath(doc.buildPath, html); err != nil {
		return err
	}
	site.verbose2(doc.String())
	return nil
}

// injectUrlprefix prefixes root-relative URLs in HTML
// href and and src attributes with the site `urlprefix`.
func (site *site) injectUrlprefix(html string) string {
	if urlprefix := site.confs[0].urlprefix; urlprefix != "" {
		// Prefix root-relative URLs with the urlprefix.
		re := regexp.MustCompile(`(?i)(href|src)="(/[^/].*?)"`)
		html = re.ReplaceAllString(html, "$1=\""+urlprefix+"$2\"")
	}
	return html
}
