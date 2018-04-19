package main

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

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

func (proj *project) watch() error {
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
		// TODO: rename to httpserver().
		done <- proj.serve()
	}()
	// Wait for error exit.
	return <-done
}
