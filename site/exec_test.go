package site

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/srackham/hindsite/fsx"
	"github.com/stretchr/testify/assert"
)

func TestExecute(t *testing.T) {
	tmpdir := filepath.Join(os.TempDir(), "hindsite-tests")
	assert := assert.New(t)
	exec := func(cmd string) (out string, code int) {
		site := NewSite()
		site.out = make(chan string, 100)
		args := strings.Split(cmd, " ")
		code = site.ExecuteArgs(args)
		close(site.out)
		for line := range site.out {
			out += line + "\n"
		}
		out = strings.Replace(out, `\`, `/`, -1) // Normalize MS Windows path separators.
		return out, code
	}

	out, code := exec("hindsite")
	assert.Equal(0, code)
	assert.Contains(out, "Hindsite is a static website generator")

	out, code = exec("hindsite build -site missing")
	assert.Equal(1, code)
	assert.Contains(out, "missing site directory")

	os.RemoveAll(tmpdir)
	fsx.MkMissingDir(tmpdir)
	out, code = exec("hindsite init -site " + tmpdir + " -from blog -v")
	assert.Equal(0, code)
	assert.Contains(out, "installing builtin template: blog")

	out, code = exec("hindsite build -site " + tmpdir + " -lint")
	assert.Equal(0, code)
	assert.Contains(out, "documents: 12\ndrafts: 0\nstatic: 6")

	out, code = exec("hindsite build " + tmpdir + " -lint") // Old v1 command syntax.
	assert.Equal(1, code)
	assert.Contains(out, "missing content directory: content")

	os.RemoveAll(tmpdir)
	fsx.MkMissingDir(tmpdir)
	out, code = exec("hindsite init -site " + tmpdir + " -from ./testdata/blog/template -v")
	assert.Equal(0, code)
	assert.Contains(out, "make directory: content/newsletters")

	wd, _ := os.Getwd()
	defer os.Chdir(wd)
	os.Chdir(tmpdir)

	out, code = exec("hindsite build")
	assert.Equal(0, code)
	assert.Contains(out, "documents: 11\ndrafts: 1\nstatic: 7")

	f := filepath.Join(tmpdir, "content", "new-test-file.md")
	out, code = exec("hindsite new -site " + tmpdir + " " + f)
	assert.Equal(0, code)
	assert.Contains(out, "")
	assert.True(fsx.FileExists(f))
	text, _ := fsx.ReadFile(f)
	assert.Contains(text, "title: New Test File")
	assert.Contains(text, "date:  "+time.Now().Format("2006-01-02"))
	assert.Contains(text, "Document content goes here.")

	out, code = exec("hindsite build")
	assert.Equal(0, code)
	assert.Contains(out, "documents: 12\ndrafts: 2\nstatic: 7")

	f = filepath.Join("content", "posts", "2018-12-01-new-test-file-two.md")
	out, code = exec("hindsite new " + f)
	assert.Equal(0, code)
	assert.Contains(out, "")
	assert.True(fsx.FileExists(f))
	text, _ = fsx.ReadFile(f)
	assert.Contains(text, "title: New Test File Two")
	assert.Contains(text, "date:  2018-12-01")
	assert.Contains(text, "Test new template.")

	out, code = exec("hindsite new " + f)
	assert.Equal(1, code)
	assert.Contains(out, "document already exists: content/posts/2018-12-01-new-test-file-two.md")

	out, code = exec("hindsite new " + tmpdir + " " + f) // Old v1 command syntax.
	assert.Equal(1, code)
	assert.Contains(out, "to many command arguments")

	out, code = exec("hindsite build -drafts")
	assert.Equal(0, code)
	assert.Contains(out, "documents: 13\ndrafts: 0\nstatic: 7")

}
