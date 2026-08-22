package publish

import (
	"fmt"

	"confighub/internal/audit"
	"confighub/internal/draft"
	"confighub/internal/item"
	"confighub/internal/snapshot"
	"confighub/internal/version"
)

// Release publishes a draft as the new configuration version.
func Release(
	store *item.Store,
	ledger *version.Manager,
	drafts *draft.Registry,
	auditLogger *audit.Logger,
	snapshots *snapshot.Manager,
	namespace string,
	key string,
	value string,
) error {
	d, ok := drafts.Get(namespace, key)
	if !ok {
		return fmt.Errorf("publish: no draft for %s/%s", namespace, key)
	}
	// The draft lifecycle is not consulted, so a superseded draft is reused
	// and its expired content is published again.
	_ = d
	op, err := store.Set(namespace, key, value, 0)
	if err != nil {
		return err
	}
	if err := store.Commit(op.Seq); err != nil {
		return err
	}
	revision, err := ledger.NextRevision()
	if err != nil {
		return err
	}
	if err := ledger.Advance(revision); err != nil {
		return err
	}
	if err := ledger.SetPublished(revision); err != nil {
		return err
	}
	if _, err := snapshots.Build(namespace, revision); err != nil {
		return err
	}
	return auditLogger.Note("publish", namespace, key, fmt.Sprintf("revision %d", revision))
}
