package main

import (
	"os"
	"path"
	"strings"
	"testing"
	"time"
)

func Test_serve(t *testing.T) {
	t.Run("serve", func(t *testing.T) {
		tmpdir := path.Join(os.TempDir(), "hindsite-tests")
		os.RemoveAll(tmpdir)
		mkMissingDir(tmpdir)
		proj := newProject()
		cmd := "hindsite init " + tmpdir + " -template ./testdata/blog/template"
		args := strings.Split(cmd, " ")
		code := execute(&proj, args)
		if code != 0 {
			t.Errorf("%s", cmd)
		}
		cmd = "hindsite serve " + tmpdir
		args = strings.Split(cmd, " ")
		if err := proj.parseArgs(args); err != nil {
			t.Errorf("%s: %v", cmd, err)
		}
		waitFor := func(output string) {
			for {
				select {
				case line := <-proj.out:
					if strings.Contains(line, output) {
						return
					}
				case <-time.After(300 * time.Millisecond):
					t.Errorf("%s: timed out waiting for: %v", cmd, output)
					return
				}
			}
		}
		updateAndWait := func(f, text, output string) {
			err := writeFile(f, text)
			if err != nil {
				t.Error(err)
			}
			waitFor(output)
		}
		removeAndWait := func(f, output string) {
			err := os.Remove(f)
			if err != nil {
				t.Error(err)
			}
			waitFor(output)
		}
		proj.out = make(chan string, 100)
		proj.in = make(chan string)
		proj.quit = make(chan error)
		// Start server.
		go func() { proj.serve() }()
		waitFor("Press Ctrl+C to stop")
		// Create new post with copy of existing post.
		existingfile := path.Join(tmpdir, "content", "posts", "2016-10-18-sed-sed.md")
		text, err := readFile(existingfile)
		if err != nil {
			t.Error(err)
		}
		newfile := path.Join(tmpdir, "content", "posts", "newfile.md")
		updateAndWait(newfile, text, "error: content/posts/newfile.md: duplicate document build path in: content/posts/2016-10-18-sed-sed.md")
		// Fix post error.
		text = strings.Replace(text, "slug: sed-sed", "slug: newfile", 1)
		updateAndWait(newfile, text, "updated: content/posts/newfile.md")
		// Change post title.
		text = strings.Replace(text, "title: Sed Sed", "title: New File", 1)
		updateAndWait(newfile, text, "updated: content/posts/newfile.md")
		// Add post tag.
		text = strings.Replace(text, "tags: [integer,est,sed,tincidunt]", "tags: [integer,est,sed,tincidunt,newfile]", 1)
		updateAndWait(newfile, text, "updated: content/posts/newfile.md")
		// Remove post.
		removeAndWait(existingfile, "removed: content/posts/2016-10-18-sed-sed.md")
		// Rebuild.
		proj.in <- "R\n"
		waitFor("rebuilding...")
		// New static file.
		newfile = path.Join(tmpdir, "content", "newfile.txt")
		text = "Hello World!"
		updateAndWait(newfile, text, "updated: content/newfile.txt")
		// Remove static file.
		removeAndWait(newfile, "removed: content/newfile.txt")
		// Stop serve command.
		proj.quit <- nil
	})
}
