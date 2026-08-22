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

func TestPublishIgnoresExpiredDraft(t *testing.T) {
	dir := t.TempDir()
	st, err := item.NewStore(item.Options{DataDir: filepath.Join(dir, "data")})
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
	if err := drafts.Supersede("ns-01", "cfg.a"); err != nil {
		t.Fatal(err)
	}
	if err := publish.Release(st, ledger, drafts, auditLogger, snap, "ns-01", "cfg.a", "1"); err == nil {
		t.Fatal("publish must reject a superseded draft")
	}
	rev, err := ledger.Revision()
	if err != nil {
		t.Fatal(err)
	}
	if rev != 0 {
		t.Fatalf("revision advanced to %d with an expired draft", rev)
	}
}
