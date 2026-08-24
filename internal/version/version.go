// Package version persists the namespace version ledger.
package version

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// State is the lifecycle state of a published version.
type State string

const (
	Draft      State = "draft"
	Publishing State = "publishing"
	Published  State = "published"
	RolledBack State = "rolled_back"
)

// Ledger is the persisted version record of one namespace.
type Ledger struct {
	Namespace  string `json:"namespace"`
	Revision   uint64 `json:"revision"`
	Published  uint64 `json:"published"`
	State      State  `json:"state"`
	DraftState string `json:"draft_state,omitempty"`
}

// Manager reads and advances the version ledger.
type Manager struct {
	mu   sync.Mutex
	path string
	led  *Ledger
}

// NewManager creates a version manager for a namespace.
func NewManager(path string) *Manager {
	return &Manager{path: path}
}

// Load reads the persisted ledger, if any.
func (m *Manager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	data, err := os.ReadFile(m.path)
	if err != nil {
		if os.IsNotExist(err) {
			m.led = &Ledger{State: Draft}
			return nil
		}
		return fmt.Errorf("version: read %s: %w", m.path, err)
	}
	var led Ledger
	if err := json.Unmarshal(data, &led); err != nil {
		return fmt.Errorf("version: parse %s: %w", m.path, err)
	}
	m.led = &led
	return nil
}

// Current returns the current ledger values.
func (m *Manager) Current() Ledger {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.led == nil {
		return Ledger{State: Draft}
	}
	return *m.led
}

// Revision returns the current revision number.
func (m *Manager) Revision() (uint64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.led == nil {
		return 0, fmt.Errorf("version: ledger not loaded")
	}
	return m.led.Revision, nil
}
