package main

import (
	"github.com/srackham/hindsite/site"
	"os"
)

func main() {
	site := site.NewSite()
	os.Exit(site.ExecuteArgs(os.Args))
}
