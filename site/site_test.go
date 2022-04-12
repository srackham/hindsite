package site

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parseArgs(t *testing.T) {
	assert := assert.New(t)
	site := NewSite()

	args := []string{"hindsite", "serve", "./testdata/blog", "-content", "./testdata/blog/template/init"}
	assert.NoError(site.parseArgs(args))
	assert.Equal(uint16(1212), site.httpport, "httpport")
	assert.Equal(uint16(35729), site.lrport, "lrport")
	assert.Equal(false, site.drafts, "drafts")
	assert.Equal(true, site.livereload, "livereload")
	assert.Equal(false, site.navigate, "navigate")

	args = []string{"hindsite", "serve", "./testdata/blog", "-port", "1234"}
	assert.NoError(site.parseArgs(args))
	assert.Equal(uint16(1234), site.httpport, "httpport")

	args = []string{"hindsite", "serve", "./testdata/blog", "-port", "1234:8000", "-drafts"}
	assert.NoError(site.parseArgs(args))
	assert.Equal(uint16(1234), site.httpport, "httpport")
	assert.Equal(uint16(8000), site.lrport, "lrport")
	assert.Equal(true, site.drafts, "drafts")

	args = []string{"hindsite", "serve", "./testdata/blog", "-port", ":-1", "-drafts", "-navigate"}
	assert.NoError(site.parseArgs(args))
	assert.Equal(true, site.drafts, "drafts")
	assert.Equal(false, site.livereload, "livereload")
	assert.Equal(true, site.navigate, "navigate")

	args = []string{"hindsite", "illegal-command"}
	assert.Equal("illegal command: illegal-command", site.parseArgs(args).Error())

	args = []string{"hindsite", "serve", "-illegal-option"}
	assert.Equal("illegal option: -illegal-option", site.parseArgs(args).Error())

	args = []string{"hindsite", "serve", "missing-site-dir"}
	assert.Contains(site.parseArgs(args).Error(), "missing site directory: ")

	args = []string{"hindsite", "serve", ".", "-content", "missing-content-dir"}
	assert.Contains(site.parseArgs(args).Error(), "missing content directory: ")

	args = []string{"hindsite", "serve", "-port"}
	assert.Equal("missing -port argument value", site.parseArgs(args).Error())

	args = []string{"hindsite", "serve", "./testdata/blog", "-port", "99999999"}
	assert.Equal("illegal -port: 99999999", site.parseArgs(args).Error())

	args = []string{"hindsite", "serve", "./testdata/blog", "-port", ":99999999"}
	assert.Equal("illegal -port: :99999999", site.parseArgs(args).Error())
}
