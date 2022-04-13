package site

import (
	"bytes"
	"path/filepath"
	"text/template"

	"github.com/srackham/hindsite/fsx"
)

type textTemplates struct {
	templateDir string
	templates   *template.Template
}

func newTextTemplates(templateDir string) textTemplates {
	tmpls := textTemplates{}
	tmpls.templateDir = templateDir
	tmpls.templates = template.New("")
	return tmpls
}

// name joins template file name elements and converts them to template name.
// The template name is relative to the site template directory and is
// slash-separated (platform independent).
func (tmpls textTemplates) name(elem ...string) string {
	name, err := filepath.Rel(tmpls.templateDir, filepath.Join(elem...))
	if err != nil {
		panic(err) // Template file should always be in template directory.
	}
	return filepath.ToSlash(name)
}

// add parses the corresponding file from the templates directory and adds it to
// templates.
func (tmpls *textTemplates) add(tmplfile string) error {
	text, err := fsx.ReadFile(tmplfile)
	if err != nil {
		return err
	}
	name := tmpls.name(tmplfile)
	if _, err = tmpls.templates.New(name).Parse(text); err != nil {
		return err
	}
	return nil
}

// render returns named template text rendered with data.
func (tmpls textTemplates) renderText(name, text string, data templateData) (string, error) {
	tmpl, err := tmpls.templates.New(name).Parse(text)
	if err != nil {
		return "", err
	}
	var output bytes.Buffer
	if err := tmpl.Execute(&output, data); err != nil {
		return "", err
	}
	return output.String(), nil
}
