// Package quota tracks configuration capacity per namespace.
package quota

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Manager enforces a hard capacity bound for a namespace.
type Manager struct {
	mu         sync.Mutex
	capacity   int64
	used       int64
	sizes      map[string]int64
	ledgerPath string
}

// NewManager creates a quota manager with the given capacity.
func NewManager(capacity int64, ledgerPath string) *Manager {
	return &Manager{capacity: capacity, sizes: make(map[string]int64), ledgerPath: ledgerPath}
}

// Capacity returns the configured capacity.
func (m *Manager) Capacity() int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.capacity
}

// Used returns the currently occupied capacity.
func (m *Manager) Used() int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.used
}

// Size returns the recorded size of a key.
func (m *Manager) Size(key string) (int64, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	size, ok := m.sizes[key]
	return size, ok
}

type releaseEntry struct {
	Key  string `json:"key"`
	Size int64  `json:"size"`
}

// Release durably records the quota release before freeing the capacity.
func (m *Manager) Release(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	size, ok := m.sizes[key]
	if !ok {
		return fmt.Errorf("quota: no usage recorded for %s", key)
	}
	entry := releaseEntry{Key: key, Size: size}
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	if m.ledgerPath != "" {
		if err := os.MkdirAll(filepath.Dir(m.ledgerPath), 0o755); err != nil {
			return fmt.Errorf("quota: ledger dir: %w", err)
		}
		file, err := os.OpenFile(m.ledgerPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			return fmt.Errorf("quota: open ledger: %w", err)
		}
		if _, err := file.Write(append(data, '\n')); err != nil {
			file.Close()
			return fmt.Errorf("quota: ledger write: %w", err)
		}
		if err := file.Sync(); err != nil {
			file.Close()
			return fmt.Errorf("quota: ledger sync: %w", err)
		}
		if err := file.Close(); err != nil {
			return fmt.Errorf("quota: ledger close: %w", err)
		}
	}
	delete(m.sizes, key)
	m.used -= size
	return nil
}
