package node_modules

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/develar/app-builder/pkg/log"
	"github.com/develar/app-builder/pkg/util"
	"github.com/json-iterator/go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type RebuildConfiguration struct {
	DependencyTreeInfo []DependencyList `json:"dependencies"`

	Platform string `json:"platform"`
	Arch     string `json:"arch"`
	BuildFromSource bool `json:"buildFromSource"`

	NodeExecPath string `json:"nodeExecPath"`

	AdditionalArgs []string `json:"additionalArgs"`
}

type DependencyList struct {
	Dir          string    `json:"dir"`
	Dependencies []DepInfo `json:"deps"`
}

type DepInfo struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	Optional bool   `json:"optional"`
	HasPrebuildInstall bool   `json:"hasPrebuildInstall"`

	parentDir string
	dir string
}

func ConfigureRebuildCommand(app *kingpin.Application) {
	command := app.Command("rebuild-node-modules", "")
	command.Action(func(context *kingpin.ParseContext) error {
		var configuration RebuildConfiguration
		err := jsoniter.NewDecoder(os.Stdin).Decode(&configuration)
		if err != nil {
			return err
		}

		err = rebuild(&configuration)
		if err != nil {
			return err
		}
		return nil
	})
}

func rebuild(configuration *RebuildConfiguration) error {
	dependencies, err := computeNativeDependencies(configuration)
	if err != nil {
		return err
	}

	if len(dependencies) == 0 {
		log.Debug("no native dependencies")
		return nil
	}

	log.Info("rebuilding native dependencies",
		zap.Array("dependencies", zapcore.ArrayMarshalerFunc(func(encoder zapcore.ArrayEncoder) error {
			for index, item := range dependencies {
				if index != 0 {
					encoder.AppendString(", ")
				}
				encoder.AppendString(item.Name)
				encoder.AppendString("@")
				encoder.AppendString(item.Version)
			}
			return nil
		})),
		zap.String("platform", configuration.Platform),
		zap.String("arch", configuration.Arch),
	)

	dependencies, err = installUsingPrebuild(dependencies, configuration)
	if err != nil {
		return err
	}

	if len(dependencies) == 0 {
		log.Debug("all native deps were installed using prebuild-install")
		return nil
	}

	execPath, execArgs, isRunningYarn, err := computeExecPath(configuration)
	if err != nil {
		return fmt.Errorf("Could not compute exec path: %w", err)
	}
	if isRunningYarn {
		err := rebuildUsingYarn(dependencies, execPath, execArgs, configuration)
		if err != nil {
			return err
		}
	} else {
		execArgs = append(execArgs, "rebuild")
		if log.IsDebugEnabled() {
			execArgs = append(execArgs, "--verbose")
		}
		if configuration.AdditionalArgs != nil {
			execArgs = append(execArgs, configuration.AdditionalArgs...)
		}

		for _, item := range dependencies {
			execArgs = append(execArgs, item.Name+"@"+item.Version)
		}

		command := exec.Command(execPath, execArgs...)
		_, err := util.Execute(command)
		if err != nil {
			return err
		}
	}

	return nil
}

func rebuildUsingYarn(dependencies []*DepInfo, execPath string, execArgs []string, configuration *RebuildConfiguration) error {
	execArgs = append(execArgs, "run", "install")
	if configuration.AdditionalArgs != nil {
		execArgs = append(execArgs, configuration.AdditionalArgs...)
	}

	err := util.MapAsyncConcurrency(len(dependencies), getRebuildConcurrency(), func(index int) (func() error, error) {
		dependency := dependencies[index]
		if dependency == nil {
			return nil, nil
		}

		return func() error {
			logger := log.LOG.With(zap.String("name", dependency.Name), zap.String("version", dependency.Version))
			logger.Info("rebuilding native dependency")

			command := exec.Command(execPath, execArgs...)
			command.Dir = dependency.dir
			_, err := util.Execute(command)
			if err != nil {
				if dependency.Optional {
					execError, _ := err.(*util.ExecError)
					logger.Warn("cannot build optional native dependency", util.CreateExecErrorLogEntry(execError)...)
				} else {
					return err
				}
			}

			return nil
		}, nil
	})
	return err
}

func getRebuildConcurrency() int {
	if util.GetCurrentOs() == util.WINDOWS {
		return 1
	} else {
		return 2
	}
}

func installUsingPrebuild(dependencies []*DepInfo, configuration *RebuildConfiguration) ([]*DepInfo, error) {
	isRebuildPossible := checkRebuildPossible(configuration)
	if configuration.BuildFromSource {
		if isRebuildPossible {
			return dependencies, nil
		}

		log.Warn("buildFromSource option is ignored",
			zap.String("reason", "platform or arch not compatible"),
			zap.String("platform", configuration.Platform),
			zap.String("arch", configuration.Arch),
		)
	}

	err := util.MapAsyncConcurrency(len(dependencies), getRebuildConcurrency(), func(index int) (func() error, error) {
		dependency := dependencies[index]
		if !dependency.HasPrebuildInstall {
			return nil, nil
		}

		return func() error {
			logger := log.LOG.With(zap.String("name", dependency.Name), zap.String("version", dependency.Version), zap.String("platform", configuration.Platform), zap.String("arch", configuration.Arch),)
			logger.Info("install prebuilt binary")

			parentDir := dependency.parentDir
			bin := filepath.Join(parentDir, "prebuild-install", "bin.js")
			for {
				_, err := os.Stat(bin)
				if err == nil {
					break
				}

				parentDir, err = findNearestNodeModuleDir(filepath.Dir(filepath.Dir(parentDir)))
				if err != nil {
					return err
				}
				if len(parentDir) == 0 {
					log.Error("cannot find prebuild-install")
					return nil
				}
				bin = filepath.Join(parentDir, "prebuild-install", "bin.js")
			}

			_, err := util.Execute(createPrebuildInstallCommand(bin, "--force", dependency, configuration))
			if err != nil {
				execError, _ := err.(*util.ExecError)
				switch {
				case isRebuildPossible:
					logger.Warn("build native dependency from sources",
						zap.String("reason", "prebuild-install failed with error (run with env DEBUG=electron-builder to get more information)"),
						zap.ByteString("error", execError.ErrorOutput),
					)
					return nil

				case dependency.Optional:
					logger.Warn("cannot install prebuilt binaries for optional native dependency", util.CreateExecErrorLogEntry(execError)...)
					return nil

				default:
					execError.Message = "cannot build native dependency"
					execError.ExtraFields = append(execError.ExtraFields, zap.String("reason", "prebuild-install failed with error and build from sources not possible because platform or arch not compatible"))
					return err
				}
			}

			dependencies[index] = nil
			return nil
		}, nil
	})

	if err != nil {
		return nil, err
	}

	result := make([]*DepInfo, 0, len(dependencies))
	for _, item := range dependencies {
		if item == nil {
			continue
		}
		result = append(result, item)
	}
	return result, nil
}

func createPrebuildInstallCommand(bin string, extraFlag string, dependency *DepInfo, configuration *RebuildConfiguration) *exec.Cmd {
	args := []string{
		bin,
		"--platform="+configuration.Platform,
		"--arch="+configuration.Arch,
		"--target="+os.Getenv("npm_config_target"),
		"--runtime="+os.Getenv("npm_config_runtime"),
		"--verbose",
		extraFlag,
	}
	command := exec.Command(getNodeExec(configuration), args...)
	command.Dir = dependency.dir
	return command
}

func checkRebuildPossible(configuration *RebuildConfiguration) bool {
	currentOs := util.GetCurrentOs()
	nodePlatform := configuration.Platform
	switch {
	case currentOs == util.WINDOWS:
		return nodePlatform == "win32"
	case currentOs == util.MAC:
		return nodePlatform == "darwin"
	default:
		return nodePlatform != "win32" && nodePlatform != "darwin"
	}
}

func computeNativeDependencies(configuration *RebuildConfiguration) ([]*DepInfo, error) {
	result := make([][]*DepInfo, len(configuration.DependencyTreeInfo))
	err := util.MapAsync(len(configuration.DependencyTreeInfo), func(index int) (func() error, error) {
		dirInfo := configuration.DependencyTreeInfo[index]
		return func() error {
			nativeDependencies, err := computeNativeDependenciesFromNameList(&dirInfo)
			if err != nil {
				return err
			}

			result[index] = nativeDependencies
			return nil
		}, nil
	})

	if err != nil {
		return nil, err
	}

	var nativeDependencies []*DepInfo
	for _, list := range result {
		if len(list) == 0 {
			continue
		}
		nativeDependencies = append(nativeDependencies, list...)
	}
	return nativeDependencies, nil
}

func computeNativeDependenciesFromNameList(dirInfo *DependencyList) ([]*DepInfo, error) {
	result := make([]*DepInfo, len(dirInfo.Dependencies))
	err := util.MapAsync(len(dirInfo.Dependencies), func(index int) (func() error, error) {
		item := dirInfo.Dependencies[index]
		item.parentDir = dirInfo.Dir
		item.dir = filepath.Join(dirInfo.Dir, item.Name)
		return func() error {
			info, err := os.Stat(filepath.Join(item.dir, "binding.gyp"))
			if err != nil || info.IsDir() {
				return nil
			}

			result[index] = &item
			return nil
		}, nil
	})

	if err != nil {
		return nil, err
	}

	var nativeDependencies []*DepInfo
	for _, dependency := range result {
		if dependency != nil {
			nativeDependencies = append(nativeDependencies, dependency)
		}
	}
	return nativeDependencies, nil
}

func computeExecPath(configuration *RebuildConfiguration) (string, []string, bool, error) {
	//noinspection SpellCheckingInspection
	execPath := os.Getenv("npm_execpath")
	if execPath == "" {
		execPath = os.Getenv("NPM_CLI_JS")
	}

	forceYarn := util.IsEnvTrue("FORCE_YARN")

	isRunningYarn := false
	if forceYarn {
		isRunningYarn = true
	} else {
		if execPath != "" && strings.HasPrefix(filepath.Base(execPath), "yarn") {
			isRunningYarn = true
		} else {
			isRunningYarn = strings.Contains(os.Getenv("npm_config_user_agent"), "yarn")
		}
	}

	var execArgs []string

	if execPath == "" {
		suffix := ""
		if util.GetCurrentOs() == util.WINDOWS {
			suffix = ".cmd"
		}

		if isRunningYarn {
			execPath = "yarn" + suffix
		} else {
			execPath = "npm" + suffix
		}
	} else {
		// Wrap with `node` interpreter if needed
		isJs, err := isJavascriptFile(execPath)
		if err != nil {
			return execPath, execArgs, isRunningYarn, err
		}
		if isJs {
			execArgs = append(execArgs, execPath)
			execPath = getNodeExec(configuration)
		}
	}

	return execPath, execArgs, isRunningYarn, nil
}

func getNodeExec(configuration *RebuildConfiguration) string {
	//noinspection SpellCheckingInspection
	execPath := os.Getenv("npm_node_execpath")
	if execPath == "" {
		execPath = os.Getenv("NODE_EXE")
		if execPath == "" {
			execPath = os.Getenv("node")
			if execPath == "" {
				execPath = configuration.NodeExecPath
			}
		}
	}
	return execPath
}

// Efficiently reads first few bytes from the given file in order to extract
// the hashbang interpreter it's used
func readHashBang(path string) (string, error) {
	logger := log.LOG.With(zap.String("place", "readHashbang"))

	r, err := os.Open(path)
	if err != nil {
		return "", err
	}
	
	defer r.Close()
	
	var header [128]byte
	bytesRead, err := io.ReadFull(r, header[:])
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		return "", err
	}

	logger.Info(fmt.Sprintf("header=%s, n=%d\n", string(header[:]), bytesRead))

	if header[0] == '#' && header[1] == '!' {
		str := string(header[2:bytesRead])
		end := strings.IndexAny(str, "\r\n\t ")
		if end == -1 {
			end = len(str)
		}
		return str[0:end], nil
		} else {
		return "", nil
	}
}

func isJavascriptFile(path string) (bool, error) {
	logger := log.LOG.With(zap.String("place", "isJavascriptFile"))
	info, err := os.Stat(path)
	if err != nil {
		return false, fmt.Errorf("Could not get info of %s: %w", path, err)
	}

	// Resolve path if this is a link
	if info.Mode() & os.ModeSymlink != 0 {
		linkPath, err := filepath.EvalSymlinks(path)
		if err != nil {
			return false, fmt.Errorf("Could not read link of %s: %w", path, err)
		}
		path = linkPath
	}

	// Check if this is a '.js' file, in which case
	// we should return `true` without further considerations
	if strings.HasSuffix(strings.ToLower(path), ".js") {
		logger.Info(fmt.Sprintf("has suffix=%s\n", path))
		return true, nil
	}

	// Otherwise read the hashbang contents and return true
	// only if the target uses node interpreter
	interpreter, err := readHashBang(path)
	logger.Info(fmt.Sprintf("interpreter=%s", interpreter))
	if err != nil {
		return false, fmt.Errorf("Could not read hash bang of %s: %w", path, err)
	}
	if strings.HasSuffix(interpreter, "node") {
		return true, nil
	}

	return false, nil
}