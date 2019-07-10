package snap

import (
	"os/exec"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/develar/app-builder/pkg/util"
)

func ConfigurePublishCommand(app *kingpin.Application) {
	command := app.Command("publish-snap", "Publish snap.")

	file := command.Flag("file", "").Short('f').String()
	channel := command.Flag("channel", "").Short('c').Strings()

	command.Action(func(context *kingpin.ParseContext) error {
		return publishToStore(*file, *channel)
	})
}

func publishToStore(file string, channels []string) error {
	args := []string{"push", file}
	if len(channels) != 0 {
		args = append(args, "--release")
		args = append(args, strings.Join(channels, ","))
	}

	err := CheckSnapcraftVersion(true)
	if err != nil {
		return err
	}

	command := exec.Command("snapcraft", args...)
	err = util.ExecuteAndPipeStdOutAndStdErr(command)
	if err != nil {
		return err
	}

	return nil
}
