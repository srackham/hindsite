package main

import (
	"embed"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// init implements the init command.
func (site *site) init() error {
	if dirExists(site.contentDir) {
		files, err := ioutil.ReadDir(site.contentDir)
		if err != nil {
			return err
		}
		if len(files) > 0 {
			return fmt.Errorf("non-empty content directory: " + site.contentDir)
		}
	}
	if site.builtin != "" {
		// Load template directory from the built-in site.
		if dirCount(site.templateDir) > 0 {
			return fmt.Errorf("non-empty template directory: " + site.templateDir)
		}
		site.verbose("installing builtin template: " + site.builtin)
		if err := restoreEmbeddedFS(embeddedFS, "builtin/"+site.builtin+"/template", site.templateDir); err != nil {
			return err
		}
		// Hoist the restored template files from the root of the restored
		// builtin directory up one level into the root of the site template
		// directory.
		files, _ := filepath.Glob(filepath.Join(site.templateDir, site.builtin, "template", "*"))
		for _, f := range files {
			if err := os.Rename(f, filepath.Join(site.templateDir, filepath.Base(f))); err != nil {
				return err
			}
		}
		// Remove the now empty restored path.
		if err := os.RemoveAll(filepath.Join(site.templateDir, site.builtin)); err != nil {
			return err
		}
	}
	// Create the template directory structure in the content directory.
	if err := mkMissingDir(site.contentDir); err != nil {
		return err
	}
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
		if info.IsDir() {
			dst := pathTranslate(f, site.templateDir, site.contentDir)
			site.verbose("make directory: " + dst)
			err = mkMissingDir(dst)
		}
		return err
	})
	if err != nil {
		return err
	}
	copyDirContents := func(fromDir, toDir string) error {
		err = filepath.Walk(fromDir, func(f string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if f == fromDir {
				return nil
			}
			dst := pathTranslate(f, fromDir, toDir)
			if info.IsDir() {
				if !dirExists(dst) {
					site.verbose("make directory: " + dst)
					err = mkMissingDir(dst)
				}
			} else {
				site.verbose2("copy: " + f)
				site.verbose("write: " + dst)
				err = copyFile(f, dst)
			}
			return err
		})
		return nil
	}
	// Copy the contents of the optional template init directory to the content directory.
	if dirExists(site.initDir) {
		if err := copyDirContents(site.initDir, site.contentDir); err != nil {
			return err
		}
	}
	// If the template directory is outside the site directory copy it to the
	// default template directory (if it does not already exist or is empty).
	defaultTemplateDir := filepath.Join(site.siteDir, "template")
	if !pathIsInDir(site.templateDir, site.siteDir) && dirCount(defaultTemplateDir) == 0 {
		site.verbose("make directory: " + defaultTemplateDir)
		if err := mkMissingDir(defaultTemplateDir); err != nil {
			return err
		}
		if err := copyDirContents(site.templateDir, defaultTemplateDir); err != nil {
			return err
		}
	}
	return nil
}

//go:embed builtin/blog/template/** builtin/minimal/template/**
var embeddedFS embed.FS

// Recursively restore embedded file system directory srcDir to disk dstDir.
func restoreEmbeddedFS(srcFS embed.FS, srcDir string, dstDir string) error {
	entries, err := srcFS.ReadDir(srcDir)
	if err != nil {
		panic(err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			if err := restoreEmbeddedFS(srcFS, srcDir+"/"+entry.Name(), dstDir+"/"+entry.Name()); err != nil {
				return err
			}
		} else {
			if err := mkMissingDir(dstDir); err != nil {
				return err
			}
			srcFile := srcDir + "/" + entry.Name()
			contents, err := srcFS.ReadFile(srcFile)
			if err != nil {
				return err
			}
			dstFile := dstDir + "/" + entry.Name()
			if err := writeFile(dstFile, string(contents)); err != nil {
				return err
			}
		}
	}
	return nil
}
