package app

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type FavoriteMetadata struct {
	ID        string `json:"id"`
	ImageURL  string `json:"image_url"`
	PageURL   string `json:"page_url"`
	ImageFile string `json:"image_file"`
	SavedAt   string `json:"saved_at"`
}

type FavoritesManager struct {
	mu           sync.Mutex
	favoritesDir string
	metadataDir  string
}

func NewFavoritesManager(favoritesDir string) *FavoritesManager {
	metaDir := filepath.Join(favoritesDir, ".metadata")
	_ = os.MkdirAll(favoritesDir, 0755)
	_ = os.MkdirAll(metaDir, 0755)
	return &FavoritesManager{
		favoritesDir: favoritesDir,
		metadataDir:  metaDir,
	}
}

func (fm *FavoritesManager) GetFavoritesDir() string {
	return fm.favoritesDir
}

func (fm *FavoritesManager) MetadataPath(id string) string {
	return filepath.Join(fm.metadataDir, fmt.Sprintf("wallhaven-%s.json", id))
}

func (fm *FavoritesManager) HasMetadata(id string) bool {
	metaPath := fm.MetadataPath(id)
	info, err := os.Stat(metaPath)
	return err == nil && !info.IsDir()
}

func (fm *FavoritesManager) Save(current *CurrentWallpaper) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if current.ID == "" {
		return fmt.Errorf("current wallpaper has no ID")
	}

	ext := strings.TrimPrefix(filepath.Ext(current.FilePath), ".")
	if ext == "" {
		ext = "jpg"
	}
	ext = strings.ToLower(ext)

	imageFile := fmt.Sprintf("wallhaven-%s.%s", current.ID, ext)
	destImage := filepath.Join(fm.favoritesDir, imageFile)
	metaPath := fm.MetadataPath(current.ID)

	// Copy wallpaper image
	tmpImage := filepath.Join(fm.favoritesDir, fmt.Sprintf(".wallhaven-%s.%d.tmp", current.ID, os.Getpid()))
	if err := copyFileContents(current.FilePath, tmpImage); err != nil {
		_ = os.Remove(tmpImage)
		return fmt.Errorf("failed to copy wallpaper image: %w", err)
	}
	if err := os.Rename(tmpImage, destImage); err != nil {
		_ = os.Remove(tmpImage)
		return fmt.Errorf("failed to save wallpaper image: %w", err)
	}

	// Write metadata
	meta := FavoriteMetadata{
		ID:        current.ID,
		ImageURL:  current.Path,
		PageURL:   current.Link,
		ImageFile: imageFile,
		SavedAt:   time.Now().Format(time.RFC3339),
	}

	bytes, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		_ = os.Remove(destImage)
		return fmt.Errorf("failed to encode favorite metadata: %w", err)
	}

	tmpMeta := filepath.Join(fm.metadataDir, fmt.Sprintf(".wallhaven-%s.%d.tmp", current.ID, os.Getpid()))
	if err := os.WriteFile(tmpMeta, bytes, 0644); err != nil {
		_ = os.Remove(destImage)
		_ = os.Remove(tmpMeta)
		return fmt.Errorf("failed to write favorite metadata: %w", err)
	}

	if err := os.Rename(tmpMeta, metaPath); err != nil {
		_ = os.Remove(destImage)
		_ = os.Remove(tmpMeta)
		return fmt.Errorf("failed to save favorite metadata: %w", err)
	}

	return nil
}

func (fm *FavoritesManager) ListEntries() ([]FavoriteMetadata, error) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	entries, err := os.ReadDir(fm.metadataDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var results []FavoriteMetadata
	for _, e := range entries {
		if e.IsDir() || !strings.HasPrefix(e.Name(), "wallhaven-") || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		filePath := filepath.Join(fm.metadataDir, e.Name())
		bytes, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		var meta FavoriteMetadata
		if err := json.Unmarshal(bytes, &meta); err == nil && meta.ID != "" {
			results = append(results, meta)
		}
	}

	// Sort newest saved_at first
	sort.Slice(results, func(i, j int) bool {
		return results[i].SavedAt > results[j].SavedAt
	})

	return results, nil
}

func (fm *FavoritesManager) getEntryUnlocked(id string) (*FavoriteMetadata, error) {
	metaPath := fm.MetadataPath(id)
	bytes, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, err
	}

	var meta FavoriteMetadata
	if err := json.Unmarshal(bytes, &meta); err != nil {
		return nil, err
	}

	return &meta, nil
}

func (fm *FavoritesManager) getImagePathUnlocked(id string) (string, error) {
	meta, err := fm.getEntryUnlocked(id)
	if err != nil {
		return "", err
	}
	if meta.ImageFile == "" || strings.Contains(meta.ImageFile, "/") || strings.Contains(meta.ImageFile, "\\") {
		return "", fmt.Errorf("invalid image filename in favorite metadata")
	}
	return filepath.Join(fm.favoritesDir, meta.ImageFile), nil
}

func (fm *FavoritesManager) GetEntry(id string) (*FavoriteMetadata, error) {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	return fm.getEntryUnlocked(id)
}

func (fm *FavoritesManager) GetImagePath(id string) (string, error) {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	return fm.getImagePathUnlocked(id)
}

func (fm *FavoritesManager) Remove(id string) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	metaPath := fm.MetadataPath(id)
	if imgPath, err := fm.getImagePathUnlocked(id); err == nil {
		_ = os.Remove(imgPath)
	}
	return os.Remove(metaPath)
}

func copyFileContents(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

func MigrateLegacyFavorites(targetFavDir string) {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	legacyFavDir := filepath.Join(home, ".local", "share", "wallchemy-rotate", "favorites")
	if _, err := os.Stat(legacyFavDir); err != nil {
		return
	}

	_ = os.MkdirAll(targetFavDir, 0755)
	targetMeta := filepath.Join(targetFavDir, ".metadata")
	_ = os.MkdirAll(targetMeta, 0755)

	// Copy image files
	if entries, err := os.ReadDir(legacyFavDir); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			src := filepath.Join(legacyFavDir, e.Name())
			dst := filepath.Join(targetFavDir, e.Name())
			if _, err := os.Stat(dst); os.IsNotExist(err) {
				_ = copyFileContents(src, dst)
			}
		}
	}

	// Copy metadata files
	legacyMeta := filepath.Join(legacyFavDir, ".metadata")
	if entries, err := os.ReadDir(legacyMeta); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			src := filepath.Join(legacyMeta, e.Name())
			dst := filepath.Join(targetMeta, e.Name())
			if _, err := os.Stat(dst); os.IsNotExist(err) {
				_ = copyFileContents(src, dst)
			}
		}
	}
}
