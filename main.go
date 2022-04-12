package main

import (
	. "github.com/srackham/hindsite/site"
	"os"
)

func main() {
	site := NewSite()
	os.Exit(site.ExecuteArgs(os.Args))
}
