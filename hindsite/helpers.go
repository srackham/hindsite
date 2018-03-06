package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"
)

// Helpers.
func die(message string) {
	if message != "" {
		fmt.Fprintln(os.Stderr, "error: "+message)
	}
	os.Exit(1)
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
	return replaceExt(path.Base(name), "")
}

// Replace the extension of name.
func replaceExt(name, ext string) string {
	return name[0:len(name)-len(path.Ext(name))] + ext
}

// TODO return error.
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
