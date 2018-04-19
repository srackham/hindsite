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
				f := evt.Name
				proj.println(time.Now().Format("15:04:05") + ": " + evt.Op.String() + ": " + f)
				if err := proj.build(); err != nil {
					done <- err
				}
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
