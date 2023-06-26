package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

type image struct {
	ImageName         string `json:"imageName"`
	KernelVersionFile string `json:"KernelVersionFile"`
}

type folderProperties struct {
	FolderPath              string  // relative path to the folder
	ThresholdMaxImagesCount int     // max number of images to keep in the folder. Define n+1 images to keep n images (due to indexing starting at 0)
	MaxFolderSizeInGiB      float64 // max folder size in GiB
}

var (
	propertiesDev = folderProperties{
		FolderPath:              "/cleaning/dev",
		ThresholdMaxImagesCount: 10,
		MaxFolderSizeInGiB:      15,
	}
	propertiesProd = folderProperties{
		FolderPath:              "/cleaning/prod",
		ThresholdMaxImagesCount: 5,
		MaxFolderSizeInGiB:      10,
	}
)

type ByModTime []fs.DirEntry

func main() {
	var err error
	ThresholdMaxImagesCountDevEnv := os.Getenv("THRESHOLD_MAX_IMAGES_COUNT_DEV")
	if ThresholdMaxImagesCountDevEnv != "" {
		propertiesDev.ThresholdMaxImagesCount, err = strconv.Atoi(ThresholdMaxImagesCountDevEnv)
		if err != nil {
			log.Fatal(err)
		}
	}

	thresholdMaxImagesCountProdEnv := os.Getenv("THRESHOLD_MAX_IMAGES_COUNT_PROD")
	if thresholdMaxImagesCountProdEnv != "" {
		propertiesProd.ThresholdMaxImagesCount, err = strconv.Atoi(thresholdMaxImagesCountProdEnv)
		if err != nil {
			log.Fatal(err)
		}
	}

	maxFolderSizeInGiBdevEnv := os.Getenv("MAX_FOLDER_SIZE_IN_GIB_DEV")
	if maxFolderSizeInGiBdevEnv != "" {
		propertiesDev.MaxFolderSizeInGiB, err = strconv.ParseFloat(maxFolderSizeInGiBdevEnv, 64)

		if err != nil {
			log.Error(err)
		}
	}

	maxFolderSizeInGiBProdEnv := os.Getenv("MAX_FOLDER_SIZE_IN_GIB_PROD")
	if maxFolderSizeInGiBProdEnv != "" {
		propertiesProd.MaxFolderSizeInGiB, err = strconv.ParseFloat(maxFolderSizeInGiBProdEnv, 64)
		if err != nil {
			log.Error(err)
		}
	}

	for {
		// Get disk usage for the root directory
		fs := syscall.Statfs_t{}
		err := syscall.Statfs("/", &fs)
		if err != nil {
			log.Fatal(err)
		}

		// Calculate disk usage percentage
		usedSpace := float64(fs.Blocks-fs.Bavail) * float64(fs.Bsize)
		totalSpace := float64(fs.Blocks) * float64(fs.Bsize)
		freeSpace := float64(fs.Bavail) * float64(fs.Bsize)
		usedPercent := (usedSpace / totalSpace) * 100
		freeSpacePercent := (freeSpace / totalSpace) * 100

		var folderProperties = []folderProperties{
			propertiesDev,
			propertiesProd,
		}

		// Delete oldest images until the folder size is below the threshold
		for _, folderProperty := range folderProperties {
			images := getImages(folderProperty.FolderPath)
			folderSizeInGiB := getCurrentFolderSizeInGiB(folderProperty.FolderPath)

			for i := len(images) - 1; folderNeedsCleanup(folderProperty, folderSizeInGiB, images); i-- {
				err := deleteImage(folderProperty.FolderPath, images[i])
				if err != nil {
					log.Errorf("Error deleting image %s: %s", images[i], err)
				}
				images = getImages(folderProperty.FolderPath)
				folderSizeInGiB = getCurrentFolderSizeInGiB(folderProperty.FolderPath)
			}
		}

		log.Infof("Disk free: %.2f%%, Disk used: %.2f%% DevImages: %d, ProdImages: %d. No cleanup necessary.", freeSpacePercent, usedPercent, len(getImages(propertiesDev.FolderPath)), len(getImages(propertiesProd.FolderPath)))
		time.Sleep(5 * time.Minute)
	}
}

func folderNeedsCleanup(folderProperties folderProperties, currentFolderSize float64, allImages []image) bool {
	if folderProperties.MaxFolderSizeInGiB < currentFolderSize || folderProperties.ThresholdMaxImagesCount < len(allImages) {
		return true
	}
	return false
}

func (b ByModTime) Len() int      { return len(b) }
func (b ByModTime) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b ByModTime) Less(i, j int) bool {
	infoI, err := b[i].Info()
	if err != nil {
		log.Errorf("Error getting file info for %s: %s", b[i].Name(), err)
	}
	infoJ, err := b[j].Info()
	if err != nil {
		log.Errorf("Error getting file info for %s: %s", b[j].Name(), err)
	}
	return infoI.ModTime().After(infoJ.ModTime())
}

func getImages(folderName string) []image {
	var images []image
	// var folderExists bool
	files, err := os.ReadDir(folderName)
	if err != nil {
		if os.IsNotExist(err) {
			log.Errorf("Folder %s does not exist", folderName)
			// return nil, false
		}
	}
	var matches []fs.DirEntry
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".squashfs") {
			matches = append(matches, file)
		}
	}

	sort.Sort(ByModTime(matches))

	for _, file := range matches {
		imageName := file.Name()
		kernelVersionFilePath := fmt.Sprintf("%s-kernel.json", strings.TrimSuffix(imageName, ".squashfs"))
		matchingKernelVersionFileExists := checkIfKernelVersionFileExists(folderName, kernelVersionFilePath)
		if !matchingKernelVersionFileExists {
			log.Errorf("KernelVersionFile %s does not exist for image %s", kernelVersionFilePath, imageName)
			continue
		}
		images = append(images, image{ImageName: imageName, KernelVersionFile: kernelVersionFilePath})
	}

	return images
}

func checkIfKernelVersionFileExists(folderName string, kernelVersionFilePath string) bool {
	relativeFilePathtoCheck := fmt.Sprintf("%s/%s", folderName, kernelVersionFilePath)
	if _, err := os.Stat(relativeFilePathtoCheck); os.IsNotExist(err) {
		return false
	}
	return true
}

func getCurrentFolderSizeInGiB(folderName string) float64 {
	var totalSize int64
	err := filepath.Walk(folderName, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})

	if err != nil {
		log.Errorf("Error calculating folder size: %s", err)
	}

	return float64(totalSize) / (1024 * 1024 * 1024)
}

func deleteImage(folderName string, image image) error {
	relativeSquashFsFilePathtoDelete := fmt.Sprintf("%s/%s", folderName, image.ImageName)
	log.Warnf("Deleting image %s", relativeSquashFsFilePathtoDelete)
	err := os.Remove(relativeSquashFsFilePathtoDelete)
	if err != nil {
		return err
	}

	//delete kernel version file
	relativeKernelVersionFilePathtoDelete := fmt.Sprintf("%s/%s", folderName, image.KernelVersionFile)
	log.Warnf("Deleting kernel version file %s", relativeKernelVersionFilePathtoDelete)
	err = os.Remove(relativeKernelVersionFilePathtoDelete)
	if err != nil {
		return err
	}

	return nil
}
