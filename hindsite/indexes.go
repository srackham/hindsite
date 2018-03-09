package main

type index struct {
	templateDir string      // The template directory that contains the index templates.
	buildDir    string      // The build directory that the index pages are written.
	docs        []*document // Parsed documents belonging to index.
}

type indexes []index

// Search templates directory for indexes and add them to indexes.
func (idxs *indexes) init(templateDir string) error {
	// TODO
	*idxs = indexes{} // Delete elements all.
	return nil
}

// Add document to all indexes that it belongs to.
func (idxs indexes) addDocument(doc *document) error {
	for i, idx := range idxs {
		if pathIsInDir(doc.templatepath, idx.templateDir) {
			idxs[i].docs = append(idx.docs, doc)
		}
	}
	return nil
}

// Build all indexes.
func (idxs indexes) build() error {
	for _, idx := range idxs {
		if err := idx.build(); err != nil {
			return err
		}
	}
	return nil
}

// Build index.
func (idx index) build() error {
	// TODO
	return nil
}
