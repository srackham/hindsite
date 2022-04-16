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
	var site site
	exec := func(cmd string) (out string, code int) {
		site = NewSite()
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

	/*
		Test help command.
	*/
	out, code := exec("hindsite")
	assert.Equal(0, code)
	assert.Contains(out, "Hindsite is a static website generator")

	out, code = exec("hindsite help")
	assert.Equal(0, code)
	assert.Contains(out, "Hindsite is a static website generator")

	out, code = exec("hindsite help foobar")
	assert.Equal(1, code)
	assert.Contains(out, "illegal command: foobar")

	out, code = exec("hindsite help foo bar")
	assert.Equal(1, code)
	assert.Contains(out, "to many command arguments")

	out, code = exec("hindsite build -site missing")
	assert.Equal(1, code)
	assert.Contains(out, "missing site directory")

	/*
		Test init and build commands.
	*/
	os.RemoveAll(tmpdir)
	fsx.MkMissingDir(tmpdir)
	out, code = exec("hindsite init -site " + tmpdir + " -from blog -v")
	assert.Equal(0, code)
	assert.Contains(out, "installing builtin template: blog")
	assert.Equal(6, fsx.DirCount(filepath.Join(tmpdir, "template")), "unexpected number of files in template directory")
	assert.Equal(7, fsx.DirCount(filepath.Join(tmpdir, "content", "posts")), "unexpected number of files in content/posts directory")
	assert.Equal(0, fsx.DirCount(filepath.Join(tmpdir, "build")), "unexpected number of files in build directory")

	out, code = exec("hindsite build -site " + tmpdir + " -lint")
	assert.Equal(0, code)
	assert.Equal(7, fsx.DirCount(filepath.Join(tmpdir, "build", "posts")), "unexpected number of files in build/posts directory")
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

	out, code = exec("hindsite build -drafts")
	assert.Equal(0, code)
	assert.Equal(7, fsx.DirCount(filepath.Join(tmpdir, "build", "posts")), "unexpected number of files in build/posts directory")
	assert.FileExists(filepath.Join("build", "posts", "2015-05-20", "tincidunt-cursus-pulvinar", "index.html"))
	assert.Contains(out, "documents: 11\ndrafts: 0\nstatic: 7")

	/*
		Test build command -lint option.
	*/
	// The draft document contains invalid links.
	out, code = exec("hindsite build -drafts -lint")
	assert.Equal(1, code)
	assert.Equal(7, fsx.DirCount(filepath.Join(tmpdir, "build", "posts")), "unexpected number of files in build/posts directory")
	assert.Contains(out, `content/posts/links-test.md: contains malformed URL: ":invalid-url"
content/posts/links-test.md: contains link to missing anchor: "#invalid-id"
content/posts/links-test.md: contains link to missing file: "posts/2015-10-13/lorem-penatibus/missing-file.html"
content/posts/links-test.md: contains link to missing anchor: "index.html#invalid-id"
documents: 11
drafts: 0
static: 7
document errors: 4
`)

	/*
		Test the new command
	*/
	out, code = exec("hindsite build")
	assert.Equal(0, code)
	assert.Equal(6, fsx.DirCount(filepath.Join(tmpdir, "build", "posts")), "unexpected number of files in build/posts directory")
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

	out, code = exec("hindsite build -drafts")
	assert.Equal(0, code)
	assert.FileExists(filepath.Join("build", "new-test-file.html"))
	assert.FileExists(filepath.Join("build", "posts", "2018-12-01", "2018-12-01-new-test-file-two", "index.html"))
	assert.Contains(out, "documents: 13\ndrafts: 0\nstatic: 7")

	out, code = exec("hindsite new " + tmpdir + " " + f) // Old v1 command syntax.
	assert.Equal(1, code)
	assert.Contains(out, "to many command arguments")
}
