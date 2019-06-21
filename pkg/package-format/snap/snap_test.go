package snap

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestCheckWineVersion(t *testing.T) {
	g := NewGomegaWithT(t)

	err := doCheckSnapVersion("3.1", "")
	g.Expect(err).NotTo(HaveOccurred())

	err = doCheckSnapVersion("snapcraft, version 3.1.1", "")
	g.Expect(err).NotTo(HaveOccurred())

	err = doCheckSnapVersion("snapcraft, version '3.1.1'", "")
	g.Expect(err).NotTo(HaveOccurred())

	err = doCheckSnapVersion(" version 3.2.1", "")
	g.Expect(err).NotTo(HaveOccurred())

	err = doCheckSnapVersion("2.12", "")
	g.Expect(err).To(HaveOccurred())
}