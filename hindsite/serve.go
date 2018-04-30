package main

import (
	"bufio"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/jaschaephraim/lrserver"
)

const (
	// watcherLullTime is the watcherFilter debounce time.
	watcherLullTime time.Duration = 50 * time.Millisecond
)

var (
	// webpage shared variable contains the path name of the most recently requested HTML webpage.
	webpage struct {
		sync.Mutex
		path string
	}
)

// logRequest server request handler logs browser requests.
func logRequest(proj *project, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proj.verbose("request: " + r.URL.Path)
		h.ServeHTTP(w, r)
	})
}

// setWebpage server request handler sets the shared webpage variable.
func setWebpage(proj *project, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		if strings.HasSuffix(p, "/") || path.Ext(p) == ".html" {
			webpage.Lock()
			webpage.path = p
			webpage.Unlock()
		}
		h.ServeHTTP(w, r)
	})
}

// stripURLPrefix server request handler strips the urlprefix from browser
// request URLs. If URL does not start with prefix serve unmodified URL.
func stripURLPrefix(proj *project, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		if p2 := strings.TrimPrefix(p, proj.rootConf.urlprefix); len(p2) < len(p) {
			r2 := new(http.Request)
			*r2 = *r
			r2.URL = new(url.URL)
			*r2.URL = *r.URL
			r2.URL.Path = p2
			h.ServeHTTP(w, r2)
		} else {
			h.ServeHTTP(w, r)
		}
	})
}

// startHTTPServer registers server request handlers and starts the HTTP server.
func (proj *project) startHTTPServer() error {
	handler := http.FileServer(http.Dir(proj.buildDir))
	handler = stripURLPrefix(proj, handler)
	handler = setWebpage(proj, handler)
	handler = logRequest(proj, handler)
	return http.ListenAndServe(":"+proj.port, &lrHandler{Handler: handler})
}

// watcherFilter filters and debounces fsnotify events. When there has been a
// lull in file system events arriving on the in input channel then forward the
// most recent accepted file system notification event to the output channel.
func (proj *project) watcherFilter(watcher *fsnotify.Watcher, out chan fsnotify.Event) {
	var prev fsnotify.Event
	timer := time.NewTimer(watcherLullTime)
	timer.Stop()
	for {
		select {
		case evt := <-watcher.Events:
			accepted := false
			var msg string
			switch {
			case evt.Op == fsnotify.Chmod:
				msg = "ignored"
			case dirExists(evt.Name):
				msg = "ignored"
				if evt.Op == fsnotify.Create {
					watcher.Add(evt.Name)
				}
			case proj.exclude(evt.Name):
				msg = "excluded"
			default:
				msg = "accepted"
				accepted = true
			}
			proj.verbose("fsnotify: " + time.Now().Format("15:04:05.000") + ": " + msg + ": " + evt.Op.String() + ": " + evt.Name)
			if accepted {
				if prev.Op == fsnotify.Rename && prev.Name != evt.Name {
					// A rename has ocurred within the watched directories
					// (Rename followed by immediately by Create) so forward the
					// Rename to ensure the original file is deleted prior to
					// the new file being created.
					out <- prev
				}
				prev = evt
				timer.Reset(watcherLullTime)
			}
		case <-timer.C:
			out <- prev
			prev = fsnotify.Event{}
		}
	}
}

// serve implements the serve comand.
func (proj *project) serve() error {
	rooturl := "http://localhost:" + proj.port + "/"
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
	// Start LiveReload server.
	lr := lrserver.New(lrserver.DefaultName, lrserver.DefaultPort)
	lr.SetLiveCSS(true)
	if proj.verbosity == 0 {
		lr.SetStatusLog(nil)
		lr.SetErrorLog(nil)
	}
	go lr.ListenAndServe()
	// Start Web server.
	go func() {
		proj.println(fmt.Sprintf("\nServing build directory %s on %s\nPress Ctrl+C to stop\n", proj.buildDir, rooturl))
		done <- proj.startHTTPServer()
	}()
	// Start watcher event filter.
	fs := make(chan fsnotify.Event, 2)
	go proj.watcherFilter(watcher, fs)
	// Start keyboard monitor.
	kb := make(chan rune)
	go kbmonitor(kb)
	// Start thread to monitor and execute build notifications.
	go func() {
		for {
			select {
			case c := <-kb:
				if c == 'r' || c == 'R' {
					err = proj.build()
					if err != nil {
						done <- err
					}
					lr.Reload(webpage.path)
					proj.println("")
				}
			case evt := <-fs:
				start := time.Now()
				switch evt.Op {
				case fsnotify.Create, fsnotify.Write:
					proj.println(start.Format("15:04:05") + ": updated: " + evt.Name)
					err = proj.writeFile(evt.Name)
					if err == nil {
						err = proj.installHomePage()
					}
				case fsnotify.Remove, fsnotify.Rename:
					proj.println(start.Format("15:04:05") + ": removed: " + evt.Name)
					err = proj.removeFile(evt.Name)
				default:
					panic("unexpected event: " + evt.Op.String() + ": " + evt.Name)
				}
				if err != nil {
					proj.logerror(err.Error())
				}
				if err == nil {
					lr.Reload(webpage.path)
				}
				fmt.Printf("elapsed: %.3fs\n", (time.Now().Sub(start) + watcherLullTime).Seconds())
				proj.println("")
			case err := <-watcher.Errors:
				done <- err
			}
		}
	}()
	// Launch browser.
	if proj.launch {
		go func() {
			proj.verbose("launching browser: " + rooturl)
			if err := launchBrowser(rooturl); err != nil {
				proj.logerror(err.Error())
			}
		}()
	}
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
