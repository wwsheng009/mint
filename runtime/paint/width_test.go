package paint

import "testing"

func TestRuneWidth_AmbiguousArrowIsSingleWidth(t *testing.T) {
	if got := RuneWidth('↑'); got != 1 {
		t.Fatalf("RuneWidth('↑') = %d, want 1", got)
	}
	if got := RuneWidth('↓'); got != 1 {
		t.Fatalf("RuneWidth('↓') = %d, want 1", got)
	}
}

func TestStringWidth_AmbiguousArrowsIsStable(t *testing.T) {
	if got := StringWidth("↑↓: Navigate"); got != 12 {
		t.Fatalf("StringWidth(\"↑↓: Navigate\") = %d, want 12", got)
	}
}

func TestStringWidth_BoxDrawingLineIsSingleCellPerRune(t *testing.T) {
	if got := StringWidth("│──│"); got != 4 {
		t.Fatalf("StringWidth(\"│──│\") = %d, want 4", got)
	}
}

func TestStringWidth_EmojiZWJCluster(t *testing.T) {
	if got := StringWidth("👨‍👩‍👧‍👦"); got != 2 {
		t.Fatalf("StringWidth(family emoji) = %d, want 2", got)
	}
}
