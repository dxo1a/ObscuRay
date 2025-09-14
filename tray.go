package main

import (
	"os"

	"github.com/getlantern/systray"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func setupTray(app *App) {
	systray.Run(func() {
		systray.SetTitle("ObscuRay")
		systray.SetTooltip("ObscuRay VLESS Proxy")
		mShow := systray.AddMenuItem("Show", "Show application")
		mQuit := systray.AddMenuItem("Quit", "Quit application")

		go func() {
			for {
				select {
				case <-mShow.ClickedCh:
					runtime.WindowShow(app.ctx)
				case <-mQuit.ClickedCh:
					runtime.Quit(app.ctx)
				}
			}
		}()
	}, func() {
		stopProfile()
		os.Exit(0)
	})
}
