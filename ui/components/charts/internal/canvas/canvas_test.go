package canvas

import "testing"

func TestRuneCanvasSetGetAndRows(t *testing.T) {
	surface := NewRuneCanvas(3, 2, ' ')
	if !surface.Set(1, 0, 'X') {
		t.Fatal("Set() returned false for in-bounds coordinate")
	}
	if got := surface.Get(1, 0); got != 'X' {
		t.Fatalf("Get() = %q, want X", got)
	}

	rows := surface.Rows()
	if len(rows) != 2 {
		t.Fatalf("Rows() len = %d, want 2", len(rows))
	}
	if rows[0] != " X " {
		t.Fatalf("Rows()[0] = %q, want %q", rows[0], " X ")
	}
	if rows[1] != "   " {
		t.Fatalf("Rows()[1] = %q, want %q", rows[1], "   ")
	}
}

func TestRuneCanvasOutOfBounds(t *testing.T) {
	surface := NewRuneCanvas(2, 2, '.')
	if surface.Set(-1, 0, 'X') {
		t.Fatal("Set() = true for out-of-bounds coordinate")
	}
	if got := surface.Get(5, 5); got != 0 {
		t.Fatalf("Get() = %q, want zero", got)
	}
}
