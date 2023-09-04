package main

import (
	"fmt"
	"io/fs"
	"math"
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
	ImageName         string
	KernelVersionFile string
}

type folderProperties struct {
	FolderPath              string
	ThresholdMaxImagesCount int     // max number of images to keep in the folder.
	MaxFolderSizeInGiB      float64 // max folder size in GiB
}

var (
	propertiesDev = folderProperties{
		FolderPath:              "/cleaning/dev",
		ThresholdMaxImagesCount: 10, // default value that will be overwritten by environment variables, if set
		MaxFolderSizeInGiB:      15, // default value that will be overwritten by environment variables, if set

	}
	propertiesProd = folderProperties{
		FolderPath:              "/cleaning/prod",
		ThresholdMaxImagesCount: 5,  // default value that will be overwritten by environment variables, if set
		MaxFolderSizeInGiB:      10, // default value that will be overwritten by environment variables, if set
	}
)

type ByModTime []fs.DirEntry

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

func main() {
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.JSONFormatter{})

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
			log.Fatal(err)
		}
	}

	maxFolderSizeInGiBProdEnv := os.Getenv("MAX_FOLDER_SIZE_IN_GIB_PROD")
	if maxFolderSizeInGiBProdEnv != "" {
		propertiesProd.MaxFolderSizeInGiB, err = strconv.ParseFloat(maxFolderSizeInGiBProdEnv, 64)
		if err != nil {
			log.Fatal(err)
		}
	}

	// display the current configuration
	log.Infof("Dev folder: %s, ThresholdMaxImagesCount: %d, MaxFolderSizeInGiB: %.2f", propertiesDev.FolderPath, propertiesDev.ThresholdMaxImagesCount, propertiesDev.MaxFolderSizeInGiB)
	log.Infof("Prod folder: %s, ThresholdMaxImagesCount: %d, MaxFolderSizeInGiB: %.2f", propertiesProd.FolderPath, propertiesProd.ThresholdMaxImagesCount, propertiesProd.MaxFolderSizeInGiB)

	var folderProperties = []folderProperties{
		propertiesDev,
		propertiesProd,
	}

	for {
		// Get disk usage for the root directory
		freeSpace, usedSpace, totalSpace, err := calculateDiskSpaceUsage()
		if err != nil {
			log.Errorf("Error calculating disk space usage: %s", err)
		} else {
			log.Infof("Disk free: %.2f%% (%.2f GiB), Disk used: %.2f%% (%.2f GiB), Disk Space total: %.2f GiB", (freeSpace/totalSpace)*100, bytesToGiB(freeSpace), (usedSpace/totalSpace)*100, bytesToGiB(usedSpace), bytesToGiB(totalSpace))
		}

		log.Infof("Image count before deletion: images dev (%d) , images prod (%d)", len(getImagesSortedByModifiedDate(propertiesDev.FolderPath)), len(getImagesSortedByModifiedDate(propertiesProd.FolderPath)))

		// Delete oldest images until the folder size is below the threshold
		for _, folderProperty := range folderProperties {
			images := getImagesSortedByModifiedDate(folderProperty.FolderPath)

			folderSizeInGiB := getCurrentFolderSizeInGiB(folderProperty.FolderPath)

			for i := len(images) - 1; folderNeedsCleanup(folderProperty, folderSizeInGiB, images); i-- {
				err := deleteImage(folderProperty.FolderPath, images[i])
				if err != nil {
					log.Errorf("Error deleting image %s: %s", images[i], err)
				}
				images = getImagesSortedByModifiedDate(folderProperty.FolderPath)
				folderSizeInGiB = getCurrentFolderSizeInGiB(folderProperty.FolderPath)
			}
		}

		log.Infof("Image count after deletion: images dev (%d) , images prod (%d)", len(getImagesSortedByModifiedDate(propertiesDev.FolderPath)), len(getImagesSortedByModifiedDate(propertiesProd.FolderPath)))

		freeSpace, usedSpace, totalSpace, err = calculateDiskSpaceUsage()
		if err != nil {
			log.Errorf("Error calculating disk space usage: %s", err)
		} else {
			log.Infof("Disk free: %.2f%% (%.2f GiB), Disk used: %.2f%% (%.2f GiB), Disk Space total: %.2f GiB", (freeSpace/totalSpace)*100, bytesToGiB(freeSpace), (usedSpace/totalSpace)*100, bytesToGiB(usedSpace), bytesToGiB(totalSpace))
		}

		time.Sleep(5 * time.Minute)
	}
}

func folderNeedsCleanup(folderProperties folderProperties, currentFolderSize float64, allImages []image) bool {
	return folderProperties.MaxFolderSizeInGiB < currentFolderSize || folderProperties.ThresholdMaxImagesCount < len(allImages)
}

func calculateDiskSpaceUsage() (float64, float64, float64, error) {
	fs := syscall.Statfs_t{}
	err := syscall.Statfs("/", &fs)
	if err != nil {
		return 0, 0, 0, err
	}

	usedSpace := float64(fs.Blocks-fs.Bavail) * float64(fs.Bsize)
	totalSpace := float64(fs.Blocks) * float64(fs.Bsize)
	freeSpace := float64(fs.Bavail) * float64(fs.Bsize)

	return freeSpace, usedSpace, totalSpace, nil
}

// func to convert bytes to GiB
func bytesToGiB(bytes float64) float64 {
	return bytes / math.Pow(1024, 3)
}

// returns all images in Folder with newest modified image first and oldest last
func getImagesSortedByModifiedDate(folderName string) []image {
	var images []image

	files, err := os.ReadDir(folderName)
	if err != nil {
		log.Errorf("Error reading folder %s: %s", folderName, err)
	}

	var squashfsFiles []fs.DirEntry
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".squashfs") {
			//during the image sync process, the syncer creates a temporary file with the name ".azDownload...", therefore we should exclude it
			if strings.HasPrefix(file.Name(), ".azDownload") {
				continue
			}
			squashfsFiles = append(squashfsFiles, file)
		}
	}

	sort.Sort(ByModTime(squashfsFiles))

	for _, file := range squashfsFiles {
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

func checkIfKernelVersionFileExists(folderName string, kernelVersionFileName string) bool {
	filePathtoCheck := fmt.Sprintf("%s/%s", folderName, kernelVersionFileName)
	if _, err := os.Stat(filePathtoCheck); os.IsNotExist(err) {
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

	return bytesToGiB(float64(totalSize))
}

func deleteImage(folderName string, image image) error {
	squashFsFilePathtoDelete := fmt.Sprintf("%s/%s", folderName, image.ImageName)
	log.Infof("Deleting image %s", squashFsFilePathtoDelete)
	err := os.Remove(squashFsFilePathtoDelete)
	if err != nil {
		return err
	}

	//delete kernel version file
	kernelVersionFilePathtoDelete := fmt.Sprintf("%s/%s", folderName, image.KernelVersionFile)
	log.Infof("Deleting kernel version file %s", kernelVersionFilePathtoDelete)
	err = os.Remove(kernelVersionFilePathtoDelete)

	return err
}
