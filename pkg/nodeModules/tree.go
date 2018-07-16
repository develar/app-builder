package nodeModules

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/develar/errors"
	"github.com/json-iterator/go"
)

func ConfigureCommand(app *kingpin.Application) {
	command := app.Command("node-dep-tree", "")

	dir := command.Flag("dir", "").Required().String()
	command.Action(func(context *kingpin.ParseContext) error {
		collector := &Collector{
			unresolvedDependencies:       make(map[string]bool),
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
	for name := range *list {
		childDependency, err := t.resolveDependency(nodeModuleDir, name, isOptional)
		if err != nil {
			return queueIndex, errors.WithStack(err)
		}

		if childDependency != nil {
			(*queue)[queueIndex] = childDependency
			queueIndex++
		}
	}
	return queueIndex, nil
}

// nill if already handled
func (t *Collector) resolveDependency(parentNodeModuleDir string, name string, isOptional bool) (*Dependency, error) {
	dependencyNameToDependency := t.NodeModuleDirToDependencyMap[parentNodeModuleDir]
	if dependencyNameToDependency != nil {
		dependency := (*dependencyNameToDependency)[name]
		if dependency != nil {
			return nil, nil
		}
	}

	dependencyDir := filepath.Join(parentNodeModuleDir, name)
	dependency, err := readPackageJson(dependencyDir)
	if err != nil {
		if os.IsNotExist(err) {
			nodeModuleDir, err := findNearestNodeModuleDir(filepath.Dir(filepath.Dir(parentNodeModuleDir)))
			if err != nil {
				return nil, errors.WithStack(err)
			}

			if len(nodeModuleDir) == 0 {
				if !isOptional {
					t.unresolvedDependencies[name] = true
				}
				return nil, nil
			} else {
				return t.resolveDependency(nodeModuleDir, name, isOptional)
			}
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
		return nil, errors.WithStack(err)
	}

	return &dependency, nil
}
