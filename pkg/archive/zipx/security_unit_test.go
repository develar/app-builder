package zipx

import (
	"os"
	"path/filepath"
	"testing"
)

// Boundary cases for the containment check. outputDir is absolute here so the results are
// independent of the test's working directory.
func TestIsLinkTargetInsideOutputDir(t *testing.T) {
	out := filepath.Clean("/tmp/extract/out")
	ex := &Extractor{outputDir: out}
	sep := string(os.PathSeparator)
	link := func(rel string) string { return filepath.Join(out, rel) }

	cases := []struct {
		name     string
		linkPath string
		target   string
		want     bool
	}{
		{"relative sibling inside", link("a/link"), "b.txt", true},
		{"relative up one still inside", link("a/link"), "../c.txt", true},
		{"relative cleaned stays inside", link("link"), "sub/../keep.txt", true},
		{"absolute inside", link("link"), out + sep + "x", true},
		{"target equals output dir", link("link"), out, true},
		{"escape one level", link("link"), "../out2/y", false},
		{"escape two levels", link("a/link"), "../../etc/passwd", false},
		{"absolute outside", link("link"), "/etc/passwd", false},
		{"sibling prefix trick", link("link"), "/tmp/extract/outevil/y", false},
		{"absolute parent dir", link("link"), "/tmp/extract", false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := ex.isLinkTargetInsideOutputDir(c.linkPath, c.target)
			if got != c.want {
				t.Fatalf("isLinkTargetInsideOutputDir(%q, %q) = %v, want %v", c.linkPath, c.target, got, c.want)
			}
		})
	}
}

// A relative outputDir plus an absolute in-tree target must still be recognized as inside
// (this is exactly the case the absolute-normalization hardening fixes).
func TestIsLinkTargetInsideOutputDir_RelativeOutputDir(t *testing.T) {
	rel := filepath.Clean("relout")
	ex := &Extractor{outputDir: rel}
	absOut, err := filepath.Abs(rel)
	if err != nil {
		t.Fatal(err)
	}
	linkPath := filepath.Join(rel, "link")
	if !ex.isLinkTargetInsideOutputDir(linkPath, filepath.Join(absOut, "inside.txt")) {
		t.Fatalf("absolute in-tree target wrongly rejected for a relative outputDir")
	}
	if ex.isLinkTargetInsideOutputDir(linkPath, "/etc/passwd") {
		t.Fatalf("absolute outside target wrongly accepted for a relative outputDir")
	}
}
