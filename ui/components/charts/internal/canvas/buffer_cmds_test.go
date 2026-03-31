package canvas

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
)

func TestBufferToDrawCmdsPreservesStyles(t *testing.T) {
	buf := paint.NewBuffer(4, 1)
	buf.SetString(0, 0, "A", style.NewStyle().Foreground(style.Red))
	buf.SetString(1, 0, "B", style.NewStyle().Foreground(style.Blue))

	cmds := BufferToDrawCmds(buf, 0, 0)
	if len(cmds) < 3 {
		t.Fatalf("BufferToDrawCmds() len = %d, want >= 3", len(cmds))
	}
	if cmds[0].Text != "A" || cmds[0].Style.FG != style.Red {
		t.Fatalf("first cmd = %#v, want red A", cmds[0])
	}
	if cmds[1].Text != "B" || cmds[1].Style.FG != style.Blue {
		t.Fatalf("second cmd = %#v, want blue B", cmds[1])
	}
}
