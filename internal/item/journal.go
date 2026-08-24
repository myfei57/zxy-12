package item

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func osNanos() int64 {
	return time.Now().UnixNano()
}

func readJournal(path string) ([]Op, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var ops []Op
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var op Op
		if err := json.Unmarshal([]byte(line), &op); err != nil {
			return nil, err
		}
		ops = append(ops, op)
	}
	return ops, nil
}

func readTombstones(path string) ([]tombstone, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var entries []tombstone
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var entry tombstone
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func readRestoreLog(path string) ([]restoreEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var entries []restoreEntry
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var entry restoreEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func (s *Store) recover() error {
	committed, err := readMetaSeq(s.metaPath("commit.meta"))
	if err != nil {
		return err
	}
	s.committed = committed
	ops, err := readJournal(s.journal.Name())
	if err != nil {
		return err
	}
	for _, op := range ops {
		if op.Seq > committed || op.Kind != "set" {
			continue
		}
		if op.Seq > s.seq {
			s.seq = op.Seq
		}
		s.index[op.Key] = &ConfigItem{
			Namespace: op.Namespace,
			Key:       op.Key,
			Value:     op.Value,
			Version:   op.Version,
			Expires:   op.Expires,
			Seq:       op.Seq,
		}
	}
	deleteSeq, err := readMetaSeq(s.deleteMetaPath())
	if err != nil {
		return err
	}
	entries, err := readTombstones(s.tombstones.Name())
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.Seq > deleteSeq {
			continue
		}
		if entry.Seq > s.seq {
			s.seq = entry.Seq
		}
		delete(s.index, entry.Key)
	}
	restoreMeta := s.opts.RestoreMetaPath
	if restoreMeta == "" {
		restoreMeta = s.metaPath("restore.meta")
	}
	restoreSeq, err := readMetaSeq(restoreMeta)
	if err != nil {
		return err
	}
	restoreEntries, err := readRestoreLog(filepath.Join(s.opts.DataDir, "restore.log"))
	if err == nil {
		for _, entry := range restoreEntries {
			if entry.Seq > restoreSeq {
				continue
			}
			if entry.Seq > s.seq {
				s.seq = entry.Seq
			}
			s.index[entry.Key] = &ConfigItem{
				Key:     entry.Key,
				Value:   entry.Value,
				Version: uint64(entry.Seq),
				Seq:     entry.Seq,
			}
		}
	}
	applied, err := readMetaSeq(s.metaPath("applied.meta"))
	if err != nil {
		return err
	}
	s.applied = applied
	return nil
}
