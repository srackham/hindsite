package main

import (
	"os"
)

func main() {
	Cmd = Command{}
	if err := Cmd.Parse(os.Args); err != nil {
		die(err.Error())
	}
	if err := Cmd.Execute(); err != nil {
		die(err.Error())
	}
}
