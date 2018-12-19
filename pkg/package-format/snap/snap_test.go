package snap

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestCheckWineVersion(t *testing.T) {
	g := NewGomegaWithT(t)

	err := doCheckSnapVersion("3.0", "")
	g.Expect(err).NotTo(HaveOccurred())

	err = doCheckSnapVersion("snapcraft, version 3.0.1", "")
	g.Expect(err).NotTo(HaveOccurred())

	err = doCheckSnapVersion(" version 3.0.1", "")
	g.Expect(err).NotTo(HaveOccurred())

	err = doCheckSnapVersion("2.12", "")
	g.Expect(err).To(HaveOccurred())
}