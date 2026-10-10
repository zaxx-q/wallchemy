package app

import (
	"encoding/json"
	"fmt"
)

func LoadCurrentWallpaper(appCtx *AppContext) (*CurrentWallpaper, error) {
	if appCtx.Current.ID != "" && appCtx.Current.FilePath != "" {
		return &appCtx.Current, nil
	}

	if appCtx.CacheTools != nil && appCtx.CacheTools.FileExists("current") {
		infoBytes, err := appCtx.CacheTools.ReadFileBytes("current")
		if err != nil {
			return nil, fmt.Errorf("failed to read current wallpaper cache: %w", err)
		}

		var current CurrentWallpaper
		if err := json.Unmarshal(infoBytes, &current); err != nil {
			return nil, fmt.Errorf("failed to parse current wallpaper cache: %w", err)
		}

		appCtx.Current = current
		return &appCtx.Current, nil
	}

	return nil, fmt.Errorf("no current wallpaper info found")
}
