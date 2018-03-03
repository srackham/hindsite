package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

// Helpers.
func die(message string) {
	if message != "" {
		fmt.Fprintln(os.Stderr, message)
	}
	os.Exit(1)
}

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
