package rcedit

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/alecthomas/kingpin"
	"github.com/develar/app-builder/pkg/download"
	"github.com/develar/app-builder/pkg/util"
	"github.com/develar/app-builder/pkg/wine"
)

func ConfigureCommand(app *kingpin.Application) {
	command := app.Command("rcedit", "")
	configuration := command.Flag("args", "").Required().String()

	command.Action(func(context *kingpin.ParseContext) error {
		var rcEditArgs []string
		err := util.DecodeBase64IfNeeded(*configuration, &rcEditArgs)
		if err != nil {
			return err
		}
		return editResources(rcEditArgs)
	})
}

func editResources(args []string) error {
	winCodeSignPath, err := download.DownloadWinCodeSign()
	if err != nil {
		return err
	}

	if util.GetCurrentOs() == util.WINDOWS || util.IsWSL() {
		var rcEditExecutable string
		if runtime.GOARCH == "amd64" {
			rcEditExecutable = "rcedit-x64.exe"
		} else {
			rcEditExecutable = "rcedit-ia32.exe"
		}

		rcEditPath := filepath.Join(winCodeSignPath, rcEditExecutable)

		if util.IsWSL() {
			err = os.Chmod(rcEditPath, 0755)
			if err != nil {
				return err
			}
		}

		command := exec.Command(rcEditPath, args...)
		_, err = util.Execute(command)
		if err != nil {
			return err
		}

		return nil
	}

	err = wine.ExecWine(filepath.Join(winCodeSignPath, "rcedit-ia32.exe"), filepath.Join(winCodeSignPath, "rcedit-x64.exe"), args)
	if err != nil {
		return err
	}
	return nil
}
