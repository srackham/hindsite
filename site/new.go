package site

import (
	"bytes"
	"fmt"
	"path/filepath"
	"text/template"
	"time"

	"github.com/srackham/hindsite/fsx"
)

const (
	defaultNewTemplate = `---
title: {{.title}}
date:  {{.date}}
draft: true
---

Document content goes here.`
)

// new implements the new command.
func (site *site) new() (err error) {
	if site.command == "new" {
		if site.newFile == "" {
			return fmt.Errorf("document has not been specified")
		}
		if fsx.DirExists(site.newFile) {
			return fmt.Errorf("document is a directory: %s", site.newFile)
		}
		if d := filepath.Dir(site.newFile); !fsx.DirExists(d) {
			return fmt.Errorf("missing document directory: %s", d)
		}
		if fsx.FileExists(site.newFile) {
			return fmt.Errorf("document already exists: %s", site.newFile)
		}
	}
	site.newFile, err = filepath.Abs(site.newFile)
	if err != nil {
		return err
	}
	if !fsx.PathIsInDir(site.newFile, site.contentDir) {
		return fmt.Errorf("document must reside in content directory: %s", site.contentDir)
	}
	conf := site.configFor(site.newFile)
	// Extract date and title into template data map.
	date := time.Now()
	d, title := extractDateTitle(site.newFile)
	if d != "" {
		if date, err = parseDate(d, conf.timezone); err != nil {
			return err
		}
	}
	data := templateData{}
	data["date"] = date.Format("2006-01-02T15:04:05-07:00")
	data["title"] = title
	site.verbose("document title: %s\ndocument date: %s", data["title"], data["date"])
	// Search up the corresponding template directory path for the closest new.md template file.
	text := defaultNewTemplate
	for d := fsx.PathTranslate(filepath.Dir(site.newFile), site.contentDir, site.templateDir); ; {
		if f := filepath.Join(d, "new.md"); fsx.FileExists(f) {
			site.verbose("document template: %s", f)
			if text, err = fsx.ReadFile(f); err != nil {
				return err
			}
			break
		}
		if d == site.templateDir {
			break // No template file found.
		}
		d = filepath.Dir(d)
	}
	// Parse and execute template.
	tmpl, err := template.New("new.md").Parse(text)
	if err != nil {
		return err
	}
	var output bytes.Buffer
	if err := tmpl.Execute(&output, data); err != nil {
		return err
	}
	site.verbose2("document text: %#v", output.String())
	// Write the new document file.
	site.verbose("document file: %s", site.newFile)
	if err := fsx.WriteFile(site.newFile, output.String()); err != nil {
		return err
	}
	return nil
}
