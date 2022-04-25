package site

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/srackham/hindsite/fsx"
	"github.com/srackham/hindsite/slice"
)

// init implements the init command.
func (site *site) init() error {
	if len(site.cmdargs) > 0 {
		return fmt.Errorf("to many command arguments")
	}
	if site.from == "" {
		return fmt.Errorf("-from option source has not been specified")
	}
	if fsx.DirCount(site.templateDir) > 0 {
		site.warning("skipping non-empty target template directory: " + site.templateDir)
	} else {
		if slice.New("blog", "docs", "hello").Has(site.from) {
			// Load template directory from the built-in site.
			site.verbose("installing builtin template: " + site.from)
			if err := restoreEmbeddedFS(embeddedFS, "builtin/"+site.from+"/template", site.templateDir); err != nil {
				return err
			}
			// Hoist the restored template files from the root of the restored
			// builtin directory up one level into the root of the site template
			// directory.
			files, _ := filepath.Glob(filepath.Join(site.templateDir, site.from, "template", "*"))
			for _, f := range files {
				if err := os.Rename(f, filepath.Join(site.templateDir, filepath.Base(f))); err != nil {
					return err
				}
			}
			// Remove the now empty restored path.
			if err := os.RemoveAll(filepath.Join(site.templateDir, site.from)); err != nil {
				return err
			}
		} else {
			// Copy the contents of the source template directory to the template
			// directory.
			if !fsx.DirExists(site.from) {
				return fmt.Errorf("missing source template '%s'", site.from)
			}
			if fsx.PathIsInDir(site.from, site.templateDir) {
				return fmt.Errorf("source template directory '%s' cannot reside inside target template directory '%s'", site.from, site.templateDir)
			}
			if !fsx.DirExists(site.templateDir) {
				site.verbose("make directory: " + site.templateDir)
				if err := fsx.MkMissingDir(site.templateDir); err != nil {
					return err
				}
			}
			if err := site.copyDirContents(site.from, site.templateDir); err != nil {
				return err
			}
		}
	}
	if fsx.DirCount(site.contentDir) > 0 {
		site.warning("skipping non-empty target content directory: " + site.contentDir)
	} else {
		// Create the template directory structure in the content directory.
		if err := fsx.MkMissingDir(site.contentDir); err != nil {
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
				dst := fsx.PathTranslate(f, site.templateDir, site.contentDir)
				site.verbose("make directory: " + dst)
				err = fsx.MkMissingDir(dst)
			}
			return err
		})
		if err != nil {
			return err
		}
		// Copy the contents of the optional template init directory to the content directory.
		if fsx.DirExists(site.initDir) {
			if err := site.copyDirContents(site.initDir, site.contentDir); err != nil {
				return err
			}
		}
	}
	return nil
}

//go:embed builtin/blog/template/** builtin/hello/template/* builtin/docs/template/***
var embeddedFS embed.FS

// Recursively restore embedded file system directory srcDir to disk dstDir.
func restoreEmbeddedFS(srcFS embed.FS, srcDir string, dstDir string) error {
	entries, err := srcFS.ReadDir(srcDir)
	if err != nil {
		panic("failed to read embedded directory: " + srcDir)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			if err := restoreEmbeddedFS(srcFS, srcDir+"/"+entry.Name(), dstDir+"/"+entry.Name()); err != nil {
				return err
			}
		} else {
			srcFile := srcDir + "/" + entry.Name()
			contents, err := srcFS.ReadFile(srcFile)
			if err != nil {
				panic("failed to read embedded file: " + srcFile)
			}
			dstFile := dstDir + "/" + entry.Name()
			if err := fsx.WritePath(dstFile, string(contents)); err != nil {
				return err
			}
		}
	}
	return nil
}

// copyDirContents copies all files and folders in `fromDir`` to `toDir`.
func (site *site) copyDirContents(fromDir, toDir string) error {
	return filepath.Walk(fromDir, func(f string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if f == fromDir {
			return nil
		}
		dst := fsx.PathTranslate(f, fromDir, toDir)
		if info.IsDir() {
			if !fsx.DirExists(dst) {
				site.verbose("make directory: " + dst)
				err = fsx.MkMissingDir(dst)
			}
		} else {
			site.verbose2("copy: " + f)
			site.verbose("write: " + dst)
			err = fsx.CopyFile(f, dst)
		}
		return err
	})
}
