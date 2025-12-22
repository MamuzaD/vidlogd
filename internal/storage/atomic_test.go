package storage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// helper to assert no temp files are left in a directory
func assertNoTempFiles(t *testing.T, dir string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir(%q): %v", dir, err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "tmp-") {
			t.Fatalf("leftover temp file: %s", filepath.Join(dir, e.Name()))
		}
	}
}

// helper to read file content as string
func readFileString(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q): %v", path, err)
	}
	return string(b)
}

func TestWriteFileAtomic_CreatesParentAndReplacesExisting(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	target := filepath.Join(tmp, "a", "b", "c.json")

	data1 := []byte(`{"v":1}`)
	if err := WriteFileAtomic(target, data1, 0o644); err != nil {
		t.Fatalf("first write: %v", err)
	}

	if got := readFileString(t, target); got != string(data1) {
		t.Fatalf("mismatch after first write: got %q want %q", got, data1)
	}

	fi, err := os.Stat(target)
	if err != nil {
		t.Fatalf("Stat(%q): %v", target, err)
	}
	if fi.Mode().Perm() != 0o644 {
		t.Fatalf("perm mismatch: got %o want %o", fi.Mode().Perm(), 0o644)
	}

	data2 := []byte(`{"v":2}`)
	if err := WriteFileAtomic(target, data2, 0o644); err != nil {
		t.Fatalf("second write: %v", err)
	}

	if got := readFileString(t, target); got != string(data2) {
		t.Fatalf("mismatch after replace: got %q want %q", got, data2)
	}

	assertNoTempFiles(t, filepath.Dir(target))
}

func TestWriteFileAtomic_CleansUpTempOnRenameError(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	parent := filepath.Join(tmp, "parent")
	if err := os.MkdirAll(parent, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	// Force rename error: make target a directory
	target := filepath.Join(parent, "target")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatalf("MkdirAll target: %v", err)
	}

	err := WriteFileAtomic(target, []byte("data"), 0o644)
	if err == nil {
		t.Fatalf("expected error but got nil")
	}

	assertNoTempFiles(t, parent)
}
