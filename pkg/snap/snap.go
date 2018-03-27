package snap

import (
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
	"github.com/develar/app-builder/pkg/linuxTools"
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
	Url: "https://github.com/electron-userland/electron-builder-binaries/releases/download/snap-template-1.1/electron-template-1.1.snap",
	Sha512: "Lk5jCYr+iNJBwhVMryie9WdZ6kwmd0XGL017DHW9AEKEpNQQpiW+CfKzDExRyfzxgmIGo944vqvnqD3Okn17jg==",
}

//noinspection SpellCheckingInspection
var electronTemplate2 = TemplateInfo {
	Url: "https://github.com/electron-userland/electron-builder-binaries/releases/download/snap-template-2.1/electron-template-2.1.snap",
	Sha512: "ITpJRtuy3QuJxGfcAD+ogCXHt3B9UsDeXqrC86elncEp5gCvfIxZ23deQFQRzkD0dWTp894z9PuiuYfProWSCA==",
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

	templateFile := command.Flag("template", "The template file.").Short('t').String()

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
		resolvedTemplateFile, err := resolveTemplateFile(*templateFile, *templateUrl, *templateSha512)
		if err != nil {
			return errors.WithStack(err)
		}

		isUseDocker, err := DetectIsUseDocker(*isUseDockerCommandArg, len(resolvedTemplateFile) != 0)
		err = Snap(resolvedTemplateFile, isUseDocker, options)
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

func resolveTemplateFile(templateFile string, templateUrl string, templateSha512 string) (string, error) {
	if len(templateFile) != 0 || len(templateUrl) == 0 {
		return templateFile, nil
	}

	var templateInfo TemplateInfo
	if templateUrl == "electron1" {
		templateInfo = electronTemplate
	} else if templateUrl == "electron2" {
		templateInfo = electronTemplate2
	} else {
		templateInfo = TemplateInfo{
			Url:    templateUrl,
			Sha512: templateSha512,
		}
	}

	result, err := download.DownloadCompressedArtifact("snap-templates", templateInfo.Url, templateInfo.Sha512)
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

	return !isUseTemplateApp, nil
}

func Snap(templateFile string, isUseDocker bool, options SnapOptions) error {
	stageDir := *options.stageDir
	isUseTemplateApp := len(templateFile) != 0
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

	if len(templateFile) != 0 {
		return buildWithoutDockerUsingTemplate(templateFile, options)
	} else if isUseDocker {
		return buildUsingDocker(options)
	} else {
		return buildWithoutDockerAndWithoutTemplate(options)
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

func buildWithoutDockerUsingTemplate(templateFile string, options SnapOptions) error {
	err := fs.CopyDirOrFile(templateFile, *options.output)
	if err != nil {
		return errors.WithStack(err)
	}

	mksquashfsPath, err := linuxTools.GetMksquashfs()
	if err != nil {
		return errors.WithStack(err)
	}

	// will be not merged into root if pass several source dirs, so, call for each source dir
	for _, sourceDir := range []string{*options.stageDir, *options.appDir} {
		err = util.Execute(exec.Command(mksquashfsPath, sourceDir, *options.output, "-no-progress", "-quiet", "-all-root", "-no-duplicates", "-no-recovery"), "")
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func buildWithoutDockerAndWithoutTemplate(options SnapOptions) error {
	stageDir := *options.stageDir

	var primeDir string
	err := CheckSnapcraftVersion(true)
	if err != nil {
		return errors.WithStack(err)
	}

	err = util.ExecuteWithInheritedStdOutAndStdErr(exec.Command("snapcraft", "prime", "--target-arch", *options.arch), stageDir)
	if err != nil {
		return errors.WithStack(err)
	}

	primeDir = filepath.Join(stageDir, "prime")
	err = cleanUpSnap(primeDir)
	if err != nil {
		return errors.WithStack(err)
	}

	err = RemoveAdapter(filepath.Join(primeDir, "meta", "snap.yaml"))
	if err != nil {
		return errors.WithStack(err)
	}

	err = fs.CopyUsingHardlink(*options.appDir, filepath.Join(primeDir, "app"))
	if err != nil {
		return errors.WithStack(err)
	}

	err = fs.CopyUsingHardlink(filepath.Join(stageDir, "command.sh"), filepath.Join(primeDir, "command.sh"))
	if err != nil {
		return errors.WithStack(err)
	}

	err = util.Execute(exec.Command("snapcraft", "pack", primeDir, "--output", *options.output), stageDir)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func buildUsingDocker(options SnapOptions) error {
	var commands []string
	// copy stage to linux fs to avoid performance issues (https://docs.docker.com/docker-for-mac/osxfs-caching/)
	commands = append(commands,
		"cp -r /stage /s/",
		"cd /s",
		"snapcraft prime --target-arch " + *options.arch,
		"rm -rf prime/"+strings.Join(unnecessaryFiles, " prime/"),
		"mv /s/prime/* /tmp/final-stage/",
		"mv /s/command.sh /tmp/final-stage/command.sh",
		"sed -i '/adapter: none/d' /tmp/final-stage/meta/snap.yaml",
		"snapcraft pack /tmp/final-stage --output /out/"+filepath.Base(*options.output),
	)

	log.WithField("command", strings.Join(commands, "\n")).Debug("build snap using docker")

	stageDir := *options.stageDir
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
