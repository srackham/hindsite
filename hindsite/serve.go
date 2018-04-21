package main

import (
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
	if err := proj.parseConfigs(); err != nil {
		return err
	}
	if !dirExists(proj.buildDir) {
		return fmt.Errorf("missing build directory: " + proj.buildDir)
	}
	// Tweaked http.StripPrefix() handler
	// (https://golang.org/pkg/net/http/#StripPrefix). If URL does not start
	// with prefix serve unmodified URL.
	stripPrefix := func(prefix string, h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			proj.verbose("request: " + r.URL.Path)
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

// watcherFilter filters and debounces fsnotify events. When there has been a
// lull in file system events arriving on the in input channel then forward the
// most recent accepted file system notification event to the output channel.
func (proj *project) watcherFilter(in chan fsnotify.Event, out chan fsnotify.Event) {
	const lull time.Duration = 100 * time.Millisecond
	var nextOut fsnotify.Event
	timer := time.NewTimer(lull)
	timer.Stop()
	for {
		select {
		case evt := <-in:
			reject := false
			var msg string
			switch {
			case proj.exclude(evt.Name):
				msg = "excluded"
				reject = true
			case evt.Op == fsnotify.Chmod:
				msg = "ignored"
				reject = true
			default:
				msg = "accepted"
			}
			proj.verbose("fsnotify: " + msg + ": " + evt.Op.String() + ": " + evt.Name)
			if !reject {
				nextOut = evt
				timer.Reset(lull)
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
		for {
			mu := sync.Mutex{}
			select {
			case evt := <-out:
				mu.Lock()
				start := time.Now()
				proj.println(start.Format("15:04:05") + ": " + evt.Op.String() + ": " + evt.Name)
				switch evt.Op {
				case fsnotify.Write:
					err = proj.updateFile(evt.Name)
				default:
					err = proj.build()
				}
				if err != nil {
					done <- err
				}
				fmt.Printf("time: %.2fs\n", time.Now().Sub(start).Seconds())
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

func (proj *project) updateFile(f string) error {
	var err error
	switch {
	case proj.isDocument(f):
		doc := proj.getDocument(f)
		if doc == nil {
			panic("updateFile: missing document: " + f)
		}
		newDoc, err := newDocument(f, proj)
		if err != nil {
			return err
		}
		oldDoc := *doc
		doc.updateFrom(newDoc)
		// If document front matter has changed rebuild affected indexes.
		if doc.primaryIndex != nil && doc.header != oldDoc.header {
			for _, idx := range proj.idxs {
				if pathIsInDir(doc.templatePath, idx.templateDir) {
					idx.docs.sortByDate()
					if idx.primary {
						idx.docs.setPrevNext()
					}
					idx.build()
				}
			}
		}
		err = proj.renderDocument(doc)
	case pathIsInDir(f, proj.contentDir):
		err = proj.buildStaticFile(f, time.Time{})
	default:
		// template directory file.
		err = proj.build()
	}
	return err
}

func (proj *project) isDocument(f string) bool {
	ext := filepath.Ext(f)
	return (ext == ".md" || ext == ".rmu") && pathIsInDir(f, proj.contentDir)
}

// UNUSED
func (proj *project) isConfigFile(f string) bool {
	base := filepath.Base(f)
	return (base == "config.toml" || base == "config.yaml") && pathIsInDir(f, proj.templateDir)
}

// getDocument returns parsed document for source file f or nil if not found.
func (proj *project) getDocument(f string) *document {
	for _, doc := range proj.docs {
		if doc.contentPath == f {
			return doc
		}
	}
	return nil
}

// setDocument replaces
func (proj *project) setDocument(old, new *document) {
}
