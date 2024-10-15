package node_modules

import (
	"fmt"
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
	remoteModule := collector.HoiestedDependencyMap["remote"]
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

	remoteModule := collector.HoiestedDependencyMap["remote"]
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

	err = collector.readDependencyTree(dependency)
	g.Expect(err).NotTo(HaveOccurred())
	collector.processHoistDependencyMap()

	r := lo.FlatMap(lo.Values(collector.NodeModuleDirToDependencyMap), func(it *map[string]*Dependency, i int) []string {
		return lo.Keys(*it)
	})
	g.Expect(len(r)).To(Equal(97))

	g.Expect(collector.HoiestedDependencyMap["tar"].dir).To(Equal(filepath.Join(dir, "node_modules/tar")))
	g.Expect(collector.HoiestedDependencyMap["tar"].conflictDependency["minipass"].Version).To(Equal("7.1.2"))
	g.Expect(collector.HoiestedDependencyMap["tar"].conflictDependency["minizlib"].Version).To(Equal("3.0.1"))

	g.Expect(collector.HoiestedDependencyMap["archiver-utils"].dir).To(Equal(filepath.Join(dir, "node_modules/archiver-utils")))
	g.Expect(collector.HoiestedDependencyMap["archiver-utils"].Version).To(Equal("5.0.2"))
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

func TestReadDependencyTreeForParse(t *testing.T) {
	g := NewGomegaWithT(t)

	collector := &Collector{
		unresolvedDependencies:       make(map[string]bool),
		excludedDependencies:         make(map[string]bool),
		NodeModuleDirToDependencyMap: make(map[string]*map[string]*Dependency),
	}

	dir := path.Join(Dirname(), "parse-demo")

	dependency, err := readPackageJson(dir)
	dependency.dir = dir

	g.Expect(err).NotTo(HaveOccurred())

	err = collector.readDependencyTree(dependency)
	g.Expect(err).NotTo(HaveOccurred())
	collector.processHoistDependencyMap()

	g.Expect(collector.HoiestedDependencyMap["parse-asn1"].dir).To(Equal(filepath.Join(Dirname(), "parse-demo/node_modules/parse-asn1")))
	g.Expect(collector.HoiestedDependencyMap["parse-asn1"].Version).To(Equal("5.1.7"))

	g.Expect(collector.HoiestedDependencyMap["asn1.js"].dir).To(Equal(filepath.Join(Dirname(), "parse-demo/node_modules/asn1.js")))
	g.Expect(collector.HoiestedDependencyMap["asn1.js"].Version).To(Equal("4.10.1"))
}

func TestReadDependencyTreeForEs5(t *testing.T) {
	g := NewGomegaWithT(t)

	collector := &Collector{
		unresolvedDependencies:       make(map[string]bool),
		excludedDependencies:         make(map[string]bool),
		NodeModuleDirToDependencyMap: make(map[string]*map[string]*Dependency),
	}

	dir := path.Join(Dirname(), "es5-demo")

	dependency, err := readPackageJson(dir)
	dependency.dir = dir

	g.Expect(err).NotTo(HaveOccurred())

	err = collector.readDependencyTree(dependency)
	g.Expect(err).NotTo(HaveOccurred())
	collector.processHoistDependencyMap()

	fmt.Println(collector.HoiestedDependencyMap)

	g.Expect(collector.HoiestedDependencyMap["d"].dir).To(Equal(filepath.Join(Dirname(), "es5-demo/node_modules/.pnpm/d@1.0.2/node_modules/d")))
	g.Expect(collector.HoiestedDependencyMap["d"].Version).To(Equal("1.0.2"))

	g.Expect(collector.HoiestedDependencyMap["d"].conflictDependency["es5-ext"].dir).To(Equal(filepath.Join(Dirname(), "es5-demo/node_modules/.pnpm/es5-ext@0.10.64/node_modules/es5-ext")))
	g.Expect(collector.HoiestedDependencyMap["d"].conflictDependency["es5-ext"].Version).To(Equal("0.10.64"))

}
