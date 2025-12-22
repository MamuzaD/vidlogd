package models

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/mamuzad/vidlogd/internal/storage"
)

// SortVideosByLogDate sorts videos by log date, most recent first
func SortVideosByLogDate(videos []Video) {
	sort.Slice(videos, func(i, j int) bool {
		return videos[i].LogDate.After(videos[j].LogDate)
	})
}

func LoadVideos() ([]Video, error) {
	videosPath, err := storage.VideosPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get videos file path: %w", err)
	}

	data, err := os.ReadFile(videosPath)
	if err != nil {
		if os.IsNotExist(err) {
			// file doesn't exist
			return []Video{}, nil
		}
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

func SaveVideo(video Video) error {
	videos, err := LoadVideos()
	if err != nil {
		return fmt.Errorf("failed to load existing videos: %w", err)
	}

	video.CreatedAt = time.Now()
	if video.ID == "" {
		video.ID = generateVideoID()
	}

	videos = append(videos, video)

	return saveAll(videos)
}

func UpdateVideo(updatedVideo Video) error {
	videos, err := LoadVideos()
	if err != nil {
		return fmt.Errorf("failed to load existing videos: %w", err)
	}

	found := false
	for i := range videos {
		if videos[i].ID == updatedVideo.ID {
			updatedVideo.CreatedAt = videos[i].CreatedAt
			videos[i] = updatedVideo
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("video with ID %s not found", updatedVideo.ID)
	}

	return saveAll(videos)
}

// CreateVideo creates a new video with the given data
func CreateVideo(url, title, channel, releaseDate, logDateStr, review string, rewatched bool, rating float64) Video {
	logDate, err := time.Parse(DateTimeFormat, logDateStr)
	if err != nil {
		panic(fmt.Sprintf("failed to parse logDate '%s' with layout '%s': %v", logDateStr, DateTimeFormat, err))
	}

	return Video{
		ID:          generateVideoID(),
		URL:         url,
		Title:       title,
		Channel:     channel,
		ReleaseDate: releaseDate,
		LogDate:     logDate,
		Review:      review,
		Rewatched:   rewatched,
		Rating:      rating,
		CreatedAt:   time.Now(),
	}
}

func FindVideoByID(id string) (*Video, error) {
	videos, err := LoadVideos()
	if err != nil {
		return nil, err
	}

	for i := range videos {
		if videos[i].ID == id {
			return &videos[i], nil
		}
	}
	return nil, fmt.Errorf("video with ID %s not found", id)
}

func DeleteVideo(id string) error {
	videos, err := LoadVideos()
	if err != nil {
		return fmt.Errorf("failed to load existing videos: %w", err)
	}

	// filter out the video with the specified ID
	filteredVideos := videos[:0]
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

	return saveAll(filteredVideos)
}

func VideoCount() (int, error) {
	videos, err := LoadVideos()
	return len(videos), err
}

func saveAll(videos []Video) error {
	// sort videos by log date, most recent first
	SortVideosByLogDate(videos)

	// marshal for pretty json
	data, err := json.MarshalIndent(videos, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal videos to JSON: %w", err)
	}

	videosPath, err := storage.VideosPath()
	if err != nil {
		return fmt.Errorf("failed to get videos file path: %w", err)
	}

	return storage.WriteFileAtomic(videosPath, data, 0o644)
}
