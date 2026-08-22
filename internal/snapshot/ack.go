package snapshot

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// AckState durably records subscriber acknowledgements for snapshot versions.
type AckState struct {
	dir string
}

// NewAckState creates an acknowledgement store under dir.
func NewAckState(dir string) *AckState {
	return &AckState{dir: dir}
}

// Record durably stores an acknowledgement for a client and version.
func (a *AckState) Record(clientID string, version uint64) error {
	path := filepath.Join(a.dir, clientID, fmt.Sprintf("ack-%d.json", version))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(map[string]any{"client": clientID, "version": version})
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// Acked reports whether a version was already durably acknowledged.
func (a *AckState) Acked(clientID string, version uint64) bool {
	_, err := os.Stat(filepath.Join(a.dir, clientID, fmt.Sprintf("ack-%d.json", version)))
	return err == nil
}
