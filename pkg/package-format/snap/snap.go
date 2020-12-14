package snap

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/alecthomas/kingpin"
	"github.com/develar/app-builder/pkg/download"
	"github.com/develar/app-builder/pkg/fs"
	"github.com/develar/app-builder/pkg/linuxTools"
	"github.com/develar/app-builder/pkg/log"
	"github.com/develar/app-builder/pkg/util"
	"github.com/develar/errors"
	"github.com/develar/go-fs-util"
	"github.com/mcuadros/go-version"
	"go.uber.org/zap"
)

type TemplateInfo struct {
	Url    string
	Sha512 string
}

// --enable-geoip leads to very slow fetching - it seems local sources are more slow.

type SnapOptions struct {
	appDir         *string
	stageDir       *string
	icon           *string
	hooksDir       *string
	executableName *string

	extraAppArgs     *string
	excludedAppFiles *[]string

	arch   *string
	output *string
}

func ConfigureCommand(app *kingpin.Application) {
	command := app.Command("snap", "Build snap.")

	templateFile := command.Flag("template", "The template file.").Short('t').String()

	templateUrl := command.Flag("template-url", "The template archive URL.").Short('u').String()
	templateSha512 := command.Flag("template-sha512", "The expected sha512 of template archive.").String()

	//noinspection SpellCheckingInspection
	options := SnapOptions{
		appDir:           command.Flag("app", "The app dir.").Short('a').Required().String(),
		stageDir:         command.Flag("stage", "The stage dir.").Short('s').Required().String(),
		icon:             command.Flag("icon", "The path to the icon.").String(),
		hooksDir:         command.Flag("hooks", "The hooks dir.").String(),
		executableName:   command.Flag("executable", "The executable file name to create command wrapper.").String(),
		extraAppArgs:     command.Flag("extraAppArgs", "The extra app launch arguments").String(),
		excludedAppFiles: command.Flag("exclude", "The excluded app files.").Strings(),

		arch: command.Flag("arch", "The arch.").Default("amd64").String(),

		output: command.Flag("output", "The output file.").Short('o').Required().String(),
	}

	isRemoveStage := util.ConfigureIsRemoveStageParam(command)

	command.Action(func(context *kingpin.ParseContext) error {
		resolvedTemplateDir, err := ResolveTemplateDir(*templateFile, *templateUrl, *templateSha512)
		if err != nil {
			return errors.WithStack(err)
		}

		err = Snap(resolvedTemplateDir, options)
		if err != nil {
			switch e := errors.Cause(err).(type) {
			case util.MessageError:
				log.LOG.Fatal(e.Error())

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

//noinspection SpellCheckingInspection
func ResolveTemplateDir(templateFile string, templateUrl string, templateSha512 string) (string, error) {
	if len(templateFile) != 0 || len(templateUrl) == 0 {
		return templateFile, nil
	}

	switch templateUrl {
	case "electron4", "electron4:amd64":
		return download.DownloadArtifact("", download.GetGithubBaseUrl()+"snap-template-4.0-2/snap-template-electron-4.0-2-amd64.tar.7z", "PYhiQQ5KE4ezraLE7TOT2aFPGkBNjHLRN7C8qAPaC6VckHU3H+0m+JA/Wmx683fKUT2ZBwo9Mp82EuhmQo5WOQ==")
	case "electron4:armhf", "electron4:arm":
		return download.DownloadArtifact("", download.GetGithubBaseUrl()+"snap-template-4.0-1/snap-template-electron-4.0-1-armhf.tar.7z", "jK+E0d0kyzBEsFmTEUIsumtikH4XZp8NVs6DBtIBJqXAmVCuNHcmvDa0wcaigk8foU4uGZXsLlJtNj11X100Bg==")
	default:
		return download.DownloadArtifact("", templateUrl, templateSha512)
	}
}

func CheckSnapcraftVersion(isRequireToBeInstalled bool) error {
	out, err := exec.Command("snapcraft", "--version").Output()

	var installMessage string
	if runtime.GOOS == "darwin" {
		installMessage = "brew install snapcraft"
	} else {
		installMessage = "sudo snap install snapcraft --classic"
	}

	if err == nil {
		return doCheckSnapVersion(strings.TrimSpace(string(out)), installMessage)
	}

	log.Debug(err.Error())

	if isRequireToBeInstalled {
		return util.NewMessageError("snapcraft is not installed, please: "+installMessage, "ERR_SNAPCRAFT_NOT_INSTALLED")
	} else {
		return nil
	}
}

func doCheckSnapVersion(rawVersion string, installMessage string) error {
	if rawVersion == "snapcraft, version edge" {
		return nil
	}

	s := strings.TrimSpace(strings.TrimPrefix(rawVersion, "snapcraft"))
	s = strings.TrimSpace(strings.TrimPrefix(s, ","))
	s = strings.TrimSpace(strings.TrimPrefix(s, "version"))
	s = strings.Trim(s, "'")
	if version.Compare(s, "3.1.0", "<") {
		return util.NewMessageError("at least snapcraft 3.1.0 is required, but "+rawVersion+" installed, please: "+installMessage, "ERR_SNAPCRAFT_OUTDATED")
	} else {
		return nil
	}
}

func Snap(templateDir string, options SnapOptions) error {
	stageDir := *options.stageDir
	isUseTemplateApp := len(templateDir) != 0
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

	scriptDir := filepath.Join(stageDir, "scripts")
	err := fsutil.EnsureEmptyDir(scriptDir)
	if err != nil {
		return errors.WithStack(err)
	}

	if len(*options.executableName) != 0 {
		err := writeCommandWrapper(options, isUseTemplateApp, scriptDir)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	chromeSandbox := filepath.Join(*options.appDir, "app", "chrome-sandbox")
	_ = syscall.Unlink(chromeSandbox)

	switch {
	case isUseTemplateApp:
		return buildUsingTemplate(templateDir, options)
	default:
		return buildWithoutTemplate(options, scriptDir)
	}
}

func writeCommandWrapper(options SnapOptions, isUseTemplateApp bool, scriptDir string) error {
	var appPrefix string
	var dir string
	if isUseTemplateApp {
		appPrefix = ""
		dir = *options.stageDir
	} else {
		appPrefix = "app/"
		dir = scriptDir
	}

	commandWrapperFile := filepath.Join(dir, "command.sh")
	text := "#!/bin/bash -e\n" + `exec "$SNAP/desktop-init.sh" "$SNAP/desktop-common.sh" "$SNAP/desktop-gnome-specific.sh" "$SNAP/` + appPrefix + *options.executableName + `" "$@"`

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

func buildUsingTemplate(templateDir string, options SnapOptions) error {
	stageDir := *options.stageDir

	mksquashfsPath, err := linuxTools.GetMksquashfs()
	if err != nil {
		return errors.WithStack(err)
	}

	var args []string

	args, err = linuxTools.ReadDirContentTo(templateDir, args, nil)
	if err != nil {
		return errors.WithStack(err)
	}

	args, err = linuxTools.ReadDirContentTo(stageDir, args, nil)
	if err != nil {
		return errors.WithStack(err)
	}

	// https://github.com/electron-userland/electron-builder/issues/3608
	// even if electron-builder will correctly unset setgid/setuid, still, quite a lot of possibilities for user to create such incorrect permissions,
	// so, just unset it using chmod right before packaging
	dirs := []string{stageDir, *options.appDir, templateDir}
	err = util.MapAsync(len(dirs), func(taskIndex int) (func() error, error) {
		dir := dirs[taskIndex]
		return func() error {
			command := exec.Command("chmod", "-R", "g-s", dir)
			command.Dir = dir
			_, err := util.Execute(command)
			if err != nil {
				log.Warn("cannot execute chmod", zap.Error(err))
			}
			return nil
		}, nil
	})

	if err != nil {
		return errors.WithStack(err)
	}

	args, err = linuxTools.ReadDirContentTo(*options.appDir, args, func(name string) bool {
		if name == "LICENSES.chromium.html" || name == "LICENSE.electron.txt" {
			return false
		}
		return options.excludedAppFiles == nil || !util.ContainsString(*options.excludedAppFiles, name)
	})

	if err != nil {
		return errors.WithStack(err)
	}

	args = append(args, *options.output, "-no-progress", "-quiet", "-noappend", "-comp", "xz", "-no-xattrs", "-no-fragments", "-all-root")

	_, err = util.Execute(exec.Command(mksquashfsPath, args...))
	if err != nil {
		return err
	}
	return nil
}

func buildWithoutTemplate(options SnapOptions, scriptDir string) error {
	err := CheckSnapcraftVersion(true)
	if err != nil {
		return errors.WithStack(err)
	}

	stageDir := *options.stageDir

	for _, name := range AssetNames() {
		if strings.HasPrefix(name, "desktop-scripts/") {
			err := ioutil.WriteFile(filepath.Join(scriptDir, path.Base(name)), MustAsset(name), 0755)
			if err != nil {
				return errors.WithStack(err)
			}
		}
	}

	// multipass cannot access files outside of stage dir
	err = fs.CopyUsingHardlink(*options.appDir, filepath.Join(stageDir, "app"))
	if err != nil {
		return errors.WithStack(err)
	}

	isDestructiveMode := util.IsEnvTrue("SNAP_DESTRUCTIVE_MODE")

	// multipass cannot access files outside of snapcraft command working dir
	var snapEffectiveOutput string
	if isDestructiveMode {
		snapEffectiveOutput = *options.output
	} else {
		snapEffectiveOutput = "out.snap"
	}

	var args []string
	args = append(args, "snap", "--output", snapEffectiveOutput)
	if len(*options.arch) != 0 {
		if *options.arch != runtime.GOARCH {
			return fmt.Errorf("snapcraft does not currently support building %s on %s", *options.arch, runtime.GOARCH)
		}
	}

	if isDestructiveMode {
		args = append(args, "--destructive-mode")
	}

	command := exec.Command("snapcraft", args...)
	command.Env = append(os.Environ(),
		"SNAPCRAFT_HAS_TTY=false",
	)

	command.Dir = stageDir
	err = util.ExecuteAndPipeStdOutAndStdErr(command)
	if err != nil {
		return err
	}

	if !isDestructiveMode {
		err := os.Rename(filepath.Join(stageDir, snapEffectiveOutput), *options.output)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}
