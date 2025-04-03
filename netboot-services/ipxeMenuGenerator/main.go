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
	BasicData                RenderBaseData
	NetbootServerIP          string
	AzureNetbootServerIP     string
	AzureBlobstorageURL      string
	AzureBlobstorageSASToken string
}

type RenderAdvancedMenuData struct {
	BasicData       RenderBaseData
	NetbootServerIP string
	devImages       []SquashfsPaths
	prodImages      []SquashfsPaths
}

type RenderBaseData struct {
	JinjaTemplateFile string
	MenusDirectory    string
	WorkingDirectory  string
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

		azureBlobstorageSASToken := os.Getenv("AZURE_SYNC_SAS_TOKEN")
		if azureBlobstorageSASToken == "" {
			log.Fatal("AZURE_SYNC_SAS_TOKEN not set")
		}

		azureBlobstorageURL := os.Getenv("AZURE_SYNC_BLOB_URL")
		if azureBlobstorageURL == "" {
			log.Fatal("AZURE_SYNC_BLOB_URL not set")
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

		err = renderMenuIpxe(
			RenderMenuData{
				BasicData: RenderBaseData{
					JinjaTemplateFile: "menu.ipxe.j2",
					MenusDirectory:    MenusDirectory,
					WorkingDirectory:  WorkingDirectory,
				},
				NetbootServerIP:          netbootServerIP,
				AzureNetbootServerIP:     azureNetbootServerIP,
				AzureBlobstorageURL:      azureBlobstorageURL,
				AzureBlobstorageSASToken: azureBlobstorageSASToken,
			}, mostRecentSquashfsImage)
		if err != nil {
			log.Fatal(err)
		}

		prodImages, err := getImages(ProdFolder)
		if err != nil {
			log.Error(err)
		}

		devImages, err := getImages(DevFolder)
		if err != nil {
			log.Error(err)
		}

		err = renderAdvancedMenu(RenderAdvancedMenuData{
			BasicData: RenderBaseData{
				JinjaTemplateFile: "advancedmenu.ipxe.j2",
				MenusDirectory:    MenusDirectory,
				WorkingDirectory:  WorkingDirectory,
			},
			NetbootServerIP: netbootServerIP,
			devImages:       devImages,
			prodImages:      prodImages,
		})
		if err != nil {
			log.Error(err)
		}

		err = renderNetinfoMenu(RenderBaseData{
			JinjaTemplateFile: "netinfo.ipxe.j2",
			MenusDirectory:    MenusDirectory,
			WorkingDirectory:  WorkingDirectory,
		})
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

func renderMenuIpxe(menuData RenderMenuData, mostRecentSquashFS SquashfsPaths) error {
	j2, err := jinja2.NewJinja2("menu.ipxe", 1,
		jinja2.WithGlobal("netbootServerIP", menuData.NetbootServerIP),
		jinja2.WithGlobal("azureNetbootServerIP", menuData.AzureNetbootServerIP),
		jinja2.WithGlobal("azureBlobstorageSASToken", menuData.AzureBlobstorageSASToken),
		jinja2.WithGlobal("azureBlobstorageURL", menuData.AzureBlobstorageURL),
		jinja2.WithGlobal("imageName", mostRecentSquashFS),
	)
	if err != nil {
		return err
	}
	defer j2.Close()

	renderedString, err := j2.RenderFile(fmt.Sprintf("%s/%s", menuData.BasicData.WorkingDirectory, menuData.BasicData.JinjaTemplateFile))
	if err != nil {
		return err
	}

	filePath := fmt.Sprintf("%s/%s", menuData.BasicData.MenusDirectory, strings.ReplaceAll(menuData.BasicData.JinjaTemplateFile, ".j2", ""))
	err = os.WriteFile(filePath, []byte(renderedString), 0644)
	if err != nil {
		return err
	}

	log.Debugf("filename: %s\nresult: %s", menuData.BasicData.JinjaTemplateFile, renderedString)

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
		if folder.Type() == os.ModeDir {
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

func renderAdvancedMenu(advancedMenuData RenderAdvancedMenuData) error {
	j2, err := jinja2.NewJinja2("advancedmenu.ipxe", 1,
		jinja2.WithGlobal("netbootServerIP", advancedMenuData.NetbootServerIP),
		jinja2.WithGlobal("prod", advancedMenuData.prodImages),
		jinja2.WithGlobal("dev", advancedMenuData.devImages),
	)

	if err != nil {
		return err
	}
	defer j2.Close()

	renderedString, err := j2.RenderFile(fmt.Sprintf("%s/%s", advancedMenuData.BasicData.WorkingDirectory, advancedMenuData.BasicData.JinjaTemplateFile))
	if err != nil {
		return err
	}

	filePath := fmt.Sprintf("%s/%s", advancedMenuData.BasicData.MenusDirectory, strings.ReplaceAll(advancedMenuData.BasicData.JinjaTemplateFile, ".j2", ""))
	err = os.WriteFile(filePath, []byte(renderedString), 0644)
	if err != nil {
		return err
	}
	log.Debugf("filename: %s\nresult: %s", advancedMenuData.BasicData.JinjaTemplateFile, renderedString)

	return nil
}

func renderNetinfoMenu(netInfoData RenderBaseData) error {
	j2, err := jinja2.NewJinja2(netInfoData.JinjaTemplateFile, 1)
	if err != nil {
		return err
	}
	defer j2.Close()

	renderedString, err := j2.RenderFile(fmt.Sprintf("%s/%s", netInfoData.WorkingDirectory, netInfoData.JinjaTemplateFile))
	if err != nil {
		return err
	}

	filePath := fmt.Sprintf("%s/%s", netInfoData.MenusDirectory, strings.ReplaceAll(netInfoData.JinjaTemplateFile, ".j2", ""))
	err = os.WriteFile(filePath, []byte(renderedString), 0644)
	if err != nil {
		return err
	}
	log.Debugf("filename: %s\nresult: %s", netInfoData.JinjaTemplateFile, renderedString)

	return nil
}
