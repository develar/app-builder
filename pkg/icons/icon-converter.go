package icons

import (
	"image"
	"path/filepath"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/develar/app-builder/pkg/fs"
	"github.com/develar/app-builder/pkg/log"
	"github.com/develar/app-builder/pkg/util"
	"github.com/develar/errors"
	"github.com/disintegration/imaging"
	"go.uber.org/zap"
)

func ConfigureCommand(app *kingpin.Application) error {
	command := app.Command("icon", "create ICNS or ICO or icon set from PNG files")

	configuration := &IconConvertRequest{
		Sources:         command.Flag("input", "input source file or directory").Short('i').Strings(),
		FallbackSources: command.Flag("fallback-input", "fallback source file or directory").Strings(),
		Roots:           command.Flag("root", "base directory to resolve relative path").Strings(),
	}

	iconOutFormat := command.Flag("format", "output format").Short('f').Required().Enum("icns", "ico", "set")
	outDir := command.Flag("out", "output directory").Required().String()

	command.Action(func(context *kingpin.ParseContext) error {
		configuration.OutputFormat = *iconOutFormat
		configuration.OutputDir = *outDir

		result, err := ConvertIcon(configuration)
		if err != nil {
			switch t := errors.Cause(err).(type) {
			case *ImageSizeError:
				log.Debug("cannot convert icon", zap.Error(err))
				return writeUserError(t)

			case *ImageFormatError:
				log.Debug("cannot convert icon", zap.Error(err))
				return writeUserError(t)

			default:
				return err
			}
		}

		return util.WriteJsonToStdOut(result)
	})

	return nil
}

func ConvertIcon(configuration *IconConvertRequest) (*IconConvertResult, error) {
	result, err := doConvertIcon(createCommonIconSources(*configuration.Sources, configuration.OutputFormat), *configuration.Roots, configuration.OutputFormat, configuration.OutputDir)
	if err != nil {
		return nil, err
	}

	isFallback := false

	// try using fallback sources
	if result == nil {
		log.Debug("no icons found, using provided fallback sources")
		result, err = doConvertIcon(*configuration.FallbackSources, *configuration.Roots, configuration.OutputFormat, configuration.OutputDir)
		if err != nil {
			return nil, err
		}

		isFallback = true
	}

	return &IconConvertResult{Icons: result, IsFallback: isFallback}, nil
}

func isFileHasImageFormatExtension(name string, outputFormat string) bool {
	return strings.HasSuffix(name, "."+outputFormat) || strings.HasSuffix(name, ".png") || strings.HasSuffix(name, ".ico") || strings.HasSuffix(name, ".svg") || strings.HasSuffix(name, ".icns")
}

func createCommonIconSources(sources []string, outputFormat string) []string {
	var result []string

	for _, source := range sources {
		// do not use filepath.Ext to ensure that dot can be used in filename
		if isFileHasImageFormatExtension(source, outputFormat) {
			result = append(result, source)
		} else {
			result = appendImageVariants(source, source, outputFormat, result)
		}
	}

	result = appendImageVariants("icon", "icons", outputFormat, result)
	return result
}

func appendImageVariants(nameWithoutExt string, nameForSetWithoutExt string, outputFormat string, list []string) []string {
	if outputFormat != "set" {
		list = append(list, nameWithoutExt+"."+outputFormat)
	}

	list = append(list, nameForSetWithoutExt)

	if outputFormat != "png" {
		list = append(list, nameWithoutExt+".png")
	}
	if outputFormat != "icns" {
		list = append(list, nameWithoutExt+".icns")
		// ico only for non icns
		if outputFormat != "ico" {
			list = append(list, nameWithoutExt+".ico")
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

func doConvertIcon(sourceFiles []string, roots []string, outputFormat string, outDir string) ([]IconInfo, error) {
	// allowed to specify path to icns without extension, so, if file not resolved, try to add ".icns" extension
	outExt := outputFormatToSingleFileExtension(outputFormat)
	resolvedPath, fileInfo, err := resolveSourceFile(sourceFiles, roots)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if resolvedPath == "" {
		return nil, nil
	}

	log.Debug("path resolved", zap.String("path", resolvedPath), zap.String("outputFormat", outputFormat))

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
			return fs.SetNormalFilePermissions(filePath)
		}, nil
	})
}

func resizePngForLinux(inputInfo *InputFileInfo, iconFileName string, outDir string) ([]IconInfo, error) {
	var result []IconInfo
	result = append(result, IconInfo{
		File: iconFileName,
		Size: inputInfo.MaxIconSize,
	})

	var sizeList []int
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
