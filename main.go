package main

import (
	"os"
)

func main() {
	proj := newProject()
	os.Exit(proj.executeArgs(os.Args))
}
