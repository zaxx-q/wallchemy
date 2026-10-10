package handlers

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	wapi "github.com/davenicholson-xyz/go-wallhaven/wallhavenapi"
	"github.com/zaxx-q/wallchemy/app"
)

type WallchemyCommand string

const (
	SingleWallpaper  WallchemyCommand = "single"
	CollectionCmd    WallchemyCommand = "collection"
	CollectionList   WallchemyCommand = "collections"
	RandomCmd        WallchemyCommand = "random"
	InfoCmd          WallchemyCommand = "info"
	CleanCmd         WallchemyCommand = "clean"
	DaemonCmd        WallchemyCommand = "daemon"
	LaunchDaemonCmd  WallchemyCommand = "launchdaemon"
	BlacklistCmd     WallchemyCommand = "blacklist"
	ListBlacklistCmd WallchemyCommand = "listblacklist"
	RotateCmd        WallchemyCommand = "rotate"
	PreviousCmd      WallchemyCommand = "previous"
	FavoriteCmd      WallchemyCommand = "favorite"
	FavoritesCmd     WallchemyCommand = "favorites"
	PruneCacheCmd    WallchemyCommand = "prune_cache"
	OpenCmd          WallchemyCommand = "open"
)

func globalFilters(appCtx *app.AppContext) error {
	if len(appCtx.Config.Categories) > 0 {
		var flags []wapi.CategoriesFlag
		for _, cat := range appCtx.Config.Categories {
			switch strings.ToLower(cat) {
			case "general":
				flags = append(flags, wapi.General)
			case "anime":
				flags = append(flags, wapi.Anime)
			case "people":
				flags = append(flags, wapi.People)
			default:
				return fmt.Errorf("invalid category: %s", cat)
			}
		}
		if len(flags) > 0 {
			appCtx.WallhavenAPI.Categories(flags...)
		}
	}

	if len(appCtx.Config.Purity) > 0 {
		var flags []wapi.PurityFlag
		for _, purity := range appCtx.Config.Purity {
			switch strings.ToLower(purity) {
			case "sfw":
				flags = append(flags, wapi.SFW)
			case "sketchy":
				flags = append(flags, wapi.Sketchy)
			case "nsfw":
				flags = append(flags, wapi.NSFW)
			default:
				return fmt.Errorf("invalid purity: %s", purity)
			}
		}
		if len(flags) > 0 {
			appCtx.WallhavenAPI.Purity(flags...)
		}
	}

	if appCtx.Config.TopRange != "" {
		var rangeType wapi.RangeType
		switch appCtx.Config.TopRange {
		case "1d":
			rangeType = wapi.OneDay
		case "3d":
			rangeType = wapi.ThreeDays
		case "1w":
			rangeType = wapi.OneWeek
		case "1M":
			rangeType = wapi.OneMonth
		case "3M":
			rangeType = wapi.ThreeMonths
		case "6M":
			rangeType = wapi.SixMonths
		case "1y":
			rangeType = wapi.OneYear
		default:
			return fmt.Errorf("invalid range: %s", appCtx.Config.TopRange)
		}
		appCtx.WallhavenAPI.Range(rangeType)
	}

	if appCtx.Config.MinResolution != "" {
		appCtx.WallhavenAPI.MinimumResolution(appCtx.Config.MinResolution)
	}

	if len(appCtx.Config.Resolutions) > 0 {
		appCtx.WallhavenAPI.Resolutions(appCtx.Config.Resolutions...)
	}

	if len(appCtx.Config.Ratios) > 0 {
		appCtx.WallhavenAPI.Ratios(appCtx.Config.Ratios...)
	}

	if appCtx.Config.Colors != "" {
		appCtx.WallhavenAPI.Colors(appCtx.Config.Colors)
	}

	if appCtx.Config.Seed != "" {
		appCtx.WallhavenAPI.Seed(appCtx.Config.Seed)
	}

	return nil
}

func parseAction(appCtx *app.AppContext) WallchemyCommand {
	if appCtx.Config.Info {
		return InfoCmd
	}
	if appCtx.Config.Clean {
		return CleanCmd
	}
	if appCtx.Config.LaunchDaemon {
		return LaunchDaemonCmd
	}
	if appCtx.Config.Daemon != "" {
		return DaemonCmd
	}
	if appCtx.Config.ListBlacklist {
		return ListBlacklistCmd
	}
	if appCtx.Config.Blacklist != "" {
		return BlacklistCmd
	}
	if appCtx.Config.OpenCurrent {
		return OpenCmd
	}
	if appCtx.Config.Favorite {
		return FavoriteCmd
	}
	if appCtx.Config.Favorites {
		return FavoritesCmd
	}
	if appCtx.Config.Previous {
		return PreviousCmd
	}
	if appCtx.Config.PruneCache {
		return PruneCacheCmd
	}
	if appCtx.Config.ID != "" {
		return SingleWallpaper
	}
	if appCtx.Config.Collections {
		return CollectionList
	}
	if appCtx.Config.Collection != "" {
		return CollectionCmd
	}
	if appCtx.Config.Random != "" || appCtx.Config.Hot || appCtx.Config.TopList {
		return RandomCmd
	}
	if appCtx.Config.Rotate {
		return RotateCmd
	}

	return RotateCmd
}

func ExecuteCommand(appCtx *app.AppContext) (string, error) {
	action := parseAction(appCtx)
	app.Logger.WithField("action", string(action)).Debug("Action received")

	switch action {
	case SingleWallpaper:
		return handleSingleWithLifecycle(appCtx)
	case CollectionCmd:
		return handleCollectionWithLifecycle(appCtx)
	case CollectionList:
		return handleCollections(appCtx)
	case RandomCmd:
		return handleRandomWithLifecycle(appCtx)
	case CleanCmd:
		return handleClean(appCtx)
	case InfoCmd:
		return handleInfo(appCtx)
	case LaunchDaemonCmd:
		return startDaemon(appCtx)
	case DaemonCmd:
		return HandleDaemon(appCtx)
	case BlacklistCmd:
		return handleBlacklist(appCtx)
	case ListBlacklistCmd:
		return handleListBlacklist(appCtx)
	case RotateCmd:
		return handleRotate(appCtx)
	case PreviousCmd:
		return handlePrevious(appCtx)
	case FavoriteCmd:
		return handleFavorite(appCtx)
	case FavoritesCmd:
		return handleFavorites(appCtx)
	case PruneCacheCmd:
		return handlePruneCache(appCtx)
	case OpenCmd:
		return handleOpen(appCtx)
	default:
		return "", fmt.Errorf("no valid action found")
	}
}

func recordCurrentAsPrevious(appCtx *app.AppContext) {
	if appCtx.HistoryManager == nil {
		return
	}
	current, err := app.LoadCurrentWallpaper(appCtx)
	if err != nil || current.ID == "" {
		return
	}

	hist, err := appCtx.HistoryManager.Load()
	if err == nil && len(hist.Entries) > 0 {
		if hist.Current >= 0 && hist.Current < len(hist.Entries) && hist.Entries[hist.Current].ID == current.ID {
			return
		}
	}

	entry := app.HistoryEntry{
		ID:        current.ID,
		ImageURL:  current.Path,
		PageURL:   current.Link,
		ImagePath: current.FilePath,
		Label:     "Previous wallpaper",
		SavedAt:   time.Now().Format(time.RFC3339),
	}
	_ = appCtx.HistoryManager.Record(entry)
}

func recordCurrent(appCtx *app.AppContext, label string) {
	if appCtx.HistoryManager == nil {
		return
	}
	current, err := app.LoadCurrentWallpaper(appCtx)
	if err != nil || current.ID == "" {
		return
	}

	entry := app.HistoryEntry{
		ID:        current.ID,
		ImageURL:  current.Path,
		PageURL:   current.Link,
		ImagePath: current.FilePath,
		Label:     label,
		SavedAt:   time.Now().Format(time.RFC3339),
	}
	_ = appCtx.HistoryManager.Record(entry)
}

func fetchApplyAndRecord(appCtx *app.AppContext, label string, fetchFn func() (string, error)) (string, error) {
	recordCurrentAsPrevious(appCtx)

	filePath, err := fetchFn()
	if err != nil {
		if appCtx.Notifier != nil {
			appCtx.Notifier.NotifyError("Wallpaper Fetch Failed", err.Error())
		}
		return "", err
	}

	recordCurrent(appCtx, label)

	if appCtx.CacheTools != nil {
		pruneCount, err := app.PruneWallpapersCache(appCtx.CacheTools.GetCacheDirectory(), appCtx.Config.CacheKeep)
		if err == nil && pruneCount > 0 && appCtx.Notifier != nil {
			appCtx.Notifier.Log(fmt.Sprintf("Pruned %d cached wallpaper(s); keeping %d newest", pruneCount, appCtx.Config.CacheKeep))
		}
	}

	if appCtx.Notifier != nil {
		appCtx.Notifier.NotifySuccess("Wallpaper Updated", label, "preferences-desktop-wallpaper")
	}
	return filePath, nil
}

func handleSingleWithLifecycle(appCtx *app.AppContext) (string, error) {
	label := fmt.Sprintf("Wallhaven: %s", appCtx.Config.ID)
	if appCtx.Notifier != nil {
		appCtx.Notifier.NotifyStart("Fetching Wallpaper...", label)
	}

	return fetchApplyAndRecord(appCtx, label, func() (string, error) {
		return handleSingle(appCtx)
	})
}

func handleCollectionWithLifecycle(appCtx *app.AppContext) (string, error) {
	if appCtx.Config.Collection == app.RandomCollectionMarker || appCtx.Config.Collection == "" {
		coll, err := appCtx.Config.GetRandomCollection()
		if err != nil {
			if appCtx.Notifier != nil {
				appCtx.Notifier.NotifyError("Wallchemy", err.Error())
			}
			return "", err
		}
		appCtx.Config.Collection = coll
	}

	label := fmt.Sprintf("Collection: %s", appCtx.Config.Collection)
	if appCtx.Notifier != nil {
		appCtx.Notifier.NotifyStart("Fetching Wallpaper...", label)
	}

	return fetchApplyAndRecord(appCtx, label, func() (string, error) {
		return handleCollection(appCtx)
	})
}

func handleRandomWithLifecycle(appCtx *app.AppContext) (string, error) {
	var label string
	var startMsg string

	if appCtx.Config.Hot {
		label = "Hot"
		startMsg = "Mode: Hot"
	} else if appCtx.Config.TopList {
		r := appCtx.Config.TopRange
		if r == "" {
			r = "1M"
			appCtx.Config.TopRange = r
		}
		label = fmt.Sprintf("Top (%s)", r)
		startMsg = fmt.Sprintf("Top wallpapers (%s)", r)
	} else {
		if appCtx.Config.Random == app.RandomSearchMarker || appCtx.Config.Random == "" {
			appCtx.Config.Random = appCtx.Config.GetRandomSearchTerm()
		}
		label = fmt.Sprintf("Random: %s", appCtx.Config.Random)
		startMsg = fmt.Sprintf("Random search: '%s'", appCtx.Config.Random)
	}

	if appCtx.Notifier != nil {
		appCtx.Notifier.NotifyStart("Fetching Wallpaper...", startMsg)
	}

	return fetchApplyAndRecord(appCtx, label, func() (string, error) {
		return handleRandom(appCtx)
	})
}

func handleRotate(appCtx *app.AppContext) (string, error) {
	modes := appCtx.Config.RotateModes
	if len(modes) == 0 {
		modes = []string{"random", "top", "hot"}
		if len(appCtx.Config.RotateCollections) > 0 {
			modes = append(modes, "collection")
		}
	}

	if appCtx.RotateManager == nil {
		return "", fmt.Errorf("rotate manager not initialized")
	}

	mode, err := appCtx.RotateManager.GetNextMode(modes)
	if err != nil {
		if appCtx.Notifier != nil {
			appCtx.Notifier.NotifyError("Rotate Failed", err.Error())
		}
		return "", err
	}

	switch mode {
	case "random":
		term := appCtx.Config.GetRandomSearchTerm()
		appCtx.Config.Random = term
		appCtx.Config.Hot = false
		appCtx.Config.TopList = false
		label := fmt.Sprintf("Random: %s", term)
		if appCtx.Notifier != nil {
			appCtx.Notifier.NotifyStart("Fetching Wallpaper...", fmt.Sprintf("Mode: Random ('%s')", term))
		}
		return fetchApplyAndRecord(appCtx, label, func() (string, error) {
			return handleRandom(appCtx)
		})

	case "top":
		r := appCtx.Config.TopRange
		if r == "" {
			r = "1M"
			appCtx.Config.TopRange = r
		}
		appCtx.Config.TopList = true
		appCtx.Config.Hot = false
		appCtx.Config.Random = ""
		label := fmt.Sprintf("Top (%s)", r)
		if appCtx.Notifier != nil {
			appCtx.Notifier.NotifyStart("Fetching Wallpaper...", fmt.Sprintf("Mode: Top (%s)", r))
		}
		return fetchApplyAndRecord(appCtx, label, func() (string, error) {
			return handleRandom(appCtx)
		})

	case "hot":
		appCtx.Config.Hot = true
		appCtx.Config.TopList = false
		appCtx.Config.Random = ""
		label := "Hot"
		if appCtx.Notifier != nil {
			appCtx.Notifier.NotifyStart("Fetching Wallpaper...", "Mode: Hot")
		}
		return fetchApplyAndRecord(appCtx, label, func() (string, error) {
			return handleRandom(appCtx)
		})

	case "collection":
		coll, err := appCtx.Config.GetRandomCollection()
		if err != nil {
			if appCtx.Notifier != nil {
				appCtx.Notifier.NotifyError("Wallchemy", err.Error())
			}
			return "", err
		}
		appCtx.Config.Collection = coll
		label := fmt.Sprintf("Collection: %s", coll)
		if appCtx.Notifier != nil {
			appCtx.Notifier.NotifyStart("Fetching Wallpaper...", fmt.Sprintf("Collection: '%s'", coll))
		}
		return fetchApplyAndRecord(appCtx, label, func() (string, error) {
			return handleCollection(appCtx)
		})

	default:
		return "", fmt.Errorf("unknown rotate mode: %s", mode)
	}
}

func handlePrevious(appCtx *app.AppContext) (string, error) {
	if appCtx.HistoryManager == nil {
		return "", fmt.Errorf("history manager not initialized")
	}

	prevEntry, prevIdx, err := appCtx.HistoryManager.GetPrevious()
	if err != nil {
		if appCtx.Notifier != nil {
			appCtx.Notifier.NotifyError("No Previous Wallpaper", err.Error())
		}
		return "", err
	}
	if prevEntry.ID == "" {
		msg := "The history entry is invalid"
		if appCtx.Notifier != nil {
			appCtx.Notifier.NotifyError("Could Not Restore Wallpaper", msg)
		}
		return "", fmt.Errorf("%s", msg)
	}

	if appCtx.Notifier != nil {
		appCtx.Notifier.NotifyStart("Restoring Wallpaper...", fmt.Sprintf("Wallhaven: %s", prevEntry.ID))
	}

	var filePath string
	if prevEntry.ImagePath != "" && app.FileExists(prevEntry.ImagePath) {
		if err := app.ApplyWallpaper(prevEntry.ImagePath, appCtx); err != nil {
			if appCtx.Notifier != nil {
				appCtx.Notifier.NotifyError("Could Not Restore Wallpaper", err.Error())
			}
			return "", err
		}
		appCtx.Current.ID = prevEntry.ID
		appCtx.Current.Path = prevEntry.ImageURL
		appCtx.Current.Link = prevEntry.PageURL
		appCtx.Current.FilePath = prevEntry.ImagePath
		_ = SaveCurrent(appCtx)
		filePath = prevEntry.ImagePath
	} else {
		appCtx.Config.ID = prevEntry.ID
		res, err := handleSingle(appCtx)
		if err != nil {
			if appCtx.Notifier != nil {
				appCtx.Notifier.NotifyError("Could Not Restore Wallpaper", err.Error())
			}
			return "", err
		}
		filePath = res
	}

	_ = appCtx.HistoryManager.SetCurrentIndex(prevIdx)

	if appCtx.CacheTools != nil {
		pruneCount, err := app.PruneWallpapersCache(appCtx.CacheTools.GetCacheDirectory(), appCtx.Config.CacheKeep)
		if err == nil && pruneCount > 0 && appCtx.Notifier != nil {
			appCtx.Notifier.Log(fmt.Sprintf("Pruned %d cached wallpaper(s); keeping %d newest", pruneCount, appCtx.Config.CacheKeep))
		}
	}

	if appCtx.Notifier != nil {
		appCtx.Notifier.NotifySuccess("Wallpaper Restored", fmt.Sprintf("Wallhaven: %s", prevEntry.ID), "preferences-desktop-wallpaper")
	}
	return filePath, nil
}

func handleFavorite(appCtx *app.AppContext) (string, error) {
	current, err := app.LoadCurrentWallpaper(appCtx)
	if err != nil {
		if appCtx.Notifier != nil {
			appCtx.Notifier.NotifyError("Could Not Read Wallpaper", err.Error())
		}
		return "", err
	}

	if current.ID == "" || current.FilePath == "" || !app.FileExists(current.FilePath) {
		msg := "Wallchemy returned incomplete current-wallpaper information"
		if appCtx.Notifier != nil {
			appCtx.Notifier.NotifyError("Could Not Read Wallpaper", msg)
		}
		return "", fmt.Errorf("%s", msg)
	}

	if appCtx.FavoritesManager == nil {
		return "", fmt.Errorf("favorites manager not initialized")
	}

	// Check if active wallpaper in noctalia matches
	activeWp := app.GetNoctaliaActiveWallpaper()
	if activeWp != "" && app.FileExists(activeWp) {
		activeReal, _ := filepath.EvalSymlinks(activeWp)
		currentReal, _ := filepath.EvalSymlinks(current.FilePath)
		if activeReal != currentReal {
			favDir := appCtx.FavoritesManager.GetFavoritesDir()
			if strings.HasPrefix(activeReal, favDir+string(filepath.Separator)) {
				if appCtx.Notifier != nil {
					appCtx.Notifier.NotifySuccess("Wallpaper Already Saved", "The active wallpaper is already a favorite", "emblem-favorite")
				}
				return "The active wallpaper is already a favorite", nil
			}
			msg := "The active wallpaper was not set by Wallchemy"
			if appCtx.Notifier != nil {
				appCtx.Notifier.NotifyError("Could Not Save Favorite", msg)
			}
			return "", fmt.Errorf("%s", msg)
		}
	}

	if appCtx.FavoritesManager.HasMetadata(current.ID) {
		if appCtx.Notifier != nil {
			appCtx.Notifier.NotifySuccess("Wallpaper Already Saved", fmt.Sprintf("Wallhaven: %s", current.ID), "emblem-favorite")
		}
		return fmt.Sprintf("Wallpaper Already Saved: Wallhaven: %s", current.ID), nil
	}

	if appCtx.Notifier != nil {
		appCtx.Notifier.NotifyStart("Saving Favorite...", fmt.Sprintf("Wallhaven: %s", current.ID))
	}

	if err := appCtx.FavoritesManager.Save(current); err != nil {
		if appCtx.Notifier != nil {
			appCtx.Notifier.NotifyError("Could Not Save Favorite", err.Error())
		}
		return "", err
	}

	if appCtx.Notifier != nil {
		appCtx.Notifier.NotifySuccess("Wallpaper Saved", fmt.Sprintf("Wallhaven: %s", current.ID), "emblem-favorite")
	}
	return fmt.Sprintf("Wallpaper Saved: Wallhaven: %s", current.ID), nil
}

func handleFavorites(appCtx *app.AppContext) (string, error) {
	if appCtx.FavoritesManager == nil {
		return "", fmt.Errorf("favorites manager not initialized")
	}

	if appCtx.Config.FavoritesGallery {
		manager, err := app.OpenFileManager(appCtx.FavoritesManager.GetFavoritesDir(), appCtx.Config.FileManager)
		if err != nil {
			if appCtx.Notifier != nil {
				appCtx.Notifier.NotifyError("Could Not Open Gallery", err.Error())
			}
			return "", err
		}
		if appCtx.Notifier != nil {
			appCtx.Notifier.NotifySuccess("Favorite Wallpapers", fmt.Sprintf("Opened gallery in %s", manager), "folder-pictures")
		}
		return fmt.Sprintf("Opened gallery in %s", manager), nil
	}

	entries, err := appCtx.FavoritesManager.ListEntries()
	if err != nil {
		return "", err
	}

	managerName := appCtx.Config.FileManager
	if managerName == "" {
		if _, err := exec.LookPath("thunar"); err == nil {
			managerName = "Thunar"
		} else {
			managerName = "gallery"
		}
	}

	menuItems := []string{fmt.Sprintf("Open gallery in %s", managerName)}
	for _, e := range entries {
		menuItems = append(menuItems, fmt.Sprintf("%s\tWallhaven %s — saved %s", e.ID, e.ID, e.SavedAt))
	}

	sel, err := app.RunDmenu("Favorite wallpapers", menuItems, appCtx.Config.DmenuCmd)
	if err != nil || sel == "" {
		return "", nil
	}

	if strings.HasPrefix(sel, "Open gallery in ") {
		manager, err := app.OpenFileManager(appCtx.FavoritesManager.GetFavoritesDir(), appCtx.Config.FileManager)
		if err != nil {
			if appCtx.Notifier != nil {
				appCtx.Notifier.NotifyError("Could Not Open Gallery", err.Error())
			}
			return "", err
		}
		if appCtx.Notifier != nil {
			appCtx.Notifier.NotifySuccess("Favorite Wallpapers", fmt.Sprintf("Opened gallery in %s", manager), "folder-pictures")
		}
		return fmt.Sprintf("Opened gallery in %s", manager), nil
	}

	parts := strings.Split(sel, "\t")
	id := strings.TrimSpace(parts[0])
	if id == "" {
		return "", nil
	}

	action, err := app.RunDmenu(fmt.Sprintf("Favorite %s", id), []string{"Set wallpaper", "Open Wallhaven page", "Remove from favorites"}, appCtx.Config.DmenuCmd)
	if err != nil || action == "" {
		return "", nil
	}

	switch action {
	case "Set wallpaper":
		imgPath, err := appCtx.FavoritesManager.GetImagePath(id)
		if err != nil || !app.FileExists(imgPath) {
			if appCtx.Notifier != nil {
				appCtx.Notifier.NotifyError("Favorite Not Found", fmt.Sprintf("Archived image is missing for Wallhaven: %s", id))
			}
			return "", fmt.Errorf("favorite image missing")
		}
		if appCtx.Notifier != nil {
			appCtx.Notifier.NotifyStart("Setting Favorite...", fmt.Sprintf("Wallhaven: %s", id))
		}
		if err := app.ApplyWallpaper(imgPath, appCtx); err != nil {
			if appCtx.Notifier != nil {
				appCtx.Notifier.NotifyError("Could Not Set Favorite", fmt.Sprintf("Wallhaven: %s", id))
			}
			return "", err
		}
		meta, _ := appCtx.FavoritesManager.GetEntry(id)
		if meta != nil {
			appCtx.Current.ID = meta.ID
			appCtx.Current.Path = meta.ImageURL
			appCtx.Current.Link = meta.PageURL
			appCtx.Current.FilePath = imgPath
			_ = SaveCurrent(appCtx)
			recordCurrent(appCtx, fmt.Sprintf("Favorite: %s", id))
		}
		if appCtx.Notifier != nil {
			appCtx.Notifier.NotifySuccess("Wallpaper Updated", fmt.Sprintf("Favorite: %s", id), "preferences-desktop-wallpaper")
		}
		return fmt.Sprintf("Wallpaper Updated: Favorite: %s", id), nil

	case "Open Wallhaven page":
		meta, err := appCtx.FavoritesManager.GetEntry(id)
		if err != nil || meta.PageURL == "" {
			if appCtx.Notifier != nil {
				appCtx.Notifier.NotifyError("Could Not Open Wallpaper", "Favorite has no valid Wallhaven link")
			}
			return "", fmt.Errorf("no valid link")
		}
		if err := app.OpenBrowser(meta.PageURL); err != nil {
			if appCtx.Notifier != nil {
				appCtx.Notifier.NotifyError("Could Not Open Wallpaper", "Failed to launch browser")
			}
			return "", err
		}
		if appCtx.Notifier != nil {
			appCtx.Notifier.NotifySuccess("Wallhaven Opened", meta.PageURL, "internet-web-browser")
		}
		return fmt.Sprintf("Wallhaven Opened: %s", meta.PageURL), nil

	case "Remove from favorites":
		if err := appCtx.FavoritesManager.Remove(id); err != nil {
			if appCtx.Notifier != nil {
				appCtx.Notifier.NotifyError("Could Not Remove Favorite", err.Error())
			}
			return "", err
		}
		if appCtx.Notifier != nil {
			appCtx.Notifier.NotifySuccess("Favorite Removed", fmt.Sprintf("Wallhaven: %s", id), "user-trash")
		}
		return fmt.Sprintf("Favorite Removed: Wallhaven: %s", id), nil
	}

	return "", nil
}

func handleOpen(appCtx *app.AppContext) (string, error) {
	if appCtx.Notifier != nil {
		appCtx.Notifier.NotifyStart("Wallchemy", "Opening the current wallpaper on Wallhaven...")
	}

	current, err := app.LoadCurrentWallpaper(appCtx)
	if err != nil || current.Link == "" {
		if appCtx.Notifier != nil {
			appCtx.Notifier.NotifyError("Could Not Open Wallpaper", "Current wallpaper has no valid Wallhaven link")
		}
		return "", fmt.Errorf("current wallpaper has no valid Wallhaven link")
	}

	if err := app.OpenBrowser(current.Link); err != nil {
		if appCtx.Notifier != nil {
			appCtx.Notifier.NotifyError("Could Not Open Wallpaper", "Failed to launch browser")
		}
		return "", err
	}

	if appCtx.Notifier != nil {
		appCtx.Notifier.NotifySuccess("Wallhaven Opened", current.Link, "internet-web-browser")
	}
	return fmt.Sprintf("Wallhaven Opened: %s", current.Link), nil
}

func handlePruneCache(appCtx *app.AppContext) (string, error) {
	count, err := app.PruneWallpapersCache(appCtx.CacheTools.GetCacheDirectory(), appCtx.Config.CacheKeep)
	if err != nil {
		return "", err
	}
	msg := fmt.Sprintf("Keeping %d newest wallpapers", appCtx.Config.CacheKeep)
	if appCtx.Notifier != nil {
		appCtx.Notifier.Log(fmt.Sprintf("Pruned %d cached wallpaper(s); %s", count, msg))
		appCtx.Notifier.NotifySuccess("Wallchemy Cache Pruned", msg, "edit-clear")
	}
	return fmt.Sprintf("Wallchemy Cache Pruned: %s", msg), nil
}

func handleClean(appCtx *app.AppContext) (string, error) {
	appCtx.CacheTools.Clean()
	app.Logger.WithField("cache dir", appCtx.CacheTools.GetCacheDirectory()).Debug("action received")
	return appCtx.CacheTools.GetCacheDirectory() + " cleaned", nil
}

func handleBlacklist(appCtx *app.AppContext) (string, error) {
	wallpaperID := appCtx.Config.Blacklist

	if wallpaperID == app.BlacklistCurrentAndRotate {
		if appCtx.Notifier != nil {
			appCtx.Notifier.NotifyStart("Wallchemy", "Blacklisting current wallpaper...")
		}

		currentWallpaperID, err := getCurrentWallpaperID(appCtx)
		if err != nil {
			if appCtx.Notifier != nil {
				appCtx.Notifier.NotifyError("Blacklist Failed", err.Error())
			}
			return "", fmt.Errorf("failed to get current wallpaper ID: %w", err)
		}

		if err := appCtx.CacheTools.AddToBlacklist(currentWallpaperID); err != nil {
			app.Logger.WithError(err).Error("Failed to add wallpaper to blacklist")
			if appCtx.Notifier != nil {
				appCtx.Notifier.NotifyError("Blacklist Failed", err.Error())
			}
			return "", fmt.Errorf("failed to add wallpaper %s to blacklist: %w", currentWallpaperID, err)
		}

		app.Logger.WithField("wallpaper_id", currentWallpaperID).Info("Added wallpaper to blacklist")
		if appCtx.Notifier != nil {
			appCtx.Notifier.NotifySuccess("Wallpaper Blacklisted", fmt.Sprintf("Added wallpaper %s to blacklist", currentWallpaperID), "")
		}

		// Immediately rotate to next wallpaper
		return handleRotate(appCtx)
	}

	if wallpaperID == "current" {
		currentWallpaperID, err := getCurrentWallpaperID(appCtx)
		if err != nil {
			return "", fmt.Errorf("failed to get current wallpaper ID: %w", err)
		}
		wallpaperID = currentWallpaperID
	}

	if err := appCtx.CacheTools.AddToBlacklist(wallpaperID); err != nil {
		app.Logger.WithError(err).Error("Failed to add wallpaper to blacklist")
		return "", fmt.Errorf("failed to add wallpaper %s to blacklist: %w", wallpaperID, err)
	}

	app.Logger.WithField("wallpaper_id", wallpaperID).Info("Added wallpaper to blacklist")
	return fmt.Sprintf("Added wallpaper %s to blacklist", wallpaperID), nil
}

func handleListBlacklist(appCtx *app.AppContext) (string, error) {
	blacklist, err := appCtx.CacheTools.GetBlacklist()
	if err != nil {
		app.Logger.WithError(err).Error("Failed to retrieve blacklist")
		return "", fmt.Errorf("failed to retrieve blacklist: %w", err)
	}

	if len(blacklist) == 0 {
		return "No wallpapers in blacklist", nil
	}

	result := fmt.Sprintf("Blacklisted wallpapers (%d):\n", len(blacklist))
	for _, id := range blacklist {
		result += fmt.Sprintf("- %s\n", id)
	}

	app.Logger.WithField("blacklist_count", len(blacklist)).Info("Listed blacklisted wallpapers")
	return result, nil
}

func getCurrentWallpaperID(appCtx *app.AppContext) (string, error) {
	current, err := app.LoadCurrentWallpaper(appCtx)
	if err != nil {
		return "", fmt.Errorf("no current wallpaper found: %w", err)
	}
	if current.ID == "" {
		return "", fmt.Errorf("current wallpaper has no ID")
	}
	return current.ID, nil
}
