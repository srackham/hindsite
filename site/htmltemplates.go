package site

import (
	"bytes"
	"html/template"
	"path/filepath"

	"github.com/srackham/hindsite/v2/fsx"
)

type templateData map[string]interface{}

type htmlTemplates struct {
	templateDir string
	layouts     []string // Layout templates file names.
	templates   *template.Template
}

func newHTMLTemplates(templateDir string) htmlTemplates {
	tmpls := htmlTemplates{}
	tmpls.templateDir = templateDir
	tmpls.templates = template.New("")
	return tmpls
}

// contains returns true if named template is in templates.
func (tmpls htmlTemplates) contains(name string) bool {
	return tmpls.templates.Lookup(name) != nil
}

// name joins template file name elements and converts them to template name.
// The template name is relative to the site template directory and is
// slash-separated (platform independent).
func (tmpls htmlTemplates) name(elem ...string) string {
	name, err := filepath.Rel(tmpls.templateDir, filepath.Join(elem...))
	if err != nil {
		panic(err) // Template file should always be in template directory.
	}
	return filepath.ToSlash(name)
}

// add parses the corresponding file from the templates directory and adds it to
// templates.
func (tmpls *htmlTemplates) add(tmplfile string) error {
	text, err := fsx.ReadFile(tmplfile)
	if err != nil {
		return err
	}
	name := tmpls.name(tmplfile)
	if _, err = tmpls.templates.New(name).Parse(text); err != nil {
		return err
	}
	if filepath.Base(tmplfile) == "layout.html" {
		tmpls.layouts = append(tmpls.layouts, tmplfile)
	}
	return nil
}

// render renders named HTML template to a string.
func (tmpls htmlTemplates) render(name string, data templateData) (string, error) {
	buf := bytes.NewBufferString("")
	if err := tmpls.templates.ExecuteTemplate(buf, name, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
