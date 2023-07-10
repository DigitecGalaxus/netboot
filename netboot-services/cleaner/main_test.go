package main

import (
	"io/fs"
	"sort"
	"testing"
	"time"

	"gotest.tools/assert"
)

func TestSortByModificationDate(t *testing.T) {

	allFiles := []fs.DirEntry{
		//in random order
		&mockDirEntry{name: "file2.txt", isDir: false, fileInfo: &mockFileInfo{name: "file2.txt", modTime: time.Now().Add(-1 * time.Hour)}},
		&mockDirEntry{name: "file1.txt", isDir: false, fileInfo: &mockFileInfo{name: "file1.txt", modTime: time.Now()}},
		&mockDirEntry{name: "file5.txt", isDir: false, fileInfo: &mockFileInfo{name: "file5.txt", modTime: time.Now().Add(-4 * time.Hour)}},
		&mockDirEntry{name: "file4.txt", isDir: false, fileInfo: &mockFileInfo{name: "file4.txt", modTime: time.Now().Add(-3 * time.Hour)}},
		&mockDirEntry{name: "file3.txt", isDir: false, fileInfo: &mockFileInfo{name: "file3.txt", modTime: time.Now().Add(-2 * time.Hour)}},
	}

	sort.Sort(ByModTime(allFiles))

	assert.Equal(t, allFiles[0].Name(), "file1.txt")
	assert.Equal(t, allFiles[1].Name(), "file2.txt")
	assert.Equal(t, allFiles[2].Name(), "file3.txt")
	assert.Equal(t, allFiles[3].Name(), "file4.txt")
	assert.Equal(t, allFiles[4].Name(), "file5.txt")
}

type mockDirEntry struct {
	name     string
	isDir    bool
	fileInfo fs.FileInfo
}

func (m *mockDirEntry) Name() string {
	return m.name
}

func (m *mockDirEntry) IsDir() bool {
	return m.isDir
}

func (m *mockDirEntry) Type() fs.FileMode {
	// Implement this method if necessary
	// Return a default or desired file mode
	return fs.ModeDir
}

func (m *mockDirEntry) Info() (fs.FileInfo, error) {
	return m.fileInfo, nil
}

type mockFileInfo struct {
	name    string
	size    int64
	mode    fs.FileMode
	modTime time.Time
	isDir   bool
}

func (m *mockFileInfo) Name() string {
	return m.name
}

func (m *mockFileInfo) Size() int64 {
	return m.size
}

func (m *mockFileInfo) Mode() fs.FileMode {
	return m.mode
}

func (m *mockFileInfo) ModTime() time.Time {
	return m.modTime
}

func (m *mockFileInfo) IsDir() bool {
	return m.isDir
}

func (m *mockFileInfo) Sys() interface{} {
	return nil
}
