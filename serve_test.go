package main

import (
	"bytes"
	"os"
	"path"
	"strings"
	"testing"
)

func Test_serve(t *testing.T) {
	t.Run("serve", func(t *testing.T) {
		tmpdir := path.Join(os.TempDir(), "hindsite-tests")
		os.RemoveAll(tmpdir)
		mkMissingDir(tmpdir)
		proj := newProject()
		var outbuf bytes.Buffer
		proj.outlog = &outbuf
		var errbuf bytes.Buffer
		proj.errlog = &errbuf
		args := strings.Split("hindsite serve "+tmpdir, " ")
		if err := proj.parseArgs(args); err != nil {
			t.Errorf("serve command error: %v", err)
		}
		if err := proj.serve(); err != nil {
			t.Errorf("serve command error: %v", err)
		}
		errbuf.Reset()
		outbuf.Reset()
	})
}
