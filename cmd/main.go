package main

import (
	"os"
)

func main() {
	site := newSite()
	os.Exit(site.executeArgs(os.Args))
}
