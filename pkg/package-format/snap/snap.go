package snap

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"syscall"

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
var electronTemplate2 = TemplateInfo{
	Url:    "https://github.com/electron-userland/electron-builder-binaries/releases/download/snap-template-2.4/snap-template-electron-2.4.tar.7z",
	Sha512: "njelQ3fVOUEa4DoUsxmuTifrnQ51hvt4OIAfiQ1zQkqY4JpnjxE0GG/+8Jc3m+lA7fNH0uBO8pxfNTJMD5UHsA==",
}

// --enable-geoip leads to very slow fetching - it seems local sources are more slow.

type SnapOptions struct {
	appDir         *string
	stageDir       *string
	icon           *string
	hooksDir       *string
	executableName *string

	extraAppArgs *string

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

	//noinspection SpellCheckingInspection
	options := SnapOptions{
		appDir:         command.Flag("app", "The app dir.").Short('a').Required().String(),
		stageDir:       command.Flag("stage", "The stage dir.").Short('s').Required().String(),
		icon:           command.Flag("icon", "The path to the icon.").String(),
		hooksDir:       command.Flag("hooks", "The hooks dir.").String(),
		executableName: command.Flag("executable", "The executable file name to create command wrapper.").String(),
		extraAppArgs:   command.Flag("extraAppArgs", "").String(),

		arch: command.Flag("arch", "The arch.").Default("amd64").Enum("amd64", "i386", "armv7l", "arm64"),

		output: command.Flag("output", "The output file.").Short('o').Required().String(),

		dockerImage: command.Flag("docker-image", "The docker image.").Default("snapcore/snapcraft:latest").String(),
	}

	isUseDockerCommandArg := command.Flag("docker", "Whether to use Docker.").Default(isUseDockerDefault).Envar("SNAP_USE_DOCKER").Bool()
	isRemoveStage := util.ConfigureIsRemoveStageParam(command)

	command.Action(func(context *kingpin.ParseContext) error {
		resolvedTemplateFile, err := ResolveTemplateFile(*templateFile, *templateUrl, *templateSha512)
		if err != nil {
			return errors.WithStack(err)
		}

		isUseDocker := DetectIsUseDocker(*isUseDockerCommandArg, len(resolvedTemplateFile) != 0)
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

func ResolveTemplateFile(templateFile string, templateUrl string, templateSha512 string) (string, error) {
	if len(templateFile) != 0 || len(templateUrl) == 0 {
		return templateFile, nil
	}

	var templateInfo TemplateInfo
	if templateUrl == "electron2" {
		templateInfo = electronTemplate2
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

	var installMessage string
	if runtime.GOOS == "darwin" {
		installMessage = "brew update snapcraft"
	} else {
		installMessage = "sudo snap install snapcraft --classic"
	}

	if err == nil {
		return doCheckSnapVersion(string(out), installMessage)
	}

	log.Debug(err.Error())

	if isRequireToBeInstalled {
		return util.NewMessageError("snapcraft is not installed, please: "+installMessage, "ERR_SNAPCRAFT_NOT_INSTALLED")
	} else {
		return nil
	}
}

func doCheckSnapVersion(rawVersion string, installMessage string) error {
	s := strings.TrimSpace(rawVersion)
	s = strings.TrimSpace(strings.TrimPrefix(s, "snapcraft"))
	s = strings.TrimSpace(strings.TrimPrefix(s, ","))
	s = strings.TrimSpace(strings.TrimPrefix(s, "version"))
	if version.Compare(s, "3.1.0", "<") {
		return util.NewMessageError("at least snapcraft 3.1.0 is required, but "+rawVersion+" installed, please: "+installMessage, "ERR_SNAPCRAFT_OUTDATED")
	} else {
		return nil
	}
}

func DetectIsUseDocker(isUseDocker bool, isUseTemplateApp bool) bool {
	if isUseDocker {
		return true
	}

	//if util.IsEnvTrue("USE_SNAPCRAFT") {
	//	return false
	//}

	if runtime.GOOS != "darwin" {
		return isUseDocker
	}

	return !isUseTemplateApp
	//return !isUseTemplateApp
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

	if len(*options.executableName) != 0 {
		err := writeCommandWrapper(options, isUseTemplateApp)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	chromeSandbox := filepath.Join(*options.stageDir, "app", "chrome-sandbox")
	_ = syscall.Unlink(chromeSandbox)

	switch {
	case isUseTemplateApp:
		return buildWithoutDockerUsingTemplate(templateFile, options)
	case isUseDocker:
		return buildUsingDocker(options)
	default:
		return buildWithoutDockerAndWithoutTemplate(options)
	}
}

func writeCommandWrapper(options SnapOptions, isUseTemplateApp bool) error {
	var appPrefix string
	if isUseTemplateApp {
		appPrefix = ""
	} else {
		appPrefix = "app/"
	}

	commandWrapperFile := filepath.Join(*options.stageDir, "command.sh")
	text := "#!/bin/bash\nexec $SNAP/bin/desktop-launch \"$SNAP/" + appPrefix + *options.executableName + `"`
	extraAppArgs := *options.extraAppArgs
	if extraAppArgs != "" {
		text += " " + extraAppArgs
	}
	err := ioutil.WriteFile(commandWrapperFile, []byte(text), 0755)
	if err != nil {
		return errors.WithStack(err)
	}

	err = os.Chmod(commandWrapperFile, 0755)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func RemoveAdapter(snapFilePath string) error {
	data, err := ioutil.ReadFile(snapFilePath)
	if err != nil {
		return errors.WithStack(err)
	}

	re := regexp.MustCompile("(?m)[\r\n]+^\\s+adapter: none.*$")

	fixedData := re.ReplaceAll(data, []byte{})
	if len(fixedData) != len(data) {
		err = ioutil.WriteFile(snapFilePath, fixedData, 0666)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func buildWithoutDockerUsingTemplate(templateFile string, options SnapOptions) error {
	stageDir := *options.stageDir

	mksquashfsPath, err := linuxTools.GetMksquashfs()
	if err != nil {
		return errors.WithStack(err)
	}

	var args []string

	args, err = linuxTools.ReadDirContentTo(templateFile, args)
	if err != nil {
		return errors.WithStack(err)
	}

	args, err = linuxTools.ReadDirContentTo(stageDir, args)
	if err != nil {
		return errors.WithStack(err)
	}

	args, err = linuxTools.ReadDirContentTo(*options.appDir, args)
	if err != nil {
		return errors.WithStack(err)
	}

	args = append(args, *options.output, "-no-progress", "-quiet", "-noappend", "-comp", "xz", "-no-xattrs", "-no-fragments", "-all-root")

	_, err = util.Execute(exec.Command(mksquashfsPath, args...), "")
	if err != nil {
		return errors.WithStack(err)
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

	_, err = util.Execute(exec.Command("snapcraft", "pack", primeDir, "--output", *options.output), stageDir)
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
		"snapcraft prime --target-arch "+*options.arch,
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
		"--mount", "type=bind,source="+*options.appDir+",destination=/tmp/final-stage/app,readonly",
		*options.dockerImage,
		"/bin/bash", "-c", strings.Join(commands, " && "),
	), stageDir)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func cleanUpSnap(dir string) error {
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
