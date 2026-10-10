package app

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Notifier struct {
	mu      sync.Mutex
	enabled bool
	lastID  string
	logFile string
}

func NewNotifier(enabled bool, logFile string) *Notifier {
	return &Notifier{
		enabled: enabled,
		logFile: logFile,
	}
}

func (n *Notifier) Log(msg string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	entry := fmt.Sprintf("[%s] %s\n", timestamp, msg)

	if n.logFile != "" {
		if err := os.MkdirAll(filepath.Dir(n.logFile), 0755); err == nil {
			f, err := os.OpenFile(n.logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err == nil {
				_, _ = f.WriteString(entry)
				_ = f.Close()
			}
		}
	}

	if Logger != nil {
		Logger.Debug(msg)
	}
}

func (n *Notifier) NotifyStart(title, msg string) {
	n.Log(fmt.Sprintf("Starting: %s — %s", title, msg))

	if !n.enabled {
		return
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	cmd := exec.Command("notify-send",
		"-p",
		"-a", "Wallchemy",
		"-i", "preferences-desktop-wallpaper",
		"-h", "string:x-canonical-private-synchronous:wallchemy",
		title, msg,
	)
	out, err := cmd.Output()
	if err == nil {
		n.lastID = strings.TrimSpace(string(out))
	}
}

func (n *Notifier) NotifySuccess(title, msg, icon string) {
	n.Log(fmt.Sprintf("Success: %s — %s", title, msg))

	if !n.enabled {
		return
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	if icon == "" {
		icon = "preferences-desktop-wallpaper"
	}

	var args []string
	if n.lastID != "" {
		args = append(args, "-r", n.lastID)
	}
	args = append(args,
		"-a", "Wallchemy",
		"-i", icon,
		"-h", "string:x-canonical-private-synchronous:wallchemy",
		title, msg,
	)

	_ = exec.Command("notify-send", args...).Run()
}

func (n *Notifier) NotifyError(title, msg string) {
	n.Log(fmt.Sprintf("ERROR: %s — %s", title, msg))

	if !n.enabled {
		return
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	var args []string
	if n.lastID != "" {
		args = append(args, "-r", n.lastID)
	}
	args = append(args,
		"-u", "critical",
		"-a", "Wallchemy",
		"-i", "dialog-error",
		"-h", "string:x-canonical-private-synchronous:wallchemy",
		title, msg,
	)

	_ = exec.Command("notify-send", args...).Run()
}
