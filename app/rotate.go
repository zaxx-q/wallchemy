package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

type RotateManager struct {
	mu       sync.Mutex
	stateDir string
}

func NewRotateManager(stateDir string) *RotateManager {
	_ = os.MkdirAll(stateDir, 0755)
	return &RotateManager{stateDir: stateDir}
}

func (rm *RotateManager) GetNextMode(availableModes []string) (string, error) {
	if len(availableModes) == 0 {
		return "", fmt.Errorf("no rotation modes available")
	}

	rm.mu.Lock()
	defer rm.mu.Unlock()

	indexPath := filepath.Join(rm.stateDir, "mode-index")
	index := 0
	if bytes, err := os.ReadFile(indexPath); err == nil {
		trimmed := strings.TrimSpace(string(bytes))
		if val, err := strconv.Atoi(trimmed); err == nil && val >= 0 {
			index = val
		}
	}

	mode := availableModes[index%len(availableModes)]
	nextIndex := (index + 1) % len(availableModes)

	_ = os.MkdirAll(rm.stateDir, 0755)
	_ = os.WriteFile(indexPath, []byte(strconv.Itoa(nextIndex)+"\n"), 0644)

	return mode, nil
}
