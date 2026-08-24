// Package rollback restores the previous published configuration.
package rollback

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"confighub/internal/item"
	"confighub/internal/version"
)

// Target describes the version to roll back to.
type Target struct {
	Namespace string
	Key       string
	Value     string
	Revision  uint64
}

// ResolveTarget finds the previous published content for a key.
func ResolveTarget(store *item.Store, ledger *version.Manager, snapshotDir string, namespace string, key string) (*Target, error) {
	current := ledger.Current()
	if current.Published == 0 {
		return nil, fmt.Errorf("rollback: nothing to roll back for %s", namespace)
	}
	prevGen := current.Published - 1
	path := filepath.Join(snapshotDir, fmt.Sprintf("gen-%d", prevGen), "snapshot.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("rollback: read previous snapshot: %w", err)
	}
	var snap snapshotContent
	if err := json.Unmarshal(data, &snap); err != nil {
		return nil, fmt.Errorf("rollback: parse previous snapshot: %w", err)
	}
	value := ""
	for _, entry := range snap.Entries {
		if entry.Key == key {
			value = entry.Value
			break
		}
	}
	next, err := ledger.NextRevision()
	if err != nil {
		return nil, err
	}
	return &Target{
		Namespace: namespace,
		Key:       key,
		Value:     value,
		Revision:  next,
	}, nil
}

type snapshotContent struct {
	Generation uint64            `json:"generation"`
	Entries    []item.ConfigItem `json:"entries"`
}
