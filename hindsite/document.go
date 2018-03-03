package main

import (
	"time"
)

// Document TODO
type Document struct {
	Title    string
	Date     time.Time
	Synopsis string
	Addendum string
	Tags     []string
	Draft    bool
	path     string // File path relative to content directory root (no leading `./`).
	content  string // Markup text (without meta-data header).
	html     string // Rendered content.
}

// NewDocument TODO
func NewDocument(path string) *Document {
	// TODO
	result := new(Document)
	return result
}
