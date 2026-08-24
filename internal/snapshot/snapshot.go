// Package snapshot builds and restores client snapshot generations.
package snapshot

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"confighub/internal/item"
)

// Manager builds snapshot generations for a namespace.
type Manager struct {
	store *item.Store
	dir   string
}

// NewManager creates a snapshot manager for a namespace store.
func NewManager(store *item.Store, dir string) *Manager {
	return &Manager{store: store, dir: dir}
}

// Build writes a snapshot generation for a published version.
func (m *Manager) Build(namespace string, published uint64) (string, error) {
	path := filepath.Join(m.dir, fmt.Sprintf("gen-%d", published), "snapshot.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", fmt.Errorf("snapshot: mkdir: %w", err)
	}
	entries := make([]item.ConfigItem, 0, m.store.Count())
	for _, key := range m.store.Keys("") {
		if entry, ok := m.store.Entry(key); ok {
			entries = append(entries, entry)
		}
	}
	data, err := json.Marshal(snapshotFile{Generation: published, Entries: entries})
	if err != nil {
		return "", err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return "", fmt.Errorf("snapshot: write: %w", err)
	}
	if err := syncPath(tmp); err != nil {
		return "", err
	}
	if err := os.Rename(tmp, path); err != nil {
		return "", fmt.Errorf("snapshot: publish: %w", err)
	}
	return path, nil
}

// Files lists the snapshot generations present on disk.
func (m *Manager) Files() []string {
	entries, err := os.ReadDir(m.dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, entry := range entries {
		if entry.IsDir() && len(entry.Name()) > 4 && entry.Name()[:4] == "gen-" {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	return names
}

type snapshotFile struct {
	Generation uint64            `json:"generation"`
	Entries    []item.ConfigItem `json:"entries"`
}

func syncPath(path string) error {
	f, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	return f.Sync()
}
