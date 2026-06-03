package main

import (
	"context"
	"embed"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/getlantern/systray"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed mini-status.html
var miniFS embed.FS

var miniAppCtx context.Context

func main() {
	app := NewMiniApp()

	// Setup systray in a separate goroutine before wails.Run blocks
	go func() {
		systray.Run(app.onSystrayReady, app.onSystrayExit)
	}()

	// Build a simple HTTP handler: /mini → HTML, everything else → embedded assets
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		// Serve mini.html for /mini or /mini.html
		if path == "/mini" || path == "/mini.html" || path == "/mini/" {
			data, err := miniFS.Open("mini-status.html")
			if err != nil {
				http.NotFound(w, r)
				return
			}
			defer data.Close()
			content, _ := io.ReadAll(data)
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(content)
			return
		}
		// Serve from embedded FS
		file, err := miniFS.Open(strings.TrimPrefix(path, "/"))
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer file.Close()
		content, _ := io.ReadAll(file)
		w.Write(content)
	})

	err := wails.Run(&options.App{
		Title:            "GPU Status",
		Width:            285,
		Height:           230,
		AlwaysOnTop:      true,
		DisableResize:    true,
		Frameless:        true,
		StartHidden:      true,
		BackgroundColour: &options.RGBA{R: 26, G: 26, B: 26, A: 255},
		OnStartup:        app.startup,
		OnDomReady:       app.onDomReady,
		Bind:             []interface{}{app},
		AssetServer: &assetserver.Options{
			Assets:  miniFS,
			Handler: handler,
		},
	})
	if err != nil {
		fmt.Println("Error:", err.Error())
	}
}

type MiniApp struct {
	ctx context.Context
}

func NewMiniApp() *MiniApp { return &MiniApp{} }

func (a *MiniApp) startup(ctx context.Context) { a.ctx = ctx }

func (a *MiniApp) onDomReady(ctx context.Context) {
	miniAppCtx = ctx
	a.ctx = ctx
	runtime.WindowSetAlwaysOnTop(ctx, true)
	runtime.WindowSetTitle(ctx, "GPU Status")

	// Position top-right after a brief delay
	go func() {
		time.Sleep(300 * time.Millisecond)
		screens, err := runtime.ScreenGetAll(ctx)
		if err == nil && len(screens) > 0 {
			s := screens[0]
			x := s.Width - 295
			if x < 0 {
				x = 0
			}
			runtime.WindowSetPosition(ctx, x, 10)
		}
	}()

	// Show the window
	runtime.WindowShow(ctx)

	// Start polling GPU status
	go a.pollStatus(ctx)
}

func (a *MiniApp) pollStatus(ctx context.Context) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	a.emitStatus(ctx)
	for {
		<-ticker.C
		a.emitStatus(ctx)
	}
}

func (a *MiniApp) emitStatus(ctx context.Context) {
	status := a.readGPUStatus()
	runtime.EventsEmit(ctx, "gpu-status-update", map[string]interface{}{
		"available":      status.Available,
		"value":          status.Value,
		"label":          status.Label,
		"pcmStatus":      status.PCMStatus,
		"pcmStatusAvail": status.PCMEnable,
		"pcmLabel":       status.PCMLabel,
	})
}

type gpuStatus struct {
	Label     string
	Value     uint32
	Available bool
	PCMStatus uint32
	PCMEnable bool
	PCMLabel  string
}

func (a *MiniApp) readGPUStatus() gpuStatus {
	s := gpuStatus{Label: "N/A", Available: false}
	exe, _ := os.Executable()
	file := exe[:len(exe)-4] + "_gpu_status.txt"
	data, err := os.ReadFile(file)
	if err != nil {
		return s
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "label:") {
			s.Label = strings.TrimSpace(strings.TrimPrefix(line, "label:"))
		} else if strings.HasPrefix(line, "value:") {
			if v, err := strconv.ParseUint(strings.TrimSpace(strings.TrimPrefix(line, "value:")), 10, 32); err == nil {
				s.Value = uint32(v)
				s.Available = true
			}
		} else if strings.HasPrefix(line, "pcm:") {
			if v, err := strconv.ParseUint(strings.TrimSpace(strings.TrimPrefix(line, "pcm:")), 10, 32); err == nil {
				s.PCMStatus = uint32(v)
				s.PCMEnable = true
			}
		} else if strings.HasPrefix(line, "pcmLabel:") {
			s.PCMLabel = strings.TrimSpace(strings.TrimPrefix(line, "pcmLabel:"))
		}
	}
	return s
}

// onSystrayReady is called when the system tray is ready
func (a *MiniApp) onSystrayReady() {
	systray.SetTooltip("GPU Status")
	itemRestore := systray.AddMenuItem("Restore Dispatcher UI", "Restore the main app window")
	itemQuit := systray.AddMenuItem("Quit", "Exit the GPU Status mini window")

	go func() {
		for {
			select {
			case <-itemRestore.ClickedCh:
				a.restoreMainWindow()
			case <-itemQuit.ClickedCh:
				systray.Quit()
				runtime.Quit(miniAppCtx)
				os.Exit(0)
			}
		}
	}()
}

func (a *MiniApp) onSystrayExit() {}

// restoreMainWindow writes a restore signal file and exits the mini exe
func (a *MiniApp) restoreMainWindow() {
	exe, _ := os.Executable()
	restoreFile := exe[:len(exe)-4] + "_restore.txt"
	_ = os.WriteFile(restoreFile, []byte("restore"), 0644)
	systray.Quit()
	runtime.Quit(miniAppCtx)
	os.Exit(0)
}

// RestoreMainWindow is exposed to the mini window frontend (button handler)
func (a *MiniApp) RestoreMainWindow() {
	a.restoreMainWindow()
}
