package main

import (
	"ObscuRay/backend"
	"context"
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	win "golang.org/x/sys/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed assets/running.ico
var runningIcon embed.FS

//go:embed assets/stopped.ico
var stoppedIcon embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()

	backend.LoadProfiles()

	//region tray
	runningIconData, error := runningIcon.ReadFile("assets/running.ico")
	if error != nil {
		println("Error loading running.ico:", error.Error())
	}
	stoppedIconData, error := stoppedIcon.ReadFile("assets/stopped.ico")
	if error != nil {
		println("Error loading stopped.ico:", error.Error())
	}
	backend.SetIcons(runningIconData, stoppedIconData)

	go setupTray(app)
	//endregion

	if err := setShutdownPriority(); err != nil {
		log.Println("Error settings shutdown priority:", err.Error())
	}

	// Create application with options
	err := wails.Run(&options.App{
		Title:     "ObscuRay",
		Width:     700,
		Height:    400,
		Frameless: true,
		Windows: &windows.Options{
			WebviewIsTransparent:              false,
			WindowIsTranslucent:               false,
			DisableFramelessWindowDecorations: false,
		},
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		OnBeforeClose: func(ctx context.Context) bool {
			if app.quitRequested {
				return false
			}
			runtime.WindowHide(ctx)
			return true
		},
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

func setShutdownPriority() error {
	kernel32 := win.NewLazyDLL("kernel32.dll")
	setProcessShutdownParameters := kernel32.NewProc("SetProcessShutdownParameters")

	ret, _, err := setProcessShutdownParameters.Call(uintptr(0x3FF), 0)
	if ret == 0 {
		return err
	}
	return nil
}
