package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// init implements the init command.
func (proj *project) init() error {
	if dirExists(proj.contentDir) {
		files, err := ioutil.ReadDir(proj.contentDir)
		if err != nil {
			return err
		}
		if len(files) > 0 {
			return fmt.Errorf("non-empty content directory: " + proj.contentDir)
		}
	}
	if proj.builtin != "" {
		// Load template directory from the built-in project.
		if dirExists(proj.templateDir) {
			files, err := ioutil.ReadDir(proj.templateDir)
			if err != nil {
				return err
			}
			if len(files) > 0 {
				return fmt.Errorf("non-empty template directory: " + proj.templateDir)
			}
		}
		proj.verbose("installing builtin template: " + proj.builtin)
		if err := RestoreAssets(proj.templateDir, proj.builtin+"/template"); err != nil {
			return err
		}
		// Hoist the restored template files from the root of the restored
		// builtin directory up one level into the root of the project template
		// directory.
		files, _ := filepath.Glob(filepath.Join(proj.templateDir, proj.builtin, "template", "*"))
		for _, f := range files {
			if err := os.Rename(f, filepath.Join(proj.templateDir, filepath.Base(f))); err != nil {
				return err
			}
		}
		// Remove the now empty restored path.
		if err := os.RemoveAll(filepath.Join(proj.templateDir, proj.builtin)); err != nil {
			return err
		}
	} else {
		if !dirExists(proj.templateDir) {
			return fmt.Errorf("missing template directory: " + proj.templateDir)
		}
	}
	// Create the template directory structure in the content directory.
	if err := mkMissingDir(proj.contentDir); err != nil {
		return err
	}
	err := filepath.Walk(proj.templateDir, func(f string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if f == proj.templateDir {
			return nil
		}
		if info.IsDir() && f == proj.initDir {
			return filepath.SkipDir
		}
		if info.IsDir() {
			dst := pathTranslate(f, proj.templateDir, proj.contentDir)
			proj.verbose("make directory: " + dst)
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
					proj.verbose("make directory: " + dst)
					err = mkMissingDir(dst)
				}
			} else {
				proj.verbose2("copy: " + f)
				proj.verbose("write: " + dst)
				err = copyFile(f, dst)
			}
			return err
		})
		return nil
	}
	// Copy the contents of the optional template init directory to the content directory.
	if dirExists(proj.initDir) {
		if err := copyDirContents(proj.initDir, proj.contentDir); err != nil {
			return err
		}
	}
	// If the template directory is outside the project directory copy it to the
	// default template directory (if it does not already exist).
	defaultTemplateDir := filepath.Join(proj.projectDir, "template")
	if !pathIsInDir(proj.templateDir, proj.projectDir) && !dirExists(defaultTemplateDir) {
		proj.verbose("make directory: " + defaultTemplateDir)
		if err := mkMissingDir(defaultTemplateDir); err != nil {
			return err
		}
		if err := copyDirContents(proj.templateDir, defaultTemplateDir); err != nil {
			return err
		}
	}
	return nil
}
