package item

import (
	"encoding/json"
	"fmt"
)

// Delete durably removes a config key from the store.
func (s *Store) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry := tombstone{Seq: s.seq + 1, Key: key}
	line, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	if _, err := s.tombstones.Write(append(line, '\n')); err != nil {
		return fmt.Errorf("item: tombstone write: %w", err)
	}
	if err := s.tombstones.Sync(); err != nil {
		return fmt.Errorf("item: tombstone sync: %w", err)
	}
	if err := writeMetaAtomic(s.deleteMetaPath(), metaSeq{Seq: entry.Seq}); err != nil {
		return fmt.Errorf("item: delete meta: %w", err)
	}
	s.seq = entry.Seq
	delete(s.index, key)
	return nil
}
