package site

import (
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/srackham/hindsite/fsx"
	"github.com/srackham/hindsite/slice"
)

/*
Miscellaneous functions.
*/

// nz returns the string pointed to by s or "" if s is nil.
func nz(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// normalizeNewlines converts \r\n (Window) and \n (Mac OS) line teminations to
// \n (UNIX) termination.
func normalizeNewlines(s string) (result string) {
	result = strings.ReplaceAll(s, "\r\n", "\n")
	result = strings.ReplaceAll(result, "\r", "\n")
	return
}

// sortedKeys returns a sorted array of map string keys.
func sortedKeys[T any](m map[string]T) (result []string) {
	for k := range m {
		result = append(result, k)
	}
	sort.Strings(result)
	return
}

// Transform text into a slug (lowercase alpha-numeric + hyphens).
func slugify(text string, exclude slice.Slice[string]) string {
	slug := text
	slug = regexp.MustCompile(`\W+`).ReplaceAllString(slug, "-") // Replace non-alphanumeric characters with dashes.
	slug = regexp.MustCompile(`-+`).ReplaceAllString(slug, "-")  // Replace multiple dashes with single dash.
	slug = strings.Trim(slug, "-")                               // Trim leading and trailing dashes.
	slug = strings.ToLower(slug)
	if slug == "" {
		slug = "x"
	}
	if exclude.IndexOf(slug) > -1 {
		i := 2
		for exclude.IndexOf(slug+"-"+fmt.Sprint(i)) > -1 {
			i++
		}
		slug += "-" + fmt.Sprint(i)
	}
	return slug
}

// launchBrowser launches the browser at the url address. Waits till launch
// completed. Credit: https://stackoverflow.com/a/39324149/1136455
func launchBrowser(url string) error {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Run()
}

// extractDateTitle extracts the date and title strings from file name.
func extractDateTitle(name string) (date string, title string) {
	title = fsx.FileName(name)
	if regexp.MustCompile(`^\d\d\d\d-\d\d-\d\d-.+`).MatchString(title) {
		date = title[0:10]
		title = title[11:]
	}
	//lint:ignore SA1019 the file name is ASCII
	title = strings.Title(strings.Replace(title, "-", " ", -1))
	return date, title
}

/*
Date/time functions.
*/
// Parse date text. If timezone is not specified Local is assumed.
func parseDate(text string, loc *time.Location) (time.Time, error) {
	if loc == nil {
		loc, _ = time.LoadLocation("Local")
	}
	text = strings.TrimSpace(text)
	d, err := time.Parse(time.RFC3339, text)
	if err != nil {
		if d, err = time.Parse("2006-01-02 15:04:05-07:00", text); err != nil {
			if d, err = time.ParseInLocation("2006-01-02 15:04:05", text, loc); err != nil {
				d, err = time.ParseInLocation("2006-01-02", text, loc)
			}
		}
	}
	if err != nil {
		err = fmt.Errorf("illegal date value: %s", text)
	}
	return d, err
}
