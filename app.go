package main

import (
	"context"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	loadProfiles()
}

func (a *App) shutdown(ctx context.Context) {
	stopProfile()
}

func (a *App) GetProfiles() []Profile {
	return profiles
}

func (a *App) AddProfileFromClipboard() error {
	return addProfileFromClipboard()
}

func (a *App) StartProfile(id string) error {
	return startProfile(id)
}

func (a *App) StopProfile() error {
	return stopProfile()
}
