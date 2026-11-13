package config

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/zalando/go-keyring"
)

// Names used before the app was renamed to "Tracker for Kimai".
const (
	legacyAppDirName     = "kimai-tracker"
	legacyKeyringService = "kimai-tracker"
)

// MigrateLegacy moves data of earlier versions (directories and keyring
// entries named "kimai-tracker") to the current names. It never overwrites
// existing data and ignores errors: the app then simply starts fresh.
// dataDir is the current model/data directory (see voice.DataDir).
func MigrateLegacy(dataDir string) {
	if p, err := DefaultPath(); err == nil {
		moveDir(legacyOf(filepath.Dir(p)), filepath.Dir(p))
	}
	if p, err := DefaultCachePath("x"); err == nil {
		moveDir(legacyOf(filepath.Dir(p)), filepath.Dir(p))
	}
	if dataDir != "" {
		moveDir(legacyOf(dataDir), dataDir)
	}
	for _, key := range []string{SecretKimaiToken, SecretTranscriptionKey} {
		if _, err := keyring.Get(KeyringService, key); !errors.Is(err, keyring.ErrNotFound) {
			continue // already present, or keyring unavailable
		}
		if v, err := keyring.Get(legacyKeyringService, key); err == nil && v != "" {
			if keyring.Set(KeyringService, key, v) == nil {
				_ = keyring.Delete(legacyKeyringService, key)
			}
		}
	}
}

func legacyOf(dir string) string {
	return filepath.Join(filepath.Dir(dir), legacyAppDirName)
}

func moveDir(from, to string) {
	if _, err := os.Stat(to); err == nil {
		return
	}
	if _, err := os.Stat(from); err != nil {
		return
	}
	_ = os.Rename(from, to)
}
