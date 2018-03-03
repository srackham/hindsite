package main

import (
	"os"
)

func main() {
	cmd := Command{}
	if err := cmd.Parse(os.Args); err != nil {
		die(err.Error())
	}
	if err := cmd.Execute(); err != nil {
		die(err.Error())
	}
}
