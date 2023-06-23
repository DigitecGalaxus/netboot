package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/kluctl/go-jinja2"
)

type image struct {
	ImageName     string `json:"imageName"`
	KernelVersion string `json:"kernelVersion"`
}

func main() {
	for {
		netbootServerIP := os.Getenv("NETBOOT_SERVER_IP")
		renderMenuIpxe("menu.ipxe.j2", "prod", netbootServerIP)
		renderAdvancedMenu("advancedmenu.ipxe.j2", netbootServerIP)
		time.Sleep(60 * time.Second)
	}
}

type ByModTime []fs.DirEntry

func (b ByModTime) Len() int      { return len(b) }
func (b ByModTime) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b ByModTime) Less(i, j int) bool {
	infoI, err := b[i].Info()
	if err != nil {
		panic(err)
	}
	infoJ, err := b[j].Info()
	if err != nil {
		panic(err)
	}
	return infoI.ModTime().After(infoJ.ModTime())
}

func getMostRecentSquashfsImage(folderName string) string {
	files, err := os.ReadDir(folderName)
	if err != nil {
		fmt.Println("Error:", err)
		panic(err)
	}

	var matches []fs.DirEntry
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".squashfs") {
			matches = append(matches, file)
		}
	}

	sort.Sort(ByModTime(matches))

	if len(matches) > 0 {
		return strings.TrimSuffix(matches[0].Name(), ".squashfs")
	}
	panic("No squashfs image found")
}

type kernelVersion struct {
	KernelVersion string `json:"version"`
}

func getMatchingKernelVersion(folderName string, imageName string) string {
	var version kernelVersion
	bytes, err := os.ReadFile(fmt.Sprintf("%s/%s-kernel.json", folderName, imageName))
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(bytes, &version)
	if err != nil {
		panic(err)
	}
	return version.KernelVersion
}

func renderMenuIpxe(filename string, folderName string, netbootServerIP string) {
	mostRecentSquashfsImageName := getMostRecentSquashfsImage(folderName)
	kernelVersionString := getMatchingKernelVersion(folderName, mostRecentSquashfsImageName)

	j2, err := jinja2.NewJinja2("menu.ipxe", 1,
		jinja2.WithGlobal("netbootServerIP", netbootServerIP),
		jinja2.WithGlobal("imageName", mostRecentSquashfsImageName),
		jinja2.WithGlobal("kernelFolderName", kernelVersionString),
	)
	if err != nil {
		panic(err)
	}
	defer j2.Close()

	renderedString, err := j2.RenderFile(filename)
	if err != nil {
		panic(err)
	}

	filePath := fmt.Sprintf("/menus/%s", strings.ReplaceAll(filename, ".j2", ""))
	os.WriteFile(filePath, []byte(renderedString), 0644)
	fmt.Printf("filename: %s\nresult: %s", filename, renderedString)
}

func getDevImages() []image {
	return getImages("dev")
}

func getProdImages() []image {
	return getImages("prod")
}

func getImages(folderName string) []image {
	var images []image

	files, err := os.ReadDir(folderName)
	if err != nil {
		fmt.Println("Error:", err)
		panic(err)
	}

	var matches []fs.DirEntry
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".squashfs") {
			matches = append(matches, file)
		}
	}

	sort.Sort(ByModTime(matches))

	for _, file := range matches {
		imageName := strings.TrimSuffix(file.Name(), ".squashfs")
		kernelVersionString := getMatchingKernelVersion(folderName, imageName)
		images = append(images, image{ImageName: imageName, KernelVersion: kernelVersionString})
	}

	return images
}

func renderAdvancedMenu(filename string, netbootServerIP string) {
	prodImages := getProdImages()
	devImages := getDevImages()
	j2, err := jinja2.NewJinja2("advancedmenu.ipxe", 1,
		jinja2.WithGlobal("netbootServerIP", netbootServerIP),
		jinja2.WithGlobal("prod", prodImages),
		jinja2.WithGlobal("dev", devImages),
	)
	if err != nil {
		panic(err)
	}
	defer j2.Close()

	renderedString, err := j2.RenderFile(filename)
	if err != nil {
		panic(err)
	}

	filePath := fmt.Sprintf("/menus/%s", strings.ReplaceAll(filename, ".j2", ""))
	os.WriteFile(filePath, []byte(renderedString), 0644)
	fmt.Printf("filename: %s\nresult: %s", filename, renderedString)
}
