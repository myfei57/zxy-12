package settings

import "path/filepath"

// NamespaceDir returns the on-disk directory owned by one namespace.
func NamespaceDir(root string, namespace string) string {
	return filepath.Join(root, "namespaces", namespace)
}

// MetaDir returns the directory used for durability watermarks.
func MetaDir(root string, namespace string) string {
	return filepath.Join(root, "meta", namespace)
}

// VersionPath returns the file that persists the namespace version ledger.
func VersionPath(root string, namespace string) string {
	return filepath.Join(root, "versions", namespace+".json")
}

// SnapshotDir returns the directory that holds client snapshot generations.
func SnapshotDir(root string, namespace string) string {
	return filepath.Join(root, "snapshots", namespace)
}

// AuditPath returns the audit log file shared by the cluster.
func AuditPath(root string) string {
	return filepath.Join(root, "audit.log")
}

// QuotaLedgerPath returns the durable quota release ledger.
func QuotaLedgerPath(root string) string {
	return filepath.Join(root, "quota-ledger.log")
}

// WatchCursorPath returns the directory that stores push cursors.
func WatchCursorPath(root string, namespace string) string {
	return filepath.Join(root, "watch", namespace)
}

// AckPath returns the directory that stores subscriber acknowledgement records.
func AckPath(root string, namespace string) string {
	return filepath.Join(root, "acks", namespace)
}
