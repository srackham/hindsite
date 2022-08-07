package site

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/srackham/hindsite/v2/assert"
	"github.com/srackham/hindsite/v2/fsx"
)

func TestParseArgs(t *testing.T) {
	var site site
	var err error
	parse := func(cmd string) {
		args := strings.Split(cmd, " ")
		site = New()
		site.confs = append(site.confs, config{})
		err = site.parseArgs(args)
	}

	parse("hindsite serve -site ./testdata/blog -content ./testdata/blog/template/init")
	assert.True(t, err == nil)
	assert.Equal(t, uint16(1212), site.httpport)
	assert.Equal(t, uint16(35729), site.lrport)
	assert.Equal(t, false, site.drafts)
	assert.Equal(t, true, site.livereload)
	assert.Equal(t, false, site.navigate)

	parse("hindsite serve -site ./testdata/blog -content ./testdata/blog/template/init -port 1234")
	assert.True(t, err == nil)
	assert.Equal(t, uint16(1234), site.httpport)

	parse("hindsite serve -site ./testdata/blog -content ./testdata/blog/template/init -port 1234:8000 -drafts")
	assert.True(t, err == nil)
	assert.Equal(t, uint16(1234), site.httpport)
	assert.Equal(t, uint16(8000), site.lrport)
	assert.Equal(t, true, site.drafts)

	parse("hindsite serve -port :-1 -drafts -navigate -site ./testdata/blog -content ./testdata/blog/template/init")
	assert.True(t, err == nil)
	assert.Equal(t, true, site.drafts)
	assert.Equal(t, false, site.livereload)
	assert.Equal(t, true, site.navigate)

	parse("hindsite illegal-command")
	assert.Equal(t, `illegal command: "illegal-command"`, err.Error())

	parse("hindsite serve -illegal-option")
	assert.Equal(t, `illegal option: "-illegal-option"`, err.Error())

	parse("hindsite serve -site missing-site-dir")
	assert.Contains(t, err.Error(), "missing site directory: ")

	parse("hindsite serve -site . -content missing-content-dir")
	assert.Contains(t, err.Error(), "missing content directory: ")

	parse("hindsite serve -port")
	assert.Equal(t, "missing -port argument value", err.Error())

	parse("hindsite serve -site ./testdata/blog -port 99999999")
	assert.Equal(t, `illegal -port: "99999999"`, err.Error())

	parse("hindsite serve -site ./testdata/blog -port :99999999")
	assert.Equal(t, `illegal -port: ":99999999"`, err.Error())

	// -var option checks.
	parse("hindsite build -site ./testdata/blog -var foobar")
	assert.Equal(t, `illegal -var syntax: "foobar"`, err.Error())

	parse("hindsite build -site ./testdata/blog -var foobar=42")
	assert.Equal(t, `illegal -var name: "foobar"`, err.Error())

	parse("hindsite build -site ./testdata/blog -var author=Joe-Bloggs")
	assert.Equal(t, "Joe-Bloggs", *site.vars.Author)

	parse("hindsite build -site ./testdata/blog -var exclude=*.tmp")
	assert.Equal(t, "*.tmp", *site.vars.Exclude)

	parse("hindsite build -site ./testdata/blog -var homepage=posts/docs-1.html")
	assert.Equal(t, "posts/docs-1.html", *site.vars.Homepage)

	parse("hindsite build -site ./testdata/blog -var id=1234")
	assert.Equal(t, "1234", *site.vars.ID)

	parse("hindsite build -site ./testdata/blog -var include=test.tmp")
	assert.Equal(t, "test.tmp", *site.vars.Include)

	parse("hindsite build -site ./testdata/blog -var longdate=Mon-Jan-2-2006")
	assert.Equal(t, "Mon-Jan-2-2006", *site.vars.LongDate)

	parse("hindsite build -site ./testdata/blog -var mediumdate=2_Jan_2006")
	assert.Equal(t, "2_Jan_2006", *site.vars.MediumDate)

	parse("hindsite build -site ./testdata/blog -var paginate=qux")
	assert.Equal(t, `illegal paginate value: "qux"`, err.Error())

	parse("hindsite build -site ./testdata/blog -var paginate=42")
	assert.Equal(t, 42, *site.vars.Paginate)

	parse("hindsite build -site ./testdata/blog -var permalink=/posts/%y/%m/%d/%f")
	assert.Equal(t, "/posts/%y/%m/%d/%f", *site.vars.Permalink)

	parse("hindsite build -site ./testdata/blog -var shortdate=2/1/2006")
	assert.Equal(t, "2/1/2006", *site.vars.ShortDate)

	parse("hindsite build -site ./testdata/blog -var templates=*.md|*.txt|*.rmu")
	assert.Equal(t, "*.md|*.txt|*.rmu", *site.vars.Templates)

	parse("hindsite build -site ./testdata/blog -var author=Joe-Bloggs")
	assert.Equal(t, "Joe-Bloggs", *site.vars.Author)

	parse("hindsite build -site ./testdata/blog -var timezone=+0600")
	assert.Equal(t, "+0600", *site.vars.Timezone)

	parse("hindsite build -site ./testdata/blog -var urlprefix=https://foobar")
	assert.Equal(t, "https://foobar", *site.vars.URLPrefix)

	parse("hindsite build -site ./testdata/blog -var user.foo=bar")
	assert.Equal(t, "bar", site.vars.User["foo"])

	// Configuration variable checks.
	// TODO refactor to TestMergeRaw() and add test cases
	config := config{}
	s := "/"
	err = config.mergeRaw(rawConfig{URLPrefix: &s})
	assert.Equal(t, `illegal urlprefix: "/"`, err.Error())

	// -config option checks.
	parse("hindsite build -site ./testdata/blog -content ./testdata/blog/template/init -var user.foo=bar -config ./testdata/blog/template/config2.toml")
	assert.True(t, err == nil)
	assert.Equal(t, "Bill Blow", *site.vars.Author)
	assert.Equal(t, "qux", site.vars.User["baz"])
	assert.Equal(t, "qux2", site.vars.User["foo"])

	parse("hindsite build -site ./testdata/blog -content ./testdata/blog/template/init -config ./testdata/blog/template/config2.toml -config ./testdata/blog/template/config2.yaml")
	assert.True(t, err == nil)
	assert.Equal(t, "Bill Blow", *site.vars.Author)
	assert.Equal(t, "qux", site.vars.User["baz"])
	assert.Equal(t, "qux3", site.vars.User["foo"])
}

func TestExecuteArgs(t *testing.T) {
	tmpdir := filepath.Join(os.TempDir(), "hindsite-tests")
	var site site
	exec := func(cmd string) (out string, err error) {
		args := strings.Split(cmd, " ")
		site = New()
		site.out = make(chan string, 100)
		err = site.Execute(args)
		close(site.out)
		for line := range site.out {
			out += line + "\n"
		}
		out = strings.Replace(out, `\`, `/`, -1) // Normalize MS Windows path separators.
		return
	}

	/*
		Test help command.
	*/
	out, err := exec("hindsite")
	assert.True(t, err == nil)
	assert.Contains(t, out, "Hindsite is a static website generator")

	out, err = exec("hindsite help")
	assert.True(t, err == nil)
	assert.Contains(t, out, "Hindsite is a static website generator")

	out, err = exec("hindsite help foobar")
	assert.True(t, err != nil)
	assert.Contains(t, out, `illegal command: "foobar"`)

	out, err = exec("hindsite help foo bar")
	assert.True(t, err != nil)
	assert.Contains(t, out, "to many command arguments")

	/*
		Test init and build commands.
	*/
	buildSiteFrom := func(from string, buildmsg string, templateCount int, contentCount int, buildCount int) {
		t.Helper()
		os.RemoveAll(tmpdir)
		fsx.MkMissingDir(tmpdir)
		cmd := "hindsite init -site " + tmpdir + " -from " + from + " -v"
		out, err = exec(cmd)
		assert.PassIf(t, err == nil, "unexpected error: \""+cmd+"\"")
		assert.PassIf(t, templateCount == fsx.DirCount(filepath.Join(tmpdir, "template")), from+": unexpected number of files in template directory")
		assert.PassIf(t, contentCount == fsx.DirCount(filepath.Join(tmpdir, "content")), from+": unexpected number of files in content directory")
		assert.PassIf(t, fsx.DirCount(filepath.Join(tmpdir, "build")) == 0, from+": unexpected number of files in build directory")
		cmd = "hindsite build -site " + tmpdir + " -lint -v"
		out, err = exec(cmd)
		if err != nil {
			t.Fatalf("unexpected error executing \"%s\": %s", cmd, err.Error())
		}
		assert.PassIf(t, buildCount == fsx.DirCount(filepath.Join(tmpdir, "build")), from+": unexpected number of files in build directory")
		assert.Contains(t, out, buildmsg)
	}

	out, err = exec("hindsite build -site MISSING")
	assert.True(t, err != nil)
	assert.Contains(t, out, "missing site directory")

	out, err = exec("hindsite build " + tmpdir) // Old v1 command syntax.
	assert.True(t, err != nil)
	assert.ContainsPattern(t, out, `missing content directory: ".*/content"`)

	// Test built-in templates.
	buildSiteFrom("hello", "documents: 1\nstatic: 0", 2, 1, 1)
	buildSiteFrom("blog", "documents: 12\nstatic: 6", 6, 7, 9)
	assert.PassIf(t, fsx.DirCount(filepath.Join(tmpdir, "content", "posts")) == 7, "unexpected number of files in content/posts directory")
	buildSiteFrom("docs", "documents: 4\nstatic: 3", 4, 7, 7)

	/*
		Initialise and build the testdata site for subsequent tests.
	*/
	buildSiteFrom("./testdata/blog/template", "documents: 11\nstatic: 7", 11, 8, 9)

	/*
		Validate the checksums of the test site's built HTML files.
		The checksums.txt file is built with the Makefile make-checksums task.
	*/
	text, err := fsx.ReadFile("./testdata/blog/checksums.txt")
	assert.True(t, err == nil)
	lines := strings.Split(strings.TrimSpace(text), "\n")
	for _, line := range lines {
		sum, f, _ := strings.Cut(line, " ")
		f = filepath.Join(tmpdir, strings.TrimSpace(f))
		if !fsx.FileExists(f) {
			t.Logf("missing file: \"%s\"", f)
			t.Fail()
			continue
		}
		if text, err = fsx.ReadFile(f); err != nil {
			t.Logf("error reading file: \"%s\": %s", f, err.Error())
			t.Fail()
			continue
		}
		text = normalizeNewlines(text)
		sha256 := fmt.Sprintf("%x", sha256.Sum256([]byte(text)))
		if sha256 != sum {
			t.Logf("invalid checksum for: \"%s\"", f)
			t.Fail()
		}
	}

	// NOTE: From here on the tests are performed in the tmp directory on the testdata site.
	wd, _ := os.Getwd()
	defer os.Chdir(wd)
	os.Chdir(tmpdir)

	/*
		Miscellaneous tests.
	*/
	assert.PassIf(t, fsx.FileExists(filepath.Join("build", "posts", "2016-08-05", "slug-test", "index.html")), "content/posts/document-5.md slug test")

	/*
		Test drafts generation.
	*/
	out, err = exec("hindsite build -drafts")
	assert.True(t, err == nil)
	assert.PassIf(t, fsx.DirCount(filepath.Join(tmpdir, "build", "posts")) == 7, "unexpected number of files in build/posts directory")
	assert.True(t, fsx.FileExists(filepath.Join("build", "posts", "2015-05-20", "document-1", "index.html")))
	assert.Contains(t, out, "documents: 11\nstatic: 7")

	/*
		Test build command -lint option.
	*/
	// The draft document contains invalid links.
	out, err = exec("hindsite build -drafts -lint")
	assert.True(t, err != nil)
	assert.PassIf(t, fsx.DirCount(filepath.Join(tmpdir, "build", "posts")) == 7, "unexpected number of files in build/posts directory")
	assert.ContainsPattern(t, out, `".*/content/posts/links-test.md": contains illicit element id: "-illicit-id"`)
	assert.ContainsPattern(t, out, `".*/content/posts/links-test.md": contains duplicate element id: "id2"`)
	assert.ContainsPattern(t, out, `".*/content/posts/links-test.md": contains illicit URL: ":invalid-url"`)
	assert.ContainsPattern(t, out, `".*/content/posts/links-test.md": contains link to missing anchor: "#invalid-id"`)
	assert.ContainsPattern(t, out, `".*/content/posts/links-test.md": contains link to missing file: ".*/posts/2015-10-13/LOREM-PENATIBUS/missing-file.html"`)
	assert.ContainsPattern(t, out, `".*/content/posts/links-test.md": contains link to missing file: ".*/missing-file-2.html"`)
	assert.ContainsPattern(t, out, `".*/content/posts/links-test.md": contains link to missing anchor: "/index.html#invalid-id"`)
	assert.ContainsPattern(t, out, `unhygienic document URL path: ".*/posts/2015-10-13/LOREM-PENATIBUS/"`)
	assert.ContainsPattern(t, out, `unhygienic document URL path: ".*/newsletters/slug with spaces.html"`)
	assert.Contains(t, out, `documents: 11`)
	assert.Contains(t, out, `static: 7`)
	assert.Contains(t, out, `errors: 7`)
	assert.Contains(t, out, `warnings: 6`)
	assert.Contains(t, out, `root config variable "homepage" in non-root config file`)
	assert.Contains(t, out, `root config variable "urlprefix" in non-root config file`)
	assert.Contains(t, out, `root config variable "exclude" in non-root config file`)
	assert.Contains(t, out, `root config variable "include" in non-root config file`)

	/*
		Test the new command
	*/
	out, err = exec("hindsite build")
	assert.True(t, err == nil)
	assert.PassIf(t, fsx.DirCount(filepath.Join(tmpdir, "build", "posts")) == 6, "unexpected number of files in build/posts directory")
	assert.Contains(t, out, "documents: 11\nstatic: 7")

	f := filepath.Join(tmpdir, "content", "new-test-file.md")
	out, err = exec("hindsite new " + tmpdir + " " + f) // Old v1 command syntax.
	assert.True(t, err != nil)
	assert.Contains(t, out, "to many command arguments")

	f = filepath.Join(tmpdir, "content", "new-test-file.md")
	out, err = exec("hindsite new -site " + tmpdir + " " + f)
	assert.True(t, err == nil)
	assert.Contains(t, out, "")
	assert.True(t, fsx.FileExists(f))
	text, err = fsx.ReadFile(f)
	assert.True(t, err == nil)
	assert.Contains(t, text, "title: New Test File")
	assert.Contains(t, text, "date:  "+time.Now().Format("2006-01-02"))
	assert.Contains(t, text, "Document content goes here.")

	out, err = exec("hindsite build")
	assert.True(t, err == nil)
	assert.Contains(t, out, "documents: 12\nstatic: 7")

	f = filepath.Join("content", "posts", "2018-12-01-new-test-file-two.md")
	out, err = exec("hindsite new " + f)
	assert.True(t, err == nil)
	assert.Contains(t, out, "")
	assert.True(t, fsx.FileExists(f))
	text, err = fsx.ReadFile(f)
	assert.True(t, err == nil)
	assert.Contains(t, text, "title: New Test File Two")
	assert.Contains(t, text, "date:  2018-12-01")
	assert.Contains(t, text, "Test new template.")

	out, err = exec("hindsite new " + f)
	assert.True(t, err != nil)
	assert.Contains(t, out, `document already exists: "content/posts/2018-12-01-new-test-file-two.md"`)

	out, err = exec("hindsite build -drafts")
	assert.True(t, err == nil)
	assert.True(t, fsx.FileExists(filepath.Join("build", "new-test-file.html")))
	assert.True(t, fsx.FileExists(filepath.Join("build", "posts", "2018-12-01", "2018-12-01-new-test-file-two", "index.html")))
	assert.Contains(t, out, "documents: 13\nstatic: 7")

	os.Remove(f)
	out, err = exec("hindsite new -from foobar " + f)
	assert.True(t, err != nil)
	assert.Contains(t, out, `missing document template file: "foobar"`)

	f = filepath.Join("content", "posts", "2018-12-01-new-test-file-two.md")
	template := filepath.Join("template", "posts", "new.md")
	out, err = exec("hindsite new -from " + template + " " + f)
	assert.True(t, err == nil)
	assert.True(t, fsx.FileExists(f))
	text, err = fsx.ReadFile(f)
	assert.True(t, err == nil)
	assert.Contains(t, text, "title: New Test File Two")
	assert.Contains(t, text, "date:  2018-12-01")
	assert.Contains(t, text, "Test new template.")

	/*
		Test document variables.
	*/
	f = filepath.Join("build", "index.html")
	out, err = exec("hindsite build")
	assert.True(t, err == nil)
	text, err = fsx.ReadFile(f)
	assert.True(t, err == nil)
	assert.Contains(t, text, `.author=Joe Bloggs
.description=&lt;p&gt;The about document.&lt;/p&gt;

.id=/about.html
.layout=layout.html
.longdate=May 21, 2015
.mediumdate=21-May-2015
.shortdate=2015-05-21
.date.Format &quot;Monday, 02-Jan-06 15:04:05&quot;=Wednesday, 20-May-15 12:15:23
.permalink=
.slug=
.templates=*
.tags=[]
.title=About Test
.url=/about.html
.urlprefix=http://example.com
.user=map[banner:hindsite | blog highlightjs:yes]`)
}
