package icons

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/develar/app-builder/pkg/log"
	"github.com/develar/app-builder/pkg/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/biessek/golang-ico"
)

func TestIcons(t *testing.T) {
	log.InitLogger()
	RegisterFailHandler(Fail)
	RunSpecs(t, "Icons Suite")
}

func TestCommonSourcesSet(t *testing.T) {
	g := NewGomegaWithT(t)

	result := createCommonIconSources([]string{"foo"}, "set")
	g.Expect(result).To(Equal([]string{"foo", "foo.png", "foo.icns", "foo.ico", "icons", "icon.png", "icon.icns", "icon.ico"}))
}

func TestCommonSourcesIcns(t *testing.T) {
	result := createCommonIconSources([]string{"foo"}, "icns")
	g := NewGomegaWithT(t)
	g.Expect(result).To(Equal([]string{"foo.icns", "foo", "foo.png", "icon.icns", "icons", "icon.png"}))
}

func TestCommonSourcesNil(t *testing.T) {
	result := createCommonIconSources(nil, "set")
	g := NewGomegaWithT(t)
	g.Expect(result).To(Equal([]string{"icons", "icon.png", "icon.icns", "icon.ico"}))
}

func getTestDataPath() string {
	testDataPath, err := filepath.Abs(filepath.Join("..", "..", "testData"))
	Expect(err).NotTo(HaveOccurred())
	return testDataPath
}

var _ = Describe("Blockmap", func() {
	var tmpDir string

	BeforeEach(func() {
		var err error
		tmpDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := os.RemoveAll(tmpDir)
		Expect(err).NotTo(HaveOccurred())
	})

	It("CheckIcoImageSize", func() {
		_, err := doConvertIcon([]string{filepath.Join(getTestDataPath(), "icon.ico")}, nil, "ico", tmpDir)
		Expect(err).NotTo(HaveOccurred())
	})

	It("IcnsToIco", func() {
		files, err := doConvertIcon([]string{filepath.Join(getTestDataPath(), "icon.icns")}, nil, "ico", tmpDir)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(files)).To(Equal(1))
		file := files[0].File

		Expect(strings.HasSuffix(file, ".ico")).To(BeTrue())

		data, err := ioutil.ReadFile(file)
		Expect(err).NotTo(HaveOccurred())
		Expect(GetIcoSizes(data)).To(Equal([]Sizes{
			{Width: 256, Height: 256},
		}))
	})

	It("IcnsToPng", func() {
		result, err := ConvertIcnsToPngUsingOpenJpeg(filepath.Join(getTestDataPath(), "icon.icns"), tmpDir)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(result)).To(Equal(5))
	})

	It("ConvertIcnsToPngUsingOpenJpeg", func() {
		// todo opj_decompress not installed on Travis
		//result, err := ConvertIcnsToPngUsingOpenJpeg(filepath.Join(getTestDataPath(), "icon-jpeg2.icns"), tmpDir)
		//Expect(err).NotTo(HaveOccurred())
		//Expect(len(result)).To(Equal(2))
	})

	It("LargePngTo256Ico", func() {
		files, err := doConvertIcon([]string{filepath.Join(getTestDataPath(), "512x512.png")}, nil, "ico", tmpDir)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(files)).To(Equal(1))
		file := files[0].File
		Expect(strings.HasSuffix(file, ".ico")).To(BeTrue())

		reader, err := os.Open(file)
		Expect(err).NotTo(HaveOccurred())
		defer util.Close(reader)
		images, err := ico.DecodeAll(reader)
		Expect(err).NotTo(HaveOccurred())

		Expect(len(images)).To(Equal(1))

		imageSize := images[0].Bounds().Max
		Expect(imageSize.X).To(Equal(256))
		Expect(imageSize.Y).To(Equal(256))
	})
})
