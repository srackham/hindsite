package main

import (
	"os"
)

func main() {
	proj := newProject()
	os.Exit(execute(&proj, os.Args))
}

// execute runs a hindsite command and returns an exit code.
func execute(proj *project, args []string) int {
	if err := proj.parseArgs(args); err != nil {
		proj.logerror(err.Error())
		return 1
	}
	if err := proj.execute(); err != nil {
		proj.logerror(err.Error())
		return 1
	}
	return 0
}
