package models

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func DataDir() (string, error) {
	var baseDir string

	// Follow XDG Base Directory Specification on Unix-like systems
	if xdgDataHome := os.Getenv("XDG_DATA_HOME"); xdgDataHome != "" {
		baseDir = xdgDataHome
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home directory: %w", err)
		}

		switch runtime.GOOS {
		case "windows":
			if appData := os.Getenv("APPDATA"); appData != "" {
				baseDir = appData
			} else {
				baseDir = filepath.Join(homeDir, "AppData", "Roaming")
			}
		default: // Linux and other Unix-like systems
			baseDir = filepath.Join(homeDir, ".local", "share")
		}
	}

	// Create application-specific directory
	dataDir := filepath.Join(baseDir, "vidlogd")

	// Ensure directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create data directory: %w", err)
	}

	return dataDir, nil
}

func VideosFilePath() (string, error) {
	dataDir, err := DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dataDir, "videos.json"), nil
}

func SettingsFilePath() (string, error) {
	dataDir, err := DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dataDir, "settings.json"), nil
}
