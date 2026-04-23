package scale

import "testing"

func TestLinearMapAscending(t *testing.T) {
	s := NewLinear(0, 100, 0, 4)
	if got := s.Map(0); got != 0 {
		t.Fatalf("Map(0) = %d, want 0", got)
	}
	if got := s.Map(50); got != 2 {
		t.Fatalf("Map(50) = %d, want 2", got)
	}
	if got := s.Map(100); got != 4 {
		t.Fatalf("Map(100) = %d, want 4", got)
	}
}

func TestLinearMapDescendingRange(t *testing.T) {
	s := NewLinear(0, 100, 4, 0)
	if got := s.Map(0); got != 4 {
		t.Fatalf("Map(0) = %d, want 4", got)
	}
	if got := s.Map(100); got != 0 {
		t.Fatalf("Map(100) = %d, want 0", got)
	}
}

func TestLinearMapCollapsedDomain(t *testing.T) {
	s := NewLinear(10, 10, 0, 4)
	if got := s.Map(10); got != 2 {
		t.Fatalf("Map(10) = %d, want 2", got)
	}
}

func TestBandPosition(t *testing.T) {
	s := NewBand(3, 0, 4)
	if got := s.Position(0); got != 0 {
		t.Fatalf("Position(0) = %d, want 0", got)
	}
	if got := s.Position(1); got != 2 {
		t.Fatalf("Position(1) = %d, want 2", got)
	}
	if got := s.Position(2); got != 4 {
		t.Fatalf("Position(2) = %d, want 4", got)
	}
}
