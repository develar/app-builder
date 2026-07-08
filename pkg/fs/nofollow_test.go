package fs

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// A regular write must refuse to follow a symlink sitting at the destination (O_NOFOLLOW),
// so the file the symlink points at outside the tree is never touched.
func TestWriteFileRefusesSymlinkDestination(t *testing.T) {
	work := t.TempDir()
	victim := filepath.Join(work, "victim.txt")
	if err := os.WriteFile(victim, []byte("ORIGINAL"), 0o644); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(work, "dest")
	if err := os.Symlink(victim, dest); err != nil {
		t.Fatal(err)
	}

	buffer := make([]byte, 32*1024)
	err := WriteFileAndRestoreNormalPermissions(bytes.NewReader([]byte("PAYLOAD")), dest, 0o644, buffer)

	if got, _ := os.ReadFile(victim); string(got) != "ORIGINAL" {
		t.Fatalf("write followed the symlink and overwrote the victim: %q (err=%v)", string(got), err)
	}
	if err == nil {
		t.Fatalf("expected an error when the destination is a symlink, got nil")
	}
}

// A benign regular write (creating parent dirs as needed) must still work.
func TestWriteFileWritesRegularFile(t *testing.T) {
	work := t.TempDir()
	dest := filepath.Join(work, "sub", "out.txt")
	buffer := make([]byte, 32*1024)
	if err := WriteFileAndRestoreNormalPermissions(bytes.NewReader([]byte("hello")), dest, 0o644, buffer); err != nil {
		t.Fatalf("benign write failed: %v", err)
	}
	if got, _ := os.ReadFile(dest); string(got) != "hello" {
		t.Fatalf("got %q, want hello", string(got))
	}
}
