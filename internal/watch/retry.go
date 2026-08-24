package watch

import "fmt"

// Retry re-pushes the latest version if the previous push was never acked.
func (m *Manager) Retry(clientID string) (uint64, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	sub, ok := m.subs[clientID]
	if !ok {
		return 0, false, fmt.Errorf("watch: client %s is not subscribed", clientID)
	}
	current := m.ledger.Current()
	if current.Published <= sub.Cursor {
		return current.Published, false, nil
	}
	if m.Acked(clientID, current.Published) {
		return current.Published, false, nil
	}
	if _, err := m.snapshotter.Build(sub.Namespace, current.Published); err != nil {
		return 0, false, err
	}
	return current.Published, true, nil
}
