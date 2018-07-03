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
		code := proj.executeArgs(args)
		if code != 0 {
			t.Fatalf("%s", cmd)
		}
		if dirCount(path.Join(tmpdir, "template")) != 8 {
			t.Fatalf("%s: unexpected number of riles in template directory", cmd)
		}
		waitFor := func(output string) {
			for {
				select {
				case line := <-proj.out:
					line = strings.Replace(line, "\\", "/", -1) // Normalize MS Windows path separators.
					if strings.Contains(line, output) {
						return
					}
				case <-time.After(500 * time.Millisecond):
					t.Fatalf("%s: timed out waiting for: %v", cmd, output)
					return
				}
			}
		}
		updateAndWait := func(f, text, output string) {
			err := writeFile(f, text)
			if err != nil {
				t.Fatal(err)
			}
			waitFor(output)
		}
		removeAndWait := func(f, output string) {
			err := os.Remove(f)
			if err != nil {
				t.Fatal(err)
			}
			waitFor(output)
		}
		// Start server.
		proj = newProject()
		proj.out = make(chan string, 100)
		proj.in = make(chan string, 1)
		cmd = "hindsite serve " + tmpdir
		args = strings.Split(cmd, " ")
		go func() {
			if code := proj.executeArgs(args); code != 0 {
				t.Fatalf("serve start error: %v", <-proj.out)
			}
		}()
		waitFor("Press Ctrl+C to stop")
		// Create new post with copy of existing post.
		existingfile := path.Join(tmpdir, "content", "posts", "2016-10-18-sed-sed.md")
		text, err := readFile(existingfile)
		if err != nil {
			t.Fatal(err)
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
		waitFor("documents: 11")
		// New static file.
		newfile = path.Join(tmpdir, "content", "newfile.txt")
		text = "Hello World!"
		updateAndWait(newfile, text, "updated: content/newfile.txt")
		// Remove static file.
		removeAndWait(newfile, "removed: content/newfile.txt")
		// Stop serve command.
		close(proj.quit)
		time.Sleep(50 * time.Millisecond) // Allow time for serve goroutines to execute cleanup code.
	})
}
