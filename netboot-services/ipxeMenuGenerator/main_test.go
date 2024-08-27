package main

import (
	"fmt"
	"io/fs"
	"os"
	"sort"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"gotest.tools/assert"
)

var (
	err                 error
	TestWorkingDir      string
	ProdFile1           string
	ProdFile2           string
	DevFile1            string
	DevFile2            string
	ExampleDataToRender RenderMenuData
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func teardown() {
	cleanupTestFiles([]string{ProdFile1, ProdFile2, DevFile1, DevFile2})
	cleanupTestDirectories(ProdFolder, DevFolder)
}

func setupTestFiles(filename string, fileContent []byte) {
	os.WriteFile(filename, fileContent, 0644)
}

func setupTestDirectories(dirs ...string) {
	for _, dir := range dirs {
		os.MkdirAll(dir, 0755)
	}
}

func cleanupTestFiles(fileNames []string) {
	for _, filename := range fileNames {
		os.Remove(filename)
	}
}

func cleanupTestDirectories(dirs ...string) {
	for _, dir := range dirs {
		os.RemoveAll(dir)
	}
}

func setup() {
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.JSONFormatter{})

	//configure correct paths for testing
	TestWorkingDir, err = os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current working directory: %v", err)
	}

	DevFolder = "debug/dev"
	ProdFolder = "debug/prod"

	setupTestDirectories(
		WorkingDirectory,
		MenusDirectory,
		fmt.Sprintf("%s/%s", ProdFolder, "24-08-27-master-a46edbc"),
		fmt.Sprintf("%s/%s", ProdFolder, "24-08-23-master-a46edbc"),
		fmt.Sprintf("%s/%s", DevFolder, "24-08-07-noissue-fixtestpipeline-b66c813"),
		fmt.Sprintf("%s/%s", DevFolder, "24-08-08-noissue-needmoreram-5f1a800"),
	)

	// Mocking TestSquashFS Images and imitating .azDownload- prefix for currently syncing images. The content is not important.
	ProdFile1 = fmt.Sprintf("%s/%s/%s", ProdFolder, "24-08-27-master-a46edbc", ".azDownload-dg-thinclient.squashfs")
	ProdFile2 = fmt.Sprintf("%s/%s/%s", ProdFolder, "24-08-23-master-a46edbc", "dg-thinclient.squashfs")
	DevFile1 = fmt.Sprintf("%s/%s/%s", DevFolder, "24-08-07-noissue-fixtestpipeline-b66c813", ".azDownload-thinclient.squashfs")
	DevFile2 = fmt.Sprintf("%s/%s/%s", DevFolder, "24-08-08-noissue-needmoreram-5f1a800", "dg-thinclient.squashfs")

	setupTestFiles(ProdFile1, []byte("dg-thinclient.squashfs"))
	setupTestFiles(ProdFile2, []byte("dg-thinclient.squashfs"))
	setupTestFiles(DevFile1, []byte("dg-thinclient.squashfs"))
	setupTestFiles(DevFile2, []byte("dg-thinclient.squashfs"))

	// Mocking TestRenderMenuData
	ExampleDataToRender = RenderMenuData{
		JinjaTemplateFile:          "overwriteme.ipxe.j2",
		NetbootServerIP:            "192.168.1.1",
		AzureNetbootServerIP:       "10.10.10.1",
		OnpremExposedNetbootServer: "onpremise.blub.com",
		AzureBlobstorageURL:        "blub.blob.core.windows.net",
		AzureBlobstorageSASToken:   "?SASStorageAccountToken",
		HTTPAuthUser:               "1234",
		HTTPAuthPassword:           "1234",
		MenusDirectory:             TestWorkingDir,
		WorkingDirectory:           TestWorkingDir,
	}
}

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

func TestRenderMenu(t *testing.T) {
	mockFileNameTemplate := "menu_test.ipxe.j2"
	mockFileNameGenerated := "menu_test.ipxe"
	mockFileContentTemplate := `chain --autofree tftp://{{ netbootServerIP }}/ipxe/MAC-${mac:hexraw}.ipxe || echo Custom boot by MAC not found, going to menu..."
iseq ${next-server} {{ azureNetbootServerIP }} && goto register_basic_auth ||
set httpAuthUser {{ httpAuthUser }} && set httpAuthPassword {{ httpAuthPassword }} && goto check_if_onprem_exposed_netboot_is_reachable ||
imgfetch https://${httpAuthUser}:${httpAuthPassword}@{{ onpremExposedNetbootServer }}:8443/healthcheck.json && goto set_onprem_exposed_netboot ||
imgfetch https://{{ azureBlobstorageURL }}/healthcheck/healthcheck.json{{ azureBlobstorageSASToken }} &&  goto set_azure_storageaccount ||
# Chaining the advanced menu.
:advanced
chain --autofree tftp://{{ netbootServerIP }}/ipxe/advancedmenu.ipxe`

	mockFileContentGenerated := `chain --autofree tftp://192.168.1.1/ipxe/MAC-${mac:hexraw}.ipxe || echo Custom boot by MAC not found, going to menu..."
iseq ${next-server} 10.10.10.1 && goto register_basic_auth ||
set httpAuthUser 1234 && set httpAuthPassword 1234 && goto check_if_onprem_exposed_netboot_is_reachable ||
imgfetch https://${httpAuthUser}:${httpAuthPassword}@onpremise.blub.com:8443/healthcheck.json && goto set_onprem_exposed_netboot ||
imgfetch https://blub.blob.core.windows.net/healthcheck/healthcheck.json?SASStorageAccountToken &&  goto set_azure_storageaccount ||
# Chaining the advanced menu.
:advanced
chain --autofree tftp://192.168.1.1/ipxe/advancedmenu.ipxe`

	setupTestFiles(mockFileNameTemplate, []byte(mockFileContentTemplate))
	defer cleanupTestFiles([]string{mockFileNameTemplate, mockFileNameGenerated})

	mostRecentSquashfsFoldername, err := getMostRecentSquashfsImageFolder(ProdFolder)
	if err != nil {
		log.Fatal(err)
	}

	mostRecentSquashfsImage := SquashfsPaths{
		SquashfsFilename:   getSquashfsFileName(ProdFolder, mostRecentSquashfsFoldername),
		SquashfsFoldername: mostRecentSquashfsFoldername,
	}

	ExampleDataToRender.JinjaTemplateFile = mockFileNameTemplate

	err = renderMenuIpxe(ExampleDataToRender, mostRecentSquashfsImage)
	assert.NilError(t, err)

	fileContentGenerated, err := os.ReadFile(mockFileNameGenerated)
	assert.Equal(t, string(fileContentGenerated), mockFileContentGenerated)
	assert.NilError(t, err)
}

func TestRenderAdvancedMenu(t *testing.T) {
	mockFileNameTemplate := "advancedmenu_test.ipxe.j2"
	mockFileNameGenerated := "advancedmenu_test.ipxe"
	mockFileContentTemplate := `
item --gap Production:
{%- for img in prod %}
item thinclient-{{ img.squashfsFoldername }} ${sp} {{ img.squashfsFoldername }}
{%- endfor %}
item --gap Development:
{%- for img in dev %}
item thinclient-{{ img.squashfsFoldername }} ${sp} {{ img.squashfsFoldername }}
{%- endfor %}

:netinfo
chain tftp://{{ netbootServerIP }}/ipxe/netinfo.ipxe
goto advanced_menu

#####################
# Production-Images #
#####################

{% for img in prod %}
:thinclient-{{ img.squashfsFoldername }}
set squash_url ${http-protocol}://${basicAuth}${url}/prod/{{ img.squashfsFoldername }}/{{ img.squashfsFilename }}${sas_token}
set kernel_url ${http-protocol}://${basicAuth}${url}/prod/{{ img.squashfsFoldername }}/
goto startboot
{% endfor %}

######################
# Development-Images #
######################

{% for img in dev %}
:thinclient-{{ img.squashfsFoldername }}
set squash_url ${http-protocol}://${basicAuth}${url}/dev/{{ img.squashfsFoldername }}/{{ img.squashfsFilename }}${sas_token}
set kernel_url ${http-protocol}://${basicAuth}${url}/dev/{{ img.squashfsFoldername }}/
goto startboot-dev
{% endfor %}`

	mockFileContentGenerated := `
item --gap Production:
item thinclient-24-08-23-master-a46edbc ${sp} 24-08-23-master-a46edbc
item --gap Development:
item thinclient-24-08-08-noissue-needmoreram-5f1a800 ${sp} 24-08-08-noissue-needmoreram-5f1a800

:netinfo
chain tftp://192.168.1.1/ipxe/netinfo.ipxe
goto advanced_menu

#####################
# Production-Images #
#####################


:thinclient-24-08-23-master-a46edbc
set squash_url ${http-protocol}://${basicAuth}${url}/prod/24-08-23-master-a46edbc/dg-thinclient.squashfs${sas_token}
set kernel_url ${http-protocol}://${basicAuth}${url}/prod/24-08-23-master-a46edbc/
goto startboot


######################
# Development-Images #
######################


:thinclient-24-08-08-noissue-needmoreram-5f1a800
set squash_url ${http-protocol}://${basicAuth}${url}/dev/24-08-08-noissue-needmoreram-5f1a800/dg-thinclient.squashfs${sas_token}
set kernel_url ${http-protocol}://${basicAuth}${url}/dev/24-08-08-noissue-needmoreram-5f1a800/
goto startboot-dev
`

	setupTestFiles(mockFileNameTemplate, []byte(mockFileContentTemplate))
	defer cleanupTestFiles([]string{mockFileNameTemplate, mockFileNameGenerated})

	prodImages, err := getImages(ProdFolder)
	if err != nil {
		log.Error(err)
	}

	devImages, err := getImages(DevFolder)
	if err != nil {
		log.Error(err)
	}

	ExampleDataToRender.JinjaTemplateFile = mockFileNameTemplate

	err = renderAdvancedMenu(ExampleDataToRender, prodImages, devImages)
	assert.NilError(t, err)

	fileContentGenerated, err := os.ReadFile(mockFileNameGenerated)
	assert.Equal(t, string(fileContentGenerated), mockFileContentGenerated)
	assert.NilError(t, err)
}

// For sake of code coverage, we are testing renderNetinfoMenu() function. It basically just rewrites the file with the same content as we do not pass custom data to it.
func TestRenderNetinfoMenu(t *testing.T) {
	mockFileNameTemplate := "netinfo_test.ipxe.j2"
	mockFileNameGenerated := "netinfo_test.ipxe"
	mockFileContentTemplate := `#!ipxe
menu Network info
item --gap MAC:
item mac ${sp} ${netX/mac}
item --gap IP/mask:
item ip ${sp} ${netX/ip}/${netX/netmask}
item --gap Gateway:
item gw ${sp} ${netX/gateway}
item --gap Domain:
item domain ${sp} ${netX/domain}
item --gap DNS:
item dns ${sp} ${netX/dns}
item --gap DHCP server:
item dhcpserver ${sp} ${netX/dhcp-server}
item --gap Next-server:
item nextserver ${sp} ${next-server}
item --gap Filename:
item filename ${sp} ${netX/filename}
choose empty ||
exit`

	mockFileContentGenerated := `#!ipxe
menu Network info
item --gap MAC:
item mac ${sp} ${netX/mac}
item --gap IP/mask:
item ip ${sp} ${netX/ip}/${netX/netmask}
item --gap Gateway:
item gw ${sp} ${netX/gateway}
item --gap Domain:
item domain ${sp} ${netX/domain}
item --gap DNS:
item dns ${sp} ${netX/dns}
item --gap DHCP server:
item dhcpserver ${sp} ${netX/dhcp-server}
item --gap Next-server:
item nextserver ${sp} ${next-server}
item --gap Filename:
item filename ${sp} ${netX/filename}
choose empty ||
exit`

	setupTestFiles(mockFileNameTemplate, []byte(mockFileContentTemplate))
	defer cleanupTestFiles([]string{mockFileNameTemplate, mockFileNameGenerated})

	ExampleDataToRender.JinjaTemplateFile = mockFileNameTemplate

	err := renderNetinfoMenu(ExampleDataToRender)
	assert.NilError(t, err)

	fileContentGenerated, err := os.ReadFile(mockFileNameGenerated)
	assert.Equal(t, string(fileContentGenerated), mockFileContentGenerated)
	assert.NilError(t, err)

}

// Test getMostRecentSquashfsImage() --> Tested with TestRenderMenu

// Test getSquashfsFileName() --> Tested with TestRenderMenu

// test getImages() --> Tested with TestRenderAdvancedMenu
