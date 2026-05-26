package main

import (
	"context"
	"embed"
	"log"

	wailsdelivery "github.com/pidisk/pidisk/internal/delivery/wails"
	"github.com/pidisk/pidisk/internal/infra/eventbus"
	"github.com/pidisk/pidisk/internal/infra/keystore"
	"github.com/pidisk/pidisk/internal/infra/knownhosts"
	"github.com/pidisk/pidisk/internal/infra/logging"
	"github.com/pidisk/pidisk/internal/infra/metrics"
	"github.com/pidisk/pidisk/internal/infra/profiles"
	"github.com/pidisk/pidisk/internal/infra/settingsrepo"
	"github.com/pidisk/pidisk/internal/infra/syncrepo"
	"github.com/pidisk/pidisk/internal/infra/trashrepo"
	"github.com/pidisk/pidisk/internal/platform"
	"github.com/pidisk/pidisk/internal/ports"
	"github.com/pidisk/pidisk/internal/usecase"

	wailsapp "github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	logDir, err := platform.LogDir()
	if err != nil {
		log.Fatalf("resolve log dir: %v", err)
	}

	configDir, err := platform.ConfigDir()
	if err != nil {
		log.Fatalf("resolve config dir: %v", err)
	}
	settingsStore, err := settingsrepo.New(configDir)
	if err != nil {
		log.Fatalf("init settings store: %v", err)
	}
	initialSettings, err := settingsStore.Load()
	if err != nil {
		log.Fatalf("load settings: %v", err)
	}
	logger, err := logging.Init(logging.ParseLevel(initialSettings.LogLevel), logDir)
	if err != nil {
		log.Fatalf("init logger: %v", err)
	}

	dataDir, err := platform.DataDir()
	if err != nil {
		logger.Fatal().Err(err).Msg("resolve data dir")
	}

	profileRepo, err := profiles.Open(dataDir)
	if err != nil {
		logger.Fatal().Err(err).Msg("open profiles store")
	}
	defer func() { _ = profileRepo.Close() }()

	trashRepo, err := trashrepo.Open(dataDir)
	if err != nil {
		logger.Fatal().Err(err).Msg("open trash store")
	}
	defer func() { _ = trashRepo.Close() }()

	syncRepo, err := syncrepo.Open(dataDir)
	if err != nil {
		logger.Fatal().Err(err).Msg("open sync store")
	}
	defer func() { _ = syncRepo.Close() }()

	khStore, err := knownhosts.New(configDir)
	if err != nil {
		logger.Fatal().Err(err).Msg("open known_hosts store")
	}

	var metricRecorder ports.MetricsRecorder = ports.NoopMetrics{}
	if initialSettings.PrometheusEnabled {
		prom := metrics.New()
		metricRecorder = prom
		go func() {
			if err := prom.Serve(context.Background(), initialSettings.PrometheusAddr); err != nil {
				logger.Warn().Err(err).Msg("prometheus server exited")
			}
		}()
	}

	secrets := keystore.New()
	bus := eventbus.New(context.Background())
	app := wailsdelivery.NewApp(logger, bus)

	profilesUC := usecase.NewProfilesUseCase(profileRepo, secrets)
	broker := usecase.NewHostKeyBroker()
	connUC := usecase.NewConnectionUseCase(profilesUC, khStore, bus, broker, logger)
	filesUC := usecase.NewFilesUseCase(connUC, logger)
	trashUC := usecase.NewTrashUseCase(connUC, trashRepo, profilesUC, bus, logger)
	transferUC := usecase.NewTransferUseCase(connUC, bus, metricRecorder, logger)
	syncUC := usecase.NewSyncUseCase(connUC, profilesUC, syncRepo, bus, metricRecorder, logger)
	settingsUC, err := usecase.NewSettingsUseCase(settingsStore, metricRecorder, syncUC)
	if err != nil {
		logger.Fatal().Err(err).Msg("init settings usecase")
	}

	profileBindings := wailsdelivery.NewProfileBindings(app, profilesUC)
	connBindings := wailsdelivery.NewConnectionBindings(app, connUC, broker)
	fileBindings := wailsdelivery.NewFileBindings(app, filesUC, trashUC, profilesUC)
	transferBindings := wailsdelivery.NewTransferBindings(app, transferUC)
	trashBindings := wailsdelivery.NewTrashBindings(app, trashUC)
	syncBindings := wailsdelivery.NewSyncBindings(app, syncUC)
	settingsBindings := wailsdelivery.NewSettingsBindings(app, settingsUC)
	dialogBindings := wailsdelivery.NewDialogBindings(app)

	err = wailsapp.Run(&options.App{
		Title:                    "PIdisk",
		Width:                    1280,
		Height:                   820,
		MinWidth:                 900,
		MinHeight:                600,
		AssetServer:              &assetserver.Options{Assets: assets},
		BackgroundColour:         &options.RGBA{R: 17, G: 20, B: 27, A: 1},
		OnStartup:                app.OnStartup,
		OnShutdown:               app.OnShutdown,
		EnableDefaultContextMenu: true,
		Bind: []interface{}{
			app,
			profileBindings,
			connBindings,
			fileBindings,
			transferBindings,
			trashBindings,
			syncBindings,
			settingsBindings,
			dialogBindings,
		},
	})
	if err != nil {
		logger.Fatal().Err(err).Msg("wails run")
	}
}
