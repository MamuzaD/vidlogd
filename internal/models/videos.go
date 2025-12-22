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
	SortVideosByLogDate(videos)

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

	if err := storage.WriteFileAtomic(videosPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write videos file: %w", err)
	}

	return nil
}

func UpdateVideo(updatedVideo Video) error {
	videos, err := LoadVideos()
	if err != nil {
		return fmt.Errorf("failed to load existing videos: %w", err)
	}

	existingVideo, err := FindVideoByID(updatedVideo.ID)
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

	if err := storage.WriteFileAtomic(videosPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write videos file: %w", err)
	}

	return nil
}

// CreateVideo creates a new video with the given data
func CreateVideo(url, title, channel, releaseDate, logDateStr, review string, rewatched bool, rating float64) Video {
	logDate, err := time.Parse(DateTimeFormat, logDateStr)
	if err != nil {
		panic(err)
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

	videosPath, err := storage.VideosPath()
	if err != nil {
		return fmt.Errorf("failed to get videos file path: %w", err)
	}

	if err := storage.WriteFileAtomic(videosPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write videos file: %w", err)
	}

	return nil
}

func VideoCount(path string) (int, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to open video file: %w", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil || info.Size() == 0 {
		return 0, nil
	}

	var count []struct{}
	if err := json.NewDecoder(file).Decode(&count); err != nil {
		return 0, fmt.Errorf("failed to parse video count: %w", err)
	}

	return len(count), nil
}
