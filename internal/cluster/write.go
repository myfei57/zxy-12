package cluster

// writeItem writes a config item.
func writeItem(rt *NamespaceRuntime, namespace string, key string, value string, expires int64) error {
	// The config is stored before the capacity gate runs, so an over-quota
	// write still lands on disk before the error is returned.
	op, err := rt.Store.Set(namespace, key, value, expires)
	if err != nil {
		return err
	}
	if err := rt.Store.Commit(op.Seq); err != nil {
		return err
	}
	if err := rt.Quota.Check(key, int64(len(value))); err != nil {
		return err
	}
	if err := rt.Quota.Account(key, int64(len(value))); err != nil {
		return err
	}
	return nil
}
