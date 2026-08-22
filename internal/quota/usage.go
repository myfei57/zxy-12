package quota

// Available returns how much capacity remains.
func (m *Manager) Available() int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.capacity - m.used
}

// Reset clears the recorded usage (used by namespace rebuild).
func (m *Manager) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sizes = make(map[string]int64)
	m.used = 0
}
