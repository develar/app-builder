package wine

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestCheckWineVersion(t *testing.T) {
	g := NewGomegaWithT(t)

	err := doCheckWineVersion("1.9.23 (Staging)")
	g.Expect(err).NotTo(HaveOccurred())

	err = doCheckWineVersion("2.0-rc2")
	g.Expect(err).NotTo(HaveOccurred())

	err = doCheckWineVersion("3")
	g.Expect(err).NotTo(HaveOccurred())

	err = doCheckWineVersion("3.5")
	g.Expect(err).NotTo(HaveOccurred())

	err = doCheckWineVersion("1.7")
	g.Expect(err).To(HaveOccurred())
}

//noinspection GoUnusedFunction
func TestCheckWineVersionReal(t *testing.T) {
	t.SkipNow()

	g := NewGomegaWithT(t)

	err := checkWineVersion()
	g.Expect(err).To(HaveOccurred())
}
