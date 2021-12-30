package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fatih/color"
)

// build implements the build command.
func (site *site) build() error {
	startTime := time.Now()
	// Parse configuration files.
	if err := site.parseConfigs(); err != nil {
		return err
	}
	// Synthesize root config.
	site.rootConf = newConfig()
	if len(site.confs) > 0 && site.confs[0].origin == site.templateDir {
		site.rootConf.merge(site.confs[0])
	}
	site.verbose2("root config: \n" + site.rootConf.String())
	if !dirExists(site.buildDir) {
		if err := os.Mkdir(site.buildDir, 0775); err != nil {
			return err
		}
	}
	site.docs = newDocumentsLookup()
	// Delete everything in the build directory forcing a complete site rebuild.
	files, _ := filepath.Glob(filepath.Join(site.buildDir, "*"))
	for _, f := range files {
		if err := os.RemoveAll(f); err != nil {
			return err
		}
	}
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
		if site.exclude(f) {
			site.verbose("exclude: " + f)
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
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
	// Parse content directory documents and copy/render static files to the build directory.
	draftsCount := 0
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
					draftsCount++
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
	// Print summary.
	if errCount == 0 {
		color.Set(color.FgGreen, color.Bold)
	}
	site.logconsole("documents: %d", docsCount)
	site.logconsole("drafts: %d", draftsCount)
	site.logconsole("static: %d", staticCount)
	site.logconsole("time: %.2fs", time.Since(startTime).Seconds())
	color.Unset()
	// Report accumulated document parse errors.
	if errCount > 0 {
		return fmt.Errorf("document parse errors: %d", errCount)
	}
	return nil
}

func (site *site) copyHomePage() error {
	if site.rootConf.homepage != "" {
		src := site.rootConf.homepage
		if !fileExists(src) {
			return fmt.Errorf("homepage file missing: %s", src)
		}
		dst := filepath.Join(site.buildDir, "index.html")
		site.verbose2("copy homepage: " + src)
		site.verbose("write homepage: " + dst)
		if err := copyFile(src, dst); err != nil {
			return err
		}
	}
	return nil
}

func (site *site) buildStaticFile(f string) error {
	conf := site.configFor(f)
	if isTemplate(f, conf.templates) {
		return site.renderStaticFile(f)
	}
	return site.copyStaticFile(f)
}

// copyStaticFile copies the content directory srcFile to corresponding build
// directory. Creates missing destination directories.
func (site *site) copyStaticFile(srcFile string) error {
	if !pathIsInDir(srcFile, site.contentDir) {
		panic("static file is outside content directory: " + srcFile)
	}
	dstFile := pathTranslate(srcFile, site.contentDir, site.buildDir)
	site.verbose("copy static:  " + srcFile)
	err := mkMissingDir(filepath.Dir(dstFile))
	if err != nil {
		return err
	}
	err = copyFile(srcFile, dstFile)
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
	// Render document markup as a text template.
	site.verbose2("render static: " + doc.contentPath)
	site.verbose2(doc.String())
	markup := doc.content
	if isTemplate(doc.contentPath, doc.templates) {
		data := doc.frontMatter()
		markup, err = site.textTemplates.renderText("staticFile", markup, data)
		if err != nil {
			return err
		}
	}
	site.verbose("write static: " + doc.buildPath)
	err = mkMissingDir(filepath.Dir(doc.buildPath))
	if err != nil {
		return err
	}
	return writeFile(doc.buildPath, markup)
}

func (site *site) renderDocument(doc *document) error {
	var err error
	data := doc.frontMatter()
	markup := doc.content
	// Render document markup as a text template.
	if isTemplate(doc.contentPath, doc.templates) {
		site.verbose2("render template: " + doc.contentPath)
		markup, err = site.textTemplates.renderText("documentMarkup", markup, data)
		if err != nil {
			return err
		}
	}
	// Convert markup to HTML then render document layout to build directory.
	site.verbose2("render document: " + doc.contentPath)
	data["body"] = doc.render(markup)
	err = site.htmlTemplates.render(doc.layout, data, doc.buildPath)
	if err != nil {
		return err
	}
	site.verbose("write document: " + doc.buildPath)
	site.verbose2(doc.String())
	return nil
}
