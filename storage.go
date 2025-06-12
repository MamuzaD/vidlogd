package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"
)

// sortVideosByLogDate sorts videos by log date, most recent first
func sortVideosByLogDate(videos []Video) {
	sort.Slice(videos, func(i, j int) bool {
		// parse log dates for comparison
		dateI, errI := time.Parse("2006-01-02 3:04 PM", videos[i].LogDate)
		dateJ, errJ := time.Parse("2006-01-02 3:04 PM", videos[j].LogDate)

		// if either date fails to parse, fall back to creation time
		if errI != nil || errJ != nil {
			return videos[i].CreatedAt.After(videos[j].CreatedAt)
		}

		// sort by log date, most recent first
		return dateI.After(dateJ)
	})
}

func loadVideos() ([]Video, error) {
	videosPath, err := getVideosFilePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get videos file path: %w", err)
	}

	if _, err := os.Stat(videosPath); os.IsNotExist(err) {
		// file doesn't exist
		return []Video{}, nil
	}

	// read file
	data, err := os.ReadFile(videosPath)
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

	// sort videos by log date, most recent first
	sortVideosByLogDate(videos)

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

	// sort videos by log date, most recent first
	sortVideosByLogDate(videos)

	// marshal for pretty json
	data, err := json.MarshalIndent(videos, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal videos to JSON: %w", err)
	}

	videosPath, err := getVideosFilePath()
	if err != nil {
		return fmt.Errorf("failed to get videos file path: %w", err)
	}

	if err := os.WriteFile(videosPath, data, 0644); err != nil {
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

	// sort videos by log date, most recent first
	sortVideosByLogDate(videos)

	// marshal for pretty json
	data, err := json.MarshalIndent(videos, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal videos to JSON: %w", err)
	}

	videosPath, err := getVideosFilePath()
	if err != nil {
		return fmt.Errorf("failed to get videos file path: %w", err)
	}

	if err := os.WriteFile(videosPath, data, 0644); err != nil {
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
		Rating:      form.GetRating(),
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

func deleteVideo(id string) error {
	videos, err := loadVideos()
	if err != nil {
		return fmt.Errorf("failed to load existing videos: %w", err)
	}

	// filter out the video with the specified ID
	filteredVideos := make([]Video, 0, len(videos))
	found := false
	for _, video := range videos {
		if video.ID != id {
			filteredVideos = append(filteredVideos, video)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("video with ID %s not found", id)
	}

	// marshal for pretty json
	data, err := json.MarshalIndent(filteredVideos, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal videos to JSON: %w", err)
	}

	videosPath, err := getVideosFilePath()
	if err != nil {
		return fmt.Errorf("failed to get videos file path: %w", err)
	}

	if err := os.WriteFile(videosPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write videos file: %w", err)
	}

	return nil
}

// ============================= settings =============================
func loadSettings() AppSettings {
	settingsPath, err := getSettingsFilePath()
	if err != nil {
		// error getting settings path, return defaults
		return getDefaultSettings()
	}

	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		// file doesn't exist, create w/ default
		defaults := getDefaultSettings()
		saveSettings(defaults)
		return defaults
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return getDefaultSettings()
	}

	var settings AppSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return getDefaultSettings()
	}

	return settings
}

// save settings to file
func saveSettings(settings AppSettings) error {
	if err := ensureSettingsDir(); err != nil {
		return err
	}

	settingsPath, err := getSettingsFilePath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(settingsPath, data, 0644)
}
