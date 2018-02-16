// +build !windows

package dmg

import (
	"path/filepath"

	"github.com/alecthomas/kingpin"
	"github.com/develar/app-builder/pkg/fs"
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

func BuildDmg(volumePath string, icon string, background string) error {
	var fileCopier fs.FileCopier
	if icon != "" {
		// cannot use hard link because volume uses different disk
		iconPath := filepath.Join(volumePath, ".VolumeIcon.icns")
		err := fileCopier.CopyDirOrFile(icon, iconPath)
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

	if background != "" {
		err := fileCopier.CopyDirOrFile(icon, filepath.Join(volumePath, ".background", filepath.Base(background)))
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
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
