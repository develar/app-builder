package icons

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strings"

	"github.com/apex/log"
	"github.com/develar/app-builder/util"
	"github.com/pkg/errors"
)

// returns file if exists, null if file not exists, or error if unknown error
func resolveSourceFileOrNull(sourceFile string, roots []string) (string, os.FileInfo, error) {
	absolutePath, err := filepath.Abs(sourceFile)
	if err == nil {
		fileInfo, err := os.Stat(absolutePath)
		if err == nil {
			return absolutePath, fileInfo, nil
		}

		log.WithFields(log.Fields{
			"path":  absolutePath,
			"error": err,
		}).Debug("tried specified path, but got error")

		if !os.IsNotExist(err) {
			return "", nil, errors.WithStack(err)
		}
	}

	log.WithFields(log.Fields{
		"path":  sourceFile,
		"error": err,
	}).Debug("tried to convert path to absolute, but got error")

	if !filepath.IsAbs(sourceFile) {
		for _, root := range roots {
			resolvedPath := filepath.Join(root, sourceFile)
			fileInfo, err := os.Stat(resolvedPath)
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
		resolvedPath, fileInfo, err = resolveSourceFileOrNull(sourceFile+extraExtension, roots)
		if err != nil {
			return "", nil, errors.WithStack(err)
		}
		if fileInfo != nil {
			return resolvedPath, fileInfo, nil
		}
	}

	return "", nil, fmt.Errorf("icon source %s not found", sourceFile)
}

type InputFileInfo struct {
	MaxIconSize int
	MaxIconPath string
	SizeToPath  map[int]string

	MaxImage image.Image
}

func validateImageSize(file string, recommendedMinSize int) error {
	firstFileBytes, err := util.ReadFile(file, 512)
	if err != nil {
		return errors.WithStack(err)
	}

	if IsIco(firstFileBytes) {
		for _, size := range GetIcoSizes(firstFileBytes) {
			if size.Width >= recommendedMinSize && size.Height >= recommendedMinSize {
				return nil
			}
		}
	} else {
		config, err := DecodeImageConfig(file)
		if err != nil {
			return errors.WithStack(err)
		}

		if config.Width >= recommendedMinSize && config.Height >= recommendedMinSize {
			return nil
		}
	}

	return NewImageSizeError(file, recommendedMinSize)
}

func ConvertIcon(sourceFile string, roots []string, outputFormat string) (string, error) {
	// allowed to specify path to icns without extension, so, if file not resolved, try to add ".icns" extension
	outExt := "." + outputFormat
	resolvedPath, fileInfo, err := resolveSourceFile(sourceFile, roots, outExt)
	if err != nil {
		return "", errors.WithStack(err)
	}

	sourceFile = resolvedPath

	var recommendedMinSize int
	if outputFormat == "ico" {
		recommendedMinSize = 256
	} else {
		recommendedMinSize = 512
	}

	if strings.HasSuffix(resolvedPath, outExt) {
		if outputFormat == "ico" {
			err = validateImageSize(resolvedPath, recommendedMinSize)
			if err != nil {
				return "", errors.WithStack(err)
			}
		}

		return resolvedPath, nil
	}

	var inputInfo InputFileInfo
	inputInfo.SizeToPath = make(map[int]string)

	if fileInfo.IsDir() {
		icons, err := CollectIcons(sourceFile)
		if err != nil {
			return "", errors.WithStack(err)
		}

		for _, file := range icons.Icons {
			inputInfo.SizeToPath[file.Size] = file.File
		}

		inputInfo.MaxIconPath = icons.MaxIconPath
		inputInfo.MaxIconSize = icons.MaxIconSize
	} else {
		inputInfo.SizeToPath[0] = sourceFile

		maxImage, err := loadImage(sourceFile, recommendedMinSize)
		if err != nil {
			return "", errors.WithStack(err)
		}

		inputInfo.MaxIconSize = maxImage.Bounds().Max.X
		inputInfo.MaxIconPath = sourceFile
	}

	switch outputFormat {
	case "icns":
		return ConvertToIcns(inputInfo)

	case "ico":
		if inputInfo.MaxImage == nil {
			inputInfo.MaxImage, err = loadImage(inputInfo.MaxIconPath, recommendedMinSize)
			if err != nil {
				return "", errors.WithStack(err)
			}
		}

		outFile, err := util.TempFile("", outExt)
		if err != nil {
			return "", errors.WithStack(err)
		}

		err = SaveImage2(inputInfo.MaxImage, outFile, ICO)
		return outFile.Name(), err

	default:
		return "", fmt.Errorf("unknown output format %s", sourceFile)
	}
}

func loadImage(sourceFile string, recommendedMinSize int) (image.Image, error) {
	result, err := LoadImage(sourceFile)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if result.Bounds().Max.X < recommendedMinSize || result.Bounds().Max.Y < recommendedMinSize {
		return nil, NewImageSizeError(sourceFile, recommendedMinSize)
	}

	return result, nil
}
