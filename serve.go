package main

import (
	"bufio"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
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

// htmlFilter server request handler injects the LiveReload script tag into the
// body and strips the urlprefix from href URLs.
func htmlFilter(proj *project, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		if strings.HasSuffix(p, "/") {
			p += "index.html"
		}
		if path.Ext(p) == ".html" {
			p = filepath.Join(proj.buildDir, filepath.FromSlash(p[1:])) // Convert URL path to file path.
			if !fileExists(p) {
				http.Error(w, "404: file not found: "+p, 404)
				return

			}
			content, err := readFile(p)
			if err != nil {
				http.Error(w, "500: "+err.Error(), 500)
				return
			}
			// Inject LiveReload script tag.
			content = strings.Replace(content, "</body>", "<script src=\"http://localhost:35729/livereload.js\"></script>\n</body>", 1)
			if proj.rootConf.urlprefix != "" {
				// Strip urlprefix from URLs.
				content = strings.Replace(content, "href=\""+proj.rootConf.urlprefix, "href=\"", -1)
				content = strings.Replace(content, "src=\""+proj.rootConf.urlprefix, "src=\"", -1)
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(content))
		} else {
			h.ServeHTTP(w, r)
		}
	})
}

// startHTTPServer registers server request handlers and starts the HTTP server.
func (proj *project) startHTTPServer() error {
	handler := http.FileServer(http.Dir(proj.buildDir))
	handler = htmlFilter(proj, handler)
	handler = setWebpage(proj, handler)
	handler = logRequest(proj, handler)
	return http.ListenAndServe(":"+proj.port, handler)
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
			case proj.exclude(evt.Name):
				msg = "excluded"
			case dirExists(evt.Name):
				msg = "ignored"
				if evt.Op == fsnotify.Create {
					watcher.Add(evt.Name)
				}
			default:
				msg = "accepted"
				accepted = true
			}
			proj.verbose("fsnotify: " + time.Now().Format("15:04:05.000") + ": " + msg + ": " + evt.Op.String() + ": " + evt.Name)
			if accepted {
				if prev.Op == fsnotify.Rename && prev.Name != evt.Name {
					// A rename has occurred within the watched directories
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
		proj.logerror(err.Error())
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
				if f == proj.initDir && dir == proj.templateDir {
					// Skip init directory when adding template directory watchers.
					return filepath.SkipDir
				}
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
	proj.done = make(chan error)
	// Start LiveReload server.
	lr := lrserver.New(lrserver.DefaultName, lrserver.DefaultPort)
	lr.SetLiveCSS(true)
	lr.StatusLog().SetPrefix("reload: ")
	lr.ErrorLog().SetPrefix("reload: ")
	if proj.verbosity == 0 {
		lr.SetStatusLog(nil)
		lr.SetErrorLog(nil)
	}
	go lr.ListenAndServe()
	// Start Web server.
	go func() {
		proj.logconsole("\nServing build directory %s on %s\nPress Ctrl+C to stop\n", proj.buildDir, rooturl)
		proj.done <- proj.startHTTPServer()
	}()
	// Start watcher event filter.
	fs := make(chan fsnotify.Event, 2)
	go proj.watcherFilter(watcher, fs)
	// Start keyboard monitor.
	kb := make(chan string)
	go kbmonitor(kb)
	// Start thread to monitor and execute build notifications.
	go func() {
		for {
			select {
			case line := <-kb:
				switch strings.ToUpper(strings.TrimSpace(line)) {
				case "R": // Rebuild.
					proj.logconsole("rebuilding...")
					err = proj.build()
					if err != nil {
						proj.logerror(err.Error())
					}
					lr.Reload(webpage.path)
					proj.logconsole("")
				}
			case evt := <-fs:
				start := time.Now()
				switch evt.Op {
				case fsnotify.Create, fsnotify.Write:
					proj.logconsole(start.Format("15:04:05") + ": updated: " + evt.Name)
					t := fileModTime(proj.rootConf.homepage)
					err = proj.writeFile(evt.Name)
					if err == nil && t.Before(fileModTime(proj.rootConf.homepage)) {
						// homepage was modified by this event.
						err = proj.copyHomePage()
					}
				case fsnotify.Remove, fsnotify.Rename:
					proj.logconsole(start.Format("15:04:05") + ": removed: " + evt.Name)
					err = proj.removeFile(evt.Name)
				default:
					panic("unexpected event: " + evt.Op.String() + ": " + evt.Name)
				}
				if err != nil {
					proj.logerror(err.Error())
				} else {
					color.Set(color.FgGreen, color.Bold)
				}
				proj.logconsole("elapsed: %.3fs\n", (time.Now().Sub(start) + watcherLullTime).Seconds())
				proj.logconsole("")
				color.Unset()
				lr.Reload(webpage.path)
			case err := <-watcher.Errors:
				proj.done <- err
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
	return <-proj.done
}

// kbmonitor sends lines of input to the out channel.
func kbmonitor(out chan string) {
	reader := bufio.NewReader(os.Stdin)
	for {
		line, _ := reader.ReadString('\n')
		out <- line
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
		return proj.buildStaticFile(f)
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
			// The document may have been a draft so can't assume this is an error.
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
		// Arrive here if an existing published document has been updated.
		if newDoc.isDraft() {
			// Document changed to draft.
			proj.verbose("skip draft: " + f)
			return proj.removeFile(f)
		}
		oldDoc := *doc
		if err = proj.docs.update(doc, newDoc); err != nil {
			return err
		}
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
		return proj.buildStaticFile(f)
	case pathIsInDir(f, proj.templateDir):
		return proj.build()
	default:
		panic("file is not in watched directories: " + f)
	}
}
