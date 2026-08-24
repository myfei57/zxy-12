// Package publish implements the config release flow.
package publish

import (
	"fmt"

	"confighub/internal/draft"
)

// Plan is one release operation.
type Plan struct {
	Namespace string
	Key       string
	Value     string
}

// BuildPlan creates a release plan from a validated draft.
func BuildPlan(d *draft.Draft) (*Plan, error) {
	if d == nil {
		return nil, fmt.Errorf("publish: nil draft")
	}
	return &Plan{Namespace: d.Namespace, Key: d.Key, Value: d.Value}, nil
}
