package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
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

// launchBrowser launches the browser at the url address. Waits till launch
// completed. Credit: https://stackoverflow.com/a/39324149/1136455
func launchBrowser(url string) error {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Run()
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

// fileModTime returns file f's modification time or zero time if it can't.
func fileModTime(f string) time.Time {
	info, err := os.Stat(f)
	if err != nil {
		return time.Time{}
	}
	return info.ModTime()
}

// dirCount returns the number of files and folders in a directory. Returns zero if directory does not exist.
func dirCount(dir string) int {
	if !dirExists(dir) {
		return 0
	}
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	return len(files)
}

/*
Date/time functions.
*/
// Parse date text. If timezone is not specified Local is assumed.
func parseDate(text string, loc *time.Location) (time.Time, error) {
	if loc == nil {
		loc, _ = time.LoadLocation("Local")
	}
	text = strings.TrimSpace(text)
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
