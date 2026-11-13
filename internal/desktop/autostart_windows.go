//go:build windows

package desktop

import (
	"errors"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

// The app starts at login through the per-user Run key; it starts hidden in
// the tray (--hidden).
const (
	runKey   = `Software\Microsoft\Windows\CurrentVersion\Run`
	runValue = "Tracker for Kimai"

	// legacyRunValue was used before the rename to "Tracker for Kimai".
	legacyRunValue = "Kimai Tracker"
)

// AutostartEnabled reports whether the Run entry exists.
func AutostartEnabled() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKey, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()
	_, _, err = k.GetStringValue(runValue)
	return err == nil
}

// SetAutostart creates or removes the Run entry.
func SetAutostart(enabled bool) error {
	k, _, err := registry.CreateKey(registry.CURRENT_USER, runKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	if !enabled {
		if err := k.DeleteValue(runValue); err != nil && !errors.Is(err, registry.ErrNotExist) {
			return err
		}
		return nil
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	return k.SetStringValue(runValue, `"`+exe+`" --hidden`)
}

// MigrateAutostart replaces the autostart entry of versions before 1.0.0
// (value "Kimai Tracker") with the current one.
func MigrateAutostart() {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKey, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		return
	}
	defer k.Close()
	if _, _, err := k.GetStringValue(legacyRunValue); err != nil {
		return
	}
	_ = k.DeleteValue(legacyRunValue)
	_ = SetAutostart(true)
}
