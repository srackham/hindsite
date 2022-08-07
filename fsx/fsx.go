package fsx

/*
Extra file system related functions.
*/

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

/*
File functions.
*/
func DirExists(name string) bool {
	info, err := os.Stat(name)
	return err == nil && info.IsDir()
}

func FileExists(name string) bool {
	info, err := os.Stat(name)
	return err == nil && !info.IsDir()
}

func ReadFile(name string) (string, error) {
	bytes, err := os.ReadFile(name)
	return string(bytes), err
}

func WriteFile(name string, text string) error {
	return os.WriteFile(name, []byte(text), 0644)
}

// WritePath writes file and creates any missing path directories.
func WritePath(path string, text string) error {
	if err := MkMissingDir(filepath.Dir(path)); err != nil {
		return err
	}
	return WriteFile(path, text)
}

// Return file name sans extension.
func FileName(name string) string {
	return ReplaceExt(filepath.Base(name), "")
}

// Replace the extension of name.
func ReplaceExt(name, ext string) string {
	return name[0:len(name)-len(filepath.Ext(name))] + ext
}

func CopyFile(from, to string) error {
	contents, err := ReadFile(from)
	if err != nil {
		return err
	}
	err = WriteFile(to, contents)
	return err
}

func MkMissingDir(dir string) error {
	if !DirExists(dir) {
		if err := os.MkdirAll(dir, 0775); err != nil {
			return err
		}
	}
	return nil
}

// PathIsInDir returns true if path p is in directory dir or if p equals dir.
func PathIsInDir(p, dir string) bool {
	p = filepath.Clean(p)
	dir = filepath.Clean(dir)
	return p == dir || strings.HasPrefix(p, dir+string(filepath.Separator))
}

// Translate srcPath to corresponding path in dstRoot.
func PathTranslate(srcPath, srcRoot, dstRoot string) string {
	if !PathIsInDir(srcPath, srcRoot) {
		panic("srcPath not in srcRoot: " + srcPath)
	}
	dstPath, err := filepath.Rel(srcRoot, srcPath)
	if err != nil {
		panic(err.Error())
	}
	return filepath.Join(dstRoot, dstPath)
}

// FileModTime returns file f's modification time or zero time if it can't.
func FileModTime(f string) time.Time {
	info, err := os.Stat(f)
	if err != nil {
		return time.Time{}
	}
	return info.ModTime()
}

// DirCount returns the number of files and folders in a directory. Returns zero if directory does not exist.
func DirCount(dir string) int {
	if !DirExists(dir) {
		return 0
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	return len(entries)
}
