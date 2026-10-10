package app

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func FileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func DirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func GetNoctaliaActiveWallpaper() string {
	cmd := exec.Command("noctalia", "msg", "wallpaper-get")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func RunDmenu(prompt string, items []string, customCmd string) (string, error) {
	if len(items) == 0 {
		return "", nil
	}

	var cmd *exec.Cmd
	if customCmd != "" {
		parts := strings.Fields(customCmd)
		cmd = exec.Command(parts[0], parts[1:]...)
	} else if _, err := exec.LookPath("noctalia"); err == nil {
		cmd = exec.Command("noctalia", "dmenu", "-p", prompt)
	} else if _, err := exec.LookPath("rofi"); err == nil {
		cmd = exec.Command("rofi", "-dmenu", "-p", prompt)
	} else if _, err := exec.LookPath("wofi"); err == nil {
		cmd = exec.Command("wofi", "-dmenu", "-p", prompt)
	} else if _, err := exec.LookPath("dmenu"); err == nil {
		cmd = exec.Command("dmenu", "-p", prompt)
	} else {
		return "", fmt.Errorf("no menu launcher found (install noctalia, rofi, wofi, or dmenu)")
	}

	cmd.Stdin = strings.NewReader(strings.Join(items, "\n"))
	out, err := cmd.Output()
	if err != nil {
		// Non-zero exit code usually means user canceled/escaped
		return "", nil
	}

	return strings.TrimSpace(string(out)), nil
}

func OpenFileManager(dir string, preferredManager string) (string, error) {
	manager := preferredManager
	if manager == "" {
		if _, err := exec.LookPath("thunar"); err == nil {
			manager = "thunar"
		} else if runtime.GOOS == "darwin" {
			manager = "open"
		} else if runtime.GOOS == "windows" {
			manager = "explorer"
		} else {
			manager = "xdg-open"
		}
	}

	cmd := exec.Command(manager, dir)
	if err := cmd.Start(); err != nil {
		return manager, err
	}
	return manager, nil
}

func OpenBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}
