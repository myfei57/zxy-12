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
	"confighub/internal/watch"
)

func TestWatchDeliversNewSnapshot(t *testing.T) {
	dir := t.TempDir()
	st, err := item.NewStore(item.Options{DataDir: filepath.Join(dir, "data")})
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
	if err := drafts.Put(draft.New("ns-01", "cfg.a", "1", 0, "constraint.max", "64")); err != nil {
		t.Fatal(err)
	}
	if err := publish.Release(st, ledger, drafts, auditLogger, snap, "ns-01", "cfg.a", "1"); err != nil {
		t.Fatal(err)
	}
	wm := watch.NewManager(st, ledger, filepath.Join(dir, "cursor"), snapshot.NewAckState(filepath.Join(dir, "acks")), snap)
	if err := wm.Subscribe("client-1", "ns-01"); err != nil {
		t.Fatal(err)
	}
	if err := drafts.Put(draft.New("ns-01", "cfg.a", "2", 0, "constraint.max", "64")); err != nil {
		t.Fatal(err)
	}
	if err := publish.Release(st, ledger, drafts, auditLogger, snap, "ns-01", "cfg.a", "2"); err != nil {
		t.Fatal(err)
	}
	v, err := wm.Push("client-1")
	if err != nil {
		t.Fatal(err)
	}
	if v != 2 {
		t.Fatalf("push must deliver the current published version, got %d", v)
	}
}
