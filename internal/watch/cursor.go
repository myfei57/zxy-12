package watch

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// AdvanceCursor durably moves a subscription cursor after its ack lands.
func (m *Manager) AdvanceCursor(clientID string, version uint64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	sub, ok := m.subs[clientID]
	if !ok {
		return fmt.Errorf("watch: client %s is not subscribed", clientID)
	}
	sub.Cursor = version
	if err := os.MkdirAll(filepath.Dir(m.cursorPath(clientID)), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(sub)
	if err != nil {
		return err
	}
	tmp := m.cursorPath(clientID) + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, m.cursorPath(clientID))
}

// Cursor returns the current cursor of a subscription.
func (m *Manager) Cursor(clientID string) (uint64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	sub, ok := m.subs[clientID]
	if !ok {
		return 0, fmt.Errorf("watch: client %s is not subscribed", clientID)
	}
	return sub.Cursor, nil
}
