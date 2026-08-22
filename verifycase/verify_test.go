package verifycase

import (
	"path/filepath"
	"testing"

	"confighub/internal/item"
	"confighub/internal/snapshot"
	"confighub/internal/version"
	"confighub/internal/watch"
)

func TestWatchRetryNoDuplicatePush(t *testing.T) {
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
	wm := watch.NewManager(st, ledger, filepath.Join(dir, "cursor"), snapshot.NewAckState(filepath.Join(dir, "acks")), snap)
	if err := wm.Subscribe("client-1", "ns-01"); err != nil {
		t.Fatal(err)
	}
	if _, err := wm.Push("client-1"); err != nil {
		t.Fatal(err)
	}
	if err := wm.Ack("client-1", 1); err != nil {
		t.Fatal(err)
	}
	_, pushed, err := wm.Retry("client-1")
	if err != nil {
		t.Fatal(err)
	}
	if pushed {
		t.Fatal("an acknowledged version must not be pushed again by retry")
	}
}
