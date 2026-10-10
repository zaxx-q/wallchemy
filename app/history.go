package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type HistoryEntry struct {
	ID        string `json:"id"`
	ImageURL  string `json:"image_url"`
	PageURL   string `json:"page_url"`
	ImagePath string `json:"image_path"`
	Label     string `json:"label"`
	SavedAt   string `json:"saved_at"`
}

type HistoryData struct {
	Current int            `json:"current"`
	Entries []HistoryEntry `json:"entries"`
}

type HistoryManager struct {
	mu       sync.Mutex
	filePath string
	maxKeep  int
}

func NewHistoryManager(stateDir string, maxKeep int) *HistoryManager {
	if maxKeep <= 0 {
		maxKeep = 100
	}
	_ = os.MkdirAll(stateDir, 0755)
	return &HistoryManager{
		filePath: filepath.Join(stateDir, "history.json"),
		maxKeep:  maxKeep,
	}
}

func (hm *HistoryManager) Load() (*HistoryData, error) {
	data := &HistoryData{
		Current: -1,
		Entries: []HistoryEntry{},
	}

	if _, err := os.Stat(hm.filePath); os.IsNotExist(err) {
		return data, nil
	}

	bytes, err := os.ReadFile(hm.filePath)
	if err != nil {
		return data, err
	}

	if err := json.Unmarshal(bytes, data); err != nil {
		return data, err
	}

	return data, nil
}

func (hm *HistoryManager) Save(data *HistoryData) error {
	dir := filepath.Dir(hm.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	tmpFile := filepath.Join(dir, fmt.Sprintf(".history.%d.tmp", os.Getpid()))
	if err := os.WriteFile(tmpFile, bytes, 0600); err != nil {
		return err
	}

	return os.Rename(tmpFile, hm.filePath)
}

func (hm *HistoryManager) Record(entry HistoryEntry) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	data, err := hm.Load()
	if err != nil {
		return err
	}

	// Truncate forward history if current is before end
	if data.Current >= 0 && data.Current < len(data.Entries) {
		data.Entries = data.Entries[:data.Current+1]
	}

	// If last entry has the same ID, update it
	if len(data.Entries) > 0 && data.Entries[len(data.Entries)-1].ID == entry.ID {
		data.Entries[len(data.Entries)-1] = entry
	} else {
		data.Entries = append(data.Entries, entry)
	}

	// Limit to maxKeep
	if len(data.Entries) > hm.maxKeep {
		data.Entries = data.Entries[len(data.Entries)-hm.maxKeep:]
	}

	data.Current = len(data.Entries) - 1
	return hm.Save(data)
}

func (hm *HistoryManager) GetPrevious() (*HistoryEntry, int, error) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	data, err := hm.Load()
	if err != nil {
		return nil, -1, err
	}

	if data.Current < 1 || data.Current >= len(data.Entries) {
		return nil, -1, fmt.Errorf("wallpaper history has no earlier entry")
	}

	prevIdx := data.Current - 1
	entry := data.Entries[prevIdx]
	return &entry, prevIdx, nil
}

func (hm *HistoryManager) GetNext() (*HistoryEntry, int, error) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	data, err := hm.Load()
	if err != nil {
		return nil, -1, err
	}

	if data.Current < 0 || data.Current >= len(data.Entries)-1 {
		return nil, -1, fmt.Errorf("wallpaper history has no later entry")
	}

	nextIdx := data.Current + 1
	entry := data.Entries[nextIdx]
	return &entry, nextIdx, nil
}

func (hm *HistoryManager) SetCurrentIndex(idx int) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	data, err := hm.Load()
	if err != nil {
		return err
	}

	if idx < 0 || idx >= len(data.Entries) {
		return fmt.Errorf("invalid history index %d", idx)
	}

	data.Current = idx
	return hm.Save(data)
}

func MigrateLegacyState(targetStateDir string) {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	legacyDir := filepath.Join(home, ".local", "state", "wallchemy-rotate")
	if _, err := os.Stat(legacyDir); err != nil {
		return
	}

	_ = os.MkdirAll(targetStateDir, 0755)

	targetHistory := filepath.Join(targetStateDir, "history.json")
	legacyHistory := filepath.Join(legacyDir, "history.json")
	if _, err := os.Stat(targetHistory); os.IsNotExist(err) {
		if data, err := os.ReadFile(legacyHistory); err == nil {
			_ = os.WriteFile(targetHistory, data, 0644)
		}
	}

	targetMode := filepath.Join(targetStateDir, "mode-index")
	legacyMode := filepath.Join(legacyDir, "mode-index")
	if _, err := os.Stat(targetMode); os.IsNotExist(err) {
		if data, err := os.ReadFile(legacyMode); err == nil {
			_ = os.WriteFile(targetMode, data, 0644)
		}
	}

	targetLog := filepath.Join(targetStateDir, "last-run.log")
	legacyLog := filepath.Join(legacyDir, "last-run.log")
	if _, err := os.Stat(targetLog); os.IsNotExist(err) {
		if data, err := os.ReadFile(legacyLog); err == nil {
			_ = os.WriteFile(targetLog, data, 0644)
		}
	}
}
