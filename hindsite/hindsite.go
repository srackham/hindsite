package main

import (
	"os"
)

func main() {
	proj := newProject()
	if err := proj.parseArgs(os.Args); err != nil {
		proj.die(err.Error())
	}
	if err := proj.execute(); err != nil {
		proj.die(err.Error())
	}
}
