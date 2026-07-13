//go:build darwin

package zipx

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// On APFS/HFS+ (case-insensitive, full Unicode case folding) a regular entry named "ß"
// opens the directory entry an earlier symlink entry "ss" created, because ß folds to ss.
// This is the exact single-archive trigger from the report: symlink "ss" -> external
// victim, padding, then the colliding regular "ß" last. Patched must not overwrite.
func TestUnzipAPFSUnicodeFoldCollision(t *testing.T) {
	work := t.TempDir()
	out := filepath.Join(work, "out")
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	victim := filepath.Join(work, "victim.txt")
	if err := os.WriteFile(victim, []byte("ORIGINAL"), 0o644); err != nil {
		t.Fatal(err)
	}

	zipPath := filepath.Join(work, "fold.zip")
	writeZip(t, zipPath, func(w *zip.Writer) {
		addSymlinkEntry(t, w, "ss", victim)
		for j := 0; j < 300; j++ {
			addFileEntry(t, w, fmt.Sprintf("pad_%04d.txt", j), "x")
		}
		addFileEntry(t, w, "ß", "PWNED_BY_APP_BUILDER") // ß folds onto ss on APFS
	})

	_ = Unzip(zipPath, out, nil)

	if b, _ := os.ReadFile(victim); string(b) != "ORIGINAL" {
		t.Fatalf("APFS ß/ss fold followed the symlink and overwrote the external victim: %q", string(b))
	}
}
