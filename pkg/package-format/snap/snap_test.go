package snap

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestCheckWineVersion(t *testing.T) {
	g := NewGomegaWithT(t)

	err := doCheckSnapVersion("4.0", "")
	g.Expect(err).NotTo(HaveOccurred())

	err = doCheckSnapVersion("snapcraft, version 4.0.0", "")
	g.Expect(err).NotTo(HaveOccurred())

	err = doCheckSnapVersion("snapcraft, version '4.0.0'", "")
	g.Expect(err).NotTo(HaveOccurred())

	err = doCheckSnapVersion(" version 4.1.1", "")
	g.Expect(err).NotTo(HaveOccurred())

	err = doCheckSnapVersion("3.1", "")
	g.Expect(err).To(HaveOccurred())
}