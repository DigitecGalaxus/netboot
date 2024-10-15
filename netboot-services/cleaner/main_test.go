package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetImagesSortedByModifiedDate(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	createTestImageFolder(t, tempDir, "image1", 1)
	createTestImageFolder(t, tempDir, "image2", 2)
	createTestImageFolder(t, tempDir, "image3", 3)

	// Act
	images := getImagesSortedByModifiedDate(tempDir)

	// Assert
	assert.Len(t, images, 3)
	assert.Equal(t, "image3", images[0].Name())
	assert.Equal(t, "image2", images[1].Name())
	assert.Equal(t, "image1", images[2].Name())
}

func TestGetCurrentFolderSizeInGiB(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	createTestImageFolder(t, tempDir, "image1", 1)
	createTestImageFolder(t, tempDir, "image2", 2)

	// Act
	size := getCurrentFolderSizeInGiB(tempDir)

	// Assert
	assert.Greater(t, size, 0.0)
	assert.Less(t, size, 1.0) // Assuming test files are small
}

func TestFolderNeedsCleanup(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	properties := folderProperties{
		FolderPath:              tempDir,
		ThresholdMaxImagesCount: 2,
		MaxFolderSizeInGiB:      0.1,
	}
	createTestImageFolder(t, tempDir, "image1", 1)
	createTestImageFolder(t, tempDir, "image2", 2)
	createTestImageFolder(t, tempDir, "image3", 3)

	// Act & Assert
	images := getImagesSortedByModifiedDate(tempDir)
	folderSize := getCurrentFolderSizeInGiB(tempDir)
	assert.True(t, folderNeedsCleanup(properties, folderSize, images))

	err := os.RemoveAll(filepath.Join(tempDir, "image3"))
	require.NoError(t, err)

	images = getImagesSortedByModifiedDate(tempDir)
	folderSize = getCurrentFolderSizeInGiB(tempDir)
	assert.False(t, folderNeedsCleanup(properties, folderSize, images))
}

func TestDeleteImage(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	imageName := "image1"
	createTestImageFolder(t, tempDir, imageName, 1)

	// Act
	images := getImagesSortedByModifiedDate(tempDir)
	require.Len(t, images, 1)
	err := deleteImage(tempDir, images[0])

	// Assert
	assert.NoError(t, err)
	_, err = os.Stat(filepath.Join(tempDir, imageName))
	assert.True(t, os.IsNotExist(err))
}

// Helper function to create test image folders
func createTestImageFolder(t *testing.T, baseDir, folderName string, timeInHours int) {
	folderPath := filepath.Join(baseDir, folderName)
	require.NoError(t, os.MkdirAll(folderPath, 0755))

	imagePath := filepath.Join(folderPath, "image.squashfs")
	require.NoError(t, os.WriteFile(imagePath, []byte("blub"), 0644))

	modTime := time.Now().Add(time.Duration(timeInHours) * time.Hour)
	require.NoError(t, os.Chtimes(folderPath, modTime, modTime))

}
