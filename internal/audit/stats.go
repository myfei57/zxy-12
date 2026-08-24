package audit

// Counts returns the number of events per type.
func (l *Logger) Counts() map[string]int {
	counts := map[string]int{}
	for _, ev := range l.Entries(0) {
		counts[ev.Type]++
	}
	return counts
}

// FilterByType returns events whose type matches the given kind.
func (l *Logger) FilterByType(kind string) []Event {
	var out []Event
	for _, ev := range l.Entries(0) {
		if ev.Type == kind {
			out = append(out, ev)
		}
	}
	return out
}
