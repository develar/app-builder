package snap

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/apex/log"
	"github.com/develar/app-builder/pkg/download"
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

type TemplateInfo struct {
	Url    string
	Sha512 string
}

//noinspection SpellCheckingInspection
var electronTemplate = TemplateInfo {
	Url: fmt.Sprintf("https://github.com/electron-userland/electron-builder-binaries/releases/download/%[1]s/%[1]s.7z", "snap-template-0.2.0"),
	Sha512: "2Uxlk/+BkZt5T4CePfi5Cbt35TLlCuO34M5kGaFeT/V1JCx5D6i+EAdMMp1AX9vi6/4zSKW/wB5Z+DZIaHacNg==",
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
		icon:     command.Flag("icon", "The path to the icon.").String(),
		hooksDir: command.Flag("hooks", "The hooks dir.").String(),

		arch: command.Flag("arch", "The arch.").Default("amd64").Enum("amd64", "i386", "armv7l", "arm64"),

		output: command.Flag("output", "The output file.").Short('o').Required().String(),

		dockerImage: command.Flag("docker-image", "The docker image.").Default("snapcore/snapcraft:latest").String(),
	}

	isUseDockerCommandArg := command.Flag("docker", "Whether to use Docker.").Default(isUseDockerDefault).Envar("SNAP_USE_DOCKER").Bool()
	isRemoveStage := util.ConfigureIsRemoveStageParam(command)

	command.Action(func(context *kingpin.ParseContext) error {
		resolvedTemplateDir, err := resolveTemplateDir(*templateDir, *templateUrl, *templateSha512)
		if err != nil {
			return errors.WithStack(err)
		}

		isUseDocker, err := DetectIsUseDocker(*isUseDockerCommandArg, len(resolvedTemplateDir) != 0)
		err = Snap(resolvedTemplateDir, isUseDocker, options)
		if err != nil {
			switch e := errors.Cause(err).(type) {
			case util.MessageError:
				log.Fatal(e.Error())

			default:
				return err
			}
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

func resolveTemplateDir(templateDir string, templateUrl string, templateSha512 string) (string, error) {
	if len(templateDir) != 0 || len(templateUrl) == 0 {
		return templateDir, nil
	}

	var templateInfo TemplateInfo
	if templateUrl == "electron1" {
		templateInfo = electronTemplate
	} else {
		templateInfo = TemplateInfo{
			Url:    templateUrl,
			Sha512: templateSha512,
		}
	}

	result, err := download.DownloadArtifact("", templateInfo.Url, templateInfo.Sha512)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return result, nil
}

func CheckSnapcraftVersion(isRequireToBeInstalled bool) error {
	out, err := exec.Command("snapcraft", "--version").Output()

	var install string
	if runtime.GOOS == "darwin" {
		install = "brew update snapcraft"
	} else {
		install = "sudo snap install snapcraft --classic"
	}

	if err == nil {
		if version.CompareSimple(strings.TrimSpace(string(out)), "2.39.0") == 1 {
			return util.NewMessageError("at least snapcraft 2.39.0 is required, please: "+install, "ERR_SNAPCRAFT_OUTDATED")
		} else {
			return nil
		}
	}

	log.Debug(err.Error())

	if isRequireToBeInstalled {
		return util.NewMessageError("snapcraft is not installed, please: "+install, "ERR_SNAPCRAFT_NOT_INSTALLED")
	} else {
		return nil
	}
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

	err := CheckSnapcraftVersion(false)
	if err != nil {
		return true, errors.WithStack(err)
	} else {
		return false, nil
	}

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
		err := fs.CopyUsingHardlink(iconPath, filepath.Join(snapMetaDir, "gui", "icon"+filepath.Ext(iconPath)))
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

func RemoveAdapter(snapFilePath string) error {
	data, err := ioutil.ReadFile(snapFilePath)
	if err != nil {
		return errors.WithStack(err)
	}

	re, err := regexp.Compile("(?m)[\r\n]+^\\s+adapter: none.*$")
	if err != nil {
		return errors.WithStack(err)
	}

	fixedData := re.ReplaceAll(data, []byte{})
	if len(fixedData) != len(data) {
		err = ioutil.WriteFile(snapFilePath, fixedData, 0)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func buildWithoutDocker(isUseTemplateApp bool, options SnapOptions) error {
	err := CheckSnapcraftVersion(true)
	if err != nil {
		return errors.WithStack(err)
	}

	stageDir := *options.stageDir

	var primeDir string
	if isUseTemplateApp {
		primeDir = stageDir
	} else {
		util.ExecuteWithInheritedStdOutAndStdErr(exec.Command("snapcraft", "prime", "--target-arch", *options.arch), stageDir)
		primeDir = filepath.Join(stageDir, "prime")
		err := cleanUpSnap(primeDir)
		if err != nil {
			return errors.WithStack(err)
		}

		err = RemoveAdapter(filepath.Join(primeDir, "meta", "snap.yaml"))
		if err != nil {
			return errors.WithStack(err)
		}
	}

	err = fs.CopyUsingHardlink(*options.appDir, filepath.Join(primeDir, "app"))
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

	err := util.ExecuteWithInheritedStdOutAndStdErr(exec.Command("docker", "run", "--rm",
		"-v", filepath.Dir(*options.output)+":/out:delegated",
		"--mount", "type=bind,source="+stageDir+",destination=/stage,readonly",
		"--mount", "type=bind,source=" + *options.appDir+",destination=/tmp/final-stage/app,readonly",
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
