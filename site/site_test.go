package site

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/srackham/hindsite/fsx"
	"github.com/stretchr/testify/assert"
)

func TestParseArgs(t *testing.T) {
	assert := assert.New(t)

	var site site
	var err error
	parse := func(cmd string) {
		args := strings.Split(cmd, " ")
		site = New()
		site.confs = append(site.confs, config{})
		err = site.parseArgs(args)
	}

	parse("hindsite serve -site ./testdata/blog -content ./testdata/blog/template/init")
	assert.NoError(err)
	assert.Equal(uint16(1212), site.httpport, "httpport")
	assert.Equal(uint16(35729), site.lrport, "lrport")
	assert.Equal(false, site.drafts, "drafts")
	assert.Equal(true, site.livereload, "livereload")
	assert.Equal(false, site.navigate, "navigate")

	parse("hindsite serve -site ./testdata/blog -content ./testdata/blog/template/init -port 1234")
	assert.NoError(err)
	assert.Equal(uint16(1234), site.httpport, "httpport")

	parse("hindsite serve -site ./testdata/blog -content ./testdata/blog/template/init -port 1234:8000 -drafts")
	assert.NoError(err)
	assert.Equal(uint16(1234), site.httpport, "httpport")
	assert.Equal(uint16(8000), site.lrport, "lrport")
	assert.Equal(true, site.drafts, "drafts")

	parse("hindsite serve -port :-1 -drafts -navigate -site ./testdata/blog -content ./testdata/blog/template/init")
	assert.NoError(err)
	assert.Equal(true, site.drafts, "drafts")
	assert.Equal(false, site.livereload, "livereload")
	assert.Equal(true, site.navigate, "navigate")

	parse("hindsite illegal-command")
	assert.Equal(`illegal command: "illegal-command"`, err.Error())

	parse("hindsite serve -illegal-option")
	assert.Equal(`illegal option: "-illegal-option"`, err.Error())

	parse("hindsite serve -site missing-site-dir")
	assert.Contains(err.Error(), "missing site directory: ")

	parse("hindsite serve -site . -content missing-content-dir")
	assert.Contains(err.Error(), "missing content directory: ")

	parse("hindsite serve -port")
	assert.Equal("missing -port argument value", err.Error())

	parse("hindsite serve -site ./testdata/blog -port 99999999")
	assert.Equal(`illegal -port: "99999999"`, err.Error())

	parse("hindsite serve -site ./testdata/blog -port :99999999")
	assert.Equal(`illegal -port: ":99999999"`, err.Error())

	// -var option checks.
	parse("hindsite build -site ./testdata/blog -var foobar")
	assert.Equal(`illegal -var syntax: "foobar"`, err.Error())

	parse("hindsite build -site ./testdata/blog -var foobar=42")
	assert.Equal(`illegal -var name: "foobar"`, err.Error())

	parse("hindsite build -site ./testdata/blog -var author=Joe-Bloggs")
	assert.Equal("Joe-Bloggs", *site.vars.Author)

	parse("hindsite build -site ./testdata/blog -var exclude=*.tmp")
	assert.Equal("*.tmp", *site.vars.Exclude)

	parse("hindsite build -site ./testdata/blog -var homepage=posts/docs-1.html")
	assert.Equal("posts/docs-1.html", *site.vars.Homepage)

	parse("hindsite build -site ./testdata/blog -var id=1234")
	assert.Equal("1234", *site.vars.ID)

	parse("hindsite build -site ./testdata/blog -var include=test.tmp")
	assert.Equal("test.tmp", *site.vars.Include)

	parse("hindsite build -site ./testdata/blog -var longdate=Mon-Jan-2-2006")
	assert.Equal("Mon-Jan-2-2006", *site.vars.LongDate)

	parse("hindsite build -site ./testdata/blog -var mediumdate=2_Jan_2006")
	assert.Equal("2_Jan_2006", *site.vars.MediumDate)

	parse("hindsite build -site ./testdata/blog -var paginate=qux")
	assert.Equal(`illegal paginate value: "qux"`, err.Error())

	parse("hindsite build -site ./testdata/blog -var paginate=42")
	assert.Equal(42, *site.vars.Paginate)

	parse("hindsite build -site ./testdata/blog -var permalink=/posts/%y/%m/%d/%f")
	assert.Equal("/posts/%y/%m/%d/%f", *site.vars.Permalink)

	parse("hindsite build -site ./testdata/blog -var shortdate=2/1/2006")
	assert.Equal("2/1/2006", *site.vars.ShortDate)

	parse("hindsite build -site ./testdata/blog -var templates=*.md|*.txt|*.rmu")
	assert.Equal("*.md|*.txt|*.rmu", *site.vars.Templates)

	parse("hindsite build -site ./testdata/blog -var author=Joe-Bloggs")
	assert.Equal("Joe-Bloggs", *site.vars.Author)

	parse("hindsite build -site ./testdata/blog -var timezone=+0600")
	assert.Equal("+0600", *site.vars.Timezone)

	parse("hindsite build -site ./testdata/blog -var urlprefix=https://foobar")
	assert.Equal("https://foobar", *site.vars.URLPrefix)

	parse("hindsite build -site ./testdata/blog -var user.foo=bar")
	assert.Equal("bar", site.vars.User["foo"])

	// Configuration variable checks.
	// TODO refactor to TestMergeRaw() and add test cases
	config := config{}
	s := "/"
	err = config.mergeRaw(rawConfig{URLPrefix: &s})
	assert.Equal(`illegal urlprefix: "/"`, err.Error())

	// -config option checks.
	parse("hindsite build -site ./testdata/blog -content ./testdata/blog/template/init -var user.foo=bar -config ./testdata/blog/template/config2.toml")
	assert.NoError(err)
	assert.Equal("Bill Blow", *site.vars.Author)
	assert.Equal("qux", site.vars.User["baz"])
	assert.Equal("qux2", site.vars.User["foo"])

	parse("hindsite build -site ./testdata/blog -content ./testdata/blog/template/init -config ./testdata/blog/template/config2.toml -config ./testdata/blog/template/config2.yaml")
	assert.NoError(err)
	assert.Equal("Bill Blow", *site.vars.Author)
	assert.Equal("qux", site.vars.User["baz"])
	assert.Equal("qux3", site.vars.User["foo"])
}

func TestExecuteArgs(t *testing.T) {
	tmpdir := filepath.Join(os.TempDir(), "hindsite-tests")
	assert := assert.New(t)
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
	assertContainsPattern := func(s string, pattern string) {
		matched, _ := regexp.MatchString(pattern, s)
		assert.True(matched, "%q does not contain pattern %q", s, pattern)
	}

	/*
		Test help command.
	*/
	out, err := exec("hindsite")
	assert.NoError(err)
	assert.Contains(out, "Hindsite is a static website generator")

	out, err = exec("hindsite help")
	assert.NoError(err)
	assert.Contains(out, "Hindsite is a static website generator")

	out, err = exec("hindsite help foobar")
	assert.Error(err)
	assert.Contains(out, `illegal command: "foobar"`)

	out, err = exec("hindsite help foo bar")
	assert.Error(err)
	assert.Contains(out, "to many command arguments")

	/*
		Test init and build commands.
	*/
	buildSiteFrom := func(from string, buildmsg string, templateCount int, contentCount int, buildCount int) {
		os.RemoveAll(tmpdir)
		fsx.MkMissingDir(tmpdir)
		cmd := "hindsite init -site " + tmpdir + " -from " + from + " -v"
		out, err = exec(cmd)
		assert.NoError(err, "unexpected error: \""+cmd+"\"")
		assert.Equal(templateCount, fsx.DirCount(filepath.Join(tmpdir, "template")), from+": unexpected number of files in template directory")
		assert.Equal(contentCount, fsx.DirCount(filepath.Join(tmpdir, "content")), from+": unexpected number of files in content directory")
		assert.Equal(0, fsx.DirCount(filepath.Join(tmpdir, "build")), from+": unexpected number of files in build directory")
		cmd = "hindsite build -site " + tmpdir + " -lint -v"
		out, err = exec(cmd)
		if err != nil {
			t.Fatalf("unexpected error executing \"%s\": %s", cmd, err.Error())
		}
		assert.Equal(buildCount, fsx.DirCount(filepath.Join(tmpdir, "build")), from+": unexpected number of files in build directory")
		assert.Contains(out, buildmsg)
	}

	out, err = exec("hindsite build -site MISSING")
	assert.Error(err)
	assert.Contains(out, "missing site directory")

	out, err = exec("hindsite build " + tmpdir) // Old v1 command syntax.
	assert.Error(err)
	assertContainsPattern(out, `missing content directory: ".*/content"`)

	// Test built-in templates.
	buildSiteFrom("hello", "documents: 1\nstatic: 0", 2, 1, 1)
	buildSiteFrom("blog", "documents: 12\nstatic: 6", 6, 7, 9)
	assert.Equal(7, fsx.DirCount(filepath.Join(tmpdir, "content", "posts")), "unexpected number of files in content/posts directory")
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
	assert.NoError(err)
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
	assert.FileExists(filepath.Join("build", "posts", "2016-08-05", "slug-test", "index.html"), "content/posts/document-5.md slug test")

	/*
		Test drafts generation.
	*/
	out, err = exec("hindsite build -drafts")
	assert.NoError(err)
	assert.Equal(7, fsx.DirCount(filepath.Join(tmpdir, "build", "posts")), "unexpected number of files in build/posts directory")
	assert.FileExists(filepath.Join("build", "posts", "2015-05-20", "document-1", "index.html"))
	assert.Contains(out, "documents: 11\nstatic: 7")

	/*
		Test build command -lint option.
	*/
	// The draft document contains invalid links.
	out, err = exec("hindsite build -drafts -lint")
	assert.Error(err)
	assert.Equal(7, fsx.DirCount(filepath.Join(tmpdir, "build", "posts")), "unexpected number of files in build/posts directory")
	assertContainsPattern(out, `".*/content/posts/links-test.md": contains illicit element id: "-illicit-id"`)
	assertContainsPattern(out, `".*/content/posts/links-test.md": contains duplicate element id: "id2"`)
	assertContainsPattern(out, `".*/content/posts/links-test.md": contains illicit URL: ":invalid-url"`)
	assertContainsPattern(out, `".*/content/posts/links-test.md": contains link to missing anchor: "#invalid-id"`)
	assertContainsPattern(out, `".*/content/posts/links-test.md": contains link to missing file: ".*/posts/2015-10-13/LOREM-PENATIBUS/missing-file.html"`)
	assertContainsPattern(out, `".*/content/posts/links-test.md": contains link to missing file: ".*/missing-file-2.html"`)
	assertContainsPattern(out, `".*/content/posts/links-test.md": contains link to missing anchor: "/index.html#invalid-id"`)
	assertContainsPattern(out, `unhygienic document URL path: ".*/posts/2015-10-13/LOREM-PENATIBUS/"`)
	assertContainsPattern(out, `unhygienic document URL path: ".*/newsletters/slug with spaces.html"`)
	assert.Contains(out, `documents: 11`)
	assert.Contains(out, `static: 7`)
	assert.Contains(out, `errors: 7`)
	assert.Contains(out, `warnings: 6`)
	assert.Contains(out, `root config variable "homepage" in non-root config file`)
	assert.Contains(out, `root config variable "urlprefix" in non-root config file`)
	assert.Contains(out, `root config variable "exclude" in non-root config file`)
	assert.Contains(out, `root config variable "include" in non-root config file`)

	/*
		Test the new command
	*/
	out, err = exec("hindsite build")
	assert.NoError(err)
	assert.Equal(6, fsx.DirCount(filepath.Join(tmpdir, "build", "posts")), "unexpected number of files in build/posts directory")
	assert.Contains(out, "documents: 11\nstatic: 7")

	f := filepath.Join(tmpdir, "content", "new-test-file.md")
	out, err = exec("hindsite new " + tmpdir + " " + f) // Old v1 command syntax.
	assert.Error(err)
	assert.Contains(out, "to many command arguments")

	f = filepath.Join(tmpdir, "content", "new-test-file.md")
	out, err = exec("hindsite new -site " + tmpdir + " " + f)
	assert.NoError(err)
	assert.Contains(out, "")
	assert.True(fsx.FileExists(f))
	text, err = fsx.ReadFile(f)
	assert.NoError(err)
	assert.Contains(text, "title: New Test File")
	assert.Contains(text, "date:  "+time.Now().Format("2006-01-02"))
	assert.Contains(text, "Document content goes here.")

	out, err = exec("hindsite build")
	assert.NoError(err)
	assert.Contains(out, "documents: 12\nstatic: 7")

	f = filepath.Join("content", "posts", "2018-12-01-new-test-file-two.md")
	out, err = exec("hindsite new " + f)
	assert.NoError(err)
	assert.Contains(out, "")
	assert.True(fsx.FileExists(f))
	text, err = fsx.ReadFile(f)
	assert.NoError(err)
	assert.Contains(text, "title: New Test File Two")
	assert.Contains(text, "date:  2018-12-01")
	assert.Contains(text, "Test new template.")

	out, err = exec("hindsite new " + f)
	assert.Error(err)
	assert.Contains(out, `document already exists: "content/posts/2018-12-01-new-test-file-two.md"`)

	out, err = exec("hindsite build -drafts")
	assert.NoError(err)
	assert.FileExists(filepath.Join("build", "new-test-file.html"))
	assert.FileExists(filepath.Join("build", "posts", "2018-12-01", "2018-12-01-new-test-file-two", "index.html"))
	assert.Contains(out, "documents: 13\nstatic: 7")

	os.Remove(f)
	out, err = exec("hindsite new -from foobar " + f)
	assert.Error(err)
	assert.Contains(out, `missing document template file: "foobar"`)

	f = filepath.Join("content", "posts", "2018-12-01-new-test-file-two.md")
	template := filepath.Join("template", "posts", "new.md")
	out, err = exec("hindsite new -from " + template + " " + f)
	assert.NoError(err)
	assert.True(fsx.FileExists(f))
	text, err = fsx.ReadFile(f)
	assert.NoError(err)
	assert.Contains(text, "title: New Test File Two")
	assert.Contains(text, "date:  2018-12-01")
	assert.Contains(text, "Test new template.")

	/*
		Test document variables.
	*/
	f = filepath.Join("build", "index.html")
	out, err = exec("hindsite build")
	assert.NoError(err)
	text, err = fsx.ReadFile(f)
	assert.NoError(err)
	assert.Contains(text, `.author=Joe Bloggs
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
