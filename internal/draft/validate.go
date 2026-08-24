package draft

import (
	"fmt"
	"strconv"

	"confighub/internal/item"
)

// Validator checks a draft against the current item state.
type Validator struct {
	store *item.Store
}

// NewValidator creates a validator bound to an item store.
func NewValidator(store *item.Store) *Validator {
	return &Validator{store: store}
}

// Validate evaluates a draft against the live item state.
func (v *Validator) Validate(d *Draft) error {
	current, ok := v.store.Entry(d.ConstraintKey)
	if !ok {
		return fmt.Errorf("draft: constraint item %s is missing", d.ConstraintKey)
	}
	maxLen, err := strconv.Atoi(current.Value)
	if err != nil {
		return fmt.Errorf("draft: constraint item %s has invalid value %q", d.ConstraintKey, current.Value)
	}
	if len(d.Value) > maxLen {
		return fmt.Errorf("draft: value length %d exceeds constraint %d for %s", len(d.Value), maxLen, d.Key)
	}
	return nil
}
