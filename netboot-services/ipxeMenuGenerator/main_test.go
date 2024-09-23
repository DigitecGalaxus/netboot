package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetMostRecentSquashfsImageFolder(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	folders := []string{"24-08-27-master-a46edbc", "24-08-28-master-a46edbc", "24-08-29-master-a46edbc"}
	for _, folder := range folders {
		folderPath := filepath.Join(tempDir, folder)
		require.NoError(t, os.Mkdir(folderPath, 0755))
		squashfsFile := filepath.Join(folderPath, "image.squashfs")
		require.NoError(t, os.WriteFile(squashfsFile, []byte("blub"), 0644))
	}

	// Act
	result, err := getMostRecentSquashfsImageFolder(tempDir)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "24-08-29-master-a46edbc", result)
}

func TestGetSquashfsFileName(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	tests := []struct {
		name           string
		files          []string
		expectedResult string
	}{
		{
			name:           "Normal case",
			files:          []string{"image.squashfs"},
			expectedResult: "image.squashfs",
		},
		{
			name:           "Multiple files",
			files:          []string{"image.squashfs", "other.txt"},
			expectedResult: "image.squashfs",
		},
		{
			name:           "No squashfs file",
			files:          []string{"other.txt"},
			expectedResult: "",
		},
		{
			name:           "With .azDownload file",
			files:          []string{".azDownload-image.squashfs", "image.squashfs"},
			expectedResult: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Arrange
			folderPath := filepath.Join(tempDir, test.name)
			require.NoError(t, os.Mkdir(folderPath, 0755))
			for _, file := range test.files {
				require.NoError(t, os.WriteFile(filepath.Join(folderPath, file), []byte("blub"), 0644))
			}

			// Act
			result := getSquashfsFileName(tempDir, test.name)

			// Assert
			assert.Equal(t, test.expectedResult, result)
		})
	}
}

func TestGetImages(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	folders := []string{"24-08-27-master-a46edbc", "24-08-28-master-a46edbc", "24-08-29-master-a46edbc", "azDownloadFolder"}
	for _, folder := range folders {
		folderPath := filepath.Join(tempDir, folder)
		require.NoError(t, os.Mkdir(folderPath, 0755))
		if folder == "azDownloadFolder" {
			require.NoError(t, os.WriteFile(filepath.Join(folderPath, ".azDownload-image.squashfs"), []byte("blub"), 0644))
		} else {
			squashfsFile := filepath.Join(folderPath, "image.squashfs")
			require.NoError(t, os.WriteFile(squashfsFile, []byte("blub"), 0644))
		}
	}

	// Act
	images, err := getImages(tempDir)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, images, 3)
	assert.Equal(t, "24-08-29-master-a46edbc", images[0].SquashfsFoldername)
	assert.Equal(t, "24-08-28-master-a46edbc", images[1].SquashfsFoldername)
	assert.Equal(t, "24-08-27-master-a46edbc", images[2].SquashfsFoldername)

	// Assert that azDownloadFolder is not included
	for _, image := range images {
		assert.NotEqual(t, "azDownloadFolder", image.SquashfsFoldername)
	}
}

func TestRenderMenuIpxe(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	menusDir := filepath.Join(tempDir, "menus")
	require.NoError(t, os.Mkdir(menusDir, 0755))
	templateContent := `netbootServerIP: {{ netbootServerIP }}`
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "menu.ipxe.j2"), []byte(templateContent), 0644))
	renderData := RenderMenuData{
		JinjaTemplateFile: "menu.ipxe.j2",
		NetbootServerIP:   "192.168.1.1",
		MenusDirectory:    menusDir,
		WorkingDirectory:  tempDir,
	}
	squashfsImage := SquashfsPaths{
		SquashfsFilename:   "image.squashfs",
		SquashfsFoldername: "folder1",
	}

	// Act
	err := renderMenuIpxe(renderData, squashfsImage)

	// Assert
	assert.NoError(t, err)
	renderedContent, err := os.ReadFile(filepath.Join(menusDir, "menu.ipxe"))
	assert.NoError(t, err)
	assert.Contains(t, string(renderedContent), "netbootServerIP: 192.168.1.1")
}
func TestRenderAdvancedMenu(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	menusDir := filepath.Join(tempDir, "menus")
	require.NoError(t, os.Mkdir(menusDir, 0755))

	// Copy the actual advancedmenu.ipxe.j2 file to the temp directory
	sourceFile := "advancedmenu.ipxe.j2"
	destFile := filepath.Join(tempDir, "advancedmenu.ipxe.j2")
	content, err := os.ReadFile(sourceFile)
	require.NoError(t, err)
	err = os.WriteFile(destFile, content, 0644)
	require.NoError(t, err)

	renderData := RenderMenuData{
		JinjaTemplateFile:          "advancedmenu.ipxe.j2",
		NetbootServerIP:            "192.168.1.1",
		AzureNetbootServerIP:       "10.0.0.1",
		OnpremExposedNetbootServer: "netboot.example.com",
		AzureBlobstorageURL:        "https://example.blob.core.windows.net",
		AzureBlobstorageSASToken:   "?sastoken",
		HTTPAuthUser:               "user",
		HTTPAuthPassword:           "pass",
		MenusDirectory:             menusDir,
		WorkingDirectory:           tempDir,
	}

	prodImages := []SquashfsPaths{
		{SquashfsFilename: "prod1.squashfs", SquashfsFoldername: "24-08-01-master-abcdef"},
		{SquashfsFilename: "prod2.squashfs", SquashfsFoldername: "24-07-31-master-123456"},
	}
	devImages := []SquashfsPaths{
		{SquashfsFilename: "dev1.squashfs", SquashfsFoldername: "24-08-02-feature-ghijkl"},
		{SquashfsFilename: "dev2.squashfs", SquashfsFoldername: "24-08-01-bugfix-789012"},
	}

	// Act
	err = renderAdvancedMenu(renderData, prodImages, devImages)

	// Assert
	assert.NoError(t, err)
	renderedContent, err := os.ReadFile(filepath.Join(menusDir, "advancedmenu.ipxe"))
	assert.NoError(t, err)

	// Check for specific content in the rendered output
	renderedString := string(renderedContent)
	assert.Contains(t, renderedString, "item --gap Production:")
	assert.Contains(t, renderedString, "item --gap Development:")
	assert.Contains(t, renderedString, "item thinclient-24-08-01-master-abcdef ${sp} 24-08-01-master-abcdef")
	assert.Contains(t, renderedString, "item thinclient-24-07-31-master-123456 ${sp} 24-07-31-master-123456")
	assert.Contains(t, renderedString, "item thinclient-24-08-02-feature-ghijkl ${sp} 24-08-02-feature-ghijkl")
	assert.Contains(t, renderedString, "item thinclient-24-08-01-bugfix-789012 ${sp} 24-08-01-bugfix-789012")
	assert.Contains(t, renderedString, "chain tftp://192.168.1.1/ipxe/netinfo.ipxe")
	assert.Contains(t, renderedString, "set squash_url ${http-protocol}://${basicAuth}${url}/prod/24-08-01-master-abcdef/prod1.squashfs${sas_token}")
	assert.Contains(t, renderedString, "set squash_url ${http-protocol}://${basicAuth}${url}/dev/24-08-02-feature-ghijkl/dev1.squashfs${sas_token}")
}

func TestRenderNetinfoMenu(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	menusDir := filepath.Join(tempDir, "menus")
	require.NoError(t, os.Mkdir(menusDir, 0755))
	templateContent := `Netinfo menu`
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "netinfo.ipxe.j2"), []byte(templateContent), 0644))
	renderData := RenderMenuData{
		JinjaTemplateFile: "netinfo.ipxe.j2",
		MenusDirectory:    menusDir,
		WorkingDirectory:  tempDir,
	}

	// Act
	err := renderNetinfoMenu(renderData)

	// Assert
	assert.NoError(t, err)
	renderedContent, err := os.ReadFile(filepath.Join(menusDir, "netinfo.ipxe"))
	assert.NoError(t, err)
	assert.Equal(t, "Netinfo menu", string(renderedContent))
}

func TestByModTime(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	files := []string{"image1.squashfs", "image2.squashfs", "image3.squashfs"}
	var fileEntries []fs.DirEntry
	for i, file := range files {
		filePath := filepath.Join(tempDir, file)
		require.NoError(t, os.WriteFile(filePath, []byte("blub"), 0644))
		entry, err := os.ReadDir(tempDir)
		require.NoError(t, err)
		fileEntries = append(fileEntries, entry[i])
	}

	// Act
	sorted := ByModTime(fileEntries)
	sort.Sort(sorted)

	// Assert and check the order. Last modified file should be first
	assert.Equal(t, "image3.squashfs", sorted[0].Name())
	assert.Equal(t, "image2.squashfs", sorted[1].Name())
	assert.Equal(t, "image1.squashfs", sorted[2].Name())
}
