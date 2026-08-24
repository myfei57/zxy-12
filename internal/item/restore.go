package item

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Restore durably applies a rollback value through its own restore channel.
func (s *Store) Restore(namespace string, key string, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry := restoreEntry{Seq: s.seq + 1, Key: key, Value: value}
	line, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	path := filepath.Join(s.opts.DataDir, "restore.log")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("item: open restore log: %w", err)
	}
	if _, err := file.Write(append(line, '\n')); err != nil {
		file.Close()
		return fmt.Errorf("item: restore log write: %w", err)
	}
	if err := file.Sync(); err != nil {
		file.Close()
		return fmt.Errorf("item: restore log sync: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("item: restore log close: %w", err)
	}
	metaPath := s.opts.RestoreMetaPath
	if metaPath == "" {
		metaPath = s.metaPath("restore.meta")
	}
	if err := writeMetaAtomic(metaPath, metaSeq{Seq: entry.Seq}); err != nil {
		return fmt.Errorf("item: restore meta: %w", err)
	}
	s.seq = entry.Seq
	s.index[key] = &ConfigItem{
		Namespace: namespace,
		Key:       key,
		Value:     value,
		Version:   uint64(entry.Seq),
		Seq:       entry.Seq,
	}
	return nil
}

type restoreEntry struct {
	Seq   uint64 `json:"seq"`
	Key   string `json:"key"`
	Value string `json:"value"`
}
