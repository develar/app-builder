package icons

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/png"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/apex/log"
	"github.com/develar/app-builder/util"
	"github.com/disintegration/imaging"
	"github.com/pkg/errors"
)

var (
	icnsHeader = []byte{0x69, 0x63, 0x6e, 0x73}

	icnsExpectedSizes = []int{16, 32, 64, 128, 256, 512, 1024}

	// all icon sizes mapped to their respective possible OSTypes, this includes old OSTypes such as ic08 recognized on 10.5
	sizeToType = map[int][]string{
		16:   {"icp4"},
		32:   {"icp5", "ic11"},
		64:   {"icp6", "ic12"},
		128:  {"ic07"},
		256:  {"ic08", "ic13"},
		512:  {"ic09", "ic14"},
		1024: {"ic10"},
	}
)

// returns file if exists, null if file not exists, or error if unknown error
func resolveSourceFileOrNull(sourceFile string, roots []string) (string, os.FileInfo, error) {
	fileInfo, err := os.Stat(sourceFile)
	if err == nil {
		return sourceFile, fileInfo, nil
	}

	log.WithFields(log.Fields{
		"path":  sourceFile,
		"error": err,
	}).Debug("tried specified path, but got error")

	if !os.IsNotExist(err) {
		return "", nil, errors.WithStack(err)
	}

	if !filepath.IsAbs(sourceFile) {
		for _, root := range roots {
			resolvedPath := filepath.Join(root, sourceFile)
			fileInfo, err = os.Stat(resolvedPath)
			if err == nil {
				return resolvedPath, fileInfo, nil
			} else {
				log.WithFields(log.Fields{
					"path":  resolvedPath,
					"error": err,
				}).Debug("tried resolved path, but got error")
			}
		}
	}

	return "", nil, nil
}

func resolveSourceFile(sourceFile string, roots []string, extraExtension string) (string, os.FileInfo, error) {
	resolvedPath, fileInfo, err := resolveSourceFileOrNull(sourceFile, roots)
	if err != nil {
		return "", nil, errors.WithStack(err)
	}
	if fileInfo != nil {
		return resolvedPath, fileInfo, nil
	}

	if extraExtension != "" {
		resolvedPath, fileInfo, err = resolveSourceFileOrNull(sourceFile + extraExtension, roots)
		if err != nil {
			return "", nil, errors.WithStack(err)
		}
		if fileInfo != nil {
			return resolvedPath, fileInfo, nil
		}
	}

	return "", nil, fmt.Errorf("icon source %s not found", sourceFile)
}

func ConvertPngToIcns(sourceFile string, roots []string) (string, error) {
	// allowed to specify path to icns without extension, so, if file not resolved, try to add ".icns" extension
	resolvedPath, fileInfo, err := resolveSourceFile(sourceFile, roots, ".icns")
	if err != nil {
		return "", errors.WithStack(err)
	}

	sourceFile = resolvedPath

	if strings.HasSuffix(sourceFile, ".icns") {
		return sourceFile, nil
	}

	sizeToPath := make(map[int]string)
	var maxIconSize int

	var maxImage image.Image
	var maxIconFile string

	if fileInfo.IsDir() {
		icons, err := CollectIcons(sourceFile)
		if err != nil {
			return "", errors.WithStack(err)
		}

		for _, file := range icons.Icons {
			sizeToPath[file.Size] = file.File
		}

		maxIconFile = icons.MaxIconPath
		maxIconSize = icons.MaxIconSize
	} else {
		sizeToPath[0] = sourceFile

		maxImage, err := LoadImage(sourceFile)
		if err != nil {
			return "", errors.WithStack(err)
		}

		if maxImage.Bounds().Max.X < 512 || maxImage.Bounds().Max.Y < 512 {
			return "", fmt.Errorf("image %s must be at least 512x512", sourceFile)
		}

		maxIconSize = maxImage.Bounds().Max.X
		maxIconFile = sourceFile
	}

	// create a new buffer to hold the series of icons generated via resizing
	icns := new(bytes.Buffer)

	for _, size := range icnsExpectedSizes {
		if size > maxIconSize {
			// do not upscale
			continue
		}

		var imageData []byte

		existingFile, exists := sizeToPath[size]
		if exists {
			imageData, err = ioutil.ReadFile(existingFile)
			if err != nil {
				return "", errors.WithStack(err)
			}
		} else {
			if maxImage == nil {
				maxImage, err = LoadImage(maxIconFile)
				if err != nil {
					return "", errors.WithStack(err)
				}
			}

			imageBuffer := new(bytes.Buffer)
			err := png.Encode(imageBuffer, imaging.Resize(maxImage, size, size, imaging.Lanczos))
			if err != nil {
				return "", errors.WithStack(err)
			}

			imageData = imageBuffer.Bytes()
		}

		// each icon type is prefixed with a 4-byte OSType marker and a 4-byte size header (which includes the ostype/size header).
		// add the size of the total icon to lengthBytes in big-endian format.
		lengthBytes := make([]byte, 4, 4)
		binary.BigEndian.PutUint32(lengthBytes, uint32(len(imageData)+8))

		// iterate through every OSType and append the icon to icns
		for _, ostype := range sizeToType[size] {
			_, err = icns.Write([]byte(ostype))
			if err != nil {
				return "", errors.WithStack(err)
			}
			_, err = icns.Write(lengthBytes)
			if err != nil {
				return "", errors.WithStack(err)
			}
			_, err = icns.Write(imageData)
			if err != nil {
				return "", errors.WithStack(err)
			}
		}
	}

	// each ICNS file is prefixed with a 4 byte header and 4 bytes marking the length of the file, MSB first
	lengthBytes := make([]byte, 4, 4)
	binary.BigEndian.PutUint32(lengthBytes, uint32(icns.Len()+8))

	outFile, err := util.TempFile("", ".icns")
	if err != nil {
		return "", errors.WithStack(err)
	}

	defer outFile.Close()

	outFile.Write(icnsHeader)
	outFile.Write(lengthBytes)
	io.Copy(outFile, icns)

	return outFile.Name(), nil
}
