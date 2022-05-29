package main

import (
	"os"

	"github.com/srackham/hindsite/v2/site"
)

func main() {
	site := site.New()
	if err := site.Execute(os.Args); err != nil {
		os.Exit(1)
	}
}
