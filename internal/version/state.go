package version

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Advance durably moves the ledger to a new revision.
func (m *Manager) Advance(revision uint64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.led == nil {
		return fmt.Errorf("version: ledger not loaded")
	}
	if revision <= m.led.Revision {
		return fmt.Errorf("version: revision %d is not newer than %d", revision, m.led.Revision)
	}
	next := *m.led
	next.Revision = revision
	next.State = Published
	if err := writeLedger(m.path, &next); err != nil {
		return err
	}
	m.led = &next
	return nil
}

// SetPublished marks the published version number.
func (m *Manager) SetPublished(published uint64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.led == nil {
		return fmt.Errorf("version: ledger not loaded")
	}
	next := *m.led
	next.Published = published
	if err := writeLedger(m.path, &next); err != nil {
		return err
	}
	m.led = &next
	return nil
}

func writeLedger(path string, led *Ledger) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("version: mkdir: %w", err)
	}
	data, err := json.Marshal(led)
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
