package site

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
)

var highlightColor = []color.Attribute{color.FgGreen, color.Bold}
var errorColor = []color.Attribute{color.FgRed, color.Bold}
var warningColor = []color.Attribute{color.FgRed}

// colorize executes a function with color attributes.
func colorize(attributes []color.Attribute, fn func()) {
	defer color.Unset()
	color.Set(attributes...)
	fn()
}

// output prints a line to `out` if `site.verbosity` is greater than equal or
// equal to `verbosity`. If `site.out` is set then the line is written to it
// instead of `out` (this feature is used for testing purposes).
func (site *site) output(out io.Writer, verbosity int, format string, v ...interface{}) {
	if site.verbosity >= verbosity {
		msg := fmt.Sprintf(format, v...)
		// Strip leading site directory from quoted path names to make message more readable.
		if filepath.IsAbs(site.siteDir) {
			msg = strings.Replace(msg, `"`+site.siteDir+string(filepath.Separator), `"`, -1)
		}
		if site.out == nil {
			fmt.Fprintln(out, msg)
		} else {
			site.out <- msg
		}
	}
}

// logConsole prints a line to stdout.
func (site *site) logConsole(format string, v ...interface{}) {
	site.output(os.Stdout, 0, format, v...)
}

// logVerbose prints a line to stdout if `-v` logVerbose option was specified.
func (site *site) logVerbose(format string, v ...interface{}) {
	site.output(os.Stdout, 1, format, v...)
}

// logVerbose2 prints a a line to stdout the `-v` verbose option was specified more
// than once.
func (site *site) logVerbose2(format string, v ...interface{}) {
	site.output(os.Stdout, 2, format, v...)
}

// logColorize prints a colorized line to stdout.
func (site *site) logColorize(attributes []color.Attribute, format string, v ...interface{}) {
	colorize(attributes, func() {
		site.logConsole(format, v...)
	})
}

// logHighlight prints a highlighted line to stdout.
func (site *site) logHighlight(format string, v ...interface{}) {
	site.logColorize(highlightColor, format, v...)
}

// logError prints a line to stderr and increments the error count.
func (site *site) logError(format string, v ...interface{}) {
	colorize(errorColor, func() {
		site.output(os.Stderr, 0, "error: "+format, v...)
	})
	site.errors++
}

// logWarning prints a line to stdout and increments the warnings count.
func (site *site) logWarning(format string, v ...interface{}) {
	colorize(warningColor, func() {
		site.output(os.Stdout, 0, "warning: "+format, v...)
	})
	site.warnings++
}
