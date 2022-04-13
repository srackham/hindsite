package fsutil

import (
	"reflect"
	"testing"
	"time"
)

func TestDirExists(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DirExists(tt.args.name); got != tt.want {
				t.Errorf("DirExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFileExists(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FileExists(tt.args.name); got != tt.want {
				t.Errorf("FileExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReadFile(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadFile(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ReadFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWriteFile(t *testing.T) {
	type args struct {
		name string
		text string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := WriteFile(tt.args.name, tt.args.text); (err != nil) != tt.wantErr {
				t.Errorf("WriteFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWritePath(t *testing.T) {
	type args struct {
		path string
		text string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := WritePath(tt.args.path, tt.args.text); (err != nil) != tt.wantErr {
				t.Errorf("WritePath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFileName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FileName(tt.args.name); got != tt.want {
				t.Errorf("FileName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReplaceExt(t *testing.T) {
	type args struct {
		name string
		ext  string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ReplaceExt(tt.args.name, tt.args.ext); got != tt.want {
				t.Errorf("ReplaceExt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCopyFile(t *testing.T) {
	type args struct {
		from string
		to   string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CopyFile(tt.args.from, tt.args.to); (err != nil) != tt.wantErr {
				t.Errorf("CopyFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMkMissingDir(t *testing.T) {
	type args struct {
		dir string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := MkMissingDir(tt.args.dir); (err != nil) != tt.wantErr {
				t.Errorf("MkMissingDir() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPathIsInDir(t *testing.T) {
	type args struct {
		p   string
		dir string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PathIsInDir(tt.args.p, tt.args.dir); got != tt.want {
				t.Errorf("PathIsInDir() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPathTranslate(t *testing.T) {
	type args struct {
		srcPath string
		srcRoot string
		dstRoot string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PathTranslate(tt.args.srcPath, tt.args.srcRoot, tt.args.dstRoot); got != tt.want {
				t.Errorf("PathTranslate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFileModTime(t *testing.T) {
	type args struct {
		f string
	}
	tests := []struct {
		name string
		args args
		want time.Time
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FileModTime(tt.args.f); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FileModTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDirCount(t *testing.T) {
	type args struct {
		dir string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DirCount(tt.args.dir); got != tt.want {
				t.Errorf("DirCount() = %v, want %v", got, tt.want)
			}
		})
	}
}
