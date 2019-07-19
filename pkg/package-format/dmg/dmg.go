// +build !windows

package dmg

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/develar/app-builder/pkg/fs"
	"github.com/develar/app-builder/pkg/log"
	"github.com/develar/app-builder/pkg/util"
	"github.com/develar/errors"
	"github.com/json-iterator/go"
	"github.com/pkg/xattr"
	"go.uber.org/zap"
)

func ConfigureCommand(app *kingpin.Application) {
	command := app.Command("dmg", "Build dmg.")

	volumePath := command.Flag("volume", "").Required().String()
	icon := command.Flag("icon", "").String()
	background := command.Flag("background", "").String()

	command.Action(func(context *kingpin.ParseContext) error {
		backgroundFileInImage, err := BuildDmg(*volumePath, *icon, *background)
		if err != nil {
			return err
		}

		jsonWriter := jsoniter.NewStream(jsoniter.ConfigFastest, os.Stdout, 32*1024)
		jsonWriter.WriteObjectStart()

		if *background != "" {
			pixelWidth, pixelHeight, err := getImageSizeUsingSips(*background)
			if err != nil {
				return err
			}

			jsonWriter.WriteObjectField("backgroundWidth")
			jsonWriter.WriteInt(pixelWidth)
			jsonWriter.WriteMore()
			jsonWriter.WriteObjectField("backgroundHeight")
			jsonWriter.WriteInt(pixelHeight)

			jsonWriter.WriteMore()
			jsonWriter.WriteObjectField("backgroundFile")
			jsonWriter.WriteString(backgroundFileInImage)
		}

		jsonWriter.WriteObjectEnd()
		err = jsonWriter.Flush()
		if err != nil {
			return err
		}

		return nil
	})
}

func getImageSizeUsingSips(background string) (int, int, error) {
	command := exec.Command("sips", "-g", "pixelHeight", "-g", "pixelWidth", background)
	result, err := util.Execute(command)
	if err != nil {
		return 0, 0, err
	}

	pixelWidth := 0
	pixelHeight := 0
	re := regexp.MustCompile(`([a-zA-Z]+):\s*(\d+)`)
	lines := bytes.Split(result, []byte("\n"))
	for _, value := range lines {
		if len(value) == 0 {
			continue
		}

		nameAndValue := re.FindStringSubmatch(string(value))
		if nameAndValue == nil {
			continue
		}

		size, err := strconv.Atoi(nameAndValue[2])
		if err != nil {
			return 0, 0, errors.WithStack(err)
		}

		switch nameAndValue[1] {
		case "pixelWidth":
			pixelWidth = size
		case "pixelHeight":
			pixelHeight = size
		}
	}
	return pixelWidth, pixelHeight, nil
}

func BuildDmg(volumePath string, icon string, backgroundPath string) (string, error) {
	if icon != "" {
		// cannot use hard link because volume uses different disk
		iconPath := filepath.Join(volumePath, ".VolumeIcon.icns")
		err := fs.CopyDirOrFile(icon, iconPath)
		if err != nil {
			return "", errors.WithStack(err)
		}

		err = setHasCustomIconAttribute(volumePath)
		if err != nil {
			return "", errors.WithStack(err)
		}

		err = setIsInvisibleAttribute(iconPath)
		if err != nil {
			return "", errors.WithStack(err)
		}
	}

	backgroundFileInImage := ""
	if backgroundPath != "" {
		backgroundPath, err := GetEffectiveBackgroundPath(backgroundPath)
		if err != nil {
			return "", err
		}

		backgroundFileInImage = filepath.Join(volumePath, ".background", filepath.Base(backgroundPath))
		err = fs.CopyDirOrFile(backgroundPath, backgroundFileInImage)
		if err != nil {
			return "", errors.WithStack(err)
		}
	}
	return backgroundFileInImage, nil
}

func GetEffectiveBackgroundPath(path string) (string, error) {
	if strings.HasSuffix(path, ".tiff") || strings.HasSuffix(path, ".TIFF") {
		return path, nil
	}

	re := regexp.MustCompile(`\.([a-z]+)$`)
	retinaFile := re.ReplaceAllString(path, "@2x.$1")
	_, err := os.Stat(retinaFile)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Debug("checking retina file", zap.Error(err))
		}
		return path, nil
	}

	tiffFile, err := util.TempFile("", ".tiff")
	if err != nil {
		return "", err
	}

	//noinspection SpellCheckingInspection
	_, err = util.Execute(exec.Command("tiffutil", "-cathidpicheck", path, retinaFile, "-out", tiffFile))
	if err != nil {
		return "", err
	}

	return tiffFile, nil
}

func setHasCustomIconAttribute(path string) error {
	data := make([]byte, 32)
	// kHasCustomIcon
	data[8] = 4
	return xattr.Set(path, "com.apple.FinderInfo", data)
}

func setIsInvisibleAttribute(path string) error {
	data := make([]byte, 32)
	data[0] = 'i'
	data[1] = 'c'
	data[2] = 'n'
	data[3] = 's'

	// kIsInvisible
	data[8] = 0x40
	return xattr.Set(path, "com.apple.FinderInfo", data)
}
