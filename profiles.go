package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/atotto/clipboard"
)

type Profile struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	VLESS    string `json:"vless"`
	IsActive bool   `json:"isActive"`
}

var profiles []Profile
var profilesFile = filepath.Join(os.Getenv("APPDATA"), "ObscuRay", "profiles.json")

func loadProfiles() {
	os.MkdirAll(filepath.Dir(profilesFile), 0755)
	data, err := os.ReadFile(profilesFile)
	if err == nil {
		json.Unmarshal(data, &profiles)
	}
}

func saveProfiles() error {
	data, err := json.MarshalIndent(profiles, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(profilesFile, data, 0644)
}

func addProfileFromClipboard() error {
	vless, err := clipboard.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read clipboard: %v", err)
	}
	if !isValidVLESS(vless) {
		return fmt.Errorf("invalid VLESS URL")
	}
	id := fmt.Sprintf("%d", len(profiles)+1)
	name := parseVLESSName(vless)
	profiles = append(profiles, Profile{
		ID:       id,
		Name:     name,
		VLESS:    vless,
		IsActive: false,
	})
	return saveProfiles()
}

func findProfile(id string) *Profile {
	for i := range profiles {
		if profiles[i].ID == id {
			return &profiles[i]
		}
	}
	return nil
}
