package icons

import (
	"fmt"
	"image"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/develar/app-builder/pkg/log"
	"github.com/develar/app-builder/pkg/util"
	"github.com/develar/errors"
	"github.com/develar/go-fs-util"
	"github.com/disintegration/imaging"
)

type Icns2PngMapping struct {
	Id   string
	Size int
}

var icnsTypeToSize = []Icns2PngMapping{
	{"is32", 16},
	{"il32", 32},
	{"ih32", 48},
	{"icp6", 64},
	{"it32", 128},
	{ICNS_256, 256},
	{ICNS_512, 512},
}

func ConvertIcnsToPng(inFile string, outDir string) ([]IconInfo, error) {
	err := fsutil.EnsureEmptyDir(outDir)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var sizeList []int
	var result []IconInfo
	if runtime.GOOS == "darwin" && os.Getenv("FORCE_ICNS2PNG") == "" {
		result, err = ConvertIcnsToPngUsingIconUtil(inFile, outDir, &sizeList)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	} else {
		result, err = ConvertIcnsToPngUsingOpenJpeg(inFile, outDir)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		sortBySize(result)
		for _, item := range icnsTypeToSize {
			if !hasSize(result, item.Size) {
				sizeList = append(sizeList, item.Size)
			}
		}
	}

	maxIconInfo := result[len(result)-1]
	err = multiResizeImage(maxIconInfo.File, filepath.Join(outDir, "icon_%dx%d.png"), &result, sizeList)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	sortBySize(result)
	return result, nil
}

func ConvertIcnsToPngUsingIconUtil(inFile string, outDir string, sizeList *[]int) ([]IconInfo, error) {
	// iconutil requires suffix .iconset
	outDir = filepath.Join(outDir, "result.iconset")
	err := os.Mkdir(outDir, 0755)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	output, err := exec.Command("iconutil", "--convert", "iconset", "--output", outDir, inFile).CombinedOutput()
	if err != nil {
		log.Info(string(output))
		return nil, errors.WithStack(err)
	}

	iconFileNames, err := fsutil.ReadDirContent(outDir)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var result []IconInfo
	for _, item := range icnsTypeToSize {
		fileName := fmt.Sprintf("icon_%dx%d.png", item.Size, item.Size)
		if util.ContainsString(iconFileNames, fileName) {
			result = append(result, IconInfo{filepath.Join(outDir, fileName), item.Size})
		} else {
			*sizeList = append(*sizeList, item.Size)
		}
	}
	return result, nil
}

func hasSize(list []IconInfo, size int) bool {
	for _, info := range list {
		if info.Size == size {
			return true
		}
	}
	return false
}

func multiResizeImage(inFile string, outFileNameFormat string, result *[]IconInfo, sizeList []int) error {
	imageCount := len(sizeList)
	if imageCount == 0 {
		return nil
	}

	originalImage, err := LoadImage(inFile)
	if err != nil {
		return errors.WithStack(err)
	}

	return multiResizeImage2(&originalImage, outFileNameFormat, result, sizeList)
}

func multiResizeImage2(originalImage *image.Image, outFileNameFormat string, result *[]IconInfo, sizeList []int) error {
	imageCount := len(sizeList)
	if imageCount == 0 {
		return nil
	}

	maxSize := (*originalImage).Bounds().Max.X

	return util.MapAsync(imageCount, func(taskIndex int) (func() error, error) {
		size := sizeList[taskIndex]
		if size > maxSize {
			return nil, nil
		}

		outFilePath := fmt.Sprintf(outFileNameFormat, size, size)
		*result = append(*result, IconInfo{
			File: outFilePath,
			Size: size,
		})

		return func() error {
			newImage := imaging.Resize(*originalImage, size, size, imaging.Lanczos)
			return SaveImage(newImage, outFilePath, PNG)
		}, nil
	})
}
