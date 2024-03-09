package dmg

import (
	"runtime"
	"testing"

	"github.com/develar/app-builder/pkg/log"
	. "github.com/onsi/gomega"
)

func TestSize(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping not finished test")
		return
	}

	g := NewGomegaWithT(t)

	log.InitLogger()

	w, h, err := getImageSizeUsingSips("/Volumes/data/Desktop/test.png")
	g.Expect(err).To(BeNil())
	g.Expect(w).To(Equal(1316))
	g.Expect(h).To(Equal(894))
}
