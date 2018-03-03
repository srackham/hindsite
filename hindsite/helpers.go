package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
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

func readFile(name string) string {
	bytes, err := ioutil.ReadFile(name)
	if err != nil {
		die(err.Error())
	}
	return string(bytes)
}

func writeFile(name string, text string) {
	err := ioutil.WriteFile(name, []byte(text), 0644)
	if err != nil {
		die(err.Error())
	}
}

// Return file name sans extension.
func fileName(name string) string {
	return replaceExt(path.Base(name), "")
}

// Replace the extension of name.
func replaceExt(name, ext string) string {
	return name[0:len(name)-len(path.Ext(name))] + ext
}
