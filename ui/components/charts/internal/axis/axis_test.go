package axis

import "testing"

func TestHorizontalLine(t *testing.T) {
	if got := HorizontalLine(5); got != "─────" {
		t.Fatalf("HorizontalLine(5) = %q, want %q", got, "─────")
	}
}

func TestLabelRune(t *testing.T) {
	if got := LabelRune("CPU", '•'); got != 'C' {
		t.Fatalf("LabelRune() = %q, want C", got)
	}
	if got := LabelRune("   ", '•'); got != '•' {
		t.Fatalf("LabelRune() = %q, want •", got)
	}
}

func TestLabelRow(t *testing.T) {
	row := LabelRow(5, []string{"A", "B", "C"}, []int{0, 2, 4}, '•')
	if row != "A B C" {
		t.Fatalf("LabelRow() = %q, want %q", row, "A B C")
	}
}

func TestGridRows(t *testing.T) {
	rows := GridRows(6, 3)
	if len(rows) != 3 {
		t.Fatalf("GridRows() len = %d, want 3", len(rows))
	}
	if rows[0] != 1 || rows[1] != 2 || rows[2] != 3 {
		t.Fatalf("GridRows() = %#v, want [1 2 3]", rows)
	}
}
