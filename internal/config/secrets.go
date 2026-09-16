package config

import (
	"errors"
	"fmt"

	"github.com/zalando/go-keyring"
)

// KeyringService is the service name under which secrets are stored.
const KeyringService = "kimai-tracker"

// Secret identifiers.
const (
	SecretKimaiToken = "kimai-api-token"
)

// ErrSecretNotFound is returned when no secret is stored under a key.
var ErrSecretNotFound = errors.New("secret not found")

// ErrKeyringUnavailable is returned when the system keyring cannot be reached
// (e.g. no Secret Service provider such as GNOME Keyring or KWallet is running).
var ErrKeyringUnavailable = errors.New("system keyring unavailable")

// Secrets stores sensitive values outside the config file.
type Secrets interface {
	Get(key string) (string, error)
	Set(key, value string) error
	Delete(key string) error
}

// Keyring stores secrets in the operating system keyring
// (Secret Service on Linux, Keychain on macOS, Credential Manager on Windows).
type Keyring struct{}

// Get returns the secret or ErrSecretNotFound.
func (Keyring) Get(key string) (string, error) {
	v, err := keyring.Get(KeyringService, key)
	return v, wrapKeyringErr(err)
}

// Set stores the secret.
func (Keyring) Set(key, value string) error {
	return wrapKeyringErr(keyring.Set(KeyringService, key, value))
}

// Delete removes the secret. Deleting a missing secret is not an error.
func (Keyring) Delete(key string) error {
	err := wrapKeyringErr(keyring.Delete(KeyringService, key))
	if errors.Is(err, ErrSecretNotFound) {
		return nil
	}
	return err
}

func wrapKeyringErr(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, keyring.ErrNotFound):
		return ErrSecretNotFound
	default:
		return fmt.Errorf("%w: %v", ErrKeyringUnavailable, err)
	}
}

// MemorySecrets is an in-memory Secrets implementation for tests.
type MemorySecrets map[string]string

// Get implements Secrets.
func (m MemorySecrets) Get(key string) (string, error) {
	v, ok := m[key]
	if !ok {
		return "", ErrSecretNotFound
	}
	return v, nil
}

// Set implements Secrets.
func (m MemorySecrets) Set(key, value string) error { m[key] = value; return nil }

// Delete implements Secrets.
func (m MemorySecrets) Delete(key string) error { delete(m, key); return nil }
