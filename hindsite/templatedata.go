package main

type templateData map[string]interface{}

// Merge in data from another data map.
func (data templateData) add(from templateData) {
	for k, v := range from {
		data[k] = v
	}
}
