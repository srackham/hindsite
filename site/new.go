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
	if len(site.cmdargs) == 0 {
		return fmt.Errorf("missing document file name")
	}
	if len(site.cmdargs) > 1 {
		return fmt.Errorf("to many command arguments")
	}
	newFile := site.cmdargs[0]
	if newFile == "" {
		return fmt.Errorf("new document has not been specified")
	}
	if fsx.DirExists(newFile) {
		return fmt.Errorf("document is a directory: %s", newFile)
	}
	if d := filepath.Dir(newFile); !fsx.DirExists(d) {
		return fmt.Errorf("missing document directory: %s", d)
	}
	if fsx.FileExists(newFile) {
		return fmt.Errorf("document already exists: %s", newFile)
	}
	newFile, err = filepath.Abs(newFile)
	if err != nil {
		return err
	}
	if !fsx.PathIsInDir(newFile, site.contentDir) {
		return fmt.Errorf("document must reside in content directory: %s", site.contentDir)
	}
	if err = site.parseConfigFiles(); err != nil {
		return err
	}
	conf := site.configFor(newFile)
	// Extract date and title into template data map.
	date := time.Now()
	d, title := extractDateTitle(newFile)
	if d != "" {
		if date, err = parseDate(d, conf.timezone); err != nil {
			return err
		}
	}
	data := templateData{}
	data["date"] = date.Format("2006-01-02T15:04:05-07:00")
	data["title"] = title
	site.verbose("document title: %s\ndocument date: %s", data["title"], data["date"])
	text := defaultNewTemplate
	if site.from != "" {
		// Read document template file specified in `-var template=<template-file>` option.
		f := site.from
		if !fsx.FileExists(f) {
			return fmt.Errorf("missing document template file: %s", f)
		}
		if text, err = fsx.ReadFile(f); err != nil {
			return err
		}
	} else {
		// Attempt to read `new.md` document template file from site template
		// directory by searching along the corresponding template directory path.
		for d := fsx.PathTranslate(filepath.Dir(newFile), site.contentDir, site.templateDir); ; {
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
	site.verbose("document file: %s", newFile)
	if err := fsx.WriteFile(newFile, output.String()); err != nil {
		return err
	}
	return nil
}
