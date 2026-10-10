package main

import (
	"fmt"
	"os"
	"path/filepath"

	wapi "github.com/davenicholson-xyz/go-wallhaven/wallhavenapi"
	"github.com/zaxx-q/wallchemy/app"
	"github.com/zaxx-q/wallchemy/handlers"
)

var version = "version"

func main() {

	appCtx := app.NewAppContext(version)

	logConfig := app.LoggerConfigFromEnv()
	if err := app.InitLogger(logConfig); err != nil {
		panic("Failed to initialize logging")
	}

	app.Logger.Info("Wallchemy started")

	config, err := app.LoadConfig()
	if err != nil {
		app.Logger.WithError(err).Error("Unable to load config")
		fmt.Println(err)
		os.Exit(1)
	}

	appCtx.AddConfig(config)

	if config.Version {
		app.Logger.WithField("version", version).Info("Version check")
		fmt.Println(version)
		os.Exit(0)
	}

	if config.Help {
		app.Logger.Info("Help requested")
		app.ShowHelp()
		os.Exit(0)
	}

	cache, err := app.NewCacheTools("wallchemy")
	if err != nil {
		panic("Failed to parse cache folder")
	}
	app.Logger.WithField("folder", cache.GetCacheDirectory()).Info("Cache folder initialized")

	if err := runApp(appCtx); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}

func runApp(appCtx *app.AppContext) error {

	var result string

	appCtx.AddWallhavenAPI(wapi.New())

	cache, err := app.NewCacheTools("wallchemy")
	if err != nil {
		return fmt.Errorf("Cache error: %w", err)
	}
	appCtx.AddCacheTools(cache)

	// Migrate legacy wallchemy-rotate state & favorites if present
	app.MigrateLegacyState(appCtx.Config.StateDir)
	app.MigrateLegacyFavorites(appCtx.Config.FavoritesDir)

	// Initialize desktop notifications & state logger
	logFile := filepath.Join(appCtx.Config.StateDir, "last-run.log")
	notificationsEnabled := true
	if appCtx.Config.Notifications != nil {
		notificationsEnabled = *appCtx.Config.Notifications
	}
	notifier := app.NewNotifier(notificationsEnabled, logFile)
	appCtx.AddNotifier(notifier)

	// Initialize history, rotate, and favorites managers
	appCtx.AddHistoryManager(app.NewHistoryManager(appCtx.Config.StateDir, appCtx.Config.HistoryKeep))
	appCtx.AddRotateManager(app.NewRotateManager(appCtx.Config.StateDir))
	appCtx.AddFavoritesManager(app.NewFavoritesManager(appCtx.Config.FavoritesDir))

	if appCtx.Config.APIKey != "" {
		appCtx.WallhavenAPI.APIKey(appCtx.Config.APIKey)
	}

	result, err = handlers.ExecuteCommand(appCtx)
	if err != nil {
		return err
	}

	if appCtx.Config.Quiet {
		app.Logger.WithField("output", result).Debug("Silencing output")
	} else {
		if result != "" {
			fmt.Println(result)
		}
	}

	return nil
}
