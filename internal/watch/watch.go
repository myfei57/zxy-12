package watch

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"confighub/internal/item"
	"confighub/internal/snapshot"
	"confighub/internal/version"
)

// Subscription is one client watching a namespace.
type Subscription struct {
	ClientID  string `json:"client_id"`
	Namespace string `json:"namespace"`
	Cursor    uint64 `json:"cursor"`
}

// Manager tracks subscriptions and their push cursors.
type Manager struct {
	mu             sync.Mutex
	subs           map[string]*Subscription
	cursorRoot     string
	store          *item.Store
	ledger         *version.Manager
	acks           *snapshot.AckState
	snapshotter    SnapshotBuilder
	cachedPublished map[string]uint64
}

// SnapshotBuilder materializes a version snapshot for a namespace.
type SnapshotBuilder interface {
	Build(namespace string, published uint64) (string, error)
}

// NewManager creates a watch manager.
func NewManager(
	store *item.Store,
	ledger *version.Manager,
	cursorRoot string,
	acks *snapshot.AckState,
	snapshotter SnapshotBuilder,
) *Manager {
	return &Manager{
		subs:            make(map[string]*Subscription),
		cursorRoot:      cursorRoot,
		store:           store,
		ledger:          ledger,
		acks:            acks,
		snapshotter:     snapshotter,
		cachedPublished: make(map[string]uint64),
	}
}

// Subscribe registers a client and captures the currently published version.
func (m *Manager) Subscribe(clientID string, namespace string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	sub := &Subscription{ClientID: clientID, Namespace: namespace}
	data, err := os.ReadFile(m.cursorPath(clientID))
	if err == nil {
		if err := json.Unmarshal(data, sub); err != nil {
			return fmt.Errorf("watch: parse cursor: %w", err)
		}
	}
	m.subs[clientID] = sub
	m.cachedPublished[clientID] = m.ledger.Current().Published
	return nil
}

func (m *Manager) cursorPath(clientID string) string {
	return filepath.Join(m.cursorRoot, clientID+".json")
}

// Subs lists the registered subscriptions.
func (m *Manager) Subs() []Subscription {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Subscription, 0, len(m.subs))
	for _, sub := range m.subs {
		out = append(out, *sub)
	}
	return out
}

// Ack records that a client durably accepted a pushed version.
func (m *Manager) Ack(clientID string, version uint64) error {
	return m.acks.Record(clientID, version)
}

// Acked reports whether a version was already acknowledged.
func (m *Manager) Acked(clientID string, version uint64) bool {
	return m.acks.Acked(clientID, version)
}

// Deliver records the subscriber acknowledgement and then advances the cursor.
func (m *Manager) Deliver(clientID string, version uint64) error {
	if err := m.Ack(clientID, version); err != nil {
		return err
	}
	return m.AdvanceCursor(clientID, version)
}
