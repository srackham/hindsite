package main

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"
)

type templates struct {
	templateDir string
	templates   *template.Template
}

type templateData map[string]interface{}

func newTemplates(templateDir string) templates {
	tmpls := templates{}
	tmpls.templateDir = templateDir
	tmpls.templates = template.New("")
	return tmpls
}

// Returns true if named template is in templates.
func (tmpls templates) contains(name string) bool {
	return tmpls.templates.Lookup(name) != nil
}

// Joins template file name elements and converts them to template name. The
// template name is relative to the project template directory and is
// slash-separated (platform independent).
func (tmpls templates) name(elem ...string) string {
	name, _ := filepath.Rel(tmpls.templateDir, filepath.Join(elem...))
	return filepath.ToSlash(name)
}

// Parses the corresponding file from the templates directory and adds it to templates.
func (tmpls templates) add(name string) error {
	tmplfile := filepath.Join(tmpls.templateDir, name)
	if !fileExists(tmplfile) {
		return fmt.Errorf("missing template file: %s", tmplfile)
	}
	text, err := readFile(tmplfile)
	if err != nil {
		return err
	}
	_, err = tmpls.templates.New(name).Parse(text)
	return err
}

// Render named template to file.
func (tmpls templates) render(name string, data templateData, outfile string) error {
	buf := bytes.NewBufferString("")
	if err := tmpls.templates.ExecuteTemplate(buf, name, data); err != nil {
		return err
	}
	html := buf.String()
	if err := mkMissingDir(filepath.Dir(outfile)); err != nil {
		return err
	}
	return writeFile(outfile, html)
}

// Merge in data from another data map.
func (data templateData) add(from templateData) {
	for k, v := range from {
		data[k] = v
	}
}
