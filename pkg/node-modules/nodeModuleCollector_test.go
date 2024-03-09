package node_modules

import (
	"path"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/samber/lo"
)

func TestReadDependencyTreeByYarn(t *testing.T) {
	g := NewGomegaWithT(t)

	collector := &Collector{
		unresolvedDependencies:       make(map[string]bool),
		excludedDependencies:         make(map[string]bool),
		NodeModuleDirToDependencyMap: make(map[string]*map[string]*Dependency),
	}

	dir := path.Join(Dirname(), "yarn-demo")
	var dependencies []*Dependency
	dependency, err := readPackageJson(dir)
	dependency.dir = dir
	dependencies = append(dependencies, dependency)

	if len(dependency.Workspaces) > 0 {
		workspaces := getAllWorkspaces(dir, dependency.Workspaces)
		if len(workspaces) > 0 {
			dependencies = append(dependencies, workspaces...)
		}
	}

	for _, dependency = range dependencies {
		err = collector.readDependencyTree(dependency)
		if err != nil {
			continue
		}
	}
	g.Expect(err).NotTo(HaveOccurred())

	err = collector.readDependencyTree(dependency)
	g.Expect(err).NotTo(HaveOccurred())
	r := lo.FlatMap(lo.Values(collector.NodeModuleDirToDependencyMap), func(it *map[string]*Dependency, i int) []string {
		return lo.Keys(*it)
	})
	g.Expect(r).To(ConsistOf([]string{
		"foo", "ms",
	}))
}

func TestReadDependencyTreeByNpm(t *testing.T) {
	g := NewGomegaWithT(t)

	collector := &Collector{
		unresolvedDependencies:       make(map[string]bool),
		excludedDependencies:         make(map[string]bool),
		NodeModuleDirToDependencyMap: make(map[string]*map[string]*Dependency),
	}

	dir := path.Join(Dirname(), "npm-demo")

	dependency, err := readPackageJson(dir)
	dependency.dir = dir
	g.Expect(err).NotTo(HaveOccurred())

	err = collector.readDependencyTree(dependency)
	g.Expect(err).NotTo(HaveOccurred())
	r := lo.FlatMap(lo.Values(collector.NodeModuleDirToDependencyMap), func(it *map[string]*Dependency, i int) []string {
		return lo.Keys(*it)
	})
	g.Expect(r).To(ConsistOf([]string{
		"js-tokens", "react", "loose-envify",
	}))
}

func TestReadDependencyTreeByPnpm(t *testing.T) {
	g := NewGomegaWithT(t)

	collector := &Collector{
		unresolvedDependencies:       make(map[string]bool),
		excludedDependencies:         make(map[string]bool),
		NodeModuleDirToDependencyMap: make(map[string]*map[string]*Dependency),
	}

	dir := path.Join(Dirname(), "pnpm-demo")

	dependency, err := readPackageJson(dir)
	dependency.dir = dir
	g.Expect(err).NotTo(HaveOccurred())

	err = collector.readDependencyTree(dependency)
	g.Expect(err).NotTo(HaveOccurred())
	r := lo.FlatMap(lo.Values(collector.NodeModuleDirToDependencyMap), func(it *map[string]*Dependency, i int) []string {
		return lo.Keys(*it)
	})
	g.Expect(r).To(ConsistOf([]string{
		"js-tokens", "react", "loose-envify",
	}))
}
