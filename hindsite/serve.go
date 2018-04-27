package main

import (
	"bufio"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// httpserver starts the HTTP server.
func (proj *project) httpserver() error {
	// Tweaked http.StripPrefix() handler
	// (https://golang.org/pkg/net/http/#StripPrefix). If URL does not start
	// with prefix serve unmodified URL.
	stripPrefix := func(prefix string, h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			proj.verbose2("request: " + r.URL.Path)
			if p := strings.TrimPrefix(r.URL.Path, prefix); len(p) < len(r.URL.Path) {
				r2 := new(http.Request)
				*r2 = *r
				r2.URL = new(url.URL)
				*r2.URL = *r.URL
				r2.URL.Path = p
				h.ServeHTTP(w, r2)
			} else {
				h.ServeHTTP(w, r)
			}
		})
	}
	http.Handle("/", stripPrefix(proj.rootConf.urlprefix, http.FileServer(http.Dir(proj.buildDir))))
	proj.println(fmt.Sprintf("\nServing build directory %s on http://localhost:%s/\nPress Ctrl+C to stop\n", proj.buildDir, proj.port))
	return http.ListenAndServe(":"+proj.port, nil)
}

// watcherLullTime is the watcherFilter debounce time.
const watcherLullTime time.Duration = 20 * time.Millisecond

// watcherFilter filters and debounces fsnotify events. When there has been a
// lull in file system events arriving on the in input channel then forward the
// most recent accepted file system notification event to the output channel.
func (proj *project) watcherFilter(in chan fsnotify.Event, out chan fsnotify.Event) {
	var nextOut fsnotify.Event
	timer := time.NewTimer(watcherLullTime)
	timer.Stop()
	for {
		select {
		case evt := <-in:
			reject := false
			var msg string
			switch {
			case evt.Op == fsnotify.Chmod:
				msg = "ignored"
				reject = true
			case proj.exclude(evt.Name):
				msg = "excluded"
				reject = true
			default:
				msg = "accepted"
			}
			// TODO: Restore verbose2.
			// proj.verbose2("fsnotify: " + time.Now().Format("15:04:05.000") + ": " + msg + ": " + evt.Op.String() + ": " + evt.Name)
			proj.verbose("fsnotify: " + time.Now().Format("15:04:05.000") + ": " + msg + ": " + evt.Op.String() + ": " + evt.Name)
			if !reject {
				nextOut = evt
				timer.Reset(watcherLullTime)
			}
		case <-timer.C:
			out <- nextOut
		}
	}
}

// serve implements the serve comand.
func (proj *project) serve() error {
	// Full rebuild to initialize document and index structures.
	if err := proj.build(); err != nil {
		return err
	}
	// Create file system watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()
	// Watch content and template directories.
	watcherAddDir := func(dir string) error {
		proj.verbose("watching: " + dir)
		return filepath.Walk(dir, func(f string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return watcher.Add(f)
			}
			return nil
		})
	}
	if err := watcherAddDir(proj.contentDir); err != nil {
		return err
	}
	if err := watcherAddDir(proj.templateDir); err != nil {
		return err
	}
	// Error channel to exit serve command.
	done := make(chan error)
	// Start thread to monitor file system notifications and rebuild website.
	go func() {
		out := make(chan fsnotify.Event)
		go proj.watcherFilter(watcher.Events, out)
		kb := make(chan rune)
		go kbmonitor(kb)
		mu := sync.Mutex{}
		for {
			select {
			case c := <-kb:
				if c == 'r' || c == 'R' {
					mu.Lock()
					err = proj.build()
					if err != nil {
						done <- err
					}
					proj.println("")
					mu.Unlock()
				}
			case evt := <-out:
				mu.Lock()
				start := time.Now()
				switch evt.Op {
				case fsnotify.Create, fsnotify.Write:
					proj.println(start.Format("15:04:05") + ": updated: " + evt.Name)
					if dirExists(evt.Name) {
						proj.verbose("watching: " + evt.Name)
						err = watcher.Add(evt.Name)
					} else {
						err = proj.writeFile(evt.Name)
						if err == nil {
							err = proj.installHomePage()
						}
					}
				case fsnotify.Remove, fsnotify.Rename:
					proj.println(start.Format("15:04:05") + ": removed: " + evt.Name)
					err = proj.removeFile(evt.Name)
					if err == nil {
						err = proj.installHomePage()
					}
				default:
					panic("unexpected event: " + evt.Op.String() + ": " + evt.Name)
				}
				if err != nil {
					proj.logerror(err.Error())
				}
				fmt.Printf("elapsed: %.3fs\n", (time.Now().Sub(start) + watcherLullTime).Seconds())
				proj.println("")
				mu.Unlock()
			case err := <-watcher.Errors:
				done <- err
			}
		}
	}()
	// Start web server thread.
	go func() {
		done <- proj.httpserver()
	}()
	// Wait for error exit.
	return <-done
}

// kbmonitor sends keyboard characters to the out channel. The input source is
// buffered so characters are only received on a lin-by-line basis.
func kbmonitor(out chan rune) {
	reader := bufio.NewReader(os.Stdin)
	for {
		c, num, err := reader.ReadRune()
		if num > 0 && err == nil {
			out <- c
		}
	}
}

// createFile handles the fsnotify Create event and adds the file to the build
// set.
func (proj *project) createFile(f string) error {
	switch {
	case proj.isDocument(f):
		if proj.docs.byContentPath[f] != nil {
			panic("document already exists")
		}
		doc, err := newDocument(f, proj)
		if err != nil {
			return err
		}
		if doc.isDraft() {
			proj.verbose("skip draft: " + f)
			return nil
		}
		if err := proj.docs.add(&doc); err != nil {
			return err
		}
		proj.idxs.addDocument(&doc)
		// Rebuild indexes containing the new document.
		for _, idx := range proj.idxs {
			if pathIsInDir(doc.templatePath, idx.templateDir) {
				if err := idx.build(nil); err != nil {
					return err
				}
			}
		}
		return proj.renderDocument(&doc)
	case pathIsInDir(f, proj.contentDir):
		return proj.buildStaticFile(f, time.Time{})
	case pathIsInDir(f, proj.templateDir):
		return proj.build()
	default:
		panic("file is not in watched directories: " + f)
	}
}

// removeFile handles fsnotify Remove events and removes the document from the
// build set.
func (proj *project) removeFile(f string) error {
	switch {
	case proj.isDocument(f):
		doc := proj.docs.byContentPath[f]
		if doc == nil {
			// The document may have been a directory or a draft so can't assume
			// this is an error.
			return nil
		}
		// Delete from documents.
		proj.docs.delete(doc)
		// Rebuild indexes containing the removed document.
		for _, idx := range proj.idxs {
			if pathIsInDir(doc.templatePath, idx.templateDir) {
				idx.docs = idx.docs.delete(doc)
				if err := idx.build(nil); err != nil {
					return err
				}
			}
		}
		proj.verbose("delete document: " + doc.buildPath)
		return os.Remove(doc.buildPath)
	case pathIsInDir(f, proj.contentDir):
		f := pathTranslate(f, proj.contentDir, proj.buildDir)
		// The deleted content may have been a directory.
		if fileExists(f) {
			proj.verbose("delete static: " + f)
			return os.Remove(f)
		}
		return nil
	case pathIsInDir(f, proj.templateDir):
		return proj.build()
	default:
		panic("file is not in watched directories: " + f)
	}
}

// writeFile handles document creation an update events. If the document is
// changed to a draft it is removed from the build set.
func (proj *project) writeFile(f string) error {
	switch {
	case proj.isDocument(f):
		newDoc, err := newDocument(f, proj)
		if err != nil {
			return err
		}
		doc := proj.docs.byContentPath[f]
		if doc == nil {
			if newDoc.isDraft() {
				// Draft document updated, don't do anything.
				proj.verbose("skip draft: " + f)
				return nil
			}
			// Document has just been created and written or was a draft and has changed to non-draft.
			return proj.createFile(f)
		}
		if newDoc.isDraft() {
			// Document changed to draft.
			proj.verbose("skip draft: " + f)
			return proj.removeFile(f)
		}
		oldDoc := *doc
		doc.updateFrom(newDoc)
		// Rebuild affected document index pages.
		for _, idx := range proj.idxs {
			if pathIsInDir(doc.templatePath, idx.templateDir) {
				if oldDoc.date.Equal(doc.date) && strings.Join(oldDoc.tags, ",") == strings.Join(doc.tags, ",") {
					// Neither date ordering or tags have changed so only rebuild document index pages containing doc.
					if err := idx.build(doc); err != nil {
						return err
					}
				} else {
					// Rebuild the index completely.
					if err := idx.build(nil); err != nil {
						return err
					}
				}
			}
		}
		return proj.renderDocument(doc)
	case pathIsInDir(f, proj.contentDir):
		return proj.buildStaticFile(f, time.Time{})
	case pathIsInDir(f, proj.templateDir):
		return proj.build()
	default:
		panic("file is not in watched directories: " + f)
	}
}

func (proj *project) isDocument(f string) bool {
	ext := filepath.Ext(f)
	return (ext == ".md" || ext == ".rmu") && pathIsInDir(f, proj.contentDir)
}
