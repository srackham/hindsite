package main

import (
	"bytes"
	"fmt"
	"html/template"
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
func die(message string) {
	if message != "" {
		fmt.Fprintln(os.Stderr, "error: "+message)
	}
	os.Exit(1)
}

// Print message if `-v` verbose option set.
func verbose(message string) {
	if Cmd.verbose {
		fmt.Println(message)
	}
}

func renderTemplate(tmplfile string, data templateData, outfile string) error {
	tmpl, err := template.ParseFiles(tmplfile)
	if err != nil {
		return err
	}
	buf := bytes.NewBufferString("")
	if err := tmpl.Execute(buf, data); err != nil {
		return err
	}
	html := buf.String()
	if err := mkMissingDir(filepath.Dir(outfile)); err != nil {
		return err
	}
	return writeFile(outfile, html)
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

/*
String lists.
*/
// TODO Should we not just add methods to sort.StringSlice https://golang.org/pkg/sort/#StringSlice
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

func pathIsInDir(p, dir string) bool {
	return strings.HasPrefix(p, dir+string(filepath.Separator))
}

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
		return files, fmt.Errorf("root is not an ancestor of base: %s: %s", root, base)
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

// Returns true if oldfile's modification time is older than newfile's rounded
// to the nearest second.
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
	diff := newtime.Sub(oldtime)
	return diff > 0 && diff.Truncate(1*time.Second) != 0, nil
}

/*
Date/time functions.
*/
func parseDate(text string, loc *time.Location) (time.Time, error) {
	if loc == nil {
		loc, _ = time.LoadLocation("Local")
	}
	// TODO handle other formats to capture time.
	return time.ParseInLocation(time.RFC3339, text[0:10]+"T00:00:00+00:00", loc)
}
