package node_modules

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/alecthomas/kingpin"
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
			return err
		}

		dependency.dir = *dir
		err = collector.readDependencyTree(dependency)
		if err != nil {
			return err
		}

		jsonWriter := jsoniter.NewStream(jsoniter.ConfigDefault, os.Stdout, 32*1024)
		writeResult(jsonWriter, collector)
		err = jsonWriter.Flush()
		if err != nil {
			return err
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
		writeDependencyList(jsonWriter, collector.NodeModuleDirToDependencyMap[nodeModulesDir])

		jsonWriter.WriteObjectEnd()
	}
	jsonWriter.WriteArrayEnd()
}

func writeDependencyList(jsonWriter *jsoniter.Stream, dependencyMap *map[string]*Dependency) {
	jsonWriter.WriteArrayStart()
	isFirst := true
	for name, info := range *dependencyMap {
		if isFirst {
			isFirst = false
		} else {
			jsonWriter.WriteMore()
		}

		jsonWriter.WriteObjectStart()

		jsonWriter.WriteObjectField("name")
		jsonWriter.WriteString(name)

		jsonWriter.WriteMore()
		jsonWriter.WriteObjectField("version")
		jsonWriter.WriteString(info.Version)

		if info.isOptional == 1 {
			jsonWriter.WriteMore()
			jsonWriter.WriteObjectField("optional")
			jsonWriter.WriteBool(true)
		}

		for name, _ := range info.Dependencies {
			if name == "prebuild-install" {
				jsonWriter.WriteMore()
				jsonWriter.WriteObjectField("hasPrebuildInstall")
				jsonWriter.WriteBool(true)
				break
			}
		}

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

func writeArray(names []string, jsonWriter *jsoniter.Stream) {
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
