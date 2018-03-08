package main

type index struct {
	// TODO
	docs []document
}

type indexes []index

// Indexes is the global indexes used by the build command.
// Each document contributes when it is parsed.
// Once all documents have been processed the indexes are built.
var Indexes indexes

// If document belongs to an index then add it.
// If necessary create and append new index.
func (idxs *indexes) add(doc *document) error {
	// TODO
	return nil
}

// Build all indexes.
func (idxs *indexes) build() error {
	for _, idx := range *idxs {
		if err := idx.build(); err != nil {
			return err
		}
	}
	return nil
}

// Build index.
func (idx *index) build() error {
	// TODO
	return nil
}
