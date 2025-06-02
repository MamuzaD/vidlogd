package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

const videosFile = "videos.json"

func loadVideos() ([]Video, error) {
	if _, err := os.Stat(videosFile); os.IsNotExist(err) {
		// file doesn't exist
		return []Video{}, nil
	}

	// read file
	data, err := os.ReadFile(videosFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read videos file: %w", err)
	}

	if len(data) == 0 {
		return []Video{}, nil
	}

	var videos []Video
	if err := json.Unmarshal(data, &videos); err != nil {
		return nil, fmt.Errorf("failed to parse videos file: %w", err)
	}

	return videos, nil
}

func generateVideoID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func saveVideo(video Video) error {
	videos, err := loadVideos()
	if err != nil {
		return fmt.Errorf("failed to load existing videos: %w", err)
	}

	video.CreatedAt = time.Now()
	if video.ID == "" {
		video.ID = generateVideoID()
	}

	videos = append(videos, video)

	// marshal for pretty json
	data, err := json.MarshalIndent(videos, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal videos to JSON: %w", err)
	}

	if err := os.WriteFile(videosFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write videos file: %w", err)
	}

	return nil
}

func updateVideo(updatedVideo Video) error {
	videos, err := loadVideos()
	if err != nil {
		return fmt.Errorf("failed to load existing videos: %w", err)
	}

	existingVideo, err := findVideoByID(updatedVideo.ID)
	if err != nil {
		return err
	}

	updatedVideo.CreatedAt = existingVideo.CreatedAt

	for i, video := range videos {
		if video.ID == updatedVideo.ID {
			videos[i] = updatedVideo
			break
		}
	}

	// marshal for pretty json
	data, err := json.MarshalIndent(videos, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal videos to JSON: %w", err)
	}

	if err := os.WriteFile(videosFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write videos file: %w", err)
	}

	return nil
}

// helper for saving for forms
func createVideoFromForm(form FormModel) Video {
	return Video{
		URL:         form.GetValue(url),
		Title:       form.GetValue(title),
		Channel:     form.GetValue(channel),
		ReleaseDate: form.GetValue(release),
		LogDate:     form.GetValue(logDate),
		Review:      form.GetValue(review),
	}
}

func findVideoByID(id string) (*Video, error) {
	videos, err := loadVideos()
	if err != nil {
		return nil, err
	}

	for _, video := range videos {
		if video.ID == id {
			return &video, nil
		}
	}

	return nil, fmt.Errorf("video with ID %s not found", id)
}
