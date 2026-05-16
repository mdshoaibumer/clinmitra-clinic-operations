package main

import (
	"embed"
	"log/slog"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"

	"clinmitra/internal/app"
)

//go:embed all:frontend/dist
var assets embed.FS

// main is the application entry point. It initializes the Application,
// configures the Wails runtime with window settings and asset server,
// and starts the desktop application event loop.
func main() {
	application, err := app.NewApplication()
	if err != nil {
		slog.Error("failed to initialize application", "error", err)
		os.Exit(1)
	}

	err = wails.Run(&options.App{
		Title:     "Clinmitra Dental - Dental Clinic Management",
		Width:     1280,
		Height:    800,
		MinWidth:  1024,
		MinHeight: 700,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup:  application.Startup,
		OnShutdown: application.Shutdown,
		Bind:       application.GetBindings(),
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
		},
	})

	if err != nil {
		slog.Error("error starting application", "error", err)
		os.Exit(1)
	}
}
