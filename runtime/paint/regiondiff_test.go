package paint

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/style"
)

func TestRegionDiff_RowChangeProducesRect(t *testing.T) {
	front := NewBuffer(8, 3)
	back := NewBuffer(8, 3)
	st := style.NewStyle()

	back.SetString(2, 1, "ABC", st)
	diff := RegionDiff(front, back)
	if !diff.HasChanges {
		t.Fatal("expected changes")
	}
	if len(diff.DirtyRegions) == 0 {
		t.Fatal("expected dirty regions")
	}
}

func TestRegionDiff_SameBufferNoChanges(t *testing.T) {
	front := NewBuffer(6, 2)
	back := NewBuffer(6, 2)
	st := style.NewStyle()

	front.SetString(0, 0, "hello", st)
	back.SetString(0, 0, "hello", st)

	diff := RegionDiff(front, back)
	if diff.HasChanges {
		t.Fatal("expected no changes")
	}
	if len(diff.DirtyRegions) != 0 {
		t.Fatalf("expected 0 dirty regions, got %d", len(diff.DirtyRegions))
	}
}
