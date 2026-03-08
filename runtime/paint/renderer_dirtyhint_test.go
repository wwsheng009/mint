package paint

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/style"
)

func TestRenderer_MarkDirtyRect_ForcesHintRedraw(t *testing.T) {
	renderer := NewRenderer(12, 3)
	st := style.NewStyle()

	back := renderer.GetBackBuffer()
	back.SetString(0, 0, "hello", st)
	if out := renderer.Render(); out == "" {
		t.Fatal("first render should produce output")
	}

	// No content change: should be empty without dirty hint.
	if out := renderer.Render(); out != "" {
		t.Fatalf("render without changes should be empty, got %d bytes", len(out))
	}

	// Explicit dirty rect hint should trigger redraw for the hinted line.
	renderer.MarkDirtyRect(Rect{X: 0, Y: 0, Width: 5, Height: 1})
	out := renderer.Render()
	if out == "" {
		t.Fatal("dirty rect hint should force redraw output")
	}
}

func TestRenderer_ForceFullRender_ForcesFullRender(t *testing.T) {
	renderer := NewRenderer(10, 4)
	st := style.NewStyle()

	back := renderer.GetBackBuffer()
	back.SetString(0, 0, "A", st)
	_ = renderer.Render()

	if out := renderer.Render(); out != "" {
		t.Fatalf("render without changes should be empty, got %d bytes", len(out))
	}

	renderer.ForceFullRender()
	out := renderer.Render()
	if out == "" {
		t.Fatal("ForceFullRender should force a full render output")
	}

	stats := renderer.GetStats()
	if stats.ChangedCells != 4 {
		t.Fatalf("ChangedCells=%d, want 4 full lines", stats.ChangedCells)
	}
}

func TestRenderer_UseLineDiffFalse_RendersFullScreenLines(t *testing.T) {
	renderer := NewRenderer(8, 3)
	st := style.NewStyle()

	back := renderer.GetBackBuffer()
	back.SetString(0, 0, "abc", st)
	_ = renderer.Render()

	// One-cell update
	back = renderer.GetBackBuffer()
	back.SetCell(2, 0, 'X', st)

	renderer.UseLineDiff(false)
	out := renderer.Render()
	if out == "" {
		t.Fatal("render should produce output after content change")
	}

	stats := renderer.GetStats()
	if stats.ChangedCells != 3 {
		t.Fatalf("ChangedCells=%d, want 3 full lines when line diff disabled", stats.ChangedCells)
	}
}
