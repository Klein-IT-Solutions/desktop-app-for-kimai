// Package config loads and stores the user configuration in
// ~/.config/kimai-tracker/config.json. Secrets (API tokens, keys) are never
// written to this file; see secrets.go.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// AppDirName is the directory name below the user config dir.
const AppDirName = "kimai-tracker"

// Config holds all non-secret settings.
type Config struct {
	// KimaiURL is the base URL of the Kimai instance, e.g. https://kimai.example.com.
	KimaiURL string `json:"kimaiUrl"`
	// Language is the UI / transcription language as ISO 639-1 code.
	Language string `json:"language"`
}

// Default returns the configuration used when no file exists yet.
func Default() Config {
	return Config{Language: "de"}
}

// Store reads and writes the configuration file. It is safe for concurrent use.
type Store struct {
	path string
	mu   sync.Mutex
	cfg  Config
}

// DefaultPath returns ~/.config/kimai-tracker/config.json (honouring XDG_CONFIG_HOME).
func DefaultPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("determine config directory: %w", err)
	}
	return filepath.Join(dir, AppDirName, "config.json"), nil
}

// Open loads the configuration at path. A missing file yields the defaults.
func Open(path string) (*Store, error) {
	s := &Store{path: path, cfg: Default()}
	data, err := os.ReadFile(path)
	switch {
	case errors.Is(err, os.ErrNotExist):
		return s, nil
	case err != nil:
		return nil, fmt.Errorf("read config: %w", err)
	}
	if err := json.Unmarshal(data, &s.cfg); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}
	if s.cfg.Language == "" {
		s.cfg.Language = Default().Language
	}
	return s, nil
}

// Path returns the location of the configuration file.
func (s *Store) Path() string { return s.path }

// Get returns a copy of the current configuration.
func (s *Store) Get() Config {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.cfg
}

// Update applies fn to the configuration and persists the result atomically.
func (s *Store) Update(fn func(*Config)) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	next := s.cfg
	fn(&next)
	if err := s.write(next); err != nil {
		return err
	}
	s.cfg = next
	return nil
}

func (s *Store) write(cfg Config) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(s.path), ".config-*.json")
	if err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(append(data, '\n')); err != nil {
		tmp.Close()
		return fmt.Errorf("write config: %w", err)
	}
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return fmt.Errorf("write config: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	if err := os.Rename(tmp.Name(), s.path); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}
