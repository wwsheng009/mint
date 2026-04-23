package paint

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/style"
)

func TestPartialFrameBuffer_ApplyFromRegions(t *testing.T) {
	pf := NewPartialFrameBuffer(10, 2)
	src := NewBuffer(10, 2)
	st := style.NewStyle()

	src.SetStringAligned(0, 0, "0123456789", st, 10)
	src.SetStringAligned(0, 1, "abcdefghij", st, 10)

	// Apply only X:[2,5) at row 1
	pf.ApplyFrom(src, []Rect{{X: 2, Y: 1, Width: 3, Height: 1}})
	dst := pf.Buffer()

	if got := dst.GetContent(2, 1).Cluster; got != "c" {
		t.Fatalf("expected 'c' at (2,1), got %q", got)
	}
	if got := dst.GetContent(4, 1).Cluster; got != "e" {
		t.Fatalf("expected 'e' at (4,1), got %q", got)
	}
	if got := dst.GetContent(1, 1).Cluster; got != " " {
		t.Fatalf("expected unchanged cell at (1,1) to be space, got %q", got)
	}
}
