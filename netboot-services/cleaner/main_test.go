package main

import (
	"io/fs"
	"os"
	"sort"
	"testing"
	"time"

	"gotest.tools/v3/assert"
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

func setupTestFiles(filename string, fileContent []byte) {
	os.WriteFile(filename, fileContent, 0644)
}

func setupTestDirectories(dirs ...string) {
	for _, dir := range dirs {
		os.MkdirAll(dir, 0755)
	}
}

func cleanupTestDirectories(dirs ...string) {
	for _, dir := range dirs {
		os.RemoveAll(dir)
	}
}

func TestDeleteImage(t *testing.T) {
	setupTestDirectories("test", "test/img1", "test/img2", "test/img3")
	setupTestFiles("test/img1/img1.squashfs", []byte("test1"))
	setupTestFiles("test/img2/img2.squashfs", []byte("test2"))
	setupTestFiles("test/img3/.azDownload-img3.squashfs", []byte("test3"))

	propertiesTest := folderProperties{
		FolderPath:              "test",
		ThresholdMaxImagesCount: 1,
		MaxFolderSizeInGiB:      0.0000000000001, // small value due to small testsize
	}

	// should only get two images as one is .azDownload prefixed
	testImages := getImagesSortedByModifiedDate(propertiesTest.FolderPath)
	assert.Equal(t, 2, len(testImages))

	folderSizeInGiB := getCurrentFolderSizeInGiB(propertiesTest.FolderPath)
	for i := len(testImages) - 1; folderNeedsCleanup(propertiesTest, folderSizeInGiB, testImages); i-- {
		err := deleteImage(propertiesTest.FolderPath, testImages[i])
		assert.NilError(t, err)

		testImages = getImagesSortedByModifiedDate(propertiesTest.FolderPath)
		folderSizeInGiB = getCurrentFolderSizeInGiB(propertiesTest.FolderPath)
	}

	// Test with a folder that has exceeded the max folder size
	setupTestFiles("test/img1/img1.squashfs", []byte("test1"))
	setupTestFiles("test/img2/img2.squashfs", []byte("test2"))
	testImages = getImagesSortedByModifiedDate(propertiesTest.FolderPath)
	assert.Equal(t, 1, len(testImages))
	for i := len(testImages) - 1; folderNeedsCleanup(propertiesTest, folderSizeInGiB, testImages); i-- {
		err := deleteImage(propertiesTest.FolderPath, testImages[i])
		assert.NilError(t, err)

		testImages = getImagesSortedByModifiedDate(propertiesTest.FolderPath)
		folderSizeInGiB = getCurrentFolderSizeInGiB(propertiesTest.FolderPath)
	}

	cleanupTestDirectories("test")
}
