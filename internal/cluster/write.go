package cluster

// writeItem writes a config item through the quota-gated path.
func writeItem(rt *NamespaceRuntime, namespace string, key string, value string, expires int64) error {
	if err := rt.Quota.Check(key, int64(len(value))); err != nil {
		return err
	}
	op, err := rt.Store.Set(namespace, key, value, expires)
	if err != nil {
		return err
	}
	if err := rt.Store.Commit(op.Seq); err != nil {
		return err
	}
	if err := rt.Quota.Account(key, int64(len(value))); err != nil {
		return err
	}
	return nil
}
