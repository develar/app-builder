package commands

import (
	"image"
	"os"

	"github.com/disintegration/imaging"
	"image/png"
	"bufio"
	"text/template"
	"bytes"
	"sync"
	"runtime"
	"os/exec"
	"io/ioutil"
	"path"
	"encoding/json"
	"fmt"
	"github.com/develar/app-builder/util"
	"strings"
)

type IconInfo struct {
	File string `json:"file"`
	Size int    `json:"size"`
}

type Icns2PngMapping struct {
	Id   string
	Size int
}

var icns2PngMappingList = []Icns2PngMapping{
	{"is32", 16},
	{"il32", 32},
	{"ih32", 48},
	{"icp6", 64},
	{"it32", 128},
	{"ic08", 256},
	{"ic09", 512},
}

type ConvertIcnsToPngResult struct {
	MaxIconPath string     `json:"maxIconPath"`
	Icons       []IconInfo `json:"icons"`
}

func ConvertIcnsToPng(inFile string) error {
	tempDir, err := util.TempDir(os.Getenv("ELECTRON_BUILDER_TMP_DIR"), ".iconset")
	if err != nil {
		return err
	}

	var maxIconPath string
	var result []IconInfo

	sizeList := []int{24, 96}
	outFileTemplate := path.Join(tempDir, "icon_{{.Width}}x{{.Height}}.png")
	if runtime.GOOS == "darwin" && os.Getenv("FORCE_ICNS2PNG") == "" {
		output, err := exec.Command("iconutil", "--convert", "iconset", "--output", tempDir, inFile).CombinedOutput()
		if err != nil {
			fmt.Println(string(output))
			return err
		}

		iconFiles, err := ioutil.ReadDir(tempDir)
		if err != nil {
			return err
		}

		for _, item := range icns2PngMappingList {
			fileName := fmt.Sprintf("icon_%dx%d.png", item.Size, item.Size)
			if contains(iconFiles, fileName) {
				// list sorted by size, so, last assignment is a max size
				maxIconPath = path.Join(tempDir, fileName)
				result = append(result, IconInfo{maxIconPath, item.Size})
			} else {
				sizeList = append(sizeList, item.Size)
			}
		}
	} else {
		outputBytes, err := exec.Command("icns2png", "--extract", "--output", tempDir, inFile).CombinedOutput()
		output := string(outputBytes)
		if err != nil {
			fmt.Println(output)
			return err
		}

		namePrefix := strings.TrimSuffix(path.Base(inFile), path.Ext(inFile))

		for _, item := range icns2PngMappingList {
			if strings.Contains(output, item.Id) {
				// list sorted by size, so, last assignment is a max size
				maxIconPath = path.Join(tempDir, fmt.Sprintf("%s_%dx%dx32.png", namePrefix, item.Size, item.Size))
				result = append(result, IconInfo{maxIconPath, item.Size})
			} else {
				sizeList = append(sizeList, item.Size)
			}
		}
	}

	err = multiResizeImage(maxIconPath, outFileTemplate, &result, sizeList, nil)
	if err != nil {
		return err
	}

	serializedResult, err := json.Marshal(ConvertIcnsToPngResult{
		MaxIconPath: maxIconPath,
		Icons:       result,
	})
	if err != nil {
		return err
	}

	_, err = os.Stdout.Write(serializedResult)
	if err != nil {
		return err
	}

	return nil
}

func contains(files []os.FileInfo, name string) bool {
	for _, fileInfo := range files {
		if fileInfo.Name() == name {
			return true
		}
	}
	return false
}

func multiResizeImage(inFile string, outFileNameTemplateString string, result *[]IconInfo, wList []int, hList []int) (error) {
	reader, err := os.Open(inFile)
	if err != nil {
		return err
	}

	originalImage, _, err := image.Decode(reader)
	if err != nil {
		return err
	}

	if hList == nil || len(hList) == 0 {
		hList = wList
	}

	outFileNameTemplate := template.Must(template.New("name").Parse(outFileNameTemplateString))

	var waitGroup sync.WaitGroup

	imageCount := len(wList)
	waitGroup.Add(imageCount)

	for i := 0; i < imageCount; i++ {
		w := wList[i]
		h := hList[i]

		outFilePath, err := computeName(outFileNameTemplate, map[string]interface{}{
			"Width":  w,
			"Height": h,
		})
		if err != nil {
			return err
		}

		*result = append(*result, IconInfo{
			File: outFilePath,
			Size: w,
		})
		go resizeImage(originalImage, w, h, outFilePath, &waitGroup)
	}

	waitGroup.Wait()
	return nil
}

func computeName(template *template.Template, data interface{}) (string, error) {
	outFileNameBuffer := &bytes.Buffer{}
	err := template.Execute(outFileNameBuffer, data)
	if err != nil {
		return "", err
	}
	return outFileNameBuffer.String(), nil
}

func resizeImage(originalImage image.Image, w int, h int, outFileName string, waitGroup *sync.WaitGroup) error {
	defer waitGroup.Done()
	newImage := imaging.Resize(originalImage, w, h, imaging.Lanczos)
	return saveImage(newImage, outFileName)
}

func saveImage(image *image.NRGBA, outFileName string) error {
	outFile, err := os.Create(outFileName)
	if err != nil {
		return err
	}

	writer := bufio.NewWriter(outFile)
	err = png.Encode(writer, image)
	if err != nil {
		return err
	}

	flushError := writer.Flush()
	closeError := outFile.Close()
	if flushError != nil {
		return flushError
	}
	if closeError != nil {
		return closeError
	}

	return nil
}
