package blockmap_test

import (
	"crypto/sha512"
	"encoding/base64"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/json-iterator/go"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/develar/app-builder/pkg/blockmap"
)

func TestBlockmap(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Blockmap Suite")
}

var _ = Describe("Blockmap", func() {
	It("append", func() {
		file, err := ioutil.TempFile("", "append")
		Expect(err).NotTo(HaveOccurred())

		_, err = file.WriteString(strings.Repeat("hello world. ", 1024))
		Expect(err).NotTo(HaveOccurred())
		err = file.Close()
		Expect(err).NotTo(HaveOccurred())

		inputInfo, err := BuildBlockMap(file.Name(), DefaultChunkerConfiguration, DEFLATE, "")
		Expect(err).NotTo(HaveOccurred())

		fileData, err := ioutil.ReadFile(file.Name())
		Expect(err).NotTo(HaveOccurred())

		hash := sha512.New()
		_, err = hash.Write(fileData)
		Expect(err).NotTo(HaveOccurred())
		Expect(inputInfo.Sha512).To(Equal(base64.StdEncoding.EncodeToString(hash.Sum(nil))))
		Expect(inputInfo.Size).To(Equal(len(fileData)))

		serializedInputInfo, err := jsoniter.ConfigFastest.Marshal(inputInfo)
		Expect(err).NotTo(HaveOccurred())
		//noinspection SpellCheckingInspection
		Expect(string(serializedInputInfo)).To(Equal("{\"size\":13423,\"sha512\":\"zPFW3WAFUKFvAfBdNXHDIuZekSW/qf33lf5OgKXBKg9oOobwVH9X/DRHExC9087Cxkp3nqFrwtreWZHLso3D6g==\",\"blockMapSize\":107}"))
	})
})
