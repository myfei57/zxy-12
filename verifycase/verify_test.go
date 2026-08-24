package verifycase

import (
	"path/filepath"
	"testing"

	"confighub/internal/audit"
	"confighub/internal/draft"
	"confighub/internal/item"
	"confighub/internal/publish"
	"confighub/internal/rollback"
	"confighub/internal/snapshot"
	"confighub/internal/version"
)

func TestRollbackRevisionAfterDurable(t *testing.T) {
	dir := t.TempDir()
	st, err := item.NewStore(item.Options{
		DataDir:         filepath.Join(dir, "data"),
		RestoreMetaPath: filepath.Join(dir, "missing-meta", "restore.meta"),
	})
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	drafts := draft.NewRegistry()
	ledger := version.NewManager(filepath.Join(dir, "version.json"))
	if err := ledger.Load(); err != nil {
		t.Fatal(err)
	}
	auditLogger := audit.NewLogger(filepath.Join(dir, "audit.log"))
	snap := snapshot.NewManager(st, filepath.Join(dir, "snapshots"))
	for _, value := range []string{"1", "2"} {
		if err := drafts.Put(draft.New("ns-01", "cfg.a", value, 0, "constraint.max", "64")); err != nil {
			t.Fatal(err)
		}
		if err := publish.Release(st, ledger, drafts, auditLogger, snap, "ns-01", "cfg.a", value); err != nil {
			t.Fatal(err)
		}
	}
	if err := rollback.Run(st, ledger, drafts, auditLogger, filepath.Join(dir, "snapshots"), "ns-01", "cfg.a"); err == nil {
		t.Fatal("rollback must fail when the restore is not durable")
	}
	rev, err := ledger.Revision()
	if err != nil {
		t.Fatal(err)
	}
	if rev != 2 {
		t.Fatalf("rollback revision advanced to %d although the restore was not durable", rev)
	}
}
