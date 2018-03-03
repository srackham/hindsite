package main

import (
	"os"
)

func main() {
	cmd := Command{}
	if !cmd.Parse(os.Args) {
		os.Exit(1)
	}
	cmd.Execute()
}
