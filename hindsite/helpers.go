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
		fmt.Fprintln(os.Stderr, message)
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

func readFile(filename string) string {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		die(err.Error())
	}
	return string(bytes)
}

func writeFile(filename string, text string) {
	err := ioutil.WriteFile(filename, []byte(text), 0644)
	if err != nil {
		die(err.Error())
	}
}

// Replace the extension of filename (ext has leading period).
func replaceExt(filename, ext string) string {
	return filename[0:len(filename)-len(path.Ext(filename))] + ext
}
