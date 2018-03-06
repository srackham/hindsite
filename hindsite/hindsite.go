package main

import (
	"os"
)

func main() {
	Cmd = command{}
	if err := Cmd.Parse(os.Args); err != nil {
		die(err.Error())
	}
	if err := Cmd.Execute(); err != nil {
		die(err.Error())
	}
}
