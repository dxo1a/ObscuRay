package backend

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

var Profiles []Profile
var profilesFile = filepath.Join(os.Getenv("APPDATA"), "ObscuRay", "profiles.json")

func LoadProfiles() {
	os.MkdirAll(filepath.Dir(profilesFile), 0755)
	data, err := os.ReadFile(profilesFile)
	if err == nil {
		json.Unmarshal(data, &Profiles)
	}
}

func saveProfiles() error {
	data, err := json.MarshalIndent(Profiles, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(profilesFile, data, 0644)
}

func AddProfileFromClipboard() error {
	vless, err := clipboard.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read clipboard: %v", err)
	}
	if !isValidVLESS(vless) {
		return fmt.Errorf("invalid VLESS URL")
	}
	id := fmt.Sprintf("%d", len(Profiles)+1)
	name := parseVLESSName(vless)
	Profiles = append(Profiles, Profile{
		ID:       id,
		Name:     name,
		VLESS:    vless,
		IsActive: false,
	})
	return saveProfiles()
}

func findProfile(id string) *Profile {
	for i := range Profiles {
		if Profiles[i].ID == id {
			return &Profiles[i]
		}
	}
	return nil
}

func DeleteProfile(id string) error {
	for i, profile := range Profiles {
		if profile.ID == id {
			if profile.IsActive {
				return fmt.Errorf("cannot delete active profile")
			}
			Profiles = append(Profiles[:i], Profiles[i+1:]...)
			return saveProfiles()
		}
	}
	return fmt.Errorf("profile not found")
}

func CopyVLESS(id string) error {
	profile := findProfile(id)
	if profile == nil {
		return fmt.Errorf("profile not found")
	}
	err := clipboard.WriteAll(profile.VLESS)
	if err != nil {
		return fmt.Errorf("failed to copy VLESS: %v", err)
	}
	return nil
}
