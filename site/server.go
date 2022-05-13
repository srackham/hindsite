package site

import (
	"bufio"
	"context"
	"fmt"
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
	"github.com/srackham/hindsite/fsx"
)

const (
	// watcherLullTime is the watcherFilter debounce time.
	watcherLullTime time.Duration = 50 * time.Millisecond
	// -navigate option LiveReload navigation plugin.
	navigatePrefix = "__hindsite_navigate:"
	navigatePlugin = `function HindsitePlugin() {}
HindsitePlugin.identifier = 'hindsitePlugin';
HindsitePlugin.version = '0.1';
HindsitePlugin.prototype.reload = function(path) {
   	var prefix = "` + navigatePrefix + `";
	if (path.lastIndexOf(prefix, 0) !== 0) {
		return false
	}
	path = path.substring(prefix.length);
    if (window.location.pathname === path) {
		window.location.reload();
	} else {
        window.location.pathname = path;
    }
    return true;
};
LiveReload.addPlugin(HindsitePlugin);`
)

// server is a site plus server specific fields and methods.
type server struct {
	*site
	mutex      *sync.Mutex
	rootURL    string
	browserURL string
	quit       chan struct{}
	err        error
}

func newServer(site *site) server {
	return server{
		site:    site,
		rootURL: "http://localhost:" + fmt.Sprintf("%d", site.httpport) + "/",
		mutex:   &sync.Mutex{},
		quit:    make(chan struct{}),
	}
}

func (svr *server) close(err error) {
	svr.mutex.Lock()
	svr.err = err
	svr.mutex.Unlock()
	close(svr.quit)
}

func (svr *server) help() {
	svr.logconsole(`Serving build directory %q on %q

Press the R key followed by the Enter key to force a complete site rebuild
Press the D key followed by the Enter key to toggle the server -drafts option
Press the N key followed by the Enter key to toggle the server -navigate option
Press the Q key followed by the Enter key to exit
Press the Enter key to print help
`, svr.buildDir, svr.rootURL)
}

// setNavigateURL sets the document navigation URL that will be processed by the
// hindsite plugin in the browser LiveReload client. Does nothing if the
// -navigate option was not specified or live reload is disabled.
func (svr *server) setNavigateURL(url string) {
	if !svr.livereload || !svr.navigate {
		return
	}
	path := strings.TrimPrefix(url, svr.urlprefix())
	svr.verbose("navigate to: " + path)
	svr.mutex.Lock()
	svr.browserURL = navigatePrefix + path
	svr.mutex.Unlock()
}

// logRequest server request handler logs browser requests.
func (svr *server) logRequest(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		svr.verbose("request: " + r.URL.Path)
		h.ServeHTTP(w, r)
	})
}

// saveBrowserURL server request handler sets the server browserURL field to the
// requested HTML page.
func (svr *server) saveBrowserURL(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		if strings.HasSuffix(p, "/") || path.Ext(p) == ".html" {
			svr.mutex.Lock()
			svr.browserURL = p
			svr.mutex.Unlock()
		}
		h.ServeHTTP(w, r)
	})
}

// htmlFilter server request handler injects the LiveReload script tag into the
// body and strips the urlprefix from href URLs.
func (svr *server) htmlFilter(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		if strings.HasSuffix(p, "/") {
			p += "index.html"
		}
		if path.Ext(p) == ".html" {
			p = filepath.Join(svr.buildDir, filepath.FromSlash(p[1:])) // Convert URL path to file path.
			if !fsx.FileExists(p) {
				http.Error(w, "404: file not found: "+p, 404)
				return

			}
			content, err := fsx.ReadFile(p)
			if err != nil {
				http.Error(w, "500: "+err.Error(), 500)
				return
			}
			if svr.livereload {
				// Inject LiveReload script tag.
				content = strings.Replace(content, "</body>", "<script src=\"http://localhost:"+fmt.Sprintf("%d", svr.lrport)+"/livereload.js\"></script>\n</body>", 1)
				// Inject navigation plugin.
				if svr.navigate {
					content = strings.Replace(content, "</body>", "<script>\n"+navigatePlugin+"\n</script>\n</body>", 1)
				}
			}
			// Strip urlprefix from URLs.
			if svr.urlprefix() != "" {
				content = strings.Replace(content, "href=\""+svr.urlprefix(), "href=\"", -1)
				content = strings.Replace(content, "src=\""+svr.urlprefix(), "src=\"", -1)
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(content))
		} else {
			h.ServeHTTP(w, r)
		}
	})
}

// watcherFilter filters and debounces fsnotify events. When there has been a
// lull in file system events arriving on the in input channel then forward the
// most recent accepted file system notification event to the output channel.
func (svr *server) watcherFilter(watcher *fsnotify.Watcher, out chan<- fsnotify.Event) {
	var prev fsnotify.Event
	timer := time.NewTimer(watcherLullTime)
	timer.Stop()
	for {
		select {
		case <-svr.quit:
			return
		case evt, ok := <-watcher.Events:
			if !ok {
				return // Watcher has closed.
			}
			svr.verbose("fsnotify: " + time.Now().Format("15:04:05.000") + ": " + evt.Op.String() + ": " + evt.Name)
			accepted := false
			var msg string
			switch {
			case evt.Op == fsnotify.Chmod:
				msg = "ignored"
			case fsx.PathIsInDir(evt.Name, svr.contentDir) && svr.exclude(evt.Name):
				msg = "excluded"
			case fsx.DirExists(evt.Name):
				msg = "ignored"
				if evt.Op == fsnotify.Create {
					watcher.Add(evt.Name)
				}
			default:
				msg = "accepted"
				accepted = true
			}
			svr.verbose("fsnotify: " + msg)
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

// serve implements the serve comand. Does not return unless and error occurs or
// the server quit channel is closed.
func (svr *server) serve() error {
	if len(svr.cmdargs) > 0 {
		return fmt.Errorf("to many command arguments")
	}
	// Full rebuild to initialize document and index structures.
	err := svr.build()
	if err != nil && err != ErrNonFatal {
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
		svr.verbose("watching: " + dir)
		return filepath.Walk(dir, func(f string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				if f == svr.initDir && dir == svr.templateDir {
					// Skip init directory when adding template directory watchers.
					return filepath.SkipDir
				}
				return watcher.Add(f)
			}
			return nil
		})
	}
	if err := watcherAddDir(svr.contentDir); err != nil {
		return err
	}
	if err := watcherAddDir(svr.templateDir); err != nil {
		return err
	}
	// Start LiveReload server.
	lr := lrserver.New(lrserver.DefaultName, svr.lrport)
	defer lr.Close()
	lr.SetLiveCSS(true)
	lr.StatusLog().SetPrefix("reload: ")
	lr.ErrorLog().SetPrefix("reload: ")
	if svr.verbosity == 0 {
		lr.SetStatusLog(nil)
		lr.SetErrorLog(nil)
	}
	if svr.livereload {
		go lr.ListenAndServe()
	}
	// Start Web server.
	go func() {
		svr.help()
		handler := http.FileServer(http.Dir(svr.buildDir))
		handler = svr.htmlFilter(handler)
		handler = svr.saveBrowserURL(handler)
		handler = svr.logRequest(handler)
		httpsvr := &http.Server{Addr: ":" + fmt.Sprintf("%d", svr.httpport), Handler: handler}
		select {
		case <-svr.quit:
			if err := httpsvr.Shutdown(context.TODO()); err != nil {
				panic(err) // Failed to shut down the server gracefully.
			}
			return
		default:
			err := httpsvr.ListenAndServe()
			svr.close(err)
		}
	}()
	// Start watcher event filter.
	fsevent := make(chan fsnotify.Event, 2)
	go svr.watcherFilter(watcher, fsevent)
	// Start keyboard monitor.
	kb := make(chan string)
	go func() {
		select {
		case <-svr.quit:
			return
		default:
			reader := bufio.NewReader(os.Stdin)
			for {
				var line string
				if svr.in == nil {
					line, _ = reader.ReadString('\n')
				} else {
					line = <-svr.in
				}
				kb <- line
			}
		}
	}()
	// Launch browser.
	if svr.launch {
		go func() {
			svr.verbose("launching browser: " + svr.rootURL)
			if err := launchBrowser(svr.rootURL); err != nil {
				svr.logerror(err.Error())
			}
		}()
	}
	// Monitor and execute build notifications from keyboard and file system.
	go func() {
		for {
			select {
			case <-svr.quit:
				return
			case line := <-kb:
				switch strings.ToUpper(strings.TrimSpace(line)) {
				case "R": // Rebuild.
					svr.logconsole("rebuilding...")
					err = svr.build()
					if err != nil && err != ErrNonFatal {
						svr.logerror(err.Error())
					}
					if svr.livereload {
						lr.Reload(svr.browserURL)
					}
					svr.logconsole("")
				case "D": // Toggle -drafts option.
					svr.drafts = !svr.drafts
					svr.logconsole("drafts: %t\n", svr.drafts)
				case "N": // Toggle -navigate option.
					svr.navigate = !svr.navigate
					svr.logconsole("navigation: %t\n", svr.navigate)
				case "Q":
					svr.close(nil)
				default:
					svr.help()
				}
			case evt := <-fsevent:
				start := time.Now()
				switch evt.Op {
				case fsnotify.Create, fsnotify.Write:
					t := fsx.FileModTime(svr.homepage())
					err = svr.writeFile(evt.Name)
					if err == nil && t.Before(fsx.FileModTime(svr.homepage())) {
						// homepage was modified by this event.
						err = svr.copyHomePage()
					}
					svr.logconsole(start.Format("15:04:05") + ": updated: " + evt.Name)
				case fsnotify.Remove, fsnotify.Rename:
					err = svr.removeFile(evt.Name)
					svr.logconsole(start.Format("15:04:05") + ": removed: " + evt.Name)
				default:
					panic("unexpected event: " + evt.Op.String() + ": " + evt.Name)
				}
				if err != nil {
					svr.logerror(err.Error())
				} else {
					color.Set(color.FgGreen, color.Bold)
				}
				svr.logconsole("time: %.3fs\n", (time.Since(start) + watcherLullTime).Seconds())
				color.Unset()
				if svr.livereload {
					lr.Reload(svr.browserURL)
				}
			case err := <-watcher.Errors:
				svr.close(err)
			}
		}
	}()
	<-svr.quit
	return svr.err
}

// createFile handles the fsnotify Create event and adds the file to the build
// set.
func (svr *server) createFile(f string) error {
	switch {
	case svr.isDocument(f):
		if svr.docs.byContentPath[f] != nil {
			panic("document already exists")
		}
		doc, err := newDocument(f, svr.site)
		if err != nil {
			return err
		}
		if doc.isDraft() {
			svr.verbose("skip draft: " + f)
			return nil
		}
		if err := svr.docs.add(&doc); err != nil {
			return err
		}
		svr.idxs.addDocument(&doc)
		// Rebuild indexes containing the new document.
		for _, idx := range svr.idxs {
			if fsx.PathIsInDir(doc.templatePath, idx.templateDir) {
				if err := idx.build(nil); err != nil {
					return err
				}
			}
		}
		svr.setNavigateURL(doc.url)
		return svr.renderDocument(&doc)
	case fsx.PathIsInDir(f, svr.contentDir):
		return svr.buildStaticFile(f)
	case fsx.PathIsInDir(f, svr.templateDir):
		return svr.build() // Rebuild the site if a file in the template directory is changed.
	default:
		panic("file is not in watched directories: " + f)
	}
}

// removeFile handles fsnotify Remove events and removes the document from the
// build set.
func (svr *server) removeFile(f string) error {
	switch {
	case svr.isDocument(f):
		doc := svr.docs.byContentPath[f]
		if doc == nil {
			// The document may have been a draft so can't assume this is an error.
			return nil
		}
		// Delete from documents.
		svr.docs.delete(doc)
		// Rebuild indexes containing the removed document.
		for _, idx := range svr.idxs {
			if fsx.PathIsInDir(doc.templatePath, idx.templateDir) {
				idx.docs = idx.docs.delete(doc)
				if err := idx.build(nil); err != nil {
					return err
				}
			}
		}
		svr.verbose("delete document: " + doc.buildPath)
		return os.Remove(doc.buildPath)
	case fsx.PathIsInDir(f, svr.contentDir):
		f := fsx.PathTranslate(f, svr.contentDir, svr.buildDir)
		// The deleted content may have been a directory.
		if fsx.FileExists(f) {
			svr.verbose("delete static: " + f)
			return os.Remove(f)
		}
		return nil
	case fsx.PathIsInDir(f, svr.templateDir):
		return svr.build()
	default:
		panic("file is not in watched directories: " + f)
	}
}

// writeFile handles document creation an update events. If the document is
// changed to a draft it is removed from the build set.
func (svr *server) writeFile(f string) error {
	switch {
	case svr.isDocument(f):
		newDoc, err := newDocument(f, svr.site)
		if err != nil {
			return err
		}
		doc := svr.docs.byContentPath[f]
		if doc == nil {
			if newDoc.isDraft() {
				// Draft document updated, don't do anything.
				svr.verbose("skip draft: " + f)
				return nil
			}
			// Document has just been created and written or was a draft and has changed to non-draft.
			return svr.createFile(f)
		}
		// Arrive here if an existing published document has been updated.
		if newDoc.isDraft() {
			// Document changed to draft.
			svr.verbose("skip draft: " + f)
			return svr.removeFile(f)
		}
		oldDoc := *doc
		if err = svr.docs.update(doc, newDoc); err != nil {
			return err
		}
		// Rebuild affected document index pages.
		for _, idx := range svr.idxs {
			if fsx.PathIsInDir(doc.templatePath, idx.templateDir) {
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
		svr.setNavigateURL(doc.url)
		return svr.renderDocument(doc)
	case fsx.PathIsInDir(f, svr.contentDir):
		return svr.buildStaticFile(f)
	case fsx.PathIsInDir(f, svr.templateDir):
		return svr.build()
	default:
		panic("file is not in watched directories: " + f)
	}
}
