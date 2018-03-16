package main

import (
	"os"
)

func main() {
	proj := newProject()
	Config = newConfig()
	if err := proj.parseArgs(os.Args); err != nil {
		die(err.Error())
	}
	if err := proj.execute(); err != nil {
		die(err.Error())
	}
}
