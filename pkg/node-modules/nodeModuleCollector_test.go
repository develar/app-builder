package node_modules

import (
	"path"
	"path/filepath"
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
	collector.processHoistDependencyMap()

	r := lo.FlatMap(lo.Values(collector.NodeModuleDirToDependencyMap), func(it *map[string]*Dependency, i int) []string {
		return lo.Keys(*it)
	})
	g.Expect(r).To(ConsistOf([]string{
		"js-tokens", "react", "remote", "loose-envify",
	}))
	remoteModule := collector.HoiestedDependencyMap["@electron/remote"]
	g.Expect(remoteModule.alias).To(Equal("remote"))
	g.Expect(remoteModule.Name).To(Equal("@electron/remote"))
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
	collector.processHoistDependencyMap()
	r := lo.FlatMap(lo.Values(collector.NodeModuleDirToDependencyMap), func(it *map[string]*Dependency, i int) []string {
		return lo.Keys(*it)
	})
	g.Expect(r).To(ConsistOf([]string{
		"js-tokens", "react", "remote", "loose-envify",
	}))

	remoteModule := collector.HoiestedDependencyMap["@electron/remote"]
	g.Expect(remoteModule.Name).To(Equal("@electron/remote"))
	g.Expect(remoteModule.alias).To(Equal("remote"))
	g.Expect(remoteModule.dir).To(Equal(filepath.Join(dir, "node_modules/.pnpm/@electron+remote@2.1.2_electron@31.0.0/node_modules/@electron/remote")))

	reactModule := collector.HoiestedDependencyMap["react"]
	g.Expect(reactModule.dir).To(Equal(filepath.Join(dir, "node_modules/.pnpm/react@18.2.0/node_modules/react")))
}

func TestReadDependencyTreeForTar(t *testing.T) {
	g := NewGomegaWithT(t)

	collector := &Collector{
		unresolvedDependencies:       make(map[string]bool),
		excludedDependencies:         make(map[string]bool),
		NodeModuleDirToDependencyMap: make(map[string]*map[string]*Dependency),
	}

	dir := path.Join(Dirname(), "tar-demo")

	dependency, err := readPackageJson(dir)
	dependency.dir = dir

	g.Expect(err).NotTo(HaveOccurred())

	collector.rootDependency = dependency
	err = collector.readDependencyTree(dependency)
	g.Expect(err).NotTo(HaveOccurred())
	collector.processHoistDependencyMap()

	r := lo.FlatMap(lo.Values(collector.NodeModuleDirToDependencyMap), func(it *map[string]*Dependency, i int) []string {
		return lo.Keys(*it)
	})
	g.Expect(len(r)).To(Equal(46))

	g.Expect(collector.HoiestedDependencyMap["tar"].dir).To(Equal(filepath.Join(dir, "node_modules/tar")))
	g.Expect(collector.HoiestedDependencyMap["minipass"].Version).To(Equal("7.1.2"))
	g.Expect(collector.HoiestedDependencyMap["minizlib"].Version).To(Equal("3.0.1"))
	g.Expect(collector.HoiestedDependencyMap["tar"].conflictDependency["ansi-regex"].Version).To(Equal("5.0.1"))
}

func TestReadDependencyTreeForYarn(t *testing.T) {
	g := NewGomegaWithT(t)

	collector := &Collector{
		unresolvedDependencies:       make(map[string]bool),
		excludedDependencies:         make(map[string]bool),
		NodeModuleDirToDependencyMap: make(map[string]*map[string]*Dependency),
	}

	dir := path.Join(Dirname(), "yarn-demo/packages/test-app")

	dependency, err := readPackageJson(dir)
	dependency.dir = dir

	g.Expect(err).NotTo(HaveOccurred())

	collector.rootDependency = dependency
	err = collector.readDependencyTree(dependency)
	g.Expect(err).NotTo(HaveOccurred())
	collector.processHoistDependencyMap()

	g.Expect(collector.HoiestedDependencyMap["foo"].dir).To(Equal(filepath.Join(Dirname(), "yarn-demo/packages/foo")))
	g.Expect(collector.HoiestedDependencyMap["foo"].Version).To(Equal("1.0.0"))
	g.Expect(collector.HoiestedDependencyMap["foo"].conflictDependency["ms"].dir).To(Equal(filepath.Join(Dirname(), "yarn-demo/node_modules/ms")))
	g.Expect(collector.HoiestedDependencyMap["foo"].conflictDependency["ms"].Version).To(Equal("2.0.0"))
	g.Expect(collector.HoiestedDependencyMap["ms"].Version).To(Equal("2.1.1"))
	g.Expect(collector.HoiestedDependencyMap["ms"].dir).To(Equal(filepath.Join(Dirname(), "yarn-demo/packages/test-app/node_modules/ms")))
}
