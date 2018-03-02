package main

import (
	"path"
)

// ConfigParams defines global configuration parameters.
type ConfigParams struct {
	projectDir  string
	contentDir  string
	templateDir string
	buildDir    string
}

// Config contains global configuration parameters.
var Config ConfigParams

// InitDirs TODO
func (conf *ConfigParams) InitDirs(projectDir, contentDir, templateDir, buildDir string) {
	conf.projectDir = path.Clean(projectDir)
	conf.contentDir = path.Join(projectDir, "content")
	conf.templateDir = path.Join(projectDir, "template")
	conf.buildDir = path.Join(projectDir, "build")
}
