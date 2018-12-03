package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_execute(t *testing.T) {
	tmpdir := filepath.Join(os.TempDir(), "hindsite-tests")
	assert := assert.New(t)
	exec := func(cmd string) (out string, code int) {
		proj := newProject()
		proj.out = make(chan string, 100)
		args := strings.Split(cmd, " ")
		code = proj.executeArgs(args)
		close(proj.out)
		for line := range proj.out {
			out += line + "\n"
		}
		return out, code
	}

	out, code := exec("hindsite")
	assert.Equal(0, code)
	assert.Contains(out, "Hindsite is a static website generator")

	out, code = exec("hindsite build missing")
	assert.Equal(1, code)
	assert.Contains(out, "error: missing project directory")

	os.RemoveAll(tmpdir)
	mkMissingDir(tmpdir)
	out, code = exec("hindsite init " + tmpdir + " -builtin blog -v")
	assert.Equal(0, code)
	assert.Contains(out, "installing builtin template: blog")

	out, code = exec("hindsite build " + tmpdir)
	assert.Equal(0, code)
	assert.Contains(out, "documents: 12\ndrafts: 0\nstatic: 6")

	os.RemoveAll(tmpdir)
	mkMissingDir(tmpdir)
	out, code = exec("hindsite init " + tmpdir + " -template ./testdata/blog/template")
	assert.Equal(0, code)
	assert.Contains(out, "")

	out, code = exec("hindsite build " + tmpdir)
	assert.Equal(0, code)
	assert.Contains(out, "documents: 11\ndrafts: 1\nstatic: 7")

	f := filepath.Join(tmpdir, "content", "new-test-file.md")
	out, code = exec("hindsite new " + tmpdir + " " + f)
	assert.Equal(0, code)
	assert.Contains(out, "")
	assert.True(fileExists(f))
	text, _ := readFile(f)
	assert.Contains(text, "title: New Test File")
	assert.Contains(text, "date:  "+time.Now().Format("2006-01-02"))
	assert.Contains(text, "Document content goes here.")

	out, code = exec("hindsite build " + tmpdir)
	assert.Equal(0, code)
	assert.Contains(out, "documents: 12\ndrafts: 2\nstatic: 7")

	f = filepath.Join(tmpdir, "content", "posts", "2018-12-01-new-test-file-two.md")
	out, code = exec("hindsite new " + tmpdir + " " + f)
	assert.Equal(0, code)
	assert.Contains(out, "")
	assert.True(fileExists(f))
	text, _ = readFile(f)
	assert.Contains(text, "title: New Test File Two")
	assert.Contains(text, "date:  2018-12-01")
	assert.Contains(text, "Test new template.")

	out, code = exec("hindsite new " + tmpdir + " " + f)
	assert.Equal(1, code)
	assert.Contains(out, "error: document already exists: content/posts/2018-12-01-new-test-file-two.md")

	out, code = exec("hindsite build " + tmpdir + " -drafts")
	assert.Equal(0, code)
	assert.Contains(out, "documents: 13\ndrafts: 0\nstatic: 7")

}
