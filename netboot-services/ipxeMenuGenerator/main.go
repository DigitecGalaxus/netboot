package main

import (
	"fmt"
	"io/fs"
	"os"
	"sort"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/kluctl/go-jinja2"
)

var (
	WorkingDirectory = "/work"
	MenusDirectory   = "/menus"
	DevFolder        = "/assets/dev"
	ProdFolder       = "/assets/prod"
)

type SquashfsPaths struct {
	SquashfsFilename   string `json:"squashfsFilename"`
	SquashfsFoldername string `json:"squashfsFoldername"`
}

type RenderMenuData struct {
	JinjaTemplateFile          string
	NetbootServerIP            string
	AzureNetbootServerIP       string
	OnpremExposedNetbootServer string
	AzureBlobstorageURL        string
	AzureBlobstorageSASToken   string
	HTTPAuthUser               string
	HTTPAuthPassword           string
	MenusDirectory             string
	WorkingDirectory           string
}

func main() {
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.JSONFormatter{})

	for {
		netbootServerIP := os.Getenv("NETBOOT_SERVER_IP")
		if netbootServerIP == "" {
			log.Fatal("NETBOOT_SERVER_IP not set")
		}
		azureNetbootServerIP := os.Getenv("AZURE_NETBOOT_SERVER_IP")
		if azureNetbootServerIP == "" {
			log.Fatal("AZURE_NETBOOT_SERVER_IP not set")
		}

		onpremExposedNetbootServer := os.Getenv("ONPREM_NETBOOT_SERVER")
		if onpremExposedNetbootServer == "" {
			log.Fatal("ONPREM_NETBOOT_SERVER not set")
		}

		azureBlobstorageSASToken := os.Getenv("AZURE_SYNC_SAS_TOKEN")
		if azureBlobstorageSASToken == "" {
			log.Fatal("AZURE_SYNC_SAS_TOKEN not set")
		}

		azureBlobstorageURL := os.Getenv("AZURE_SYNC_BLOB_URL")
		if azureBlobstorageURL == "" {
			log.Fatal("AZURE_SYNC_BLOB_URL not set")
		}

		httpAuthUser := os.Getenv("HTTP_AUTH_USER")
		if httpAuthUser == "" {
			log.Fatal("HTTP_AUTH_USER not set")
		}

		httpAuthPassword := os.Getenv("HTTP_AUTH_PASSWORD")
		if httpAuthPassword == "" {
			log.Fatal("HTTP_AUTH_PASSWORD not set")
		}

		mostRecentSquashfsFoldername, err := getMostRecentSquashfsImageFolder(ProdFolder)
		if err != nil {
			log.Fatal(err)
		}

		mostRecentSquashfsImage := SquashfsPaths{
			SquashfsFilename:   getSquashfsFileName(ProdFolder, mostRecentSquashfsFoldername),
			SquashfsFoldername: mostRecentSquashfsFoldername,
		}

		if mostRecentSquashfsImage.SquashfsFoldername == "" || mostRecentSquashfsImage.SquashfsFilename == "" {
			log.Fatalf("No recent SquashFS File or Folder found on %s", ProdFolder)
		}

		RenderMenuData := RenderMenuData{
			JinjaTemplateFile:          "menu.ipxe.j2",
			NetbootServerIP:            netbootServerIP,
			AzureNetbootServerIP:       azureNetbootServerIP,
			OnpremExposedNetbootServer: onpremExposedNetbootServer,
			AzureBlobstorageURL:        azureBlobstorageURL,
			AzureBlobstorageSASToken:   azureBlobstorageSASToken,
			HTTPAuthUser:               httpAuthUser,
			HTTPAuthPassword:           httpAuthPassword,
			MenusDirectory:             MenusDirectory,
			WorkingDirectory:           WorkingDirectory,
		}

		err = renderMenuIpxe(RenderMenuData, mostRecentSquashfsImage)
		if err != nil {
			log.Fatal(err)
		}

		RenderMenuData.JinjaTemplateFile = "advancedmenu.ipxe.j2"

		prodImages, err := getImages(ProdFolder)
		if err != nil {
			log.Error(err)
		}

		devImages, err := getImages(DevFolder)
		if err != nil {
			log.Error(err)
		}

		err = renderAdvancedMenu(RenderMenuData, prodImages, devImages)
		if err != nil {
			log.Error(err)
		}

		RenderMenuData.JinjaTemplateFile = "netinfo.ipxe.j2"

		err = renderNetinfoMenu(RenderMenuData)
		if err != nil {
			log.Error(err)
		}
		time.Sleep(60 * time.Second)
	}
}

type ByModTime []fs.DirEntry

func (b ByModTime) Len() int      { return len(b) }
func (b ByModTime) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b ByModTime) Less(i, j int) bool {
	infoI, err := b[i].Info()
	if err != nil {
		log.Fatal(err)
	}
	infoJ, err := b[j].Info()
	if err != nil {
		log.Fatal(err)
	}
	return infoI.ModTime().After(infoJ.ModTime())
}

func getMostRecentSquashfsImageFolder(folderName string) (string, error) {
	files, err := os.ReadDir(folderName)
	if err != nil {
		log.Fatal(err)
	}

	var matches []fs.DirEntry
	for _, file := range files {
		squashfsFileName := getSquashfsFileName(folderName, file.Name())
		if strings.HasSuffix(squashfsFileName, ".squashfs") {
			matches = append(matches, file)
		}
	}

	sort.Sort(ByModTime(matches))

	if len(matches) > 0 {
		return strings.TrimSuffix(matches[0].Name(), ".squashfs"), nil
	}
	return "", nil
}

func renderMenuIpxe(renderMenuData RenderMenuData, mostRecentSquashFS SquashfsPaths) error {
	j2, err := jinja2.NewJinja2("menu.ipxe", 1,
		jinja2.WithGlobal("netbootServerIP", renderMenuData.NetbootServerIP),
		jinja2.WithGlobal("azureNetbootServerIP", renderMenuData.AzureNetbootServerIP),
		jinja2.WithGlobal("onpremExposedNetbootServer", renderMenuData.OnpremExposedNetbootServer),
		jinja2.WithGlobal("azureBlobstorageSASToken", renderMenuData.AzureBlobstorageSASToken),
		jinja2.WithGlobal("azureBlobstorageURL", renderMenuData.AzureBlobstorageURL),
		jinja2.WithGlobal("httpAuthUser", renderMenuData.HTTPAuthUser),
		jinja2.WithGlobal("httpAuthPassword", renderMenuData.HTTPAuthPassword),
		jinja2.WithGlobal("imageName", mostRecentSquashFS),
	)
	if err != nil {
		return err
	}
	defer j2.Close()

	renderedString, err := j2.RenderFile(fmt.Sprintf("%s/%s", renderMenuData.WorkingDirectory, renderMenuData.JinjaTemplateFile))
	if err != nil {
		return err
	}

	filePath := fmt.Sprintf("%s/%s", renderMenuData.MenusDirectory, strings.ReplaceAll(renderMenuData.JinjaTemplateFile, ".j2", ""))
	err = os.WriteFile(filePath, []byte(renderedString), 0644)
	if err != nil {
		return err
	}

	log.Debugf("filename: %s\nresult: %s", renderMenuData.JinjaTemplateFile, renderedString)

	return nil
}

func getImages(folderName string) ([]SquashfsPaths, error) {
	folders, err := os.ReadDir(folderName)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}

	var squashfsFiles []fs.DirEntry
	for _, folder := range folders {
		if folder.Type().String() == os.ModeDir.String() {
			squashfsFilename := getSquashfsFileName(folderName, folder.Name())
			if squashfsFilename == "" {
				fmt.Println("not APPENDING folder due to active .azDownload Sync: ", folder.Name())
				continue
			}
			squashfsFiles = append(squashfsFiles, folder)
		}
	}

	sort.Sort(ByModTime(squashfsFiles))

	var squashfsPaths []SquashfsPaths
	for _, file := range squashfsFiles {
		SquashfsPath := SquashfsPaths{
			SquashfsFilename:   getSquashfsFileName(folderName, file.Name()),
			SquashfsFoldername: file.Name(),
		}
		squashfsPaths = append(squashfsPaths, SquashfsPath)
	}

	return squashfsPaths, nil
}

func getSquashfsFileName(folderName string, newImageFolderName string) string {
	newFolderToSearch := fmt.Sprintf("%s/%s", folderName, newImageFolderName)
	files, err := os.ReadDir(newFolderToSearch)
	if err != nil {
		log.Error("Error:", err)
	}

	for _, file := range files {
		if strings.HasPrefix(file.Name(), ".azDownload") {
			return ""
		}
		if strings.HasSuffix(file.Name(), ".squashfs") {
			return file.Name()
		}
	}
	return ""
}

func renderAdvancedMenu(renderMenuData RenderMenuData, prodImages []SquashfsPaths, devImages []SquashfsPaths) error {
	j2, err := jinja2.NewJinja2("advancedmenu.ipxe", 1,
		jinja2.WithGlobal("netbootServerIP", renderMenuData.NetbootServerIP),
		jinja2.WithGlobal("azureNetbootServerIP", renderMenuData.AzureNetbootServerIP),
		jinja2.WithGlobal("onpremExposedNetbootServer", renderMenuData.OnpremExposedNetbootServer),
		jinja2.WithGlobal("azureBlobstorageSASToken", renderMenuData.AzureBlobstorageSASToken),
		jinja2.WithGlobal("azureBlobstorageURL", renderMenuData.AzureBlobstorageURL),
		jinja2.WithGlobal("httpAuthUser", renderMenuData.HTTPAuthUser),
		jinja2.WithGlobal("httpAuthPassword", renderMenuData.HTTPAuthPassword),
		jinja2.WithGlobal("prod", prodImages),
		jinja2.WithGlobal("dev", devImages),
	)
	if err != nil {
		return err
	}
	defer j2.Close()

	renderedString, err := j2.RenderFile(fmt.Sprintf("%s/%s", renderMenuData.WorkingDirectory, renderMenuData.JinjaTemplateFile))
	if err != nil {
		return err
	}

	filePath := fmt.Sprintf("%s/%s", renderMenuData.MenusDirectory, strings.ReplaceAll(renderMenuData.JinjaTemplateFile, ".j2", ""))
	err = os.WriteFile(filePath, []byte(renderedString), 0644)
	if err != nil {
		return err
	}
	log.Debugf("filename: %s\nresult: %s", renderMenuData.JinjaTemplateFile, renderedString)

	return nil
}

func renderNetinfoMenu(renderMenuData RenderMenuData) error {
	j2, err := jinja2.NewJinja2(renderMenuData.JinjaTemplateFile, 1)
	if err != nil {
		return err
	}
	defer j2.Close()

	renderedString, err := j2.RenderFile(fmt.Sprintf("%s/%s", renderMenuData.WorkingDirectory, renderMenuData.JinjaTemplateFile))
	if err != nil {
		return err
	}

	filePath := fmt.Sprintf("%s/%s", renderMenuData.MenusDirectory, strings.ReplaceAll(renderMenuData.JinjaTemplateFile, ".j2", ""))
	err = os.WriteFile(filePath, []byte(renderedString), 0644)
	if err != nil {
		return err
	}
	log.Debugf("filename: %s\nresult: %s", renderMenuData.JinjaTemplateFile, renderedString)

	return nil
}
