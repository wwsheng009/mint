package scroll

import "testing"

func TestVerticalViewport_Normalize(t *testing.T) {
	v := NewVerticalViewport(3, 0, 10)

	if v.ViewSize != 1 {
		t.Fatalf("view size = %d, want 1", v.ViewSize)
	}
	if v.Offset != 2 {
		t.Fatalf("offset = %d, want 2", v.Offset)
	}
}

func TestVerticalViewport_VisibleRange(t *testing.T) {
	v := NewVerticalViewport(10, 4, 3)
	start, end := v.VisibleRange()

	if start != 3 || end != 7 {
		t.Fatalf("range = (%d,%d), want (3,7)", start, end)
	}
}

func TestVerticalViewport_EnsureVisible(t *testing.T) {
	v := NewVerticalViewport(20, 5, 0)

	if !v.EnsureVisible(9) {
		t.Fatal("EnsureVisible should change offset")
	}
	if v.Offset != 5 {
		t.Fatalf("offset = %d, want 5", v.Offset)
	}

	if !v.EnsureVisible(2) {
		t.Fatal("EnsureVisible should change offset when target above")
	}
	if v.Offset != 2 {
		t.Fatalf("offset = %d, want 2", v.Offset)
	}
}

func TestVerticalViewport_PageUpDown(t *testing.T) {
	v := NewVerticalViewport(30, 6, 0)
	if !v.PageDown() {
		t.Fatal("PageDown should move offset")
	}
	if v.Offset != 6 {
		t.Fatalf("offset = %d, want 6", v.Offset)
	}

	if !v.PageUp() {
		t.Fatal("PageUp should move offset")
	}
	if v.Offset != 0 {
		t.Fatalf("offset = %d, want 0", v.Offset)
	}
}
