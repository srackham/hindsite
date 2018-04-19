package main

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

func (proj *project) watch() error {
	if err := proj.build(); err != nil {
		return err
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()
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
	done := make(chan error)
	go func() {
		mu := sync.Mutex{}
		prevEvent := fsnotify.Event{}
		var prevTime time.Time
		isValid := func(evt fsnotify.Event) bool {
			result := true
			concurrent := time.Now().Sub(prevTime) < time.Millisecond*100
			switch {
			case proj.exclude(evt.Name):
				proj.println(0, "EXCLUDED\n")
				result = false
			case evt.Op == fsnotify.Chmod:
				proj.println(0, "IGNORED\n")
				result = false
			case evt == prevEvent && concurrent:
				proj.println(0, "CONCURRENT\n")
				result = false
			}
			prevEvent = evt
			prevTime = time.Now()
			return result
		}
		for {
			select {
			case evt := <-watcher.Events:
				mu.Lock()
				f := evt.Name
				proj.println(0, evt.Op.String()+": "+f)
				if isValid(evt) {
					proj.println(0, "START BUILD")
					if err := proj.build(); err != nil {
						done <- err
					}
					proj.println(0, "END BUILD")
					proj.println(0, "")
				}
				mu.Unlock()
			case err := <-watcher.Errors:
				done <- err
			}
		}
	}()
	go func() {
		done <- proj.serve()
	}()
	return <-done
}
