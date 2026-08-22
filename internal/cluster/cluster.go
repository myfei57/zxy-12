// Package cluster wires namespaces, stores, versions, watch and the console.
package cluster

import (
	"fmt"
	"os"
	"sync"

	"confighub/internal/audit"
	"confighub/internal/console"
	"confighub/internal/draft"
	"confighub/internal/item"
	"confighub/internal/ns"
	"confighub/internal/publish"
	"confighub/internal/pull"
	"confighub/internal/quota"
	"confighub/internal/rollback"
	"confighub/internal/settings"
	"confighub/internal/snapshot"
	"confighub/internal/version"
	"confighub/internal/watch"
)

// NamespaceRuntime bundles the per-namespace components.
type NamespaceRuntime struct {
	Store    *item.Store
	Ledger   *version.Manager
	Snapshot *snapshot.Manager
	Quota    *quota.Manager
	Pull     *pull.Manager
	Watch    *watch.Manager
}

func newItemStore(dataDir string, metaDir string) (*item.Store, error) {
	return item.NewStore(item.Options{DataDir: dataDir, MetaDir: metaDir})
}

// Cluster is the assembled ConfigHub process.
type Cluster struct {
	Cfg     settings.Settings
	NS      *ns.Registry
	Drafts  *draft.Registry
	Audit   *audit.Logger
	Runtime map[string]*NamespaceRuntime
	Server  *console.Server
	mu      sync.Mutex
}

// Build creates the cluster and its on-disk state.
func Build(cfg settings.Settings) (*Cluster, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		return nil, fmt.Errorf("cluster: mkdir data: %w", err)
	}
	reg := ns.NewRegistry()
	for _, n := range ns.BuildInitial(cfg.NamespaceCount, cfg.NodeID) {
		if err := reg.Register(n); err != nil {
			return nil, err
		}
	}
	auditLogger := audit.NewLogger(settings.AuditPath(cfg.DataDir))
	drafts := draft.NewRegistry()
	c := &Cluster{
		Cfg:     cfg,
		NS:      reg,
		Drafts:  drafts,
		Audit:   auditLogger,
		Runtime: make(map[string]*NamespaceRuntime),
	}
	for _, id := range reg.IDs() {
		rt, err := buildNamespaceRuntime(cfg, id, auditLogger, drafts)
		if err != nil {
			return nil, err
		}
		c.Runtime[id] = rt
	}
	c.Server = console.NewServer(c)
	return c, nil
}

func buildNamespaceRuntime(cfg settings.Settings, id string, auditLogger *audit.Logger, drafts *draft.Registry) (*NamespaceRuntime, error) {
	dataDir := settings.NamespaceDir(cfg.DataDir, id)
	metaDir := settings.MetaDir(cfg.DataDir, id)
	if err := os.MkdirAll(metaDir, 0o755); err != nil {
		return nil, err
	}
	store, err := newItemStore(dataDir, metaDir)
	if err != nil {
		return nil, err
	}
	ledger := version.NewManager(settings.VersionPath(cfg.DataDir, id))
	if err := ledger.Load(); err != nil {
		return nil, err
	}
	snapDir := settings.SnapshotDir(cfg.DataDir, id)
	snapshotManager := snapshot.NewManager(store, snapDir)
	quotaManager := quota.NewManager(cfg.DefaultQuota, settings.QuotaLedgerPath(cfg.DataDir))
	pullManager := pull.NewManager(store, ledger, snapDir)
	watchManager := watch.NewManager(
		store,
		ledger,
		settings.WatchCursorPath(cfg.DataDir, id),
		snapshot.NewAckState(settings.AckPath(cfg.DataDir, id)),
		snapshotManager,
	)
	return &NamespaceRuntime{
		Store:    store,
		Ledger:   ledger,
		Snapshot: snapshotManager,
		Quota:    quotaManager,
		Pull:     pullManager,
		Watch:    watchManager,
	}, nil
}

// Start serves the console until the process exits.
func (c *Cluster) Start() error {
	_ = c.Audit.Note("start", "", "", c.Cfg.HTTPAddr)
	return c.Server.Start(c.Cfg.HTTPAddr)
}

// Close flushes every namespace store.
func (c *Cluster) Close() error {
	for _, id := range c.NS.IDs() {
		if rt := c.Runtime[id]; rt != nil {
			if err := rt.Store.Close(); err != nil {
				return err
			}
		}
	}
	return nil
}

// Publish delegates to the publish flow for a namespace.
func (c *Cluster) Publish(namespace string, key string, value string) error {
	rt, ok := c.Runtime[namespace]
	if !ok {
		return fmt.Errorf("cluster: unknown namespace %s", namespace)
	}
	return publish.Release(rt.Store, rt.Ledger, c.Drafts, c.Audit, rt.Snapshot, namespace, key, value)
}

// Rollback delegates to the rollback flow for a namespace.
func (c *Cluster) Rollback(namespace string, key string) error {
	rt, ok := c.Runtime[namespace]
	if !ok {
		return fmt.Errorf("cluster: unknown namespace %s", namespace)
	}
	return rollback.Run(rt.Store, rt.Ledger, c.Drafts, c.Audit, settings.SnapshotDir(c.Cfg.DataDir, namespace), namespace, key)
}
