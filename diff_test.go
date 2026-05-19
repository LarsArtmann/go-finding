package finding

import "testing"

const diffTestKeep = "keep"

func TestDiff_Empty(t *testing.T) {
	t.Parallel()

	result := Diff(nil, nil)
	if len(result.Added) != 0 || len(result.Removed) != 0 || len(result.Unchanged) != 0 {
		t.Errorf("Diff(nil, nil) = added=%d removed=%d unchanged=%d, want all 0",
			len(result.Added), len(result.Removed), len(result.Unchanged))
	}
}

func TestDiff_AllAdded(t *testing.T) {
	t.Parallel()

	after := []Finding{
		{ID: "a"}, {ID: "b"},
	}

	result := Diff(nil, after)
	if len(result.Added) != 2 {
		t.Errorf("Added = %d, want 2", len(result.Added))
	}

	if len(result.Removed) != 0 || len(result.Unchanged) != 0 {
		t.Errorf("Removed=%d Unchanged=%d, want 0", len(result.Removed), len(result.Unchanged))
	}
}

func TestDiff_AllRemoved(t *testing.T) {
	t.Parallel()

	before := []Finding{
		{ID: "a"}, {ID: "b"},
	}

	result := Diff(before, nil)
	if len(result.Removed) != 2 {
		t.Errorf("Removed = %d, want 2", len(result.Removed))
	}

	if len(result.Added) != 0 || len(result.Unchanged) != 0 {
		t.Errorf("Added=%d Unchanged=%d, want 0", len(result.Added), len(result.Unchanged))
	}
}

func TestDiff_Mixed(t *testing.T) {
	t.Parallel()

	before := []Finding{
		{ID: diffTestKeep}, {ID: "remove"},
	}
	after := []Finding{
		{ID: diffTestKeep}, {ID: "add"},
	}

	result := Diff(before, after)

	if len(result.Unchanged) != 1 || result.Unchanged[0].ID != "keep" {
		t.Errorf("Unchanged = %v, want [{keep}]", result.Unchanged)
	}

	if len(result.Added) != 1 || result.Added[0].ID != "add" {
		t.Errorf("Added = %v, want [{add}]", result.Added)
	}

	if len(result.Removed) != 1 || result.Removed[0].ID != "remove" {
		t.Errorf("Removed = %v, want [{remove}]", result.Removed)
	}
}

func TestDiff_SortedOutput(t *testing.T) {
	t.Parallel()

	before := []Finding{{ID: "z"}, {ID: "a"}}
	after := []Finding{{ID: "z"}, {ID: "a"}, {ID: "c"}, {ID: "b"}}

	result := Diff(before, after)

	for _, list := range [][]Finding{result.Added, result.Removed, result.Unchanged} {
		for i := 1; i < len(list); i++ {
			if list[i].ID < list[i-1].ID {
				t.Errorf("result not sorted: %v", list)
			}
		}
	}
}
