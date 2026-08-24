package verifycase

import (
	"path/filepath"
	"testing"

	"confighub/internal/draft"
	"confighub/internal/item"
)

func TestDraftValidatesCurrentItemState(t *testing.T) {
	dir := t.TempDir()
	st, err := item.NewStore(item.Options{DataDir: filepath.Join(dir, "data")})
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	op1, _ := st.Set("ns-01", "constraint.max", "10", 0)
	if err := st.Commit(op1.Seq); err != nil {
		t.Fatal(err)
	}
	d := draft.New("ns-01", "cfg.timeout", "1234567890", 1, "constraint.max", "10")
	op2, _ := st.Set("ns-01", "constraint.max", "5", 0)
	if err := st.Commit(op2.Seq); err != nil {
		t.Fatal(err)
	}
	validator := draft.NewValidator(st)
	if err := validator.Validate(d); err == nil {
		t.Fatal("validation must read the current constraint, not the stale draft copy")
	}
}
