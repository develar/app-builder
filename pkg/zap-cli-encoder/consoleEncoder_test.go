package zap_cli_encoder

import (
	"strings"
	"testing"

	. "github.com/onsi/gomega"
	"go.uber.org/zap/buffer"
)

func TestAppend(t *testing.T) {
	var linePool = buffer.NewPool()
	buf := linePool.Get()
	appendPaddedString("a\nb", buf)

	g := NewGomegaWithT(t)
	g.Expect(buf.String()).To(Equal("a\n" + strings.Repeat(" ", levelIndicatorRuneCount) + "b"))
}
