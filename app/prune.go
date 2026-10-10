package app

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func PruneWallpapersCache(cacheBaseDir string, keep int) (int, error) {
	if keep <= 0 {
		keep = 150
	}

	wallpapersDir := filepath.Join(cacheBaseDir, "wallpapers")
	entries, err := os.ReadDir(wallpapersDir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}

	type fileWithTime struct {
		path    string
		modTime time.Time
	}

	var files []fileWithTime
	for _, e := range entries {
		if e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		files = append(files, fileWithTime{
			path:    filepath.Join(wallpapersDir, e.Name()),
			modTime: info.ModTime(),
		})
	}

	if len(files) <= keep {
		return 0, nil
	}

	// Sort newest first
	sort.Slice(files, func(i, j int) bool {
		return files[i].modTime.After(files[j].modTime)
	})

	removed := 0
	for _, f := range files[keep:] {
		if err := os.Remove(f.path); err == nil {
			removed++
		}
	}

	return removed, nil
}
