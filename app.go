package main

import (
	"ObscuRay/backend"
	"context"
)

// App struct
type App struct {
	ctx           context.Context
	quitRequested bool
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	backend.LoadProfiles()
}

func (a *App) shutdown(ctx context.Context) {
	backend.StopProfile()
}

func (a *App) GetProfiles() []backend.Profile {
	return backend.Profiles
}

func (a *App) AddProfileFromClipboard() error {
	return backend.AddProfileFromClipboard()
}

func (a *App) StartProfile(id string) error {
	return backend.StartProfile(id)
}

func (a *App) StopProfile() error {
	return backend.StopProfile()
}

func (a *App) DeleteProfile(id string) error {
	return backend.DeleteProfile(id)
}

func (a *App) GetTrafficStats() (map[string]int64, error) {
	return backend.GetTrafficStats()
}

func (a *App) CopyVLESS(id string) error {
	return backend.CopyVLESS(id)
}
