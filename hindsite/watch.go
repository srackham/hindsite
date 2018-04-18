package main

import (
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

func (proj *project) watch() error {
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
		for {
			select {
			case event := <-watcher.Events:
				f := event.Name
				proj.verbose(event.Op.String() + ": " + f)
			case err := <-watcher.Errors:
				done <- err
			}
		}
	}()
	return <-done
}
