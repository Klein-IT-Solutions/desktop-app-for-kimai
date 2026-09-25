package hotkey

import (
	"context"
	"sync"
	"time"
)

// Backends.
const (
	BackendX11    = "x11"
	BackendPortal = "portal"
)

// Status describes whether the shortcut works and how.
type Status struct {
	Active bool `json:"active"`
	// Pending is set while the desktop asks the user to confirm the shortcut.
	Pending bool   `json:"pending"`
	Backend string `json:"backend,omitempty"`
	// Shortcut is the canonical accelerator, e.g. "Super+Shift+T".
	Shortcut string `json:"shortcut"`
	// Trigger is what the desktop actually bound (portal), if reported.
	Trigger string `json:"trigger,omitempty"`
	// Message explains a failure or limitation (German, for the UI).
	Message string `json:"message,omitempty"`
}

// Manager owns the active shortcut registration.
type Manager struct {
	onPress  func()
	onStatus func(Status)

	mu     sync.Mutex
	cancel context.CancelFunc
	status Status
	last   time.Time
}

// NewManager creates a manager. onPress runs on every activation (debounced),
// onStatus whenever the registration status changes.
func NewManager(onPress func(), onStatus func(Status)) *Manager {
	return &Manager{onPress: onPress, onStatus: onStatus}
}

// Status returns the current registration status.
func (m *Manager) Status() Status {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.status
}

// Register replaces the current shortcut. An empty accelerator disables it.
func (m *Manager) Register(accelerator string) Status {
	m.Stop()
	if accelerator == "" {
		return m.set(Status{Message: "Kein globales Tastenkürzel festgelegt."})
	}
	acc, err := Parse(accelerator)
	if err != nil {
		return m.set(Status{Shortcut: accelerator, Message: "Ungültiges Tastenkürzel: " + err.Error()})
	}
	ctx, cancel := context.WithCancel(context.Background())
	m.mu.Lock()
	m.cancel = cancel
	m.mu.Unlock()
	return m.set(register(ctx, acc, m.press, func(s Status) {
		if ctx.Err() == nil {
			m.set(s)
		}
	}))
}

// Stop releases the shortcut.
func (m *Manager) Stop() {
	m.mu.Lock()
	cancel := m.cancel
	m.cancel = nil
	m.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

// press debounces key repeat and duplicate events.
func (m *Manager) press() {
	m.mu.Lock()
	if time.Since(m.last) < 300*time.Millisecond {
		m.mu.Unlock()
		return
	}
	m.last = time.Now()
	m.mu.Unlock()
	if m.onPress != nil {
		m.onPress()
	}
}

func (m *Manager) set(s Status) Status {
	m.mu.Lock()
	m.status = s
	m.mu.Unlock()
	if m.onStatus != nil {
		m.onStatus(s)
	}
	return s
}
