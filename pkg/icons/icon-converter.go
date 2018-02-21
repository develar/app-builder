package icons

import (
	"fmt"
	"image"
	"path/filepath"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/apex/log"
	"github.com/develar/app-builder/pkg/fs"
	"github.com/develar/app-builder/pkg/util"
	"github.com/develar/errors"
	"github.com/disintegration/imaging"
)

func ConfigureCommand(app *kingpin.Application) {
	command := app.Command("icon", "create ICNS or ICO or icon set from PNG files")

	sources := command.Flag("input", "input directory or file").Short('i').Required().Strings()
	iconOutFormat := command.Flag("format", "output format").Short('f').Required().Enum("icns", "ico", "set")
	outDir := command.Flag("out", "output directory").Required().String()
	iconRoots := command.Flag("root", "base directory to resolve relative path").Short('r').Strings()

	command.Action(func(context *kingpin.ParseContext) error {
		resultFile, err := ConvertIcon(*sources, *iconRoots, *iconOutFormat, *outDir)
		if err != nil {
			log.Debugf("%+v\n", err)

			switch t := errors.Cause(err).(type) {
			default:
				return err

			case *ImageSizeError:
				return writeUserError(t)

			case *ImageFormatError:
				return writeUserError(t)
			}
		}

		return util.WriteJsonToStdOut(IconConvertResult{Icons: resultFile})
	})
}

func writeUserError(error ImageError) error {
	return util.WriteJsonToStdOut(MisConfigurationError{Message: error.Error(), Code: error.ErrorCode()})
}

func (t InputFileInfo) GetMaxImage() (image.Image, error) {
	if t.maxImage == nil {
		var err error
		t.maxImage, err = loadImage(t.MaxIconPath, t.recommendedMinSize)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}
	return t.maxImage, nil
}

func validateImageSize(file string, recommendedMinSize int) error {
	firstFileBytes, err := fs.ReadFile(file, 512)
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

func outputFormatToSingleFileExtension(outputFormat string) string {
	if outputFormat == "set" {
		return ".png"
	}
	return "." + outputFormat
}

func ConvertIcon(sourceFiles []string, roots []string, outputFormat string, outDir string) ([]IconInfo, error) {
	// allowed to specify path to icns without extension, so, if file not resolved, try to add ".icns" extension
	outExt := outputFormatToSingleFileExtension(outputFormat)
	resolvedPath, fileInfo, err := resolveSourceFile(sourceFiles, roots, outExt)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	log.WithFields(log.Fields{
		"path": resolvedPath,
		"outputFormat": outputFormat,
	}).Debug("path resolved")

	var inputInfo InputFileInfo
	inputInfo.SizeToPath = make(map[int]string)

	if outputFormat == "icns" {
		inputInfo.recommendedMinSize = 512
	} else {
		inputInfo.recommendedMinSize = 256
	}

	isOutputFormatIco := outputFormat == "ico"
	if strings.HasSuffix(resolvedPath, outExt) {
		if outputFormat != "icns" {
			err = validateImageSize(resolvedPath, inputInfo.recommendedMinSize)
			if err != nil {
				return nil, errors.WithStack(err)
			}
		}

		// size not required in this case
		return []IconInfo{{File: resolvedPath}}, nil
	}

	if fileInfo.IsDir() {
		icons, err := CollectIcons(resolvedPath)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		if outputFormat == "set" {
			return icons, nil
		}

		for _, file := range icons {
			inputInfo.SizeToPath[file.Size] = file.File
		}

		maxIcon := icons[len(icons)-1]
		inputInfo.MaxIconPath = maxIcon.File
		inputInfo.MaxIconSize = maxIcon.Size
	} else {
		if outputFormat == "set" && strings.HasSuffix(resolvedPath, ".icns") {
			result, err := ConvertIcnsToPng(resolvedPath, outDir)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			return result, nil
		}

		maxImage, err := loadImage(resolvedPath, inputInfo.recommendedMinSize)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		if isOutputFormatIco && maxImage.Bounds().Max.X > 256 {
			image256 := imaging.Resize(maxImage, 256, 256, imaging.Lanczos)
			maxImage = image256
		}

		inputInfo.MaxIconSize = maxImage.Bounds().Max.X
		inputInfo.maxImage = maxImage
		inputInfo.SizeToPath[inputInfo.MaxIconSize] = resolvedPath
	}

	outFile := filepath.Join(outDir, "icon" + outExt)
	switch outputFormat {
	case "icns":
		err := ConvertToIcns(inputInfo, outFile)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return []IconInfo{{File: outFile}}, err

	case "ico":
		maxImage, err := inputInfo.GetMaxImage()
		if err != nil {
			return nil, errors.WithStack(err)
		}

		err = SaveImage(maxImage, outFile, ICO)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return []IconInfo{{File: outFile}}, nil

	default:
		return nil, fmt.Errorf("unknown output format %s", resolvedPath)
	}
}

func loadImage(sourceFile string, recommendedMinSize int) (image.Image, error) {
	result, err := LoadImage(sourceFile)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if result.Bounds().Max.X < recommendedMinSize || result.Bounds().Max.Y < recommendedMinSize {
		return nil, errors.WithStack(NewImageSizeError(sourceFile, recommendedMinSize))
	}

	return result, nil
}
