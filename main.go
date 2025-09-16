package main

import (
	"ObscuRay/backend"
	"context"
	"crypto/sha256"
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"

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

//go:embed assets/sing-box.exe
var singBoxEmbed embed.FS

func main() {
	logFile, err := os.OpenFile(filepath.Join(os.Getenv("APPDATA"), "ObscuRay", "app.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		println("Failed to open log file:", err.Error())
	}
	log.SetOutput(logFile)

	singBoxPath, err := extractSingBox()
	if err != nil {
		log.Fatalf("Failed to extract sing-box.exe: %v", err)
	}
	backend.SetSingBoxPath(singBoxPath)

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
	err = wails.Run(&options.App{
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

func extractSingBox() (string, error) {
	cacheDir := filepath.Join(os.Getenv("APPDATA"), "ObscuRay", "cache")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %v", err)
	}
	cachePath := filepath.Join(cacheDir, "sing-box.exe")

	singBoxData, err := singBoxEmbed.ReadFile("assets/sing-box.exe")
	if err != nil {
		return "", fmt.Errorf("failed to read sing-box.exe from assets: %v", err)
	}

	expectedHash := sha256.Sum256(singBoxData)

	// Check if file exists and hash matches
	if _, err := os.Stat(cachePath); err == nil {
		fileData, err := os.ReadFile(cachePath)
		if err != nil {
			log.Printf("Failed to read cached sing-box.exe: %v", err)
		} else {
			actualHash := sha256.Sum256(fileData)
			if actualHash == expectedHash {
				log.Println("Using existing cached sing-box.exe:", cachePath)
				return cachePath, nil
			}
			log.Println("Cached sing-box.exe hash mismatch, overwriting...")
		}
	} else {
		log.Println("Cached sing-box.exe not found, creating new...")
	}

	// Write or overwrite the file
	if err := os.WriteFile(cachePath, singBoxData, 0755); err != nil {
		return "", fmt.Errorf("failed to write sing-box.exe to cache: %v", err)
	}
	if err := os.Chmod(cachePath, 0755); err != nil {
		return "", fmt.Errorf("failed to chmod sing-box.exe: %v", err)
	}

	log.Println("sing-box.exe cached at:", cachePath)
	return cachePath, nil
}
