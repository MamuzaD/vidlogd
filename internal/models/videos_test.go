package models

import (
	"strings"
	"testing"
	"time"

	"github.com/mamuzad/vidlogd/internal/storage"
)

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

	t1, _ := time.Parse(DateTimeFormat, "2025-01-01 1:00 PM")
	t2, _ := time.Parse(DateTimeFormat, "2025-01-02 1:00 PM")

	v1 := Video{
		ID:      "v1",
		URL:     "https://example.com/1",
		Title:   "one",
		LogDate: t1,
		Rating:  4.5,
	}
	v2 := Video{
		ID:      "v2",
		URL:     "https://example.com/2",
		Title:   "two",
		LogDate: t2,
		Rating:  5,
	}

	// --- Check count before anything exists
	count, _ := VideoCount(videosPath)
	if count != 0 {
		t.Errorf("expected 0 videos initially, got %d", count)
	}

	// --- Save two videos
	if err := SaveVideo(v1); err != nil {
		t.Fatalf("SaveVideo(v1): %v", err)
	}
	if err := SaveVideo(v2); err != nil {
		t.Fatalf("SaveVideo(v2): %v", err)
	}

	// -- Verify VideoCount sees 2 videos on disk
	count, err = VideoCount(videosPath)
	if err != nil {
		t.Fatalf("VideoCount error: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 videos on disk, got %d", count)
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

	//  Verify VideoCount drops to 1
	count, _ = VideoCount(videosPath)
	if count != 1 {
		t.Errorf("expected 1 video after delete, got %d", count)
	}
	loaded, err = LoadVideos()
	if err != nil {
		t.Fatalf("LoadVideos: %v", err)
	}
	if loaded[0].ID != "v1" {
		t.Fatalf("expected v1 first, got %s", loaded[0].ID)
	}

	// Test finding a non-existent ID
	_, err = FindVideoByID("non-existent")
	if err == nil {
		t.Error("expected error when finding non-existent ID, got nil")
	}

	// Test deleting a non-existent ID
	err = DeleteVideo("non-existent")
	if err == nil {
		t.Error("expected error when deleting non-existent ID, got nil")
	}
}
