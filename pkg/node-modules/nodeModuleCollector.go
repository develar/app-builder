package node_modules

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/develar/app-builder/pkg/fs"
	"github.com/develar/app-builder/pkg/log"
	"github.com/develar/errors"
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"
)

type DependencyBinary struct {
	NapiVersions []uint `json:"napi_versions"`
}

type Dependency struct {
	Name                 string            `json:"name"`
	Version              string            `json:"version"`
	Dependencies         map[string]string `json:"dependencies"`
	OptionalDependencies map[string]string `json:"optionalDependencies"`
	Binary               *DependencyBinary `json:"binary"`

	conflictDependency map[string]*Dependency
	dir                string
	isOptional         int
}

type Collector struct {
	unresolvedDependencies map[string]bool

	excludedDependencies map[string]bool

	NodeModuleDirToDependencyMap map[string]*map[string]*Dependency `json:"nodeModuleDirToDependencyMap"`
	DependencyMap                map[string]*Dependency             `json:"dependencyMap"`
}

func (t *Collector) readDependencyTree(dependency *Dependency) error {
	maxQueueSize := len(dependency.Dependencies) + len(dependency.OptionalDependencies)

	if maxQueueSize == 0 {
		return nil
	}

	nodeModuleDir, err := findNearestNodeModuleDir(dependency.dir)
	if err != nil {
		return err
	}

	if len(nodeModuleDir) == 0 {
		for name := range dependency.Dependencies {
			t.unresolvedDependencies[name] = true
		}
		return nil
	}

	// process direct children first
	queue := make([]*Dependency, maxQueueSize)
	queueIndex := 0

	queueIndex, err = t.processDependencies(&dependency.Dependencies, nodeModuleDir, false, &queue, queueIndex)
	if err != nil {
		return err
	}

	queueIndex, err = t.processDependencies(&dependency.OptionalDependencies, nodeModuleDir, true, &queue, queueIndex)
	if err != nil {
		return err
	}

	if queueIndex == 0 {
		return nil
	}

	// do not sort - final result will be sorted
	for i := 0; i < queueIndex; i++ {
		err = t.readDependencyTree(queue[i])
		if err != nil {
			return err
		}
		t.AddDependencyMap(queue[i], dependency)
	}

	return nil
}

func (t *Collector) AddDependencyMap(childDependency *Dependency, parentDependency *Dependency) {
	if t.DependencyMap == nil {
		t.DependencyMap = make(map[string]*Dependency)
	}

	name := childDependency.Name
	if d, ok := t.DependencyMap[name]; ok {
		if d.Version != childDependency.Version {
			if parentDependency.conflictDependency == nil {
				parentDependency.conflictDependency = make(map[string]*Dependency)
			}
			parentDependency.conflictDependency[name] = childDependency
		}
	} else {
		t.DependencyMap[name] = childDependency
	}
}

func (t *Collector) processDependencies(list *map[string]string, nodeModuleDir string, isOptional bool, queue *[]*Dependency, queueIndex int) (int, error) {
	unresolved := make([]string, 0)
	for name := range *list {
		if strings.HasPrefix(name, "@types/") {
			continue
		}

		if t.excludedDependencies != nil {
			_, isExcluded := t.excludedDependencies[name]
			if isExcluded {
				continue
			}
		}

		childDependency, err := t.resolveDependency(nodeModuleDir, name)
		if err != nil {
			return queueIndex, err
		}

		if childDependency == nil {
			unresolved = append(unresolved, name)
		} else {
			(*queue)[queueIndex] = childDependency
			correctOptionalState(isOptional, childDependency)
			queueIndex++
		}
	}

	var err error
	guardCount := 0
	for len(unresolved) > 0 {
		nodeModuleDir, err = findNearestNodeModuleDir(getParentDir(getParentDir(nodeModuleDir)))
		if err != nil {
			return queueIndex, err
		}

		if len(nodeModuleDir) == 0 {
			if !isOptional {
				for _, name := range unresolved {
					if len(name) != 0 {
						t.unresolvedDependencies[name] = true
					}
				}
			}
			return queueIndex, nil
		}

		if log.IsDebugEnabled() {
			log.Debug("unresolved deps", zap.Strings("unresolved", unresolved), zap.String("nodeModuleDir", nodeModuleDir), zap.Int("round", guardCount))
		}

		hasUnresolved := false
		for index, name := range unresolved {
			if len(name) == 0 {
				continue
			}

			childDependency, err := t.resolveDependency(nodeModuleDir, name)
			if err != nil {
				return queueIndex, err
			}

			if childDependency == nil {
				hasUnresolved = true
			} else {
				(*queue)[queueIndex] = childDependency
				correctOptionalState(isOptional, childDependency)
				queueIndex++
				unresolved[index] = ""
			}
		}

		if !hasUnresolved {
			break
		}

		guardCount++
		if guardCount > 999 {
			return queueIndex, errors.New("Infinite loop: " + nodeModuleDir)
		}
	}

	return queueIndex, nil
}

func correctOptionalState(isOptional bool, childDependency *Dependency) {
	if isOptional {
		if childDependency.isOptional == 0 {
			childDependency.isOptional = 1
		}
	} else {
		childDependency.isOptional = 2
	}
}

// nil if already handled
func (t *Collector) resolveDependency(parentNodeModuleDir string, name string) (*Dependency, error) {
	dependencyNameToDependency := t.NodeModuleDirToDependencyMap[parentNodeModuleDir]
	if dependencyNameToDependency != nil {
		dependency := (*dependencyNameToDependency)[name]
		if dependency != nil {
			return nil, nil
		}
	}

	realParentNodeModuleDir := fs.FindParentWithFile(parentNodeModuleDir, name)
	if realParentNodeModuleDir == "" {
		return nil, nil
	}
	dependencyDir := filepath.Join(realParentNodeModuleDir, name)
	dependency, err := readPackageJson(dependencyDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		} else {
			return nil, errors.WithStack(err)
		}
	}

	if name == "libui-node" {
		// remove because production app doesn't need to download libui
		//noinspection SpellCheckingInspection
		delete(dependency.Dependencies, "libui-download")
	}

	if dependencyNameToDependency == nil {
		m := make(map[string]*Dependency)
		t.NodeModuleDirToDependencyMap[parentNodeModuleDir] = &m
		dependencyNameToDependency = &m
	}

	(*dependencyNameToDependency)[name] = dependency
	dependency.dir = dependencyDir
	return dependency, nil
}

func findNearestNodeModuleDir(dir string) (string, error) {
	if len(dir) == 0 {
		return "", nil
	}

	realDir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return "", errors.WithStack(err)
	}
	dir = realDir

	guardCount := 0
	for {
		nodeModuleDir := filepath.Join(dir, "node_modules")
		fileInfo, err := os.Stat(nodeModuleDir)
		if err != nil {
			if !os.IsNotExist(err) {
				return "", errors.WithStack(err)
			}
		} else if fileInfo.IsDir() {
			return nodeModuleDir, nil
		}

		dir = getParentDir(dir)
		if len(dir) == 0 {
			return "", nil
		}

		guardCount++
		if guardCount > 999 {
			return "", errors.New("infinite loop: " + dir)
		}
	}
}

func getParentDir(file string) string {
	if len(file) == 0 {
		return file
	}

	dir := filepath.Dir(file)
	// https://github.com/develar/app-builder/pull/3
	if len(dir) > 1 /* . or / or empty */ && dir != file {
		return dir
	} else {
		return ""
	}
}

func readPackageJson(dir string) (*Dependency, error) {
	packageFile := filepath.Join(dir, "package.json")
	data, err := ioutil.ReadFile(packageFile)
	if err != nil {
		return nil, err
	}

	var dependency Dependency
	err = jsoniter.Unmarshal(data, &dependency)
	if err != nil {
		return nil, errors.WithMessage(err, "Error reading package.json: "+packageFile)
	}

	return &dependency, nil
}
