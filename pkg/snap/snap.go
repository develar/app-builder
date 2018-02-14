package snap

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/apex/log"
	"github.com/develar/app-builder/download"
	"github.com/develar/app-builder/pkg/fs"
	"github.com/develar/app-builder/pkg/util"
	"github.com/develar/errors"
	"github.com/mcuadros/go-version"
)

// usr/share/fonts is required, cannot run otherwise
var unnecessaryFiles = []string{
	"usr/share/fonts",
	"usr/share/doc",
	"usr/share/man",
	"usr/share/icons",
	"usr/share/bash-completion",
	"usr/share/lintian",
	"usr/share/dh-python",
	"usr/share/python3",

	"usr/lib/python*",
	"usr/bin/python*",
}

// --enable-geoip leads to very slow fetching - it seems local sources are more slow.

type SnapOptions struct {
	appDir   *string
	stageDir *string
	icon     *string
	hooksDir *string

	dockerImage *string

	arch   *string
	output *string
}

func ConfigureCommand(app *kingpin.Application) {
	command := app.Command("snap", "Build snap.")

	templateDir := command.Flag("template", "The template dir.").Short('t').String()

	templateUrl := command.Flag("template-url", "The template archive URL.").Short('u').String()
	templateSha512 := command.Flag("template-sha512", "The expected sha512 of template archive.").String()

	var isUseDockerDefault string
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		isUseDockerDefault = "false"
	} else {
		isUseDockerDefault = "true"
	}

	options := SnapOptions{
		appDir:   command.Flag("app", "The app dir.").Short('a').Required().String(),
		stageDir: command.Flag("stage", "The stage dir.").Short('s').Required().String(),
		icon: command.Flag("icon", "The path to the icon.").String(),
		hooksDir: command.Flag("hooks", "The hooks dir.").String(),

		arch:          command.Flag("arch", "The arch.").Default("amd64").Enum("amd64", "i386", "armv7l", "arm64"),

		output: command.Flag("output", "The output file.").Short('o').Required().String(),

		dockerImage: command.Flag("docker-image", "The docker image.").Default("snapcore/snapcraft:latest").String(),
	}

	isUseDockerCommandArg := command.Flag("docker", "Whether to use Docker.").Default(isUseDockerDefault).Envar("SNAP_USE_DOCKER").Bool()
	isRemoveStage := command.Flag("remove-stage", "Whether to remove stage after build.").Default("true").Bool()

	command.Action(func(context *kingpin.ParseContext) error {
		resolvedTemplateDir := *templateDir
		if len(resolvedTemplateDir) == 0 && len(*templateUrl) != 0 {
			var err error
			resolvedTemplateDir, err = download.DownloadArtifact("", *templateUrl, *templateSha512)
			if err != nil {
				return errors.WithStack(err)
			}
		}

		isUseDocker, err := DetectIsUseDocker(*isUseDockerCommandArg, len(	resolvedTemplateDir) != 0)
		err = Snap(resolvedTemplateDir, isUseDocker, options)
		if err != nil {
			return errors.WithStack(err)
		}

		if *isRemoveStage {
			err = os.RemoveAll(*options.stageDir)
			if err != nil {
				return errors.WithStack(err)
			}
		}

		return nil
	})
}

func DetectIsUseDocker(isUseDocker bool, isUseTemplateApp bool) (bool, error) {
	if isUseDocker {
		return true, nil
	}

	if runtime.GOOS != "darwin" {
		return isUseDocker, nil
	}

	if !isUseTemplateApp {
		return true, nil
	}

	out, err := exec.Command("snapcraft", "--version").Output()
	if err == nil {
		if version.CompareSimple(strings.TrimSpace(string(out)), "2.39.0") == 1 {
			return true, errors.Errorf("at least snapcraft 2.39.0 is required, please 'brew update snapcraft'")
		} else {
			return false, nil
		}
	}

	log.WithError(err).Debug("snapcraft not installed")
	log.WithFields(log.Fields{
		"reason":   "snapcraft not installed",
		"solution": "brew install snapcraft",
	}).Warn("docker is used to build snap")

	return true, nil
}

func Snap(templateDir string, isUseDocker bool, options SnapOptions) error {
	stageDir := *options.stageDir
	isUseTemplateApp := len(templateDir) != 0
	if isUseTemplateApp {
		err := fs.CopyUsingHardlink(templateDir, stageDir)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	var snapMetaDir string
	if isUseTemplateApp {
		snapMetaDir = filepath.Join(stageDir, "meta")
	} else {
		snapMetaDir = filepath.Join(stageDir, "snap")
	}

	iconPath := *options.icon
	if len(iconPath) != 0 {
		err := fs.CopyUsingHardlink(iconPath, filepath.Join(snapMetaDir, "gui", "icon" + filepath.Ext(iconPath)))
		if err != nil {
			return errors.WithStack(err)
		}
	}

	if len(*options.hooksDir) != 0 {
		err := fs.CopyUsingHardlink(*options.hooksDir, filepath.Join(snapMetaDir, "hooks"))
		if err != nil {
			return errors.WithStack(err)
		}
	}

	if isUseDocker {
		return buildUsingDocker(isUseTemplateApp, options)
	} else {
		return buildWithoutDocker(isUseTemplateApp, options)
	}
}

func buildWithoutDocker(isUseTemplateApp bool, options SnapOptions) error {
	stageDir := *options.stageDir

	var primeDir string
	if isUseTemplateApp {
		primeDir = stageDir
	} else {
		util.Execute(exec.Command("snapcraft", "prime", "--target-arch", *options.arch), stageDir)
		primeDir = filepath.Join(stageDir, "prime")
		err := cleanUpSnap(primeDir)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	err := fs.CopyUsingHardlink(*options.appDir, filepath.Join(primeDir, "app"))
	if err != nil {
		return errors.WithStack(err)
	}

	err = util.Execute(exec.Command("snapcraft", "pack", primeDir, "--output", *options.output), stageDir)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func buildUsingDocker(isUseTemplateApp bool, options SnapOptions) error {
	stageDir := *options.stageDir

	if isUseTemplateApp {
		err := util.Execute(exec.Command("docker", "run", "--rm",
			"-v", filepath.Dir(*options.output)+":/out:delegated",
			// cannot be "ro" because we mount stage/app, so, "delegated"
			"-v", stageDir+":/stage:delegated",
			"-v", *options.appDir+":/stage/app:ro",
			*options.dockerImage,
			"snapcraft", "pack", "/stage", "--output", "/out/"+filepath.Base(*options.output),
		), stageDir)
		return errors.WithStack(err)
	}

	var commands []string
	// copy stage to linux fs to avoid performance issues (https://docs.docker.com/docker-for-mac/osxfs-caching/)
	commands = append(commands,
		"cp -r /stage /s/",
		"cd /s",
		"snapcraft prime --target-arch " + *options.arch,
		"rm -rf prime/"+strings.Join(unnecessaryFiles, " prime/"),
		"mv /s/prime/* /tmp/final-stage/",
		"snapcraft pack /tmp/final-stage --output /out/"+filepath.Base(*options.output),
	)

	log.WithField("command", strings.Join(commands, "\n")).Debug("build snap using docker")

	err := util.Execute(exec.Command("docker", "run", "--rm",
		"-v", filepath.Dir(*options.output)+":/out:delegated",
		"-v", stageDir+":/stage:ro",
		"-v", *options.appDir+":/tmp/final-stage/app:ro",
		*options.dockerImage,
		"/bin/bash", "-c", strings.Join(commands, " && "),
	), stageDir)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func cleanUpSnap(dir string) (error) {
	return util.MapAsync(len(unnecessaryFiles), func(taskIndex int) (func() error, error) {
		file := filepath.Join(dir, unnecessaryFiles[taskIndex])
		return func() error {
			err := fs.RemoveByGlob(file)
			if err != nil {
				return errors.WithStack(err)
			}
			return nil
		}, nil
	})
}