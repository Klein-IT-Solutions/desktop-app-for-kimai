package main

import (
	"embed"
	"log"
	"os"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"

	"github.com/klein-it-solutions/kimai-desktop-app/internal/app"
	"github.com/klein-it-solutions/kimai-desktop-app/internal/config"
	"github.com/klein-it-solutions/kimai-desktop-app/internal/desktop"
)

//go:embed all:frontend/dist
var assets embed.FS

// Window size including the transparent margin that holds the CSS drop shadow.
// The visible card is 360×480 plus 2×16 px (frontend/src/app.css, --shadow-gutter).
const (
	windowWidth  = 392
	windowHeight = 512
)

func init() {
	application.RegisterEvent[bool]("theme:changed")
	application.RegisterEvent[app.ConnectionResult](app.EventConnectionChanged)
	application.RegisterEvent[app.TrackerState](app.EventTrackerChanged)
}

// openConfig loads the configuration. A corrupt file is moved aside so the
// app still starts (with defaults) instead of failing.
func openConfig() *config.Store {
	path, err := config.DefaultPath()
	if err != nil {
		log.Fatal(err)
	}
	store, err := config.Open(path)
	if err == nil {
		return store
	}
	log.Printf("config: %v – moving it to %s.broken and starting with defaults", err, path)
	_ = os.Rename(path, path+".broken")
	store, err = config.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	return store
}

// openCatalog opens the list cache (~/.cache) and UI state (~/.config).
func openCatalog() *app.Catalog {
	cachePath, err := config.DefaultCachePath("cache.json")
	if err != nil {
		log.Fatal(err)
	}
	statePath, err := config.DefaultStatePath("state.json")
	if err != nil {
		log.Fatal(err)
	}
	return app.NewCatalog(cachePath, statePath)
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--check-keyring" {
		os.Exit(checkKeyring())
	}

	var mainWindow *application.WebviewWindow

	secretsPath, err := config.DefaultSecretsPath()
	if err != nil {
		log.Fatal(err)
	}
	conn := app.NewConnection(openConfig(), config.Keyring{}, &config.FileSecrets{Path: secretsPath})
	settings := app.NewSettingsService(conn)
	tracker := app.NewTrackerService(conn, settings, openCatalog())

	a := application.New(application.Options{
		Name:        "Kimai Tracker",
		Description: "Minimalist desktop time tracker for Kimai",
		Services: []application.Service{
			application.NewService(app.NewAppService()),
			application.NewService(settings),
			application.NewService(tracker),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Linux: application.LinuxOptions{
			ApplicationID: "solutions.klein_it.KimaiTracker",
			ProgramName:   "kimai-tracker",
		},
		// A second launch brings the running instance to the front instead of
		// starting another tracker.
		SingleInstance: &application.SingleInstanceOptions{
			UniqueID: "solutions.klein-it.kimai-tracker",
			OnSecondInstanceLaunch: func(application.SecondInstanceData) {
				if mainWindow != nil {
					mainWindow.Show()
					mainWindow.Focus()
				}
			},
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	mainWindow = a.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:            "main",
		Title:           "Kimai Tracker",
		Width:           windowWidth,
		Height:          windowHeight,
		MinWidth:        windowWidth,
		MinHeight:       windowHeight,
		DisableResize:   true,
		Frameless:       true,
		BackgroundType:  application.BackgroundTypeTransparent,
		InitialPosition: application.WindowCentered,
		Mac: application.MacWindow{
			Backdrop: application.MacBackdropTranslucent,
			TitleBar: application.MacTitleBarHiddenInset,
		},
		Windows: application.WindowsWindow{
			BackdropType: application.Acrylic,
		},
		Linux: application.LinuxWindow{
			WindowIsTranslucent: true,
		},
		URL: "/",
	})

	a.Event.OnApplicationEvent(events.Common.ApplicationStarted, func(*application.ApplicationEvent) {
		application.InvokeSync(desktop.MakeFramelessWindowsTransparent)
	})

	a.Event.OnApplicationEvent(events.Common.ThemeChanged, func(*application.ApplicationEvent) {
		a.Event.Emit("theme:changed", a.Env.IsDarkMode())
	})

	if err := a.Run(); err != nil {
		log.Fatal(err)
	}
}
