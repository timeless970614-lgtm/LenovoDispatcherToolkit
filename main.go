package main

import (
	"embed"
	"log"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Redirect log output to a file for debugging
	logDir := filepath.Join(os.Getenv("LOCALAPPDATA"), "LenovoToolkit")
	os.MkdirAll(logDir, 0755)
	logFile, err := os.OpenFile(filepath.Join(logDir, "toolkit.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		log.SetOutput(logFile)
		defer logFile.Close()
	}
	log.Println("=== LenovoToolkit starting ===")

	app := NewApp()

	// Dedicated WebView2 user data directory — avoids bloat from the default
	// %APPDATA% folder which can accumulate stale cache from other WebView2 apps.
	webviewDataPath := filepath.Join(logDir, "WebView2")

	err = wails.Run(&options.App{
		Title:     "Lenovo Toolkit",
		Width:     1150,
		Height:    750,
		MinWidth:  800,
		MinHeight: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 20, G: 20, B: 20, A: 1},
		StartHidden:      true, // hide until DOM is ready → eliminates flash/skeleton
		OnStartup:        app.startup,
		OnDomReady:       app.onDomReady,
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
			WebviewUserDataPath:  webviewDataPath,
		},
	})
	if err != nil {
		println("Error:", err.Error())
	}
}
