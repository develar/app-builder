package zipx

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/develar/app-builder/pkg/log"
)

// createSymlink logs a warning on rejection; the CLI inits the logger in main, so tests
// must do the same or log.Warn dereferences a nil *zap.Logger.
func TestMain(m *testing.M) {
	log.InitLogger()
	os.Exit(m.Run())
}

func writeZip(t *testing.T, path string, build func(w *zip.Writer)) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	w := zip.NewWriter(f)
	build(w)
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}

func addSymlinkEntry(t *testing.T, w *zip.Writer, name, target string) {
	t.Helper()
	h := &zip.FileHeader{Name: name, Method: zip.Store}
	h.SetMode(os.ModeSymlink | 0o777)
	fw, err := w.CreateHeader(h)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fw.Write([]byte(target)); err != nil {
		t.Fatal(err)
	}
}

func addFileEntry(t *testing.T, w *zip.Writer, name, content string) {
	t.Helper()
	h := &zip.FileHeader{Name: name, Method: zip.Deflate}
	h.SetMode(0o644)
	fw, err := w.CreateHeader(h)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fw.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
}

// Defense 1: a symlink entry whose target escapes the output dir must never be materialized.
func TestUnzipRefusesEscapingSymlink(t *testing.T) {
	work := t.TempDir()
	out := filepath.Join(work, "out")
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	victim := filepath.Join(work, "victim.txt")
	if err := os.WriteFile(victim, []byte("ORIGINAL"), 0o644); err != nil {
		t.Fatal(err)
	}

	zipPath := filepath.Join(work, "escape.zip")
	writeZip(t, zipPath, func(w *zip.Writer) {
		addSymlinkEntry(t, w, "escape", victim) // absolute target outside out/
	})

	// the shared async helper may swallow the returned error, so assert on disk state.
	_ = Unzip(zipPath, out, nil)

	link := filepath.Join(out, "escape")
	if fi, err := os.Lstat(link); err == nil && fi.Mode()&os.ModeSymlink != 0 {
		tgt, _ := os.Readlink(link)
		t.Fatalf("escaping symlink was materialized: %s -> %s", link, tgt)
	}
}

// Defense 2: a regular write must not follow a symlink already sitting at the destination.
func TestUnzipDoesNotWriteThroughSymlink(t *testing.T) {
	work := t.TempDir()
	out := filepath.Join(work, "out")
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	victim := filepath.Join(work, "victim.txt")
	if err := os.WriteFile(victim, []byte("ORIGINAL"), 0o644); err != nil {
		t.Fatal(err)
	}
	// pre-plant a symlink at the destination name (stands in for one an earlier entry made).
	if err := os.Symlink(victim, filepath.Join(out, "pwn")); err != nil {
		t.Fatal(err)
	}

	zipPath := filepath.Join(work, "reg.zip")
	writeZip(t, zipPath, func(w *zip.Writer) {
		addFileEntry(t, w, "pwn", "PAYLOAD")
	})

	_ = Unzip(zipPath, out, nil)

	got, _ := os.ReadFile(victim)
	if string(got) != "ORIGINAL" {
		t.Fatalf("victim outside out/ was overwritten through a symlink: %q", string(got))
	}
}

// The single-archive race (a symlink entry plus a colliding regular entry). Patched must
// never overwrite, across many concurrent schedulings.
func TestUnzipSymlinkRegularRace(t *testing.T) {
	const iterations = 100
	overwrites := 0
	for i := 0; i < iterations; i++ {
		work := t.TempDir()
		out := filepath.Join(work, "out")
		if err := os.MkdirAll(out, 0o755); err != nil {
			t.Fatal(err)
		}
		victim := filepath.Join(work, "victim.txt")
		if err := os.WriteFile(victim, []byte("ORIGINAL"), 0o644); err != nil {
			t.Fatal(err)
		}
		zipPath := filepath.Join(work, "race.zip")
		writeZip(t, zipPath, func(w *zip.Writer) {
			addSymlinkEntry(t, w, "collide", victim)
			for j := 0; j < 50; j++ {
				addFileEntry(t, w, fmt.Sprintf("pad_%03d.txt", j), "x")
			}
			addFileEntry(t, w, "collide", "PAYLOAD")
		})
		_ = Unzip(zipPath, out, nil)
		if b, _ := os.ReadFile(victim); string(b) == "PAYLOAD" {
			overwrites++
		}
	}
	t.Logf("race overwrites: %d/%d", overwrites, iterations)
	if overwrites > 0 {
		t.Fatalf("victim overwritten %d/%d times across schedulings", overwrites, iterations)
	}
}

// Benign archives (files, a subdir, an in-tree relative symlink) must still extract.
func TestUnzipBenignStillWorks(t *testing.T) {
	work := t.TempDir()
	out := filepath.Join(work, "out")
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	zipPath := filepath.Join(work, "benign.zip")
	writeZip(t, zipPath, func(w *zip.Writer) {
		addFileEntry(t, w, "a.txt", "hello")
		addFileEntry(t, w, "sub/b.txt", "world")
		addSymlinkEntry(t, w, "link_to_a", "a.txt") // relative, stays inside
	})
	if err := Unzip(zipPath, out, nil); err != nil {
		t.Fatalf("benign extract failed: %v", err)
	}
	if b, _ := os.ReadFile(filepath.Join(out, "a.txt")); string(b) != "hello" {
		t.Fatalf("a.txt = %q", string(b))
	}
	if b, _ := os.ReadFile(filepath.Join(out, "sub", "b.txt")); string(b) != "world" {
		t.Fatalf("sub/b.txt = %q", string(b))
	}
	if b, _ := os.ReadFile(filepath.Join(out, "link_to_a")); string(b) != "hello" {
		t.Fatalf("in-tree symlink did not resolve: %q", string(b))
	}
}
