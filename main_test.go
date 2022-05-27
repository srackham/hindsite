package main

import (
	"os"
	"testing"
)

// TestMain passes unless os.Exit is called with a non-zero exit code.
// See https://tip.golang.org/doc/go1.16#go-test
func TestMain(t *testing.T) {
	defer quiet()()
	os.Args = []string{"hindsite"}
	main()
}

// quiet helper suppress stdout.
func quiet() func() {
	s := os.Stdout
	null, _ := os.Open(os.DevNull)
	os.Stdout = null
	return func() { os.Stdout = s }
}
