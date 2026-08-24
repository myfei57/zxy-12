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
	// A snapshot generation is a full snapshot of every key for its version,
	// so switching to the target version is a single load. Concatenating older
	// generations on top would let stale values overwrite the target version,
	// mixing old and new values for the same key; the store must cut over to
	// the target generation in one shot rather than merging file by file.
	path := m.SnapshotPath(namespace, generation)
	if err := m.Load(path); err != nil {
		return 0, err
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
