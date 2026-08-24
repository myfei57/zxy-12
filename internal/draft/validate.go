package draft

import (
	"fmt"
	"strconv"

	"confighub/internal/item"
)

// Validator checks a draft against the item state.
type Validator struct {
	store *item.Store
}

// NewValidator creates a validator bound to an item store.
func NewValidator(store *item.Store) *Validator {
	return &Validator{store: store}
}

// Validate evaluates a draft against the constraint captured at creation.
func (v *Validator) Validate(d *Draft) error {
	// The validation uses the constraint copied when the draft was created, so
	// an item update while the draft is open is never reflected.
	maxLen, err := strconv.Atoi(d.ConstraintMax)
	if err != nil {
		return fmt.Errorf("draft: invalid constraint %q", d.ConstraintMax)
	}
	if len(d.Value) > maxLen {
		return fmt.Errorf("draft: value length %d exceeds constraint %d for %s", len(d.Value), maxLen, d.Key)
	}
	return nil
}
