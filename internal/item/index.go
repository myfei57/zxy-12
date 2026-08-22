package item

import "sort"

// Get returns the value stored for a key that has not expired.
func (s *Store) Get(key string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.index[key]
	if !ok {
		return "", false
	}
	if entry.Expires != 0 && entry.Expires <= nowNanos() {
		return "", false
	}
	return entry.Value, true
}

// Has reports whether a key currently exists in the index.
func (s *Store) Has(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.index[key]
	return ok
}

// Entry returns the full stored item for a key.
func (s *Store) Entry(key string) (ConfigItem, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.index[key]
	if !ok {
		return ConfigItem{}, false
	}
	return *entry, true
}

// Keys lists keys with the given prefix, sorted.
func (s *Store) Keys(prefix string) []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	var keys []string
	for key := range s.index {
		if len(prefix) == 0 || len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}

// Search returns keys whose stored value contains the given term.
func (s *Store) Search(term string) []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	var keys []string
	for key, entry := range s.index {
		if term != "" && len(entry.Value) >= len(term) && contains(entry.Value, term) {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}

func contains(value string, term string) bool {
	for i := 0; i+len(term) <= len(value); i++ {
		if value[i:i+len(term)] == term {
			return true
		}
	}
	return false
}

// LastSeq returns the highest operation sequence seen by the store.
func (s *Store) LastSeq() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.seq
}

// Count returns the number of keys in the index.
func (s *Store) Count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.index)
}

// JournalPath returns the path of the write-ahead journal.
func (s *Store) JournalPath() string {
	return s.journal.Name()
}
