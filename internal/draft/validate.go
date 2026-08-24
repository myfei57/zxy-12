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

// Validate evaluates a draft against the constraint's current state.
func (v *Validator) Validate(d *Draft) error {
	// Read the constraint live from the store so that an update, expiry, or
	// deletion of the constraint item while the draft is open is reflected at
	// validation time. Using the value cached on the draft when it was created
	// would validate against a stale constraint and let through values that the
	// current configuration no longer permits.
	constraint, ok := v.store.Get(d.ConstraintKey)
	if !ok {
		return fmt.Errorf("draft: constraint %s not available for %s", d.ConstraintKey, d.Key)
	}
	maxLen, err := strconv.Atoi(constraint)
	if err != nil {
		return fmt.Errorf("draft: invalid constraint %q", constraint)
	}
	if len(d.Value) > maxLen {
		return fmt.Errorf("draft: value length %d exceeds constraint %d for %s", len(d.Value), maxLen, d.Key)
	}
	return nil
}
