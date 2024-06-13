package node_modules

import (
	"path"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/samber/lo"
)

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
		"js-tokens", "react", "remote", "loose-envify",
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
		"js-tokens", "react", "remote", "loose-envify",
	}))
}
