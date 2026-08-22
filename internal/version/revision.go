package version

// NextRevision returns the revision that follows the current one.
func (m *Manager) NextRevision() (uint64, error) {
	current, err := m.Revision()
	if err != nil {
		return 0, err
	}
	return current + 1, nil
}
