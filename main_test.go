package main

import (
	"os"
	"path"
	"strings"
	"testing"
)

func Test_execute(t *testing.T) {
	type test struct {
		name string
		proj project
		cmd  string
		code int
		out  string
	}
	tmpdir := path.Join(os.TempDir(), "hindsite-tests")
	tests := []test{
		{
			"help command",
			newProject(),
			"hindsite",
			0,
			"Hindsite is a static website generator",
		},
		{
			"build missing",
			newProject(),
			"hindsite build missing",
			1,
			"error: missing project directory",
		},
		{
			"build blog",
			newProject(),
			"hindsite build ./testdata/blog -build " + tmpdir,
			0,
			"documents: 11\ndrafts: 0\nstatic: 6",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.RemoveAll(tmpdir)
			mkMissingDir(tmpdir)
			tt.proj.out = make(chan string, 100)
			args := strings.Split(tt.cmd, " ")
			code := execute(&tt.proj, args)
			close(tt.proj.out)
			var out string
			for line := range tt.proj.out {
				out += line + "\n"
			}
			if code != tt.code {
				t.Errorf("%s: exit code: got %v, want %v", tt.name, code, tt.code)
			}
			if !strings.Contains(out, tt.out) {
				t.Errorf("%s: output does not contain: %v", tt.name, tt.out)
			}
		})
	}
}
