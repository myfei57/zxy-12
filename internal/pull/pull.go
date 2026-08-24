// Package pull serves client snapshot pulls for a namespace.
package pull

import (
	"fmt"
	"path/filepath"

	"confighub/internal/item"
	"confighub/internal/version"
)

// Manager resolves and loads client snapshots.
type Manager struct {
	store  *item.Store
	ledger *version.Manager
	dir    string
}

// NewManager creates a pull manager.
func NewManager(store *item.Store, ledger *version.Manager, dir string) *Manager {
	return &Manager{store: store, ledger: ledger, dir: dir}
}

// Resolve returns the generation that a client should pull.
func (m *Manager) Resolve(namespace string) (uint64, error) {
	current := m.ledger.Current()
	if current.Published == 0 {
		return 0, fmt.Errorf("pull: namespace %s has no published version", namespace)
	}
	return current.Published, nil
}

// SnapshotPath returns the snapshot file for a generation.
func (m *Manager) SnapshotPath(namespace string, generation uint64) string {
	return filepath.Join(m.dir, fmt.Sprintf("gen-%d", generation), "snapshot.json")
}

// Load applies one snapshot generation into the store.
func (m *Manager) Load(path string) error {
	return loadSnapshot(m.store, path)
}
