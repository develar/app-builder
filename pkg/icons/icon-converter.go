package icons

import (
	"image"
	"path/filepath"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/apex/log"
	"github.com/develar/app-builder/pkg/fs"
	"github.com/develar/app-builder/pkg/util"
	"github.com/develar/errors"
	"github.com/disintegration/imaging"
	"github.com/phayes/permbits"
)

func ConfigureCommand(app *kingpin.Application) {
	command := app.Command("icon", "create ICNS or ICO or icon set from PNG files")

	// first is used as is and file ext can be missed,
	// then common sources added,
	sources := command.Flag("input", "input directory or file").Short('i').Strings()
	// then fallback sources
	fallbackSources := command.Flag("fallback-input", "input directory or file").Strings()
	iconOutFormat := command.Flag("format", "output format").Short('f').Required().Enum("icns", "ico", "set")
	outDir := command.Flag("out", "output directory").Required().String()
	iconRoots := command.Flag("root", "base directory to resolve relative path").Short('r').Strings()

	command.Action(func(context *kingpin.ParseContext) error {
		icons, err := ConvertIcon(createCommonIconSources(*sources, *fallbackSources, *iconOutFormat), *iconRoots, *iconOutFormat, *outDir)
		if err != nil {
			switch t := errors.Cause(err).(type) {
			case *ImageSizeError:
				log.Debugf("%+v\n", err)
				return writeUserError(t)

			case *ImageFormatError:
				log.Debugf("%+v\n", err)
				return writeUserError(t)

			default:
				return err
			}
		}

		return util.WriteJsonToStdOut(IconConvertResult{Icons: icons})
	})
}

func isFileHasImageFormatExtension(name string, outputFormat string) bool {
	return strings.HasSuffix(name, "."+outputFormat) || strings.HasSuffix(name, ".png") || strings.HasSuffix(name, ".ico") || strings.HasSuffix(name, ".svg") || strings.HasSuffix(name, ".icns")
}

func createCommonIconSources(sources []string, fallbackSources []string, outputFormat string) []string {
	var list []string

	if len(sources) != 0 {
		source := sources[0]
		// do not use filepath.Ext to ensure that dot can be used in filename
		if outputFormat != "set" && !isFileHasImageFormatExtension(source, outputFormat) {
			list = append(list, source+"."+outputFormat)
			appendImageVariants(source, outputFormat, list)
		} else {
			list = append(list, source)
		}
	}

	if outputFormat != "set" {
		list = append(list, "icon." + outputFormat)
	}

	list = append(list, "icons")

	list = appendImageVariants("icon", outputFormat, list)

	if sources != nil && len(sources) > 1 {
		list = append(list, sources[1:]...)
	}
	if fallbackSources != nil && len(fallbackSources) > 0 {
		list = append(list, fallbackSources...)
	}
	return list
}

func appendImageVariants(name string, outputFormat string, list []string) []string {
	if outputFormat != "png" {
		list = append(list, name + ".png")
	}
	if outputFormat != "icns" {
		list = append(list, name + ".icns")
		// ico only for non icns
		if outputFormat != "ico" {
			list = append(list, name+".ico")
		}
	}
	return list
}

func writeUserError(error util.MessageError) error {
	return util.WriteJsonToStdOut(MisConfigurationError{Message: error.Error(), Code: error.ErrorCode()})
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
	resolvedPath, fileInfo, err := resolveSourceFile(sourceFiles, roots)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if resolvedPath == "" {
		return nil, nil
	}

	log.WithFields(log.Fields{
		"path":         resolvedPath,
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
		icons, iconFileName, err := CollectIcons(resolvedPath)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		if len(icons) == 0 {
			err = configureInputInfoFromSingleFile(iconFileName, isOutputFormatIco, &inputInfo)
			if err != nil {
				return nil, errors.WithStack(err)
			}

			if outputFormat == "set" {
				return resizePngForLinux(&inputInfo, iconFileName, outDir)
			}
		} else {
			err = checkAndFixIconPermissions(icons)
			if err != nil {
				return nil, errors.WithStack(err)
			}
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
		if outputFormat == "set" {
			if strings.HasSuffix(resolvedPath, ".icns") {
				result, err := ConvertIcnsToPng(resolvedPath, outDir)
				if err != nil {
					return nil, errors.WithStack(err)
				}
				return result, nil
			} else if strings.HasSuffix(resolvedPath, ".svg") {
				return []IconInfo{
					{
						File: resolvedPath,
						Size: 1024,
					},
				}, nil
			}
		}

		err = configureInputInfoFromSingleFile(resolvedPath, isOutputFormatIco, &inputInfo)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}

	return convertSingleFile(&inputInfo, filepath.Join(outDir, "icon"+outExt), outputFormat)
}

// https://github.com/electron-userland/electron-builder/issues/2654#issuecomment-369972916
func checkAndFixIconPermissions(icons []IconInfo) error {
	return util.MapAsync(len(icons), func(taskIndex int) (func() error, error) {
		filePath := icons[taskIndex].File
		return func() error {
			permissions, err := permbits.Stat(filePath)
			if err != nil {
				return errors.WithStack(err)
			}

			if permissions.GroupRead() && permissions.OtherRead() {
				return nil
			}

			log.WithFields(log.Fields{
				"file":   filePath,
				"reason": "group or other cannot read",
			}).Error("fix permissions")
			permissions.SetGroupWrite(true)
			permissions.SetOtherRead(true)
			err = permbits.Chmod(filePath, permissions)
			if err != nil {
				return errors.WithStack(err)
			}
			return nil
		}, nil
	})
}

func resizePngForLinux(inputInfo *InputFileInfo, iconFileName string, outDir string) ([]IconInfo, error) {
	var result []IconInfo
	result = append(result, IconInfo{
		File: iconFileName,
		Size: inputInfo.MaxIconSize,
	})

	sizeList := []int{24, 96}
	for _, item := range icnsTypeToSize {
		if item.Size < inputInfo.MaxIconSize {
			sizeList = append(sizeList, item.Size)
		}
	}

	err := multiResizeImage2(&inputInfo.maxImage, filepath.Join(outDir, "icon_%dx%d.png"), &result, sizeList)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	sortBySize(result)
	return result, nil
}

func convertSingleFile(inputInfo *InputFileInfo, outFile string, outputFormat string) ([]IconInfo, error) {
	switch outputFormat {
	case "icns":
		err := ConvertToIcns(*inputInfo, outFile)
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
		return nil, errors.Errorf("unknown output format %s", outputFormat)
	}
}

func configureInputInfoFromSingleFile(file string, isOutputFormatIco bool, inputInfo *InputFileInfo) error {
	if strings.HasSuffix(file, ".svg") {
		inputInfo.MaxIconSize = 1024
		inputInfo.SizeToPath[inputInfo.MaxIconSize] = file
		return nil
	}

	maxImage, err := loadImage(file, inputInfo.recommendedMinSize)
	if err != nil {
		return errors.WithStack(err)
	}

	if isOutputFormatIco && maxImage.Bounds().Max.X > 256 {
		image256 := imaging.Resize(maxImage, 256, 256, imaging.Lanczos)
		maxImage = image256
	}

	inputInfo.MaxIconSize = maxImage.Bounds().Max.X
	inputInfo.maxImage = maxImage
	inputInfo.SizeToPath[inputInfo.MaxIconSize] = file

	return nil
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
