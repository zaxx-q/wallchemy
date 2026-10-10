package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHistoryManager(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "wallchemy_hist_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	hm := NewHistoryManager(tempDir, 3)

	// Initial load should be empty
	data, err := hm.Load()
	if err != nil {
		t.Fatalf("unexpected error loading empty history: %v", err)
	}
	if data.Current != -1 || len(data.Entries) != 0 {
		t.Fatalf("expected empty history, got current=%d len=%d", data.Current, len(data.Entries))
	}

	// Cannot get previous on empty
	_, _, err = hm.GetPrevious()
	if err == nil {
		t.Fatalf("expected error getting previous on empty history")
	}

	// Add entry 1
	e1 := HistoryEntry{ID: "w1", ImageURL: "http://example.com/1.jpg", Label: "test1"}
	if err := hm.Record(e1); err != nil {
		t.Fatalf("failed to record entry 1: %v", err)
	}

	data, err = hm.Load()
	if err != nil || data.Current != 0 || len(data.Entries) != 1 {
		t.Fatalf("expected 1 entry at current 0, got current=%d len=%d", data.Current, len(data.Entries))
	}

	// Still cannot get previous when current is 0
	_, _, err = hm.GetPrevious()
	if err == nil {
		t.Fatalf("expected error getting previous when current is 0")
	}

	// Add entry 2
	e2 := HistoryEntry{ID: "w2", ImageURL: "http://example.com/2.jpg", Label: "test2"}
	if err := hm.Record(e2); err != nil {
		t.Fatalf("failed to record entry 2: %v", err)
	}

	// At the end, GetNext should fail
	_, _, err = hm.GetNext()
	if err == nil {
		t.Fatalf("expected error getting next when at the edge")
	}

	prev, prevIdx, err := hm.GetPrevious()
	if err != nil {
		t.Fatalf("failed to get previous: %v", err)
	}
	if prevIdx != 0 || prev.ID != "w1" {
		t.Fatalf("expected previous to be w1 at 0, got %s at %d", prev.ID, prevIdx)
	}

	// Set current index back to 0
	if err := hm.SetCurrentIndex(prevIdx); err != nil {
		t.Fatalf("failed to set current index: %v", err)
	}

	// Now that current is 0 and we have 2 entries, GetNext should succeed and return w2 at 1
	next, nextIdx, err := hm.GetNext()
	if err != nil {
		t.Fatalf("failed to get next from index 0: %v", err)
	}
	if nextIdx != 1 || next.ID != "w2" {
		t.Fatalf("expected next to be w2 at 1, got %s at %d", next.ID, nextIdx)
	}

	// Add entry 3 while at 0 (should truncate forward history)
	e3 := HistoryEntry{ID: "w3", ImageURL: "http://example.com/3.jpg", Label: "test3"}
	if err := hm.Record(e3); err != nil {
		t.Fatalf("failed to record entry 3: %v", err)
	}

	data, err = hm.Load()
	if err != nil || len(data.Entries) != 2 || data.Entries[1].ID != "w3" {
		t.Fatalf("expected entries [w1, w3], got %v", data.Entries)
	}

	// Add entry 4 and 5 (testing maxKeep=3)
	e4 := HistoryEntry{ID: "w4", ImageURL: "http://example.com/4.jpg", Label: "test4"}
	e5 := HistoryEntry{ID: "w5", ImageURL: "http://example.com/5.jpg", Label: "test5"}
	_ = hm.Record(e4)
	_ = hm.Record(e5)

	data, err = hm.Load()
	if err != nil || len(data.Entries) != 3 {
		t.Fatalf("expected length capped at 3, got %d", len(data.Entries))
	}
	if data.Entries[0].ID != "w3" || data.Entries[1].ID != "w4" || data.Entries[2].ID != "w5" {
		t.Fatalf("expected entries [w3, w4, w5], got %v", data.Entries)
	}
}

func TestRotateManager(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "wallchemy_rot_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	rm := NewRotateManager(tempDir)
	modes := []string{"random", "top", "hot"}

	m1, err := rm.GetNextMode(modes)
	if err != nil || m1 != "random" {
		t.Fatalf("expected random, got %s (err: %v)", m1, err)
	}

	m2, err := rm.GetNextMode(modes)
	if err != nil || m2 != "top" {
		t.Fatalf("expected top, got %s (err: %v)", m2, err)
	}

	m3, err := rm.GetNextMode(modes)
	if err != nil || m3 != "hot" {
		t.Fatalf("expected hot, got %s (err: %v)", m3, err)
	}

	// Wrap around
	m4, err := rm.GetNextMode(modes)
	if err != nil || m4 != "random" {
		t.Fatalf("expected wrap around to random, got %s (err: %v)", m4, err)
	}
}

func TestPruneWallpapersCache(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "wallchemy_prune_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	wallpapersDir := filepath.Join(tempDir, "wallpapers")
	_ = os.MkdirAll(wallpapersDir, 0755)

	// Create 5 dummy wallpaper files
	for i := 1; i <= 5; i++ {
		p := filepath.Join(wallpapersDir, filepath.Base(filepath.Join(wallpapersDir, "wallhaven-file"+string(rune('0'+i))+".jpg")))
		_ = os.WriteFile(p, []byte("data"), 0644)
	}

	// Keep 3
	removed, err := PruneWallpapersCache(tempDir, 3)
	if err != nil {
		t.Fatalf("unexpected prune error: %v", err)
	}
	if removed != 2 {
		t.Fatalf("expected 2 removed, got %d", removed)
	}

	entries, _ := os.ReadDir(wallpapersDir)
	if len(entries) != 3 {
		t.Fatalf("expected 3 remaining, got %d", len(entries))
	}
}

func TestFavoritesManager(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "wallchemy_fav_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fm := NewFavoritesManager(tempDir)

	// Create a dummy source wallpaper
	srcPath := filepath.Join(tempDir, "source.png")
	_ = os.WriteFile(srcPath, []byte("dummy image content"), 0644)

	current := &CurrentWallpaper{
		ID:       "test123",
		Path:     "https://w.wallhaven.cc/full/te/wallhaven-test123.png",
		Link:     "https://wallhaven.cc/w/test123",
		FilePath: srcPath,
	}

	if fm.HasMetadata("test123") {
		t.Fatalf("expected HasMetadata false before save")
	}

	if err := fm.Save(current); err != nil {
		t.Fatalf("failed to save favorite: %v", err)
	}

	if !fm.HasMetadata("test123") {
		t.Fatalf("expected HasMetadata true after save")
	}

	entries, err := fm.ListEntries()
	if err != nil || len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d (err: %v)", len(entries), err)
	}
	if entries[0].ID != "test123" {
		t.Fatalf("expected ID test123, got %s", entries[0].ID)
	}

	imgPath, err := fm.GetImagePath("test123")
	if err != nil {
		t.Fatalf("failed to get image path: %v", err)
	}
	if !FileExists(imgPath) {
		t.Fatalf("expected saved image to exist at %s", imgPath)
	}

	// Remove
	if err := fm.Remove("test123"); err != nil {
		t.Fatalf("failed to remove favorite: %v", err)
	}
	if fm.HasMetadata("test123") {
		t.Fatalf("expected HasMetadata false after remove")
	}
}
