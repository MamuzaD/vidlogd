package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"
	"github.com/mamuzad/vidlogd/internal/models"
)

type YouTubeMetadata struct {
	Title       string
	Creator     string
	ReleaseDate string
}

type MetadataFetchedMsg struct {
	Metadata YouTubeMetadata
	Error    string
}

type YouTubeAPIResponse struct {
	Items []struct {
		Snippet struct {
			Title        string `json:"title"`
			ChannelTitle string `json:"channelTitle"`
			PublishedAt  string `json:"publishedAt"`
		} `json:"snippet"`
	} `json:"items"`
}

var youtubeAPIKey string

func loadYouTubeAPI() {
	if youtubeAPIKey != "" {
		return
	}

	godotenv.Load()
	youtubeAPIKey = os.Getenv("YOUTUBE_API_KEY")

	// if not found in environment, try settings
	if youtubeAPIKey == "" {
		settings := models.LoadSettings()
		youtubeAPIKey = settings.APIKey
	}
}

func getYouTubeAPIKey() string {
	// use latest settings
	settings := models.LoadSettings()
	if settings.APIKey != "" {
		return settings.APIKey
	}

	// fallback to env
	if youtubeAPIKey == "" {
		loadYouTubeAPI()
	}
	return youtubeAPIKey
}

func IsValidYouTubeURL(urlStr string) bool {
	if urlStr == "" {
		return false
	}

	u, err := neturl.Parse(urlStr)
	if err != nil {
		return false
	}

	domain := strings.ToLower(u.Host)
	if domain != "www.youtube.com" && domain != "youtube.com" && domain != "youtu.be" && domain != "m.youtube.com" {
		return false
	}

	if domain == "youtu.be" {
		return len(u.Path) > 1
	}

	if strings.HasPrefix(u.Path, "/watch") {
		return u.Query().Get("v") != ""
	}

	if strings.HasPrefix(u.Path, "/embed/") {
		return len(u.Path) > 7
	}

	return false
}

func extractVideoID(urlStr string) string {
	u, err := neturl.Parse(urlStr)
	if err != nil {
		return ""
	}

	domain := strings.ToLower(u.Host)

	if domain == "youtu.be" {
		return strings.TrimPrefix(u.Path, "/")
	}

	if strings.HasPrefix(u.Path, "/watch") {
		return u.Query().Get("v")
	}

	if strings.HasPrefix(u.Path, "/embed/") {
		return strings.TrimPrefix(u.Path, "/embed/")
	}

	return ""
}

// fetch youtube metadata and parse it
//
// returns MetadataFetchedMsg containing the Metadata (video title, channel, release date)
func FetchYouTubeMetadata(urlStr string) tea.Cmd {
	return func() tea.Msg {
		youtubeAPIKey := getYouTubeAPIKey()

		if youtubeAPIKey == "" {
			return MetadataFetchedMsg{Error: "add YOUTUBE_API_KEY to your .env file or set it in settings"}
		}

		if !IsValidYouTubeURL(urlStr) {
			return MetadataFetchedMsg{Error: "invalid YouTube URL"}
		}

		videoID := extractVideoID(urlStr)
		if videoID == "" {
			return MetadataFetchedMsg{Error: "could not extract video ID"}
		}

		apiURL := fmt.Sprintf(
			"https://www.googleapis.com/youtube/v3/videos?part=snippet&id=%s&key=%s",
			videoID,
			youtubeAPIKey,
		)

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Get(apiURL)
		if err != nil {
			return MetadataFetchedMsg{Error: "failed to fetch video data: " + err.Error()}
		}
		defer resp.Body.Close()

		if resp.StatusCode == 403 {
			return MetadataFetchedMsg{Error: "quota exceeded or invalid key"}
		}

		if resp.StatusCode != http.StatusOK {
			return MetadataFetchedMsg{Error: fmt.Sprintf("youtube error: %d", resp.StatusCode)}
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return MetadataFetchedMsg{Error: "failed to read response"}
		}

		var apiResponse YouTubeAPIResponse
		if err := json.Unmarshal(body, &apiResponse); err != nil {
			return MetadataFetchedMsg{Error: "failed to parse response"}
		}

		if len(apiResponse.Items) == 0 {
			return MetadataFetchedMsg{Error: "video not found"}
		}

		snippet := apiResponse.Items[0].Snippet

		publishedDate := ""
		if snippet.PublishedAt != "" {
			if t, err := time.Parse(time.RFC3339, snippet.PublishedAt); err == nil {
				publishedDate = t.Format(models.ISODateFormat)
			}
		}

		metadata := YouTubeMetadata{
			Title:       snippet.Title,
			Creator:     snippet.ChannelTitle,
			ReleaseDate: publishedDate,
		}

		return MetadataFetchedMsg{Metadata: metadata}
	}
}

