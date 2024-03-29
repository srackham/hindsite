package site

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/srackham/hindsite/v2/fsx"
)

func TestServer(t *testing.T) {
	tmpdir := filepath.Join(os.TempDir(), "hindsite-tests")
	// Initialize temporary directory with test blog.
	os.RemoveAll(tmpdir)
	fsx.MkMissingDir(tmpdir)
	cmd := "hindsite init -site " + tmpdir + " -from ./testdata/blog/template"
	args := strings.Split(cmd, " ")
	site := New()
	err := site.Execute(args)
	if err != nil {
		t.Fatalf("%s: %s", cmd, err.Error())
	}
	if fsx.DirCount(filepath.Join(tmpdir, "template")) != 11 {
		t.Fatalf("%s: unexpected number of files in template directory", cmd)
	}
	// Start server.
	cmd = "hindsite serve -site " + tmpdir
	args = strings.Split(cmd, " ")
	site = New()
	err = site.parseArgs(args)
	if err != nil {
		t.Fatalf("serve error: %v", err.Error())
	}
	site.out = make(chan string, 100)
	site.in = make(chan string, 1)
	svr := newServer(&site)
	go func() {
		if err := svr.serve(); err != nil {
			t.Errorf("serve error: %v", err.Error())
		}
	}()
	waitFor := func(pattern string) {
		for {
			select {
			case line := <-svr.out:
				line = strings.Replace(line, `\`, `/`, -1) // Normalize MS Windows path separators.
				matched, _ := regexp.MatchString(pattern, line)
				if matched {
					return
				}
			case <-time.After(1000 * time.Millisecond):
				t.Fatalf("%s: timed out waiting for: %v", cmd, pattern)
				return
			}
		}
	}
	updateAndWait := func(f, text, pattern string) {
		err := fsx.WriteFile(f, text)
		if err != nil {
			t.Fatal(err)
		}
		waitFor(pattern)
	}
	removeAndWait := func(f, output string) {
		err := os.Remove(f)
		if err != nil {
			t.Fatal(err)
		}
		waitFor(output)
	}
	waitFor("Press the Enter key to print help")
	// Create new post with copy of existing post.
	existingfile := filepath.Join(tmpdir, "content", "posts", "document-3.md")
	text, err := fsx.ReadFile(existingfile)
	if err != nil {
		t.Fatal(err)
	}
	newfile := filepath.Join(tmpdir, "content", "posts", "newfile.md")
	updateAndWait(newfile, text, `".*/content/posts/newfile.md": duplicate document build path in: ".*/content/posts/document-3.md"`)
	// Fix post error.
	text = strings.Replace(text, "slug: sed-sed", "slug: newfile", 1)
	updateAndWait(newfile, text, `updated: ".*/content/posts/newfile.md"`)
	// Change post title.
	text = strings.Replace(text, "title: Sed Sed", "title: New File", 1)
	updateAndWait(newfile, text, `updated: ".*/content/posts/newfile.md"`)
	// Add post tag.
	text = strings.Replace(text, "tags: [integer,est,sed,tincidunt]", "tags: [integer,est,sed,tincidunt,newfile]", 1)
	updateAndWait(newfile, text, `updated: ".*/content/posts/newfile.md"`)
	// Remove post.
	removeAndWait(existingfile, `removed: ".*/content/posts/document-3.md"`)
	// Rebuild.
	svr.in <- "R\n"
	waitFor("rebuilding...")
	waitFor("documents: 11")
	// New static file.
	newfile = filepath.Join(tmpdir, "content", "newfile.txt")
	text = "Hello World!"
	updateAndWait(newfile, text, `updated: ".*/content/newfile.txt"`)
	// Remove static file.
	removeAndWait(newfile, `removed: ".*/content/newfile.txt"`)
	// Stop serve command.
	svr.close(nil)
	time.Sleep(50 * time.Millisecond) // Allow time for serve goroutines to execute cleanup code.
}

// quiet helper suppresses stdout.
func quiet() func() {
	s := os.Stdout
	null, _ := os.Open(os.DevNull)
	os.Stdout = null
	return func() { os.Stdout = s }
}

// Based on https://blog.questionable.services/article/testing-http-handlers-go/
func TestHTTPHandlers(t *testing.T) {
	defer quiet()()
	cmd := "hindsite nop -site ./testdata/blog -content ./testdata/blog/template/init -var urlprefix=http://example.com"
	args := strings.Split(cmd, " ")
	site := New()
	err := site.parseArgs(args)
	if err != nil {
		t.Fatalf("%s: %s", cmd, err.Error())
	}
	err = site.parseConfigFiles()
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}
	svr := newServer(&site)

	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.FileServer(http.Dir(site.buildDir))
	handler = svr.htmlFilter(handler)
	handler = svr.saveBrowserURL(handler)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body is what we expect.
	if svr.browserURL != "/" {
		t.Errorf("saveBrowserURL handler: url: got %v want %v", svr.browserURL, "/")
	}
	wanted := "<script src=\"http://localhost:35729/livereload.js\"></script>"
	got := rr.Body.String()
	if !strings.Contains(got, wanted) {
		t.Errorf("htmlFilter handler: response did not contain: %#v", wanted)
	}
	if strings.Contains(got, site.urlprefix()) {
		t.Errorf("htmlFilter handler: response contains urlprefix: %#v", site.urlprefix())
	}
}
