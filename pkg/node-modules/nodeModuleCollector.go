package node_modules

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

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

	parent             *Dependency
	conflictDependency map[string]*Dependency
	dir                string
	isOptional         int
	alias              string
}

type Collector struct {
	rootDependency         *Dependency
	unresolvedDependencies map[string]bool

	excludedDependencies map[string]bool
	allDependencies      []*Dependency
	allDependenciesMap   map[string]*Dependency

	NodeModuleDirToDependencyMap map[string]*map[string]*Dependency `json:"nodeModuleDirToDependencyMap"`

	HoiestedDependencyMap map[string]*Dependency `json:"hoiestedDependencyMap"`
}

func (t *Collector) readDependencyTree(dependency *Dependency) error {
	if t.rootDependency == nil {
		t.rootDependency = dependency
		t.allDependenciesMap = make(map[string]*Dependency)
	} else {
		key := dependency.alias + dependency.Version + dependency.dir
		if _, ok := t.allDependenciesMap[key]; ok {
			return nil
		}
		t.allDependenciesMap[key] = dependency
		t.allDependencies = append(t.allDependencies, dependency)
	}

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
		queue[i].parent = dependency
	}
	return nil
}

func IsParentPath(parentPath, childPath string) bool {
	parentPath = filepath.Clean(parentPath)
	childPath = filepath.Clean(childPath)

	relPath, err := filepath.Rel(parentPath, childPath)
	if err != nil {
		return false
	}

	if relPath == "." || strings.HasPrefix(relPath, "..") {
		return false
	}
	return true
}

func (t *Collector) writeToParentConflicDependency(d *Dependency) {
	for p := d.parent; p != t.rootDependency; p = p.parent {
		if IsParentPath(p.dir, d.dir) {
			if p.conflictDependency == nil {
				p.conflictDependency = make(map[string]*Dependency)
			}
			p.conflictDependency[d.alias] = d
			return
		}
	}

	if h, ok := t.HoiestedDependencyMap[d.alias]; ok {
		if h.Version == d.Version {
			return
		}

		// for pnpm
		p := d.parent
		last := d
		for p != t.rootDependency {
			if t.HoiestedDependencyMap[p.alias] != nil {
				last = p
				break
			}

			if p.conflictDependency != nil {
				if c, ok := p.conflictDependency[d.alias]; ok {
					if c.Version == d.Version {
						return
					}
					break
				}
			}
			last = p
			p = p.parent
		}

		if last.conflictDependency == nil {
			last.conflictDependency = make(map[string]*Dependency)
		}
		last.conflictDependency[d.alias] = d
		return
	}

	t.HoiestedDependencyMap[d.alias] = d

}

func (t *Collector) processHoistDependencyMap() {
	t.HoiestedDependencyMap = make(map[string]*Dependency)
	for _, d := range t.allDependencies {
		t.writeToParentConflicDependency(d)
	}
}

func (t *Collector) processDependencies(list *map[string]string, nodeModuleDir string, isOptional bool, queue *[]*Dependency, queueIndex int) (int, error) {
	unresolved := make([]string, 0)

	names := make([]string, 0, len(*list))
	for k := range *list {
		names = append(names, k)
	}
	sort.Strings(names)

	for _, name := range names {
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
			return dependency, nil
		}
	}

	dependencyDir := filepath.Join(parentNodeModuleDir, name)
	info, err := os.Stat(dependencyDir)
	if err == nil && !info.IsDir() {
		return nil, nil
	}

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
	dependency.alias = name
	dependency.dir = resolvePath(dependencyDir)
	return dependency, nil
}

func resolvePath(dir string) string {
	// Check if the path is a symlink
	info, err := os.Lstat(dir)
	if err != nil {
		return dir
	}

	if info.Mode()&os.ModeSymlink != 0 {
		// It's a symlink, resolve the real path
		realPath, err := filepath.EvalSymlinks(dir)
		if err != nil {
			return dir
		}
		return realPath
	}

	// Not a symlink, return the original path
	return dir
}

func findNearestNodeModuleDir(dir string) (string, error) {
	if len(dir) == 0 {
		return "", nil
	}

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
	data, err := os.ReadFile(packageFile)
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
