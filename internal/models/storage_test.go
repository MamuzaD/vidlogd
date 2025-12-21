package models

import (
	"encoding/json"
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

func TestVideoLifecycle_SaveUpdateDelete(t *testing.T) {
	xdg := t.TempDir()
	t.Setenv("XDG_DATA_HOME", xdg)

	videosPath, err := getVideosFilePath()
	if err != nil {
		t.Fatalf("getVideosFilePath: %v", err)
	}
	if !strings.Contains(videosPath, xdg) {
		t.Fatalf("expected videos path inside %q, got %q", xdg, videosPath)
	}

	v1 := Video{
		ID:      "v1",
		URL:     "https://example.com/1",
		Title:   "one",
		LogDate: "2025-01-01 1:00 PM",
		Rating:  4.5,
	}
	v2 := Video{
		ID:      "v2",
		URL:     "https://example.com/2",
		Title:   "two",
		LogDate: "2025-01-02 1:00 PM",
		Rating:  5,
	}

	// --- Save two videos
	if err := SaveVideo(v1); err != nil {
		t.Fatalf("SaveVideo(v1): %v", err)
	}
	if err := SaveVideo(v2); err != nil {
		t.Fatalf("SaveVideo(v2): %v", err)
	}

	raw := readFileString(t, videosPath)
	var onDisk []Video
	if err := json.Unmarshal([]byte(raw), &onDisk); err != nil {
		t.Fatalf("invalid JSON: %v\nraw:\n%s", err, raw)
	}
	if len(onDisk) != 2 {
		t.Fatalf("expected 2 videos, got %d", len(onDisk))
	}

	// --- Load & verify order
	loaded, err := LoadVideos()
	if err != nil {
		t.Fatalf("LoadVideos: %v", err)
	}
	if len(loaded) != 2 {
		t.Fatalf("expected 2 videos, got %d", len(loaded))
	}
	if loaded[0].ID != "v2" {
		t.Fatalf("expected v2 first, got %s", loaded[0].ID)
	}

	// --- Update v1 title
	v1.Title = "updated title"
	if err := UpdateVideo(v1); err != nil {
		t.Fatalf("UpdateVideo: %v", err)
	}

	updated, err := FindVideoByID("v1")
	if err != nil {
		t.Fatalf("FindVideoByID: %v", err)
	}
	if updated.Title != "updated title" {
		t.Fatalf("expected updated title, got %q", updated.Title)
	}

	// --- Delete v2
	if err := DeleteVideo("v2"); err != nil {
		t.Fatalf("DeleteVideo: %v", err)
	}

	remaining, err := LoadVideos()
	if err != nil {
		t.Fatalf("LoadVideos after delete: %v", err)
	}
	if len(remaining) != 1 || remaining[0].ID != "v1" {
		t.Fatalf("unexpected remaining videos: %+v", remaining)
	}
}
