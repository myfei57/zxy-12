package watch

import (
	"fmt"

	"confighub/internal/version"
)

// Push delivers the currently published version to a subscriber.
func (m *Manager) Push(clientID string) (uint64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	sub, ok := m.subs[clientID]
	if !ok {
		return 0, fmt.Errorf("watch: client %s is not subscribed", clientID)
	}
	current := m.ledger.Current()
	if current.State != version.Published {
		return 0, fmt.Errorf("watch: namespace %s has no published version", sub.Namespace)
	}
	if _, err := m.snapshotter.Build(sub.Namespace, current.Published); err != nil {
		return 0, err
	}
	return current.Published, nil
}
