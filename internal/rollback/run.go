package rollback

import (
	"fmt"

	"confighub/internal/audit"
	"confighub/internal/draft"
	"confighub/internal/item"
	"confighub/internal/version"
)

// Run rolls a key back to the target content and advances the revision.
func Run(
	store *item.Store,
	ledger *version.Manager,
	drafts *draft.Registry,
	auditLogger *audit.Logger,
	snapshotDir string,
	namespace string,
	key string,
) error {
	target, err := ResolveTarget(store, ledger, snapshotDir, namespace, key)
	if err != nil {
		return err
	}
	// The old config must be durably restored before the revision advances;
	// otherwise a failed restore leaves the revision pointing at a value that
	// never landed on disk.
	if err := store.Restore(namespace, key, target.Value); err != nil {
		return err
	}
	if err := ledger.Advance(target.Revision); err != nil {
		return err
	}
	if err := ledger.SetPublished(target.Revision); err != nil {
		return err
	}
	if err := drafts.Supersede(namespace, key); err != nil {
		return err
	}
	return auditLogger.Note("rollback", namespace, key, fmt.Sprintf("revision %d", target.Revision))
}
