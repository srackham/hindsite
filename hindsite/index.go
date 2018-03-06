package main

type index struct {
	// TODO
}

// Indexes is the global indexes used by the build command.
// Each document contributes when it is parsed.
// Once all documents have been processed the indexes are built.
var Indexes []index

func (idx *index) build() {
	// TODO
}
