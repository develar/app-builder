package node_modules

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/apex/log"
	"github.com/develar/app-builder/pkg/util"
	"github.com/develar/errors"
	"github.com/json-iterator/go"
)

func ConfigureCommand(app *kingpin.Application) {
	command := app.Command("node-dep-tree", "")

	dir := command.Flag("dir", "").Required().String()
	excludedDependencies := command.Flag("exclude-dep", "").Strings()

	command.Action(func(context *kingpin.ParseContext) error {
		var excluded map[string]bool
		if excludedDependencies == nil || len(*excludedDependencies) == 0 {
			excluded = nil
		} else {
			excluded = make(map[string]bool, len(*excludedDependencies))
			for _, name := range *excludedDependencies {
				excluded[name] = true
			}
		}

		collector := &Collector{
			unresolvedDependencies:       make(map[string]bool),
			excludedDependencies:         excluded,
			NodeModuleDirToDependencyMap: make(map[string]*map[string]*Dependency),
		}
		dependency, err := readPackageJson(*dir)
		if err != nil {
			return errors.WithStack(err)
		}

		dependency.dir = *dir
		err = collector.readDependencyTree(dependency)
		if err != nil {
			return errors.WithStack(err)
		}

		jsonWriter := jsoniter.NewStream(jsoniter.ConfigDefault, os.Stdout, 32*1024)
		writeResult(jsonWriter, collector)
		err = jsonWriter.Flush()
		if err != nil {
			return errors.WithStack(err)
		}

		return nil
	})
}

func writeResult(jsonWriter *jsoniter.Stream, collector *Collector) {
	moduleDirs := make([]string, len(collector.NodeModuleDirToDependencyMap))
	index := 0
	for k := range collector.NodeModuleDirToDependencyMap {
		moduleDirs[index] = k
		index++
	}

	if len(moduleDirs) > 1 {
		sort.Slice(moduleDirs, func(i, j int) bool {
			return pathSorter(strings.Split(moduleDirs[i], string(filepath.Separator)), strings.Split(moduleDirs[j], string(filepath.Separator)))
		})
	}

	jsonWriter.WriteArrayStart()
	isFirst := true
	for _, nodeModulesDir := range moduleDirs {
		if isFirst {
			isFirst = false
		} else {
			jsonWriter.WriteMore()
		}

		jsonWriter.WriteObjectStart()
		jsonWriter.WriteObjectField("dir")
		jsonWriter.WriteString(nodeModulesDir)

		jsonWriter.WriteMore()
		jsonWriter.WriteObjectField("deps")
		writeArray(jsonWriter, collector.NodeModuleDirToDependencyMap[nodeModulesDir])
		jsonWriter.WriteObjectEnd()
	}
	jsonWriter.WriteArrayEnd()
}

func pathSorter(a []string, b []string) bool {
	aL := len(a)
	l := aL
	bL := len(b)
	if bL > l {
		l = bL
	}

	for i := 0; i < l; i++ {
		if i == aL {
			return true
		}
		if i == bL {
			return false
		}
		if a[i] > b[i] {
			return false
		}
		if a[i] < b[i] {
			return true
		}
		if aL < bL {
			return true
		}
		if aL > bL {
			return false
		}
	}

	return false
}

func writeArray(jsonWriter *jsoniter.Stream, v *map[string]*Dependency) {
	names := make([]string, len(*v))
	index := 0
	for k := range *v {
		names[index] = k
		index++
	}

	if len(names) > 1 {
		sort.Strings(names)
	}

	isComma := false
	jsonWriter.WriteArrayStart()
	for _, depName := range names {
		if isComma {
			jsonWriter.WriteMore()
		} else {
			isComma = true
		}
		jsonWriter.WriteString(depName)
	}
	jsonWriter.WriteArrayEnd()
}

type Dependency struct {
	Name                 string            `json:"name"`
	Version              string            `json:"version"`
	Dependencies         map[string]string `json:"dependencies"`
	OptionalDependencies map[string]string `json:"optionalDependencies"`

	dir string
}

type Collector struct {
	unresolvedDependencies map[string]bool

	excludedDependencies map[string]bool

	NodeModuleDirToDependencyMap map[string]*map[string]*Dependency `json:"nodeModuleDirToDependencyMap"`
}

func (t *Collector) readDependencyTree(dependency *Dependency) error {
	maxQueueSize := len(dependency.Dependencies) + len(dependency.OptionalDependencies)

	if maxQueueSize == 0 {
		return nil
	}

	nodeModuleDir, err := findNearestNodeModuleDir(dependency.dir)
	if err != nil {
		return errors.WithStack(err)
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
			return errors.WithStack(err)
		}
	}

	return nil
}

func (t *Collector) processDependencies(list *map[string]string, nodeModuleDir string, isOptional bool, queue *[]*Dependency, queueIndex int) (int, error) {
	unresolved := make([]string, 0)
	for name := range *list {
		if t.excludedDependencies != nil {
			_, isExcluded := t.excludedDependencies[name]
			if isExcluded {
				continue
			}
		}

		childDependency, err := t.resolveDependency(nodeModuleDir, name)
		if err != nil {
			return queueIndex, errors.WithStack(err)
		}

		if childDependency != nil {
			(*queue)[queueIndex] = childDependency
			queueIndex++
		} else {
			unresolved = append(unresolved, name)
		}
	}

	var err error
	guardCount := 0
	for len(unresolved) > 0 {
		nodeModuleDir, err = findNearestNodeModuleDir(getParentDir(getParentDir(nodeModuleDir)))
		if err != nil {
			return queueIndex, errors.WithStack(err)
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

		if util.IsDebugEnabled() {
			log.WithField("unresolved", strings.Join(unresolved, ", ")).WithField("nodeModuleDir", nodeModuleDir).WithField("round", guardCount).Debug("unresolved deps")
		}

		hasUnresolved := false
		for index, name := range unresolved {
			if len(name) == 0 {
				continue
			}

			childDependency, err := t.resolveDependency(nodeModuleDir, name)
			if err != nil {
				return queueIndex, errors.WithStack(err)
			}

			if childDependency != nil {
				(*queue)[queueIndex] = childDependency
				queueIndex++
				unresolved[index] = ""
			} else {
				hasUnresolved = true
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

// nil if already handled
func (t *Collector) resolveDependency(parentNodeModuleDir string, name string) (*Dependency, error) {
	dependencyNameToDependency := t.NodeModuleDirToDependencyMap[parentNodeModuleDir]
	if dependencyNameToDependency != nil {
		dependency := (*dependencyNameToDependency)[name]
		if dependency != nil {
			return nil, nil
		}
	}

	dependencyDir := filepath.Join(parentNodeModuleDir, name)
	dependency, err := readPackageJson(dependencyDir)

	if //noinspection SpellCheckingInspection
	name == "libui-node" {
		// remove because production app doesn't need to download libui
		//noinspection SpellCheckingInspection
		delete(dependency.Dependencies, "libui-download")
	}

	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		} else {
			return nil, errors.WithStack(err)
		}
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
			return "", errors.New("Infinite loop: " + dir)
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
	data, err := ioutil.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return nil, err
	}

	var dependency Dependency
	err = jsoniter.Unmarshal(data, &dependency)
	if err != nil {
		log.Error("Error reading package.json: " + filepath.Join(dir, "package.json"))
		return nil, errors.WithStack(err)
	}

	return &dependency, nil
}
