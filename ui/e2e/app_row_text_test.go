package e2e

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
)

func TestRowTextAndPositionsMapsEveryRuneInCluster(t *testing.T) {
	buffer := paint.NewBuffer(8, 1)
	buffer.SetString(0, 0, "ab👨cd", style.Style{})

	rowText, positions := rowTextAndPositions(buffer, 0)
	runeCount := len([]rune(rowText))
	if runeCount != len(positions) {
		t.Fatalf("positions length must match row rune length: row=%q runes=%d positions=%d", rowText, runeCount, len(positions))
	}

	index := findSubstring(rowText, "cd")
	if index < 0 {
		t.Fatalf("expected substring cd in row %q", rowText)
	}
	if index >= len(positions) {
		t.Fatalf("substring index %d out of positions length %d", index, len(positions))
	}
	if positions[index] != 4 {
		t.Fatalf("expected c to map to cell x=4, got %d in row %q", positions[index], rowText)
	}
}
