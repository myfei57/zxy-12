package item

import "fmt"

// Commit durably records that every journal op up to seq is committed.
func (s *Store) Commit(seq uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if seq > s.seq {
		return fmt.Errorf("item: commit seq %d beyond journal %d", seq, s.seq)
	}
	if seq <= s.committed {
		return nil
	}
	if err := s.journal.Sync(); err != nil {
		return fmt.Errorf("item: sync journal: %w", err)
	}
	if err := writeMetaAtomic(s.metaPath("commit.meta"), metaSeq{Seq: seq}); err != nil {
		return fmt.Errorf("item: write commit meta: %w", err)
	}
	s.committed = seq
	return nil
}
