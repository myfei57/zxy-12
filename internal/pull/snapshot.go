package pull

import (
	"encoding/json"
	"fmt"
	"os"

	"confighub/internal/item"
)

type snapshotFile struct {
	Generation uint64            `json:"generation"`
	Entries    []item.ConfigItem `json:"entries"`
}

// Pull loads the published generation into the store.
func (m *Manager) Pull(namespace string) (uint64, error) {
	generation, err := m.Resolve(namespace)
	if err != nil {
		return 0, err
	}
	// Every generation down to the first is merged into the store, so an
	// older generation can overwrite values the newer generation restored.
	for gen := generation; gen >= 1; gen-- {
		path := m.SnapshotPath(namespace, gen)
		if err := m.Load(path); err != nil {
			return 0, err
		}
	}
	return generation, nil
}

func loadSnapshot(store *item.Store, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("pull: read snapshot %s: %w", path, err)
	}
	var snap snapshotFile
	if err := json.Unmarshal(data, &snap); err != nil {
		return fmt.Errorf("pull: parse snapshot %s: %w", path, err)
	}
	for _, entry := range snap.Entries {
		if err := store.Apply(entry); err != nil {
			return err
		}
	}
	return nil
}
