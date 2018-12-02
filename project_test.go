package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parseArgs(t *testing.T) {
	assert := assert.New(t)
	proj := newProject()

	args := []string{"hindsite", "serve", "./testdata/blog", "-content", "./testdata/blog/template/init"}
	assert.NoError(proj.parseArgs(args))
	assert.Equal(uint16(1212), proj.httpport, "httpport")
	assert.Equal(uint16(35729), proj.lrport, "lrport")
	assert.Equal(false, proj.drafts, "drafts")
	assert.Equal(true, proj.livereload, "livereload")
	assert.Equal(false, proj.navigate, "navigate")

	args = []string{"hindsite", "serve", "./testdata/blog", "-port", "1234"}
	assert.NoError(proj.parseArgs(args))
	assert.Equal(uint16(1234), proj.httpport, "httpport")

	args = []string{"hindsite", "serve", "./testdata/blog", "-port", "1234:8000", "-drafts"}
	assert.NoError(proj.parseArgs(args))
	assert.Equal(uint16(1234), proj.httpport, "httpport")
	assert.Equal(uint16(8000), proj.lrport, "lrport")
	assert.Equal(true, proj.drafts, "drafts")

	args = []string{"hindsite", "serve", "./testdata/blog", "-port", ":-1", "-drafts", "-navigate"}
	assert.NoError(proj.parseArgs(args))
	assert.Equal(true, proj.drafts, "drafts")
	assert.Equal(false, proj.livereload, "livereload")
	assert.Equal(true, proj.navigate, "navigate")

	args = []string{"hindsite", "illegal-command"}
	assert.Equal("illegal command: illegal-command", proj.parseArgs(args).Error())

	args = []string{"hindsite", "serve", "-illegal-option"}
	assert.Equal("illegal option: -illegal-option", proj.parseArgs(args).Error())

	args = []string{"hindsite", "serve", "missing-project-dir"}
	assert.Contains(proj.parseArgs(args).Error(), "missing project directory: ")

	args = []string{"hindsite", "serve", ".", "-content", "missing-content-dir"}
	assert.Contains(proj.parseArgs(args).Error(), "missing content directory: ")

	args = []string{"hindsite", "serve", "-port"}
	assert.Equal("missing -port argument value", proj.parseArgs(args).Error())

	args = []string{"hindsite", "serve", "./testdata/blog", "-port", "99999999"}
	assert.Equal("illegal -port: 99999999", proj.parseArgs(args).Error())

	args = []string{"hindsite", "serve", "./testdata/blog", "-port", ":99999999"}
	assert.Equal("illegal -port: :99999999", proj.parseArgs(args).Error())
}
