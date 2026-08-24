package draft

import "fmt"

// MarkValidated moves a draft into the validated state.
func (r *Registry) MarkValidated(namespace string, key string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	d, ok := r.drafts[namespace+"/"+key]
	if !ok {
		return fmt.Errorf("draft: unknown draft %s/%s", namespace, key)
	}
	if d.State == Superseded {
		return fmt.Errorf("draft: %s/%s is superseded", namespace, key)
	}
	d.State = Validated
	return nil
}

// Supersede invalidates a draft (for example after a rollback).
func (r *Registry) Supersede(namespace string, key string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	d, ok := r.drafts[namespace+"/"+key]
	if !ok {
		return fmt.Errorf("draft: unknown draft %s/%s", namespace, key)
	}
	d.State = Superseded
	return nil
}

// StateOf returns the current state of a draft.
func (r *Registry) StateOf(namespace string, key string) (State, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	d, ok := r.drafts[namespace+"/"+key]
	if !ok {
		return "", false
	}
	return d.State, true
}

// All lists the drafts of a namespace.
func (r *Registry) All(namespace string) []*Draft {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []*Draft
	for _, d := range r.drafts {
		if d.Namespace == namespace {
			out = append(out, d)
		}
	}
	return out
}

// Count returns the number of drafts in a namespace.
func (r *Registry) Count(namespace string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	count := 0
	for _, d := range r.drafts {
		if d.Namespace == namespace {
			count++
		}
	}
	return count
}
