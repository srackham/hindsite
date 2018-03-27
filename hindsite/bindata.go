// Code generated by go-bindata.
// sources:
// ../examples/builtin/blog/template/config.toml
// ../examples/builtin/blog/template/favicon.ico
// ../examples/builtin/blog/template/layout.html
// ../examples/builtin/blog/template/main.css
// ../examples/builtin/blog/template/newsletters/config.toml
// ../examples/builtin/blog/template/newsletters/docs.html
// ../examples/builtin/blog/template/newsletters/layout.html
// ../examples/builtin/blog/template/partials.html
// ../examples/builtin/blog/template/posts/docs.html
// ../examples/builtin/blog/template/posts/layout.html
// ../examples/builtin/blog/template/posts/tag.html
// ../examples/builtin/blog/template/posts/tags.html
// ../examples/builtin/minimal/template/index.md
// ../examples/builtin/minimal/template/layout.html
// DO NOT EDIT!

package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (fi bindataFileInfo) Name() string {
	return fi.name
}
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}
func (fi bindataFileInfo) IsDir() bool {
	return false
}
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _blogTemplateConfigToml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x1c\xc9\xb1\x4e\xc3\x30\x10\x06\xe0\xdd\x4f\x61\x39\x73\xe3\xb6\x03\x1b\x03\x3b\x62\x40\x62\x60\x42\x26\xfe\x9b\x9c\xb0\x7d\xd6\xdd\x05\x02\xe2\xe1\xab\x64\xfd\xbe\xc1\x77\x41\x47\xcb\xc8\xde\xd8\xeb\x6f\xb3\x05\x4a\x8a\xec\x85\xd9\x4e\x82\x92\x8c\xbe\xe1\xdf\x5e\x9f\x75\x74\xc3\x2a\xa5\x0b\x6e\xb4\xf9\x47\x1f\xe2\x67\xe1\x39\x38\x37\x60\x9b\xca\x9a\xb1\xdb\xc4\xed\x46\xf3\x28\x75\xfd\xaf\x89\x5a\xdc\xdb\xa8\xe2\x8f\xdb\xf1\x4f\x15\x42\x53\x8a\x2f\xf8\xf9\x78\x67\xf9\x0a\x6e\xd0\x85\xc5\x72\xb2\xe3\xaf\xf1\x12\xaf\xe7\xf3\x43\x70\x6e\xe1\x8a\x9e\xe6\x83\xa9\x65\x6c\xd0\xd8\x59\x4d\x63\xe6\x49\x4f\x97\x71\xb1\x5a\xc2\x3d\x00\x00\xff\xff\x73\xbe\x8e\xc0\xc2\x00\x00\x00")

func blogTemplateConfigTomlBytes() ([]byte, error) {
	return bindataRead(
		_blogTemplateConfigToml,
		"blog/template/config.toml",
	)
}

func blogTemplateConfigToml() (*asset, error) {
	bytes, err := blogTemplateConfigTomlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "blog/template/config.toml", size: 194, mode: os.FileMode(436), modTime: time.Unix(1521966860, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _blogTemplateFaviconIco = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x94\x93\x7d\x68\x5b\x65\x18\xc5\x4f\xee\x6d\xda\xb5\x1b\x33\xb3\xb8\xb6\x58\x5d\xe7\x98\x94\x42\xd5\x3f\x74\x42\x3b\xd4\x62\x61\x9d\xd2\x8a\xd0\xd6\x2f\xb4\xa8\xb3\x5b\xe6\xcc\xc2\x1c\xb4\x93\xab\xf9\x68\xf7\x91\xac\xcd\x58\xb0\x59\x2c\x31\x53\x3b\xd6\x99\xb8\x39\x24\x8d\x32\x37\xc8\x36\x2c\x41\x62\x35\x19\x6d\x26\x0e\xc9\xb2\x25\x5d\xee\xd2\xae\x95\x38\xb6\x56\x77\x24\x09\x17\x86\x50\xa9\x07\x1e\x1e\x78\xe1\x77\xde\xf7\xe5\x79\x0e\xa0\x82\x0a\x1a\x4d\xb6\x57\x61\x4b\x01\xb0\x12\x40\x35\x00\x0d\x80\x2a\xe4\xcf\x15\xcd\x94\xe4\x6b\x21\x15\x0a\x58\x52\xb3\x14\xaf\x97\x08\x78\x09\x40\x45\x0e\xff\x1f\x6a\x2c\x13\xb5\xe6\xa7\x1e\xa0\xf9\xb1\xe2\x4c\xe7\x43\x42\xf8\x85\xfb\xe0\x59\x5d\x80\x9d\x02\xd0\x04\xa0\x12\x80\x7a\x21\xf6\xf1\xe5\xa8\xdf\xfb\xc4\xf2\xf4\xdb\x9d\x9b\xb9\x5b\xea\xe6\xe7\xbd\x5d\xfc\x52\xda\xcc\x2f\xb6\x6c\xe0\xc1\x67\x4b\x67\xdf\xaf\x15\x2e\xb4\xde\x8f\x93\x6b\x0b\xd1\xa5\x06\x5a\x44\x15\x1e\xbc\x9b\xb7\xaf\xc7\xa9\x6f\x5a\x0b\xd9\xfe\x62\x1b\x77\x99\x76\xf3\x63\xe7\x20\x5d\x43\xc3\x3c\xea\x3d\x41\xdf\xf1\x61\x06\x8e\xda\x39\xea\x36\xf2\x7c\xdf\x9b\x74\xb6\x54\xc8\x95\x6a\x34\x2a\xec\x73\x65\xd8\x74\x49\x0f\xa6\x2d\xcb\x38\xd2\x51\xc1\x03\xba\x57\x69\x39\x30\xc0\x7e\x9b\x8d\x0e\x87\x83\x2e\xf7\x61\x1e\x39\xe6\xe1\xd7\x23\xa7\x38\xfa\x53\x98\xda\x37\xda\x5d\x0a\xab\x11\xb0\xc2\xdb\x8c\x89\x79\x47\x31\x63\x96\xb5\xbc\x2a\x15\x30\xac\x55\xf3\x84\xe1\x2d\x9e\x1c\xf9\x9e\xc7\xbc\x5f\xd1\xe7\xf3\xd1\xef\xf7\xd3\xe3\xf1\xd0\xef\xff\x96\x6d\xed\x2f\x7f\xa6\xf0\x5b\x6b\xb0\xef\xa6\x05\x4c\x1e\xac\xe1\x6f\x86\x72\x26\x3e\x02\x13\x1f\x82\x17\xb7\x0b\x3c\x27\x3d\xcf\xe0\xd9\xb3\xfc\xf1\xe7\x08\xc7\xc6\xc6\x98\x4c\x26\x39\x3e\x3e\x7e\xa7\xa1\xa1\xa1\x33\xcb\x3e\xb2\x02\x0f\x47\xf5\x90\x6f\xbb\xab\x78\xd9\xb2\x86\xd7\x7b\xc1\x6b\x26\x30\x61\x40\xce\xe7\x77\x3d\x18\xdc\x51\xcd\xc8\xf9\x33\xfc\x35\x96\x60\x38\x1c\x66\x47\x47\x87\x5b\x99\xab\xab\x11\x7d\xf3\x83\xf7\x72\xd2\xb1\x8e\x69\xcb\x12\xa6\x7b\xc1\xa4\x51\x60\xca\x9c\xed\xf9\x8a\x77\x81\xa1\xf7\x56\xf2\xe2\xe9\xe3\x9c\xca\xcc\x73\xeb\xbb\xdb\x86\x95\x39\x86\xb6\x09\x1e\xf9\xd3\x3a\x4e\xd9\x2b\x99\xb1\x82\x09\x4b\x39\xaf\xba\x37\x32\xd9\x53\x92\xf3\x98\x34\xe5\x3d\x12\x12\x18\xd5\x97\xf0\x3b\x43\xeb\xdc\xa3\xb5\xb5\xbb\x94\xbf\x4f\xae\x13\xd6\x47\xbb\x2b\xe2\x73\x4e\x91\x57\x24\xf0\xd2\xa1\x26\xde\x38\xbd\x83\xb2\xf7\x35\x5e\xdb\x57\x4a\xd9\x04\xa6\x7a\xf2\x1e\x72\x0f\x18\xdb\x0e\xbe\x52\x85\xc3\x77\xef\xe5\x0f\x2d\x58\x1d\xdd\x84\xc0\xac\x0d\x94\xad\x1a\xc6\xdd\xcd\xbc\x71\x66\x27\xa7\x7d\x9d\x4c\xd9\xd7\xe4\xb8\xa9\xbd\xe0\xb4\x19\x74\x6d\xc4\x68\x59\x31\x56\xfd\x7b\xf7\x22\x1b\xb0\xf4\x97\x77\x30\x90\x7d\xf3\xcc\xfe\x42\x5e\xf9\xe4\x19\x5e\xf7\xeb\x98\x74\x3e\xc9\x59\x2b\xf8\xc7\x1e\xd0\x50\x8f\xa1\x65\x42\x2e\x5e\x0b\x2a\xd4\x86\xe6\xb8\x04\x39\x63\x53\x33\x65\x2b\xe7\x9f\xfd\x02\x6f\x5a\x41\xa9\x0e\xae\x02\xa0\x78\x11\x11\xc2\x50\x13\xaa\x2f\xe8\x10\xfc\x7b\x00\xbc\xdd\x0f\x7e\x50\x8f\x43\x22\x50\xb4\x18\x56\xd1\xb9\x56\x94\x4e\x68\x31\xd8\x5d\x07\xa3\xf8\x1f\xf7\x92\xc0\x2d\xc3\x1e\xce\x14\x05\x53\x31\xd5\xdc\x74\x4c\x35\x91\x08\xa8\x8e\x44\x8c\xe2\xfe\x88\x51\xd4\x85\x20\xf6\x5d\x46\x91\x8d\x46\x11\x42\xa0\x08\x88\xdd\x03\xdc\x5a\x05\xfc\xf5\x34\x70\xc7\x00\xfc\x13\x00\x00\xff\xff\x45\x44\x84\x4f\x7e\x04\x00\x00")

func blogTemplateFaviconIcoBytes() ([]byte, error) {
	return bindataRead(
		_blogTemplateFaviconIco,
		"blog/template/favicon.ico",
	)
}

func blogTemplateFaviconIco() (*asset, error) {
	bytes, err := blogTemplateFaviconIcoBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "blog/template/favicon.ico", size: 1150, mode: os.FileMode(436), modTime: time.Unix(1521964427, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _blogTemplateLayoutHtml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xaa\xae\x2e\x49\xcd\x2d\xc8\x49\x2c\x49\x55\x50\xca\x48\x4d\x4c\x49\x2d\x52\x52\xd0\xab\xad\xe5\xb2\xc9\x30\xb4\xab\xae\xd6\x2b\xc9\x2c\xc9\x49\xad\xad\xb5\xd1\xcf\x30\xb4\xe3\xaa\xae\xd6\x4b\xca\x4f\xa9\xac\xad\xe5\x42\xd6\x95\x96\x9f\x5f\x02\xd5\x05\x08\x00\x00\xff\xff\xd3\x36\x34\x46\x4d\x00\x00\x00")

func blogTemplateLayoutHtmlBytes() ([]byte, error) {
	return bindataRead(
		_blogTemplateLayoutHtml,
		"blog/template/layout.html",
	)
}

func blogTemplateLayoutHtml() (*asset, error) {
	bytes, err := blogTemplateLayoutHtmlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "blog/template/layout.html", size: 77, mode: os.FileMode(436), modTime: time.Unix(1521964908, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _blogTemplateMainCss = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xd4\x58\x5d\x6f\xdb\xbc\x15\xbe\xf7\xaf\x38\x73\x50\xa0\x0d\x24\x55\x76\xec\x26\x95\x81\x60\x45\x2f\xb6\x8b\x6d\x28\xb6\x5e\xf6\x86\x16\x69\x8b\x28\x45\x6a\x14\x1d\xdb\x2d\xfc\xdf\x07\x7e\x4a\x94\x25\x27\x7b\xef\x5e\x07\x48\x62\x9e\x0f\x1e\x92\xe7\x3c\xe7\x21\xef\xe1\xf7\x0c\x00\xa0\x46\x72\x4f\x79\x01\xf9\xc6\x7c\x6d\x10\xc6\x94\xef\xc3\xf7\x63\x45\x15\x49\xdb\x06\x95\xa4\x00\xca\x2b\x22\xa9\xb2\x92\xad\x38\xa5\x2d\xfd\x65\x94\xb7\x42\x62\x22\xd3\xad\x38\x6d\x66\x97\xd9\x56\xe0\xb3\xf3\xbe\x13\x5c\x69\x2d\x52\xc0\x32\x6f\x4e\xd6\xb2\x14\x4c\xc8\x02\xee\x1e\xcc\xc7\x79\x43\xe5\xcf\xbd\x14\x07\x8e\x53\x27\x36\x33\x6f\x3a\x2f\x3b\x54\x53\x76\x2e\x60\xfe\xed\x3b\xfc\x87\x48\xba\x9b\x27\xf0\x37\x22\xe4\x9e\xa2\x04\xe6\xdf\x69\x4d\x5a\xf8\x17\x39\xc2\xbf\x45\x8d\xf8\x3c\x81\x56\xeb\xe8\x70\xaa\x45\x02\xd5\x32\x81\xea\x21\x81\x6a\x95\x40\xb5\x4e\xa0\xfa\x34\x58\xfe\x22\x6f\x4e\x7e\xcd\x3e\x3e\x54\xae\x16\xab\xe5\x54\x08\x88\xb7\xf3\x04\xfe\x4e\xd8\x0b\x51\xb4\x44\x09\x7c\x91\x14\xb1\x04\x5a\xc4\xdb\xb4\x37\x39\xfc\x8e\x76\x21\x5b\x4a\x52\x6f\xe0\x32\xab\x96\xb1\x64\x91\xad\xbd\xe4\x61\x28\x09\x36\xab\xa1\x64\xe1\x25\xeb\xa1\x24\xf7\x92\x4f\xb1\x24\xcf\x3e\x87\x08\x6e\x6e\x4a\xaa\x44\x13\x1c\xf5\xc7\xb7\x42\x29\x51\x6b\x4f\x26\xae\x99\x5d\xcb\x88\xe5\x3a\x58\x86\x04\xb1\x96\x8b\xe6\x04\xad\x60\x14\xc3\x1d\x36\x1f\xed\x84\xd6\xfb\xe0\xe5\x94\x1e\x29\x56\x95\x3e\x97\xfc\x9d\x16\x2a\xb4\x65\xc4\x89\x9d\xb3\x52\x30\x86\x9a\x96\x14\xe0\xff\xb3\x73\x19\xd5\x94\xa1\xb3\x38\xa8\x02\x76\xf4\x44\xb0\xcb\xe5\x81\x4b\x9c\x80\xaa\x9c\x4f\x45\x4e\x2a\x45\x8c\xee\x79\x01\x8c\xec\x5c\x8e\xbf\x10\xa9\xcf\x96\x79\x89\x12\x4d\x08\x26\xb3\x61\x10\x0c\xcf\xa0\x4c\xca\x3f\x83\x92\xfa\x57\x95\xbc\xa6\x81\xa3\x95\x4c\xec\x07\x4a\x00\x15\x95\x78\x21\xd2\x69\xfb\xbc\x5c\x2d\x9f\xb6\x25\x32\x2a\xfd\xe8\x31\x29\x85\x44\x8a\x0a\x5e\x00\x17\x9c\x18\x85\xc8\xc1\x95\xda\x81\x63\x22\x19\xb5\xba\x4d\x02\x8d\x24\x09\x30\x9a\x00\x56\x09\x60\x9c\xc0\x96\x89\xf2\xe7\x7f\x0f\x42\x11\x68\x9c\x13\xad\x9e\x56\x84\xee\x2b\xa5\xcf\x78\xb5\xde\x5c\x9d\x7c\xde\x3b\xf9\xab\x9c\x59\xbb\x9c\x39\xb0\x04\x04\x73\x4e\x87\xb0\x13\x50\x09\x72\x5b\x9a\xcb\x75\x63\xb0\x05\xab\x38\xd3\x82\xdf\x5e\x99\xb6\xea\xcc\x34\x5a\x29\xc4\x68\x69\x8c\x70\x6c\xa4\x0f\x78\x34\xb3\x6d\xf4\x06\xc3\xba\x85\x0f\x23\xb4\x41\xad\x3d\x9a\xf9\x33\xb4\x5b\xde\xcb\x4f\x3b\xcb\xba\x3b\x5a\x62\x3e\x51\x2e\x3e\xe9\x54\x8c\x70\x28\x5b\x93\x7a\x08\x44\x0f\x4f\xfa\x47\x87\x95\x95\x34\x44\xe4\xa5\x8f\x8f\x8f\x03\xec\x5e\x5b\x2c\x73\xea\x45\xb1\x25\x3b\x21\x3b\x33\xae\x08\x57\x05\xcc\x7f\x2c\xf3\xc5\x0a\x7e\x2c\xf3\xfc\xf3\x5c\x6b\x97\x02\x93\x3e\x6e\x7b\xb8\xfb\x26\xd1\xbe\x46\x0a\x7d\x93\x22\x81\x7f\x12\xce\xf4\x1f\xc1\x51\x29\x12\x98\x7f\xe1\x18\x31\xa2\xbf\x8b\x79\x02\xf3\x7f\x1c\x4a\x8a\x11\x7c\x15\xbc\x15\x8c\xe8\x91\xaf\xe2\x20\x29\x91\x1a\x9b\xe7\x09\xd4\x82\x0b\xd3\x4d\x6c\xc8\x1f\xef\xe3\xa9\xac\x1d\x6a\xaf\x1c\xfb\xf9\x3a\x7b\xb8\xff\x38\xd1\x39\xee\x76\x4f\xfa\xc7\xe4\xb4\xfc\x13\xae\xa8\x07\xd6\x9f\x7d\x7e\x7c\xbc\x8f\x0b\x8f\x0b\x59\x23\x16\x4c\xa2\x36\xdd\x48\x92\x1e\x25\x6a\xa6\x3a\x6b\xd8\x9f\x57\x20\xa8\x97\xcb\x12\x61\x7a\x68\xaf\x59\xc2\x27\x5b\x95\x19\xd3\x51\xa5\xdd\x54\x1e\xdf\xde\x36\x77\x57\x3a\xc3\x75\x68\xe7\xf7\x05\x17\xea\x7d\x23\xc9\x07\x78\x86\x5e\x8e\x46\xba\x5c\xd8\x15\x5f\x66\x19\x66\x69\x25\x24\xfd\x25\xb8\x42\x0c\x9e\x61\x58\xfd\x51\x5b\xb3\x29\x12\xf9\xbd\x0e\x3a\xe2\x3d\x7d\xa8\xb2\xd6\x6d\x83\xb8\xb3\x15\x0d\x2a\xa9\x3a\x17\xb0\x80\xbf\xd0\xba\x11\x52\x21\xae\x4c\x54\x2f\x44\xb6\x64\x48\x3a\x4c\xb1\x5b\x62\x74\x99\x61\xfa\xe2\xb4\x34\x12\x47\x06\x51\x46\x05\xd2\x63\x38\x4f\x20\x3a\xaf\xa4\x41\x8c\xcf\x8f\x37\x00\xfa\xd1\x23\x74\xd6\x52\x4c\xb6\x48\xbe\xde\xac\xa6\x33\xa5\xdb\x4c\x7d\xf6\xbb\xdd\xce\x03\x60\xd8\xc6\x45\xe0\x85\x57\x18\x68\xf7\xc4\x47\x71\x5f\xec\xa8\x6c\x55\x5a\x56\x94\x8d\x1d\x69\x47\x47\xb2\x06\xed\x49\xba\x95\x04\xfd\x0c\xe8\xed\x07\x52\x0b\x86\x05\x20\x76\x44\xe7\xd6\xa8\x73\x91\xde\xb2\xa0\x5c\x87\x50\x00\x7a\x11\xd4\x74\xe6\xcc\xd0\x01\x03\xf0\x93\xf4\x21\x68\x95\x84\xab\xb8\x03\x3b\x3d\x2b\xe8\x69\x4a\x5d\x44\x23\x8a\x66\xdc\xae\x4b\x92\x96\xc8\x17\x17\x58\x3b\x56\x08\xae\x68\x32\x83\x16\x37\xd5\xd2\xf1\x8a\x71\xdd\x22\x81\xc1\x30\xda\x75\xcb\xc0\xb4\x6d\x18\x3a\x17\x96\x6b\x6d\x06\x7d\x65\xbe\xb9\x26\x0a\xf9\xc8\x44\x7d\x8f\x25\x23\x48\xea\x8b\x84\xaa\xc6\x8b\xd8\x6f\xcc\x8e\x09\xa4\xfa\x1c\xcd\x59\x76\x03\x9a\xf0\xec\x98\x38\x16\x50\x51\x8c\x09\xdf\x74\x1b\xda\x89\x08\x63\xb4\x69\x69\x3b\x52\x37\x1e\x4a\xa6\x69\xe1\x14\x3f\x76\x4d\x7d\xb9\x7e\xf7\x26\x20\xb2\x04\xc1\x68\xf7\x2a\x62\xc0\x4e\x2e\x33\xcc\x42\xe7\x3e\xe8\x8c\x49\x75\x12\xa8\x02\x30\x4b\xdd\x88\xd3\xea\xed\x92\x57\xa5\xbc\x94\xa4\x36\xc7\x12\xab\x8b\x09\xa7\x62\xa8\xf5\x0c\x8c\x4e\x3b\x8d\xd5\x0f\x13\x4e\x0f\x43\xad\xdb\x4e\x63\xf5\xac\x0b\x7c\x82\xc8\x38\xe9\xfb\x4e\xf1\x03\xcc\xc1\x50\x9a\x4c\xbc\xd5\x58\x8c\x18\x1f\xde\x6a\x7c\x18\x31\xc6\x2c\xe5\x87\x7a\xeb\xf8\x3f\x56\xff\x4f\xf0\x99\xf5\xf1\xd7\x9a\x60\x8a\xa0\x91\x94\xfb\x83\x35\x58\xa5\xbf\x27\xdd\xbf\xe0\xef\xf2\x51\x69\xea\xa6\x1a\x35\x21\x2d\xbd\x98\xdf\xf1\xad\x60\xf2\x02\x31\x62\x7d\x99\xdd\xd5\x88\xfa\x6e\xf7\x2b\xa5\x1c\x93\x53\x01\x0b\x57\x17\x94\x77\x37\x83\xbc\xa3\xb6\xe1\x36\xb7\xbe\x6a\x3b\x36\xd9\xf3\xb1\x0a\x58\x85\xc6\xe0\xc7\xa5\x75\xbd\x72\x0d\xf3\x0e\xe9\xdb\x19\x1b\x30\x74\x57\x96\x8e\x98\xd4\x84\x1f\xc2\xc5\xa5\x75\x77\x83\x54\x9d\x1b\xd2\xa7\x1d\xaf\x3c\x82\x8c\x23\xca\x08\xb9\xf1\x2f\x06\x7e\xe2\x90\xe5\x57\xa8\x35\x68\x5c\x37\xda\xf1\x32\xb4\x63\xe7\x72\xba\x0d\x7a\x40\xe9\xaf\x9d\x51\x40\x43\xd4\x36\x97\x9b\xe8\x8a\xd1\x7b\x6c\x19\x6b\x51\xf1\xa6\x34\xa7\xde\xf5\x67\xf2\xea\xd9\x4d\x1f\x65\xdb\xd8\xae\xed\xd6\xcb\xf5\xca\xd8\x6c\x11\xe7\x41\x75\x70\x2a\x63\xcf\x32\xd1\x95\xa7\x4b\x98\xfe\x83\xcb\x2a\x1a\x3c\xba\xf4\xdc\x0a\x86\xfb\x33\x36\xd7\x73\x06\x61\x02\xfd\x34\xfa\xa3\x2f\x41\x99\x69\x54\x13\xed\xae\x47\x25\xa2\x54\xb9\xcc\xb2\x3e\x2b\x70\xb2\x40\x08\x1c\x40\xb4\xa5\x24\x84\x03\xe2\x18\xde\xeb\x1a\xf4\x0f\x1d\xcd\xe9\x83\x1b\xec\x4a\xf0\x29\xcf\xf5\xb0\xa7\x74\xe1\xa5\x6e\xb8\x6d\x61\x2f\x2d\x64\xf4\xcb\x7e\xa4\x59\x05\xed\x91\x72\x5d\x0c\x5c\xf5\xd1\xca\x57\x96\xe1\x22\x05\x58\xbe\x75\x14\x12\xf7\x0d\x6e\x24\xfe\x14\x92\x38\xc3\x28\x9d\x6e\x85\x7d\x19\x47\xdb\xfe\xaa\xc7\xa7\xb9\xfc\x2f\x00\x00\xff\xff\x27\x4e\xe9\x8a\x48\x15\x00\x00")

func blogTemplateMainCssBytes() ([]byte, error) {
	return bindataRead(
		_blogTemplateMainCss,
		"blog/template/main.css",
	)
}

func blogTemplateMainCss() (*asset, error) {
	bytes, err := blogTemplateMainCssBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "blog/template/main.css", size: 5448, mode: os.FileMode(436), modTime: time.Unix(1522111112, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _blogTemplateNewslettersConfigToml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x2a\x48\x4c\xcf\xcc\x4b\x2c\x49\x55\xb0\x55\xd0\x35\xe4\x02\x04\x00\x00\xff\xff\x4f\x54\x72\x02\x0e\x00\x00\x00")

func blogTemplateNewslettersConfigTomlBytes() ([]byte, error) {
	return bindataRead(
		_blogTemplateNewslettersConfigToml,
		"blog/template/newsletters/config.toml",
	)
}

func blogTemplateNewslettersConfigToml() (*asset, error) {
	bytes, err := blogTemplateNewslettersConfigTomlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "blog/template/newsletters/config.toml", size: 14, mode: os.FileMode(436), modTime: time.Unix(1522110443, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _blogTemplateNewslettersDocsHtml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x74\x90\x41\x6e\x03\x21\x0c\x45\xf7\x39\x85\xc5\x01\x40\xc9\xda\xe5\x08\xbd\x83\x1b\x9c\x21\x12\x1d\x46\xe0\xa8\x0b\xcb\x77\xaf\xa0\x55\x66\x5a\xb5\x6c\x80\xe7\x67\xc4\xb7\xaa\xf0\xfb\x56\x48\x18\x5c\x66\x4a\xdc\x1c\x78\xb3\x13\xe6\x73\x7c\xe5\x8f\x5e\x58\x84\x5b\xc7\x90\xcf\xf1\x04\x00\x80\x8f\xf2\x75\x18\x4b\xb5\xd1\xba\x30\xf8\x54\xaf\xdd\xec\xc9\xb1\xdc\x77\x69\x82\x7c\x89\x48\x90\x1b\xdf\x5e\x9c\xaa\x7f\xb4\x62\xe6\xa2\xaa\x97\xbb\x14\x36\xc3\x40\x11\x43\xbe\xfc\x6a\xeb\x1b\xad\xc3\x2a\x75\x5d\x12\xc9\x14\x27\xfb\x53\xdb\xa3\xa4\x7a\x15\x5a\xba\x03\x3f\xb6\x7f\xba\xde\xda\x0e\x30\x1c\xbf\xac\xca\x6b\xfa\xce\x83\x61\x24\x3e\x3e\xbe\xd1\xc2\xb7\x5a\x65\xce\x6a\x5c\xcc\x7e\xd4\x9f\x35\xb3\xcf\x00\x00\x00\xff\xff\x7b\x2f\x5a\x23\x5f\x01\x00\x00")

func blogTemplateNewslettersDocsHtmlBytes() ([]byte, error) {
	return bindataRead(
		_blogTemplateNewslettersDocsHtml,
		"blog/template/newsletters/docs.html",
	)
}

func blogTemplateNewslettersDocsHtml() (*asset, error) {
	bytes, err := blogTemplateNewslettersDocsHtmlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "blog/template/newsletters/docs.html", size: 351, mode: os.FileMode(436), modTime: time.Unix(1522111303, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _blogTemplateNewslettersLayoutHtml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x84\x91\xc1\x6d\xeb\x30\x0c\x86\xef\x9e\x82\x4f\x77\xcb\x48\x8e\x86\xc2\x11\x5e\x67\x50\x2a\x3a\x32\x40\x5b\x86\xc2\xa6\x0d\x04\x4d\xd4\x11\x7a\xcb\x64\x85\x14\xbb\x08\x5a\x14\xd5\x8d\xe2\x4f\x7c\x1f\xc8\x94\x84\xa6\x85\xad\x10\x28\x4f\xd6\x51\x54\xa0\x73\x6e\xcc\xbf\xb6\x05\xe3\x77\x98\x92\x96\x51\x98\x72\x36\x9d\xdf\x21\xb4\x2d\x36\xc6\xef\xcb\x3f\x87\xf9\xe4\xac\xdc\x5b\x7b\x6c\x52\xd2\xc7\xe0\xae\x65\xda\x8d\x17\x6c\x00\x00\xcc\x79\xb1\x33\x9c\xe5\xca\x74\x50\x03\x07\x2b\x3d\x30\x0d\xa2\xf0\xf6\x5e\x03\x35\x54\x61\x34\xe1\x7f\x7a\xa5\xd8\x9b\x8e\xa6\x3b\xe8\x2b\x60\xc1\x47\x1a\x0e\x2a\x25\xbd\x44\xba\xe8\x97\xc8\x39\x2b\xdc\xca\xcd\xc2\xae\xd0\xae\x50\x7f\x17\x88\xe3\xc9\x8b\xc2\x9f\xfc\x27\x76\x7f\xf2\x67\x7a\x93\x07\x7e\x2d\xbf\xf1\xcb\xbb\x7d\x6c\x16\xa6\xab\xdb\x28\x3b\xd9\x3c\x9e\x99\x6c\xec\xe1\x18\xc4\x2b\x5c\xfb\x8f\x97\x18\x42\x90\xf5\x12\x9f\x01\x00\x00\xff\xff\x93\x4a\xcc\xec\xa1\x01\x00\x00")

func blogTemplateNewslettersLayoutHtmlBytes() ([]byte, error) {
	return bindataRead(
		_blogTemplateNewslettersLayoutHtml,
		"blog/template/newsletters/layout.html",
	)
}

func blogTemplateNewslettersLayoutHtml() (*asset, error) {
	bytes, err := blogTemplateNewslettersLayoutHtmlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "blog/template/newsletters/layout.html", size: 417, mode: os.FileMode(436), modTime: time.Unix(1522052079, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _blogTemplatePartialsHtml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xac\x53\xcd\x6e\xdc\x20\x18\xbc\xfb\x29\x08\xf7\x35\xea\xb5\xc2\x48\x51\xd3\x6b\xba\x52\x73\xe9\x11\x9b\x6f\x6d\x14\x0c\x16\x7c\xde\x1f\xb9\x3e\xf5\xb1\xfa\x38\x7d\x91\x0a\xff\xc5\xbb\xd9\xad\xda\x55\x73\x09\x7c\xeb\x19\x86\x99\xa1\xeb\x14\xec\xb4\x05\x42\x2b\x90\x0a\x3c\xed\xfb\x84\x3f\x3c\x7d\xf9\xf4\xf2\x6d\xfb\x99\x54\x58\x1b\x91\xf0\xf9\x1f\x48\x25\x12\x8e\x1a\x0d\x88\xae\x4b\x87\x45\xdf\x73\x36\x4e\x12\x5e\x03\x4a\x62\x65\x0d\x19\xdd\x6b\x38\x34\xce\x23\x25\x85\xb3\x08\x16\x33\x7a\xd0\x0a\xab\x4c\xc1\x5e\x17\xb0\x19\x36\x94\x89\x84\x1b\x6d\x5f\x09\x9e\x1a\xc8\x28\xc2\x11\x59\x11\x02\x25\x1e\x4c\x46\x03\x9e\x0c\x84\x0a\x00\x29\xa9\x3c\xec\x32\xda\x75\x69\xd0\x08\x69\xeb\x4d\xe3\x61\xa7\x8f\x7d\xcf\x6a\xa9\x6d\x1a\x41\x22\xe1\x6c\x92\x98\x3b\x75\x12\x09\x57\x7a\x4f\x0a\x23\x43\xc8\x68\x2e\xad\x05\x4f\x45\x42\x08\x21\xbc\x11\xc1\xcb\xe2\xb5\x92\x35\xf9\x4e\x72\xe3\x4a\xce\x9a\x08\x57\x7a\x2f\x12\xde\x9a\x19\x55\x83\x6d\x89\x75\x9b\xc6\x6b\x8b\x03\x98\x1b\x2d\xb8\xfc\x83\x1c\x6d\x15\x1c\x21\xb0\xc6\x05\x0c\x4c\xb9\x22\x6c\x3e\xa4\xd1\x40\x2a\xb6\x71\xc4\x99\x14\x9c\x19\x7d\x07\x19\xca\x32\x4c\x54\x2f\xb2\xbc\x8f\xc9\xc2\x21\x18\x40\x04\x7f\x21\xee\x19\x0e\xff\xc6\x28\x73\xd7\xe2\x04\x7e\x8c\xeb\x37\x34\x67\xad\x99\xec\xd7\x2a\xa3\x31\xa1\xd9\xf9\x79\x26\x3d\xea\xc2\x00\x15\x49\xd7\x81\x55\x7d\x9f\x24\x6f\x55\xdc\x39\x87\x63\x15\x07\xcc\x18\xcb\x6a\xc9\xd9\x14\x30\x1b\x9b\x79\x85\x42\xb9\x22\xda\x35\x73\x74\x9d\x97\xb6\x04\x92\x4e\xfb\x5f\x3f\x7e\x92\xf5\xed\x5a\x6f\xfa\x9e\x0e\xa5\x96\x65\xac\xb4\x14\x13\x6e\x64\xbe\x7e\x42\x54\xb1\xa8\x6c\x46\xc4\xb0\x0e\xe8\x9d\x2d\xc5\xb6\xcd\x8d\x0e\x15\xa8\x8f\x9c\x4d\x23\xd2\x75\xa9\x71\xb6\x54\x12\xe3\xd3\xc9\xfd\x7b\x54\xcc\xf6\x0c\x80\x50\x37\x46\xe2\xea\x56\x24\xca\x0c\x37\xf0\x8f\x2d\x56\xce\x9f\x1f\x29\x87\xd9\x0d\xc0\xd6\xc3\x7e\xf5\xf9\xda\x97\xc6\xc3\x7e\x65\xce\xb0\x5d\x9e\xbd\x14\x57\xe9\x9e\xe1\x88\x37\xe8\x2c\x1c\x71\x45\x37\x6c\x57\x74\x53\xc6\xcd\xec\x7d\x1a\x0d\xbe\xe9\xbe\xd1\x01\x17\xf7\x63\xe1\x66\x21\x97\x61\x0f\x1f\x8c\xad\x7e\xfb\x5b\xdc\x8e\xc7\xdf\x90\xbb\x6a\xc5\x8d\x3b\xaf\x99\x9e\x24\xc2\x5f\x05\xfd\x3f\xc2\x1e\x38\x1e\x36\x9b\x85\xe8\xeb\xc9\xba\x26\xe8\x73\xb2\x34\x4c\xd3\x11\x4d\x36\x9b\x0b\x86\xb3\xfc\xd8\xda\xa3\xd9\xf3\xf1\x97\xf6\xfa\x2b\x6b\x64\x09\x17\x8f\x75\xee\x94\x2c\x2f\xcc\xb0\x6d\x9d\x83\xef\x7b\xe2\x76\x83\x37\x32\xe0\x32\x5b\x74\xdc\x59\xc9\x85\x67\x9d\xcf\x9d\x7d\x7c\xc7\x35\xdd\xfb\x77\x00\x00\x00\xff\xff\xa9\x04\xad\x54\x2a\x07\x00\x00")

func blogTemplatePartialsHtmlBytes() ([]byte, error) {
	return bindataRead(
		_blogTemplatePartialsHtml,
		"blog/template/partials.html",
	)
}

func blogTemplatePartialsHtml() (*asset, error) {
	bytes, err := blogTemplatePartialsHtmlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "blog/template/partials.html", size: 1834, mode: os.FileMode(436), modTime: time.Unix(1522110862, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _blogTemplatePostsDocsHtml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xaa\xae\x2e\x49\xcd\x2d\xc8\x49\x2c\x49\x55\x50\xca\x48\x4d\x4c\x49\x2d\x52\x52\xd0\xab\xad\xe5\xb2\xc9\x30\xb4\x73\xcc\xc9\x51\x70\xca\xc9\x4f\x57\x08\xc8\x2f\x2e\x29\xb6\xd1\xcf\x30\xb4\xe3\x42\x56\x9f\x92\x9f\x9c\x93\x59\x5c\xa2\xa4\xa0\x97\x92\x9f\x5c\x5c\x5b\x8b\x22\x59\x90\x98\x9e\x9a\x96\x9f\x5f\x02\x36\x10\xc4\x41\x93\x87\xcb\xd5\xd6\x02\x02\x00\x00\xff\xff\xa5\xd3\x6c\xda\x84\x00\x00\x00")

func blogTemplatePostsDocsHtmlBytes() ([]byte, error) {
	return bindataRead(
		_blogTemplatePostsDocsHtml,
		"blog/template/posts/docs.html",
	)
}

func blogTemplatePostsDocsHtml() (*asset, error) {
	bytes, err := blogTemplatePostsDocsHtmlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "blog/template/posts/docs.html", size: 132, mode: os.FileMode(436), modTime: time.Unix(1522004962, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _blogTemplatePostsLayoutHtml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xaa\xae\x2e\x49\xcd\x2d\xc8\x49\x2c\x49\x55\x50\xca\x48\x4d\x4c\x49\x2d\x52\x52\xd0\xab\xad\xe5\xb2\xc9\x30\xb4\xab\xae\xd6\x2b\xc9\x2c\xc9\x49\xad\xad\xb5\xd1\xcf\x30\xb4\xe3\x42\x56\x9b\x92\x9f\x9c\x94\x9f\x52\x09\x51\x8c\x2c\x91\x96\x9f\x5f\x02\x35\x04\x10\x00\x00\xff\xff\x2d\x68\xd0\x94\x5c\x00\x00\x00")

func blogTemplatePostsLayoutHtmlBytes() ([]byte, error) {
	return bindataRead(
		_blogTemplatePostsLayoutHtml,
		"blog/template/posts/layout.html",
	)
}

func blogTemplatePostsLayoutHtml() (*asset, error) {
	bytes, err := blogTemplatePostsLayoutHtmlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "blog/template/posts/layout.html", size: 92, mode: os.FileMode(436), modTime: time.Unix(1521967333, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _blogTemplatePostsTagHtml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xaa\xae\x2e\x49\xcd\x2d\xc8\x49\x2c\x49\x55\x50\xca\x48\x4d\x4c\x49\x2d\x52\x52\xd0\xab\xad\xe5\xb2\xc9\x30\xb4\xab\xae\xd6\x2b\x49\x4c\xaf\xad\x55\x70\xca\xc9\x4f\x57\x08\xc8\x2f\x2e\x29\xb6\xd1\xcf\x30\xb4\xe3\x42\xd6\x94\x92\x9f\x9c\x93\x59\x5c\xa2\xa4\xa0\x97\x92\x9f\x5c\x5c\x5b\x8b\x22\x59\x90\x98\x9e\x9a\x96\x9f\x5f\x02\x36\x15\xc4\x41\x93\x87\xca\xd5\xd6\x02\x02\x00\x00\xff\xff\xd5\xeb\x42\xbc\x87\x00\x00\x00")

func blogTemplatePostsTagHtmlBytes() ([]byte, error) {
	return bindataRead(
		_blogTemplatePostsTagHtml,
		"blog/template/posts/tag.html",
	)
}

func blogTemplatePostsTagHtml() (*asset, error) {
	bytes, err := blogTemplatePostsTagHtmlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "blog/template/posts/tag.html", size: 135, mode: os.FileMode(436), modTime: time.Unix(1522004995, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _blogTemplatePostsTagsHtml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x4c\xcd\xb1\xad\xc2\x30\x14\x85\xe1\xde\x53\x5c\xb9\x7a\xaf\x71\x94\xde\x78\x0a\x16\xb8\x22\x37\x76\x11\x6c\xe4\x38\xd5\xd1\xa9\x18\x8b\x71\x58\x04\x45\xa2\xa0\xff\xf5\x7f\xc0\xb0\xfb\x63\xd3\x61\xe2\x8b\xe9\x62\xdd\x4b\x20\x5d\x2c\x73\xba\x6a\xde\xe3\x54\xe6\xe4\x80\xae\x35\x9b\x84\xa1\x79\x27\xdd\xfb\xf9\x92\xa8\x52\xba\xad\x17\x0f\x84\xa3\x6f\xa4\x4f\xc0\x19\x90\xf2\x07\x84\x5b\x3b\xea\x20\xff\xe3\xa4\xe7\xc0\xea\x42\xba\x5f\x6e\x6d\x6d\x7c\xb9\x4f\x00\x00\x00\xff\xff\x3a\xf3\xb1\xd5\x86\x00\x00\x00")

func blogTemplatePostsTagsHtmlBytes() ([]byte, error) {
	return bindataRead(
		_blogTemplatePostsTagsHtml,
		"blog/template/posts/tags.html",
	)
}

func blogTemplatePostsTagsHtml() (*asset, error) {
	bytes, err := blogTemplatePostsTagsHtmlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "blog/template/posts/tags.html", size: 134, mode: os.FileMode(436), modTime: time.Unix(1522004422, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _minimalTemplateIndexMd = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xf2\x48\xcd\xc9\xc9\x57\xd0\x0a\xcf\x2f\xca\x49\xd1\x52\xe4\x02\x04\x00\x00\xff\xff\xac\xd1\x81\x2d\x0f\x00\x00\x00")

func minimalTemplateIndexMdBytes() ([]byte, error) {
	return bindataRead(
		_minimalTemplateIndexMd,
		"minimal/template/index.md",
	)
}

func minimalTemplateIndexMd() (*asset, error) {
	bytes, err := minimalTemplateIndexMdBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "minimal/template/index.md", size: 15, mode: os.FileMode(436), modTime: time.Unix(1522006558, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _minimalTemplateLayoutHtml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xb2\x51\x4c\xc9\x4f\x2e\xa9\x2c\x48\x55\xc8\x28\xc9\xcd\xb1\xe3\xb2\xc9\x48\x4d\x4c\xb1\xb3\x29\xc9\x2c\xc9\x49\xb5\xab\xae\xd6\x03\x33\x6a\x6b\x6d\xf4\x21\x22\x36\xfa\x60\x79\x2e\x9b\xa4\xfc\x94\x4a\x90\x3c\x88\x06\x49\x83\xf9\x5c\x80\x00\x00\x00\xff\xff\xf7\x95\xf9\xaa\x4e\x00\x00\x00")

func minimalTemplateLayoutHtmlBytes() ([]byte, error) {
	return bindataRead(
		_minimalTemplateLayoutHtml,
		"minimal/template/layout.html",
	)
}

func minimalTemplateLayoutHtml() (*asset, error) {
	bytes, err := minimalTemplateLayoutHtmlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "minimal/template/layout.html", size: 78, mode: os.FileMode(436), modTime: time.Unix(1520106920, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"blog/template/config.toml": blogTemplateConfigToml,
	"blog/template/favicon.ico": blogTemplateFaviconIco,
	"blog/template/layout.html": blogTemplateLayoutHtml,
	"blog/template/main.css": blogTemplateMainCss,
	"blog/template/newsletters/config.toml": blogTemplateNewslettersConfigToml,
	"blog/template/newsletters/docs.html": blogTemplateNewslettersDocsHtml,
	"blog/template/newsletters/layout.html": blogTemplateNewslettersLayoutHtml,
	"blog/template/partials.html": blogTemplatePartialsHtml,
	"blog/template/posts/docs.html": blogTemplatePostsDocsHtml,
	"blog/template/posts/layout.html": blogTemplatePostsLayoutHtml,
	"blog/template/posts/tag.html": blogTemplatePostsTagHtml,
	"blog/template/posts/tags.html": blogTemplatePostsTagsHtml,
	"minimal/template/index.md": minimalTemplateIndexMd,
	"minimal/template/layout.html": minimalTemplateLayoutHtml,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}
var _bintree = &bintree{nil, map[string]*bintree{
	"blog": &bintree{nil, map[string]*bintree{
		"template": &bintree{nil, map[string]*bintree{
			"config.toml": &bintree{blogTemplateConfigToml, map[string]*bintree{}},
			"favicon.ico": &bintree{blogTemplateFaviconIco, map[string]*bintree{}},
			"layout.html": &bintree{blogTemplateLayoutHtml, map[string]*bintree{}},
			"main.css": &bintree{blogTemplateMainCss, map[string]*bintree{}},
			"newsletters": &bintree{nil, map[string]*bintree{
				"config.toml": &bintree{blogTemplateNewslettersConfigToml, map[string]*bintree{}},
				"docs.html": &bintree{blogTemplateNewslettersDocsHtml, map[string]*bintree{}},
				"layout.html": &bintree{blogTemplateNewslettersLayoutHtml, map[string]*bintree{}},
			}},
			"partials.html": &bintree{blogTemplatePartialsHtml, map[string]*bintree{}},
			"posts": &bintree{nil, map[string]*bintree{
				"docs.html": &bintree{blogTemplatePostsDocsHtml, map[string]*bintree{}},
				"layout.html": &bintree{blogTemplatePostsLayoutHtml, map[string]*bintree{}},
				"tag.html": &bintree{blogTemplatePostsTagHtml, map[string]*bintree{}},
				"tags.html": &bintree{blogTemplatePostsTagsHtml, map[string]*bintree{}},
			}},
		}},
	}},
	"minimal": &bintree{nil, map[string]*bintree{
		"template": &bintree{nil, map[string]*bintree{
			"index.md": &bintree{minimalTemplateIndexMd, map[string]*bintree{}},
			"layout.html": &bintree{minimalTemplateLayoutHtml, map[string]*bintree{}},
		}},
	}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}

