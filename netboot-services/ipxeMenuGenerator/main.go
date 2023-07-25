package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"sort"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/kluctl/go-jinja2"
)

type image struct {
	ImageName     string `json:"imageName"`
	KernelVersion string `json:"kernelVersion"`
}

const (
	DevFolder  = "/assets/dev"
	ProdFolder = "/assets/prod"
)

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

		renderMenuIpxe("menu.ipxe.j2", ProdFolder, netbootServerIP, azureNetbootServerIP, onpremExposedNetbootServer, azureBlobstorageURL, azureBlobstorageSASToken, httpAuthUser, httpAuthPassword)

		err := renderAdvancedMenu("advancedmenu.ipxe.j2", netbootServerIP, azureNetbootServerIP, onpremExposedNetbootServer, azureBlobstorageURL, azureBlobstorageSASToken, httpAuthUser, httpAuthPassword)
		if err != nil {
			log.Error(err)
		}

		err = renderNetinfoMenu("netinfo.ipxe.j2")
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

func getMostRecentSquashfsImage(folderName string) (string, error) {
	files, err := os.ReadDir(folderName)
	if err != nil {
		log.Fatal(err)
	}

	var matches []fs.DirEntry
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".squashfs") {
			if strings.HasPrefix(file.Name(), ".azDownload") {
				continue
			}
			matches = append(matches, file)
		}
	}

	sort.Sort(ByModTime(matches))

	if len(matches) > 0 {
		return strings.TrimSuffix(matches[0].Name(), ".squashfs"), nil
	}
	return "", nil
}

type kernelVersion struct {
	KernelVersion string `json:"version"`
}

func getMatchingKernelVersion(folderName string, imageName string) (string, error) {
	var version kernelVersion

	if strings.HasPrefix(imageName, ".azDownload") {
		return "", nil
	}

	bytes, err := os.ReadFile(fmt.Sprintf("%s/%s-kernel.json", folderName, imageName))
	if err != nil {
		return "", err
	}
	err = json.Unmarshal(bytes, &version)
	if err != nil {
		return "", err
	}
	return version.KernelVersion, err
}

func renderMenuIpxe(filename string, folderName string, netbootServerIP string, azureNetbootServerIP string, onpremExposedNetbootServer string, azureBlobstorageURL string, azureBlobstorageSASToken string, httpAuthUser string, httpAuthPassword string) {
	mostRecentSquashfsImageName, err := getMostRecentSquashfsImage(folderName)
	if err != nil {
		log.Error(err)
	}

	if mostRecentSquashfsImageName != "" {
		kernelVersionString, err := getMatchingKernelVersion(folderName, mostRecentSquashfsImageName)
		if err != nil {
			log.Fatalf("could not find matching kernel version to the provided squashFSImage %s. Error: %s", mostRecentSquashfsImageName, err)
		}

		j2, err := jinja2.NewJinja2("menu.ipxe", 1,
			jinja2.WithGlobal("netbootServerIP", netbootServerIP),
			jinja2.WithGlobal("azureNetbootServerIP", azureNetbootServerIP),
			jinja2.WithGlobal("onpremExposedNetbootServer", onpremExposedNetbootServer),
			jinja2.WithGlobal("azureBlobstorageSASToken", azureBlobstorageSASToken),
			jinja2.WithGlobal("azureBlobstorageURL", azureBlobstorageURL),
			jinja2.WithGlobal("httpAuthUser", httpAuthUser),
			jinja2.WithGlobal("httpAuthPassword", httpAuthPassword),
			jinja2.WithGlobal("imageName", mostRecentSquashfsImageName),
			jinja2.WithGlobal("kernelFolderName", kernelVersionString),
		)
		if err != nil {
			log.Fatal(err)
		}
		defer j2.Close()

		renderedString, err := j2.RenderFile(filename)
		if err != nil {
			log.Fatal(err)
		}

		filePath := fmt.Sprintf("/menus/%s", strings.ReplaceAll(filename, ".j2", ""))
		err = os.WriteFile(filePath, []byte(renderedString), 0644)
		if err != nil {
			log.Fatal(err)
		}

		log.Debugf("filename: %s\nresult: %s", filename, renderedString)
	}
}

func getDevImages() []image {
	return getImages(DevFolder)
}

func getProdImages() []image {
	return getImages(ProdFolder)
}

func getImages(folderName string) []image {
	var images []image

	files, err := os.ReadDir(folderName)
	if err != nil {
		fmt.Println("Error:", err)
		panic(err)
	}

	var squashfsFiles []fs.DirEntry
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".squashfs") {
			if strings.HasPrefix(file.Name(), ".azDownload") {
				continue
			}
			squashfsFiles = append(squashfsFiles, file)
		}
	}

	sort.Sort(ByModTime(squashfsFiles))

	for _, file := range squashfsFiles {
		imageName := strings.TrimSuffix(file.Name(), ".squashfs")
		kernelVersionString, err := getMatchingKernelVersion(folderName, imageName)
		if err != nil {
			log.Errorf("could not find matching kernel version to the provided squashFSImage %s. Error: %s", imageName, err)
		}

		images = append(images, image{ImageName: imageName, KernelVersion: kernelVersionString})
	}

	return images
}

func renderAdvancedMenu(filename string, netbootServerIP string, azureNetbootServerIP string, onpremExposedNetbootServer string, azureBlobstorageURL string, azureBlobstorageSASToken string, httpAuthUser string, httpAuthPassword string) error {
	prodImages := getProdImages()
	devImages := getDevImages()
	j2, err := jinja2.NewJinja2("advancedmenu.ipxe", 1,
		jinja2.WithGlobal("netbootServerIP", netbootServerIP),
		jinja2.WithGlobal("azureNetbootServerIP", azureNetbootServerIP),
		jinja2.WithGlobal("onpremExposedNetbootServer", onpremExposedNetbootServer),
		jinja2.WithGlobal("azureBlobstorageSASToken", azureBlobstorageSASToken),
		jinja2.WithGlobal("azureBlobstorageURL", azureBlobstorageURL),
		jinja2.WithGlobal("httpAuthUser", httpAuthUser),
		jinja2.WithGlobal("httpAuthPassword", httpAuthPassword),
		jinja2.WithGlobal("prod", prodImages),
		jinja2.WithGlobal("dev", devImages),
	)
	if err != nil {
		return err
	}
	defer j2.Close()

	renderedString, err := j2.RenderFile(filename)
	if err != nil {
		return err
	}

	filePath := fmt.Sprintf("/menus/%s", strings.ReplaceAll(filename, ".j2", ""))
	err = os.WriteFile(filePath, []byte(renderedString), 0644)
	if err != nil {
		return err
	}
	log.Debugf("filename: %s\nresult: %s", filename, renderedString)

	return nil
}

func renderNetinfoMenu(filename string) error {
	j2, err := jinja2.NewJinja2("netinfo.ipxe", 1)
	if err != nil {
		return err
	}
	defer j2.Close()

	renderedString, err := j2.RenderFile(filename)
	if err != nil {
		return err
	}

	filePath := fmt.Sprintf("/menus/%s", strings.ReplaceAll(filename, ".j2", ""))
	err = os.WriteFile(filePath, []byte(renderedString), 0644)
	if err != nil {
		return err
	}
	log.Debugf("filename: %s\nresult: %s", filename, renderedString)

	return nil
}
