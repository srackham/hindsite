package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// build implements the build command.
func (proj *project) build() error {
	startTime := time.Now()
	// Parse configuration files.
	if err := proj.parseConfigs(); err != nil {
		return err
	}
	if !dirExists(proj.buildDir) {
		if err := os.Mkdir(proj.buildDir, 0775); err != nil {
			return err
		}
	}
	proj.docs = newDocumentsLookup()
	// Delete everything in the build directory forcing a complete site rebuild.
	files, _ := filepath.Glob(filepath.Join(proj.buildDir, "*"))
	for _, f := range files {
		if err := os.RemoveAll(f); err != nil {
			return err
		}
	}
	// Parse all template files.
	proj.htmlTemplates = newHtmlTemplates(proj.templateDir)
	proj.textTemplates = newTextTemplates(proj.templateDir)
	err := filepath.Walk(proj.templateDir, func(f string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if f == proj.templateDir {
			return nil
		}
		if info.IsDir() && f == filepath.Join(proj.templateDir, "init") {
			return filepath.SkipDir
		}
		if proj.exclude(f) {
			proj.verbose("exclude: " + f)
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
				proj.verbose("parse template: " + f)
				err = proj.htmlTemplates.add(f)
			case ".txt":
				// Compile text template.
				proj.verbose("parse template: " + f)
				err = proj.textTemplates.add(f)
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
	err = filepath.Walk(proj.contentDir, func(f string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if f == proj.contentDir {
			return nil
		}
		if proj.exclude(f) {
			proj.verbose("exclude: " + f)
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
				doc, err := newDocument(f, proj)
				if err != nil {
					return err
				}
				if doc.isDraft() {
					draftsCount++
					proj.verbose("skip draft: " + f)
					return nil
				}
				if err := proj.docs.add(&doc); err != nil {
					return err
				}
			default:
				staticCount++
				proj.buildStaticFile(f)
			}
		}
		return err
	})
	if err != nil {
		return err
	}
	// Create indexes.
	proj.idxs, err = newIndexes(proj)
	if err != nil {
		return err
	}
	for _, doc := range proj.docs.byContentPath {
		proj.idxs.addDocument(doc)
	}
	// Build index pages.
	err = proj.idxs.build()
	if err != nil {
		return err
	}
	// Render documents.
	for _, doc := range proj.docs.byContentPath {
		if err = proj.renderDocument(doc); err != nil {
			return err
		}
	}
	// Install home page.
	if err := proj.installHomePage(); err != nil {
		return err
	}
	fmt.Printf("documents: %d\n", docsCount)
	fmt.Printf("drafts: %d\n", draftsCount)
	fmt.Printf("static: %d\n", staticCount)
	fmt.Printf("time: %.2fs\n", time.Now().Sub(startTime).Seconds())
	return nil
}

// upToDate returns false target file is newer than the prerequisite file or if
// target does not exist.
func upToDate(target, prerequisite string) bool {
	result, err := fileIsOlder(prerequisite, target)
	if err != nil {
		return false
	}
	return result
}

func (proj *project) installHomePage() error {
	if proj.rootConf.homepage != "" {
		src := proj.rootConf.homepage
		if !fileExists(src) {
			return fmt.Errorf("homepage file missing: %s", src)
		}
		dst := filepath.Join(proj.buildDir, "index.html")
		if !fileExists(dst) || upToDate(src, dst) {
			proj.verbose2("copy homepage: " + src)
			proj.verbose("write homepage: " + dst)
			if err := copyFile(src, dst); err != nil {
				return err
			}
		}
	}
	return nil
}

func (proj *project) buildStaticFile(f string) error {
	conf := proj.configFor(f)
	if isTemplate(f, conf.templates) {
		return proj.renderStaticFile(f)
	}
	return proj.copyStaticFile(f)
}

// copyStaticFile copies the content directory srcFile to corresponding build
// directory. Skips if the destination file is up to date. Creates missing
// destination directories.
func (proj *project) copyStaticFile(srcFile string) error {
	if !pathIsInDir(srcFile, proj.contentDir) {
		panic("static file is outside content directory: " + srcFile)
	}
	dstFile := pathTranslate(srcFile, proj.contentDir, proj.buildDir)
	if upToDate(dstFile, srcFile) {
		return nil
	}
	proj.verbose("copy static:  " + srcFile)
	err := mkMissingDir(filepath.Dir(dstFile))
	if err != nil {
		return err
	}
	err = copyFile(srcFile, dstFile)
	if err != nil {
		return err
	}
	proj.verbose2("write static: " + dstFile)
	return nil
}

// renderStaticFile renders file f from the content directory as a text template
// and writes it to the corresponding build directory. Skips if the destination
// file is newer than f and is newer than the modified time. Creates missing
// destination directories.
func (proj *project) renderStaticFile(f string) error {
	// Parse document.
	doc, err := newDocument(f, proj)
	if err != nil {
		return err
	}
	// Render document markup as a text template.
	proj.verbose2("render static: " + doc.contentPath)
	proj.verbose2(doc.String())
	markup := doc.content
	if isTemplate(doc.contentPath, doc.templates) {
		data := doc.frontMatter()
		markup, err = proj.textTemplates.renderText("staticFile", markup, data)
		if err != nil {
			return err
		}
	}
	proj.verbose("write static: " + doc.buildPath)
	err = mkMissingDir(filepath.Dir(doc.buildPath))
	if err != nil {
		return err
	}
	return writeFile(doc.buildPath, markup)
}

func (proj *project) renderDocument(doc *document) error {
	var err error
	data := doc.frontMatter()
	markup := doc.content
	// Render document markup as a text template.
	if isTemplate(doc.contentPath, doc.templates) {
		proj.verbose2("render template: " + doc.contentPath)
		markup, err = proj.textTemplates.renderText("documentMarkup", markup, data)
		if err != nil {
			return err
		}
	}
	// Convert markup to HTML then render document layout to build directory.
	proj.verbose2("render document: " + doc.contentPath)
	data["body"] = doc.render(markup)
	err = proj.htmlTemplates.render(doc.layout, data, doc.buildPath)
	if err != nil {
		return err
	}
	proj.verbose("write document: " + doc.buildPath)
	proj.verbose2(doc.String())
	return nil
}
