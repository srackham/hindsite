package main

import (
	"os"
	"testing"
)

// TestMain passes unless os.Exit is called with a non-zero exit code.
// See https://tip.golang.org/doc/go1.16#go-test
func TestMain(t *testing.T) {
	os.Args = []string{"hindsite"}
	main()
}
