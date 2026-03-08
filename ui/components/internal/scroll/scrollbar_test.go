package scroll

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/style"
)

func TestDrawVerticalScrollbar_HideWhenNotScrollable(t *testing.T) {
	v := NewVerticalViewport(3, 5, 0)
	cmds := DrawVerticalScrollbar(0, 0, 5, v, style.Style{}, DefaultVerticalScrollbarConfig())
	if len(cmds) != 0 {
		t.Fatalf("draw cmds = %d, want 0", len(cmds))
	}
}

func TestDrawVerticalScrollbar_Scrollable(t *testing.T) {
	v := NewVerticalViewport(20, 5, 10)
	cfg := DefaultVerticalScrollbarConfig()
	cmds := DrawVerticalScrollbar(1, 2, 5, v, style.Style{}, cfg)
	if len(cmds) != 5 {
		t.Fatalf("draw cmds = %d, want 5", len(cmds))
	}

	thumbCount := 0
	for _, cmd := range cmds {
		if cmd.Text == cfg.Thumb {
			thumbCount++
		}
	}
	if thumbCount < 1 || thumbCount > 5 {
		t.Fatalf("thumb count = %d, want [1,5]", thumbCount)
	}
}
