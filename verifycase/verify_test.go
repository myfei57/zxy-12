package verifycase

import (
	"path/filepath"
	"testing"

	"confighub/internal/audit"
	"confighub/internal/draft"
	"confighub/internal/item"
	"confighub/internal/publish"
	"confighub/internal/snapshot"
	"confighub/internal/version"
)

func TestAuditAfterPublishDurable(t *testing.T) {
	dir := t.TempDir()
	st, err := item.NewStore(item.Options{
		DataDir: filepath.Join(dir, "data"),
		MetaDir: filepath.Join(dir, "missing-meta"),
	})
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	drafts := draft.NewRegistry()
	if err := drafts.Put(draft.New("ns-01", "cfg.a", "1", 0, "constraint.max", "64")); err != nil {
		t.Fatal(err)
	}
	ledger := version.NewManager(filepath.Join(dir, "version.json"))
	if err := ledger.Load(); err != nil {
		t.Fatal(err)
	}
	auditLogger := audit.NewLogger(filepath.Join(dir, "audit.log"))
	snap := snapshot.NewManager(st, filepath.Join(dir, "snapshots"))
	if err := publish.Release(st, ledger, drafts, auditLogger, snap, "ns-01", "cfg.a", "1"); err == nil {
		t.Fatal("publish must fail when the config write is not durable")
	}
	for _, ev := range auditLogger.Entries(0) {
		if ev.Type == "publish" {
			t.Fatal("audit must not record success when the publish write is not durable")
		}
	}
}
