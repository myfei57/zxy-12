package verifycase

import (
	"testing"

	"confighub/internal/cluster"
	"confighub/internal/settings"
)

func TestItemQuotaRejectsBeforeWrite(t *testing.T) {
	cfg := settings.Default()
	cfg.DataDir = t.TempDir()
	cfg.NamespaceCount = 1
	cfg.DefaultQuota = 10
	cl, err := cluster.Build(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer cl.Close()
	if err := cl.PutItem("ns-01", "cfg.a", "12345678901", 0); err == nil {
		t.Fatal("over-quota write must be rejected")
	}
	if _, ok := cl.Item("ns-01", "cfg.a"); ok {
		t.Fatal("over-quota config must not be stored")
	}
}
