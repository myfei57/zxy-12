package cluster

// writeItem writes a config item.
func writeItem(rt *NamespaceRuntime, namespace string, key string, value string, expires int64) error {
	size := int64(len(value))
	// The capacity gate must run before the write reaches the journal, so an
	// over-quota write is rejected before any data is persisted.
	if err := rt.Quota.Check(key, size); err != nil {
		return err
	}
	op, err := rt.Store.Set(namespace, key, value, expires)
	if err != nil {
		return err
	}
	if err := rt.Store.Commit(op.Seq); err != nil {
		return err
	}
	if err := rt.Quota.Account(key, size); err != nil {
		return err
	}
	return nil
}
