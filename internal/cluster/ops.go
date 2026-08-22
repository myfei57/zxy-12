package cluster

import (
	"fmt"
	"path/filepath"
	"time"

	"confighub/internal/audit"
	"confighub/internal/console"
	"confighub/internal/draft"
	"confighub/internal/settings"
)

// Namespaces lists the namespace registry for the console.
func (c *Cluster) Namespaces() []console.NamespaceInfo {
	var out []console.NamespaceInfo
	for _, id := range c.NS.IDs() {
		n, _ := c.NS.Get(id)
		out = append(out, console.NamespaceInfo{ID: n.ID, Name: n.Name, Owner: n.Owner})
	}
	return out
}

// Items lists stored config items in a namespace.
func (c *Cluster) Items(namespace string, prefix string) []console.ItemInfo {
	rt, ok := c.Runtime[namespace]
	if !ok {
		return nil
	}
	var out []console.ItemInfo
	for _, key := range rt.Store.Keys(prefix) {
		if entry, ok := rt.Store.Entry(key); ok {
			out = append(out, console.ItemInfo{Key: key, Value: entry.Value, Version: entry.Version, Expires: entry.Expires})
		}
	}
	return out
}

// Item returns one stored config item.
func (c *Cluster) Item(namespace string, key string) (console.ItemInfo, bool) {
	rt, ok := c.Runtime[namespace]
	if !ok {
		return console.ItemInfo{}, false
	}
	entry, ok := rt.Store.Entry(key)
	if !ok {
		return console.ItemInfo{}, false
	}
	return console.ItemInfo{Key: key, Value: entry.Value, Version: entry.Version, Expires: entry.Expires}, true
}

// PutItem writes a config item through the quota-gated path.
func (c *Cluster) PutItem(namespace string, key string, value string, ttlMs int64) error {
	rt, ok := c.Runtime[namespace]
	if !ok {
		return fmt.Errorf("cluster: unknown namespace %s", namespace)
	}
	expires := int64(0)
	if ttlMs > 0 {
		expires = time.Now().UnixNano() + ttlMs*int64(time.Millisecond)
	}
	return writeItem(rt, namespace, key, value, expires)
}

// DeleteItem removes a config item from its namespace.
func (c *Cluster) DeleteItem(namespace string, key string) error {
	rt, ok := c.Runtime[namespace]
	if !ok {
		return fmt.Errorf("cluster: unknown namespace %s", namespace)
	}
	if !rt.Store.Has(key) {
		return fmt.Errorf("cluster: item %s not found", key)
	}
	if err := rt.Store.Delete(key); err != nil {
		return err
	}
	_ = rt.Quota.Release(key)
	return c.Audit.Note("delete", namespace, key, "")
}

// Version returns the version ledger of a namespace.
func (c *Cluster) Version(namespace string) console.VersionInfo {
	rt, ok := c.Runtime[namespace]
	if !ok {
		return console.VersionInfo{}
	}
	led := rt.Ledger.Current()
	return console.VersionInfo{Revision: led.Revision, Published: led.Published, State: string(led.State)}
}

// Snapshots lists the snapshot generations of a namespace.
func (c *Cluster) Snapshots(namespace string) []console.SnapshotInfo {
	rt, ok := c.Runtime[namespace]
	if !ok {
		return nil
	}
	names := rt.Snapshot.Files()
	out := make([]console.SnapshotInfo, 0, len(names))
	for _, name := range names {
		out = append(out, console.SnapshotInfo{
			Generation: name,
			Path:       filepath.Join(settings.SnapshotDir(c.Cfg.DataDir, namespace), name),
		})
	}
	return out
}

// Pull loads the published snapshot generation into the namespace store.
func (c *Cluster) Pull(namespace string) (uint64, error) {
	rt, ok := c.Runtime[namespace]
	if !ok {
		return 0, fmt.Errorf("cluster: unknown namespace %s", namespace)
	}
	return rt.Pull.Pull(namespace)
}

// Subscriptions lists the registered watch subscriptions.
func (c *Cluster) Subscriptions() []console.SubscriptionInfo {
	var out []console.SubscriptionInfo
	for _, id := range c.NS.IDs() {
		rt := c.Runtime[id]
		for _, sub := range rt.Watch.Subs() {
			out = append(out, console.SubscriptionInfo{ClientID: sub.ClientID, Namespace: sub.Namespace, Cursor: sub.Cursor})
		}
	}
	return out
}

// DraftList lists the pending drafts of a namespace.
func (c *Cluster) DraftList(namespace string) []console.DraftInfo {
	var out []console.DraftInfo
	for _, d := range c.Drafts.All(namespace) {
		out = append(out, console.DraftInfo{Namespace: d.Namespace, Key: d.Key, Value: d.Value, State: string(d.State)})
	}
	return out
}

// Search finds config keys whose value contains the term.
func (c *Cluster) Search(namespace string, term string) []string {
	rt, ok := c.Runtime[namespace]
	if !ok {
		return nil
	}
	return rt.Store.Search(term)
}

// EnsureDraft creates a draft for a namespace and key.
func (c *Cluster) EnsureDraft(namespace string, key string, value string) error {
	rt, ok := c.Runtime[namespace]
	if !ok {
		return fmt.Errorf("cluster: unknown namespace %s", namespace)
	}
	entry, ok := rt.Store.Entry(key)
	if !ok {
		return fmt.Errorf("cluster: item %s missing", key)
	}
	constraint, _ := rt.Store.Get("constraint.max")
	d := draft.New(namespace, key, value, entry.Seq, "constraint.max", constraint)
	return c.Drafts.Put(d)
}

// ValidateDraft validates a draft against the current item state.
func (c *Cluster) ValidateDraft(namespace string, key string) error {
	rt, ok := c.Runtime[namespace]
	if !ok {
		return fmt.Errorf("cluster: unknown namespace %s", namespace)
	}
	d, ok := c.Drafts.Get(namespace, key)
	if !ok {
		return fmt.Errorf("cluster: no draft for %s/%s", namespace, key)
	}
	validator := draft.NewValidator(rt.Store)
	return validator.Validate(d)
}

// Push pushes the published version to a client.
func (c *Cluster) Push(clientID string) (uint64, error) {
	rt, err := c.runtimeForClient(clientID)
	if err != nil {
		return 0, err
	}
	return rt.Watch.Push(clientID)
}

// Retry re-pushes the latest version when the previous push was not acked.
func (c *Cluster) Retry(clientID string) (uint64, bool, error) {
	rt, err := c.runtimeForClient(clientID)
	if err != nil {
		return 0, false, err
	}
	return rt.Watch.Retry(clientID)
}

// Deliver records a subscriber acknowledgement and advances the cursor.
func (c *Cluster) Deliver(clientID string, version uint64) error {
	rt, err := c.runtimeForClient(clientID)
	if err != nil {
		return err
	}
	return rt.Watch.Deliver(clientID, version)
}

// AuditEntries returns the newest audit events.
func (c *Cluster) AuditEntries(limit int) []audit.Event {
	return c.Audit.Entries(limit)
}

// AuditCount returns the number of events recorded in this process.
func (c *Cluster) AuditCount() int {
	return c.Audit.Count()
}

// AuditCounts returns event counts grouped by type.
func (c *Cluster) AuditCounts() map[string]int {
	return c.Audit.Counts()
}

// Quota returns the capacity usage of a namespace.
func (c *Cluster) Quota(namespace string) console.QuotaInfo {
	rt, ok := c.Runtime[namespace]
	if !ok {
		return console.QuotaInfo{}
	}
	return console.QuotaInfo{Capacity: rt.Quota.Capacity(), Used: rt.Quota.Used()}
}

func (c *Cluster) runtimeForClient(clientID string) (*NamespaceRuntime, error) {
	for _, id := range c.NS.IDs() {
		rt := c.Runtime[id]
		for _, sub := range rt.Watch.Subs() {
			if sub.ClientID == clientID {
				return rt, nil
			}
		}
	}
	return nil, fmt.Errorf("cluster: client %s is not subscribed", clientID)
}
