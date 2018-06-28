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
		cmd := "hindsite init " + tmpdir + " -builtin blog"
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
		waitFor := func(wanted string) {
		L:
			for {
				select {
				case line := <-proj.logger:
					if strings.Contains(line, wanted) {
						break L
					}
				case <-time.After(300 * time.Millisecond):
					t.Errorf("%s: timed out waiting for: %v", cmd, wanted)
					break L
				}
			}
		}
		proj.logger = make(chan string, 100)
		proj.done = make(chan error)
		// Start server.
		go func() { proj.serve() }()
		waitFor("Press Ctrl+C to stop")
		existingfile := path.Join(tmpdir, "content", "posts", "2016-10-18-sed-sed.md")
		// Create new post with copy of existing post.
		s, err := readFile(existingfile)
		if err != nil {
			t.Error(err)
		}
		newfile := path.Join(tmpdir, "content", "posts", "newfile.md")
		err = writeFile(newfile, s)
		if err != nil {
			t.Error(err)
		}
		waitFor("error: content/posts/newfile.md: duplicate document build path in: content/posts/2016-10-18-sed-sed.md")
		// Edit post.
		s = strings.Replace(s, "slug: sed-sed", "slug: newfile", 1)
		err = writeFile(newfile, s)
		if err != nil {
			t.Error(err)
		}
		// Remove post.
		waitFor("updated: content/posts/newfile.md")
		err = os.Remove(existingfile)
		if err != nil {
			t.Error(err)
		}
		waitFor("removed: content/posts/2016-10-18-sed-sed.md")
		// TODO: The done channel abruptly terminates the server without waiting for processing to complete.
		proj.done <- nil
	})
}
