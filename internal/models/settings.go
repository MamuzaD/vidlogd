package models

import (
	"encoding/json"
	"os"

	"github.com/mamuzad/vidlogd/internal/storage"
)

// LoadSettings loads settings from file
func LoadSettings() AppSettings {
	settingsPath, err := storage.SettingsPath()
	if err != nil {
		// error getting settings path, return defaults
		return GetDefaultSettings()
	}

	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		// file doesn't exist, create w/ default
		defaults := GetDefaultSettings()
		SaveSettings(defaults)
		return defaults
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return GetDefaultSettings()
	}

	var settings AppSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return GetDefaultSettings()
	}

	return settings
}

// SaveSettings saves settings to file
func SaveSettings(settings AppSettings) error {
	settingsPath, err := storage.SettingsPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	return storage.WriteFileAtomic(settingsPath, data, 0o644)
}
