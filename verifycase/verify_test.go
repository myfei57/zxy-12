package verifycase

import (
	"os"
	"path/filepath"
	"testing"

	"confighub/internal/item"
	"confighub/internal/snapshot"
	"confighub/internal/version"
	"confighub/internal/watch"
)

func TestWatchCursorAdvancesAfterAckDurable(t *testing.T) {
	dir := t.TempDir()
	st, err := item.NewStore(item.Options{DataDir: filepath.Join(dir, "data")})
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
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
	snap := snapshot.NewManager(st, filepath.Join(dir, "snapshots"))
	if _, err := snap.Build("ns-01", 1); err != nil {
		t.Fatal(err)
	}
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	acks := snapshot.NewAckState(filepath.Join(blocker, "acks"))
	wm := watch.NewManager(st, ledger, filepath.Join(dir, "cursor"), acks, snap)
	if err := wm.Subscribe("client-1", "ns-01"); err != nil {
		t.Fatal(err)
	}
	if err := wm.Deliver("client-1", 1); err == nil {
		t.Fatal("deliver must fail when the ack cannot be recorded")
	}
	cur, err := wm.Cursor("client-1")
	if err != nil {
		t.Fatal(err)
	}
	if cur != 0 {
		t.Fatalf("cursor advanced to %d although the ack was not durable", cur)
	}
}
