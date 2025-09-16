package main

import (
	"ObscuRay/backend"
	"context"
	"os"

	"github.com/getlantern/systray"
	"github.com/wailsapp/wails/v2/pkg/runtime"
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

func setupTray(app *App) {
	lang := backend.GetLang()
	t := backend.Translations[lang]

	systray.Run(func() {
		systray.SetTitle("ObscuRay")
		systray.SetTooltip(t["title"])

		mShow := systray.AddMenuItem(t["show"], t["showDesc"])
		mQuit := systray.AddMenuItem(t["quit"], t["quitDesc"])

		activeProfileExists := false
		for _, p := range backend.Profiles {
			if p.IsActive {
				activeProfileExists = true
				break
			}
		}

		if activeProfileExists && backend.RunningIconData != nil {
			systray.SetIcon(backend.RunningIconData)
		} else if backend.StoppedIconData != nil {
			systray.SetIcon(backend.StoppedIconData)
		}

		go func() {
			for {
				select {
				case <-mShow.ClickedCh:
					runtime.WindowShow(app.ctx)
				case <-mQuit.ClickedCh:
					app.quitRequested = true
					runtime.Quit(app.ctx)
				}
			}
		}()
	}, func() {
		backend.StopProfile()
		os.Exit(0)
	})
}
