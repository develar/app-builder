// +build !windows

package dmg

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/apex/log"
	"github.com/develar/app-builder/pkg/fs"
	"github.com/develar/app-builder/pkg/util"
	"github.com/develar/errors"
	"github.com/pkg/xattr"
)

func ConfigureCommand(app *kingpin.Application) {
	command := app.Command("dmg", "Build dmg.")

	volumePath := command.Flag("volume", "").Required().String()
	icon := command.Flag("icon", "").String()
	background := command.Flag("background", "").String()

	command.Action(func(context *kingpin.ParseContext) error {
		err := BuildDmg(*volumePath, *icon, *background)
		if err != nil {
			return errors.WithStack(err)
		}

		return nil
	})
}

func BuildDmg(volumePath string, icon string, backgroundPath string) error {
	if icon != "" {
		// cannot use hard link because volume uses different disk
		iconPath := filepath.Join(volumePath, ".VolumeIcon.icns")
		err := fs.CopyDirOrFile(icon, iconPath)
		if err != nil {
			return errors.WithStack(err)
		}

		err = setHasCustomIconAttribute(volumePath)
		if err != nil {
			return errors.WithStack(err)
		}

		err = setIsInvisibleAttribute(iconPath)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	if backgroundPath != "" {
		backgroundPath, err := GetEffectiveBackgroundPath(backgroundPath)
		if err != nil {
			return errors.WithStack(err)
		}

		err = fs.CopyDirOrFile(backgroundPath, filepath.Join(volumePath, ".background", filepath.Base(backgroundPath)))
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func GetEffectiveBackgroundPath(path string) (string, error) {
	if strings.HasSuffix(path, ".tiff") || strings.HasSuffix(path, ".TIFF") {
		return path, nil
	}

	re, err := regexp.Compile("\\.([a-z]+)$")
	if err != nil {
		return "", err
	}

	retinaFile := re.ReplaceAllString(path, "@2x.$1")
	_, err = os.Stat(retinaFile)
	if err != nil {
		if !os.IsNotExist(err) {
			log.WithError(err).Debug("checking retina file")
		}
		return path, nil
	}

	tiffFile, err := util.TempFile("", ".tiff")
	if err != nil {
		return "", err
	}

	err = tiffFile.Close()
	if err != nil {
		return "", err
	}

	err = util.Execute(exec.Command("tiffutil", "-cathidpicheck", path, retinaFile, "-out", tiffFile.Name()), "")
	if err != nil {
		return "", err
	}

	return tiffFile.Name(), nil
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
