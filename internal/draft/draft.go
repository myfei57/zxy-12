// Package draft manages configuration drafts and their lifecycle.
package draft

import (
	"fmt"
	"sync"
)

// State is the lifecycle state of a draft.
type State string

const (
	Drafting   State = "drafting"
	Validated  State = "validated"
	Superseded State = "superseded"
)

// Draft is one pending configuration change.
type Draft struct {
	ID            string `json:"id"`
	Namespace     string `json:"namespace"`
	Key           string `json:"key"`
	Value         string `json:"value"`
	State         State  `json:"state"`
	ItemSeq       uint64 `json:"item_seq"`
	ConstraintKey string `json:"constraint_key"`
	ConstraintMax string `json:"constraint_max"`
}

// New creates a draft bound to the item snapshot it was created against.
func New(namespace string, key string, value string, itemSeq uint64, constraintKey string, constraintMax string) *Draft {
	return &Draft{
		ID:            namespace + "/" + key,
		Namespace:     namespace,
		Key:           key,
		Value:         value,
		State:         Drafting,
		ItemSeq:       itemSeq,
		ConstraintKey: constraintKey,
		ConstraintMax: constraintMax,
	}
}

// Registry tracks drafts per namespace.
type Registry struct {
	mu     sync.Mutex
	drafts map[string]*Draft
}

// NewRegistry returns an empty draft registry.
func NewRegistry() *Registry {
	return &Registry{drafts: make(map[string]*Draft)}
}

// Put stores a draft.
func (r *Registry) Put(d *Draft) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if d == nil || d.ID == "" {
		return fmt.Errorf("draft: cannot store an empty draft")
	}
	r.drafts[d.ID] = d
	return nil
}

// Get returns the draft for a namespace and key.
func (r *Registry) Get(namespace string, key string) (*Draft, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	d, ok := r.drafts[namespace+"/"+key]
	return d, ok
}
