package verifycase

import (
	"path/filepath"
	"testing"

	"confighub/internal/item"
	"confighub/internal/pull"
	"confighub/internal/snapshot"
	"confighub/internal/version"
)

func TestPullUsesSingleSnapshotVersion(t *testing.T) {
	dir := t.TempDir()
	snapDir := filepath.Join(dir, "snapshots")
	st1, err := item.NewStore(item.Options{DataDir: filepath.Join(dir, "s1")})
	if err != nil {
		t.Fatal(err)
	}
	defer st1.Close()
	op1, _ := st1.Set("ns-01", "cfg.a", "1", 0)
	if err := st1.Commit(op1.Seq); err != nil {
		t.Fatal(err)
	}
	snap1 := snapshot.NewManager(st1, snapDir)
	if _, err := snap1.Build("ns-01", 1); err != nil {
		t.Fatal(err)
	}
	st2, err := item.NewStore(item.Options{DataDir: filepath.Join(dir, "s2")})
	if err != nil {
		t.Fatal(err)
	}
	defer st2.Close()
	op2, _ := st2.Set("ns-01", "cfg.a", "2", 0)
	if err := st2.Commit(op2.Seq); err != nil {
		t.Fatal(err)
	}
	snap2 := snapshot.NewManager(st2, snapDir)
	if _, err := snap2.Build("ns-01", 2); err != nil {
		t.Fatal(err)
	}
	target, err := item.NewStore(item.Options{DataDir: filepath.Join(dir, "target")})
	if err != nil {
		t.Fatal(err)
	}
	defer target.Close()
	ledger := version.NewManager(filepath.Join(dir, "version.json"))
	if err := ledger.Load(); err != nil {
		t.Fatal(err)
	}
	if err := ledger.Advance(1); err != nil {
		t.Fatal(err)
	}
	if err := ledger.SetPublished(1); err != nil {
		t.Fatal(err)
	}
	if err := ledger.Advance(2); err != nil {
		t.Fatal(err)
	}
	if err := ledger.SetPublished(2); err != nil {
		t.Fatal(err)
	}
	pm := pull.NewManager(target, ledger, snapDir)
	gen, err := pm.Pull("ns-01")
	if err != nil {
		t.Fatal(err)
	}
	if gen != 2 {
		t.Fatalf("pull generation = %d, want 2", gen)
	}
	got, ok := target.Get("cfg.a")
	if !ok {
		t.Fatal("cfg.a missing after pull")
	}
	if got != "2" {
		t.Fatalf("pull must load only the current generation, got %q", got)
	}
}
