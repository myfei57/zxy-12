package item

import "fmt"

// Apply applies a snapshot entry to this store and records the applied offset.
func (s *Store) Apply(entry ConfigItem) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if entry.Seq > s.seq {
		s.seq = entry.Seq
	}
	version := entry.Version
	if version == 0 {
		version = uint64(1)
		if prev, ok := s.index[entry.Key]; ok {
			version = prev.Version + 1
		}
	}
	s.index[entry.Key] = &ConfigItem{
		Namespace: entry.Namespace,
		Key:       entry.Key,
		Value:     entry.Value,
		Version:   version,
		Expires:   entry.Expires,
		Seq:       entry.Seq,
	}
	if entry.Seq > s.applied {
		s.applied = entry.Seq
	}
	if err := writeMetaAtomic(s.metaPath("applied.meta"), metaSeq{Seq: s.applied}); err != nil {
		return fmt.Errorf("item: write applied meta: %w", err)
	}
	return nil
}

// AppliedOffset returns the highest snapshot seq applied to this store.
func (s *Store) AppliedOffset() (uint64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.applied, nil
}
