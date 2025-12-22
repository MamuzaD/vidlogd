package models

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/mamuzad/vidlogd/internal/storage"
)

// helper to read file content as string
func readFileString(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q): %v", path, err)
	}
	return string(b)
}

func TestVideoLifecycle_SaveUpdateDelete(t *testing.T) {
	xdg := t.TempDir()
	t.Setenv("XDG_DATA_HOME", xdg)

	videosPath, err := storage.VideosPath()
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
