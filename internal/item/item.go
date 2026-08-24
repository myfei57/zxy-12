// Package item implements the file-backed configuration item store.
package item

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// ConfigItem is one stored configuration entry.
type ConfigItem struct {
	Namespace string `json:"namespace"`
	Key       string `json:"key"`
	Value     string `json:"value"`
	Version   uint64 `json:"version"`
	Expires   int64  `json:"expires"`
	Seq       uint64 `json:"seq"`
}

// Op is one journaled mutation.
type Op struct {
	Seq       uint64 `json:"seq"`
	Kind      string `json:"kind"`
	Namespace string `json:"namespace"`
	Key       string `json:"key"`
	Value     string `json:"value"`
	Version   uint64 `json:"version"`
	Expires   int64  `json:"expires"`
}

// Options controls where a store keeps its files.
type Options struct {
	DataDir         string
	MetaDir         string
	TombstonePath   string
	DeleteMetaPath  string
	RestoreMetaPath string
}

// Store is a namespace-local file-backed configuration store.
type Store struct {
	opts       Options
	mu         sync.Mutex
	seq        uint64
	committed  uint64
	applied    uint64
	index      map[string]*ConfigItem
	journal    *os.File
	tombstones *os.File
}

// NewStore opens or recreates a store at the given paths.
func NewStore(opts Options) (*Store, error) {
	if opts.DataDir == "" {
		return nil, fmt.Errorf("item: data dir required")
	}
	if err := os.MkdirAll(opts.DataDir, 0o755); err != nil {
		return nil, fmt.Errorf("item: mkdir %s: %w", opts.DataDir, err)
	}
	journal, err := os.OpenFile(filepath.Join(opts.DataDir, "journal.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("item: open journal: %w", err)
	}
	tombstonePath := opts.TombstonePath
	if tombstonePath == "" {
		tombstonePath = filepath.Join(opts.DataDir, "tombstone.log")
	}
	if err := os.MkdirAll(filepath.Dir(tombstonePath), 0o755); err != nil {
		journal.Close()
		return nil, fmt.Errorf("item: mkdir tombstones: %w", err)
	}
	tombstones, err := os.OpenFile(tombstonePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		journal.Close()
		return nil, fmt.Errorf("item: open tombstones: %w", err)
	}
	s := &Store{
		opts:       opts,
		index:      make(map[string]*ConfigItem),
		journal:    journal,
		tombstones: tombstones,
	}
	if err := s.recover(); err != nil {
		journal.Close()
		tombstones.Close()
		return nil, err
	}
	return s, nil
}

// Close flushes and closes the store files.
func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.journal.Sync(); err != nil {
		return err
	}
	if err := s.journal.Close(); err != nil {
		return err
	}
	return s.tombstones.Close()
}

// Set buffers a config write in the journal and updates the in-memory index.
func (s *Store) Set(namespace string, key string, value string, ttl int64) (Op, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seq++
	version := uint64(1)
	if prev, ok := s.index[key]; ok {
		version = prev.Version + 1
	}
	op := Op{
		Seq:       s.seq,
		Kind:      "set",
		Namespace: namespace,
		Key:       key,
		Value:     value,
		Version:   version,
		Expires:   ttl,
	}
	line, err := json.Marshal(op)
	if err != nil {
		return Op{}, err
	}
	if _, err := s.journal.Write(append(line, '\n')); err != nil {
		return Op{}, fmt.Errorf("item: journal write: %w", err)
	}
	s.index[key] = &ConfigItem{
		Namespace: namespace,
		Key:       key,
		Value:     value,
		Version:   version,
		Expires:   ttl,
		Seq:       op.Seq,
	}
	return op, nil
}

func (s *Store) metaPath(name string) string {
	if s.opts.MetaDir == "" {
		return filepath.Join(s.opts.DataDir, name)
	}
	return filepath.Join(s.opts.MetaDir, name)
}

func (s *Store) deleteMetaPath() string {
	if s.opts.DeleteMetaPath != "" {
		return s.opts.DeleteMetaPath
	}
	return s.metaPath("delete.meta")
}
