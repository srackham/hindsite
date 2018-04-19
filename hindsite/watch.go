package main

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// debounce filters and debounces fsnotify events.
// If there have been no file system events on the in input channel for the time
// interval then forward the most recent accepted file system notification event
// to the output channel.
func (proj *project) debounce(interval time.Duration, in chan fsnotify.Event, out chan fsnotify.Event) {
	skip := func(evt fsnotify.Event) bool {
		result := false
		var msg string
		switch {
		case proj.exclude(evt.Name):
			msg = "excluded"
			result = true
		case evt.Op == fsnotify.Chmod:
			msg = "ignored"
			result = true
		default:
			msg = "accepted"
		}
		proj.verbose("fsnotify: " + msg + ": " + evt.Op.String() + ": " + evt.Name)
		return result
	}
	var nextOut fsnotify.Event
	timer := time.NewTimer(interval)
	timer.Stop()
	for {
		select {
		case evt := <-in:
			if !skip(evt) {
				nextOut = evt
				timer.Reset(interval)
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
		go proj.debounce(100*time.Millisecond, watcher.Events, out)
		for {
			mu := sync.Mutex{}
			select {
			case evt := <-out:
				mu.Lock()
				f := evt.Name
				proj.println(0, evt.Op.String()+": "+f)
				if err := proj.build(); err != nil {
					done <- err
				}
				proj.println(0, "")
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
