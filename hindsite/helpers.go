package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

/*
Miscellaneous functions.
*/

// nz returns the string pointed to by s or "" if s is nil.
func nz(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// Transform text into a slug (lowercase alpha-numeric + hyphens).
func slugify(text string, exclude stringlist) string {
	slug := text
	slug = regexp.MustCompile(`\W+`).ReplaceAllString(slug, "-") // Replace non-alphanumeric characters with dashes.
	slug = regexp.MustCompile(`-+`).ReplaceAllString(slug, "-")  // Replace multiple dashes with single dash.
	slug = strings.Trim(slug, "-")                               // Trim leading and trailing dashes.
	slug = strings.ToLower(slug)
	if slug == "" {
		slug = "x"
	}
	if exclude.IndexOf(slug) > -1 {
		i := 2
		for exclude.IndexOf(slug+"-"+fmt.Sprint(i)) > -1 {
			i++
		}
		slug += "-" + fmt.Sprint(i)
	}
	return slug
}

func injectLiveReload(html string) string {
	split := strings.Split(html, "</body>")
	if len(split) == 2 {
		const scripttag = "<script src=\"http://localhost:35729/livereload.js\"></script>\n"
		return split[0] + scripttag + "</body>" + split[1]
	}
	return html
}

/*
String lists.
*/
type stringlist []string

// Returns the first index of the target string `t`, or
// -1 if no match is found.
func (list stringlist) IndexOf(t string) int {
	for i, v := range list {
		if v == t {
			return i
		}
	}
	return -1
}

// Returns `true` if the target string t is in the
// slice.
func (list stringlist) Contains(t string) bool {
	return list.IndexOf(t) >= 0
}

/*
File functions.
*/
func dirExists(name string) bool {
	info, err := os.Stat(name)
	return err == nil && info.IsDir()
}

func fileExists(name string) bool {
	info, err := os.Stat(name)
	return err == nil && !info.IsDir()
}

func readFile(name string) (string, error) {
	bytes, err := ioutil.ReadFile(name)
	return string(bytes), err
}

func writeFile(name string, text string) error {
	return ioutil.WriteFile(name, []byte(text), 0644)
}

// Return file name sans extension.
func fileName(name string) string {
	return replaceExt(filepath.Base(name), "")
}

// Replace the extension of name.
func replaceExt(name, ext string) string {
	return name[0:len(name)-len(filepath.Ext(name))] + ext
}

func copyFile(from, to string) error {
	contents, err := readFile(from)
	if err != nil {
		return err
	}
	err = writeFile(to, contents)
	return err
}

func mkMissingDir(dir string) error {
	if !dirExists(dir) {
		if err := os.MkdirAll(dir, 0775); err != nil {
			return err
		}
	}
	return nil
}

// pathIsInDir returns true if path p is in directory dir or if p equals dir.
func pathIsInDir(p, dir string) bool {
	return p == dir || strings.HasPrefix(p, dir+string(filepath.Separator))
}

// Translate srcPath to corresponding path in dstRoot.
func pathTranslate(srcPath, srcRoot, dstRoot string) string {
	if !pathIsInDir(srcPath, srcRoot) {
		panic("srcPath not in srcRoot: " + srcPath)
	}
	dstPath, err := filepath.Rel(srcRoot, srcPath)
	if err != nil {
		panic(err.Error())
	}
	return filepath.Join(dstRoot, dstPath)
}

// TODO: UNUSED
// Search for files from base directory up to root directory.
// Return found files.
// Files are ordered by location (base to root).
// If a directory in the base path does not exist it is silent skipped.
// If n >= 0 the function returns first n matched files.
func filesInPath(base, root string, patterns []string, n int) (files []string, err error) {
	if !filepath.IsAbs(base) {
		return files, fmt.Errorf("base path is not absolute: %s", base)
	}
	if !filepath.IsAbs(root) {
		return files, fmt.Errorf("root path is not absolute: %s", root)
	}
	if base != root && !pathIsInDir(base, root) {
		return files, fmt.Errorf("base is not a child of root of base: %s: %s", base, root)
	}
	p := base
	count := 0
	for {
		for _, pat := range patterns {
			matches, err := filepath.Glob(filepath.Join(p, pat))
			if err != nil {
				return files, err
			}
			for _, match := range matches {
				if count >= n {
					break
				}
				files = append(files, match)
				count++
			}
		}
		if p == root {
			break
		}
		p = filepath.Dir(p)
	}
	return files, nil
}

// isOlder returns true if oldtime is older than newtime by at least 0.5s.
func isOlder(oldtime, newtime time.Time) bool {
	delta := time.Second / 2
	diff := newtime.Sub(oldtime)
	return diff > 0 && diff > delta
}

// fileIsOlder returns true if oldfile's modification time is older than
// newfile's by at least 0.5s. Returns error is one or both files do not exist.
func fileIsOlder(oldfile, newfile string) (bool, error) {
	info, err := os.Stat(newfile)
	if err != nil {
		return false, err
	}
	newtime := info.ModTime()
	info, err = os.Stat(oldfile)
	if err != nil {
		return false, err
	}
	oldtime := info.ModTime()
	return isOlder(oldtime, newtime), nil
}

/*
Date/time functions.
*/
// Parse date text. If timezone is not specified Local is assumed.
func parseDate(text string, loc *time.Location) (time.Time, error) {
	if loc == nil {
		loc, _ = time.LoadLocation("Local")
	}
	d, err := time.Parse(time.RFC3339, text)
	if err != nil {
		if d, err = time.Parse("2006-01-02 15:04:05-07:00", text); err != nil {
			if d, err = time.ParseInLocation("2006-01-02 15:04:05", text, loc); err != nil {
				d, err = time.ParseInLocation("2006-01-02", text, loc)
			}
		}
	}
	return d, err
}
