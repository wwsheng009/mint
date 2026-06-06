package render

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
)

type clipPaintNode struct {
	id   string
	cmds []paint.DrawCmd
	text string
}

func (n clipPaintNode) ID() string                     { return n.id }
func (n clipPaintNode) NodeType() paint.NodeType       { return paint.NodeTypeElement }
func (n clipPaintNode) Tag() string                    { return "clip-test" }
func (n clipPaintNode) Style() style.Style             { return style.Style{} }
func (n clipPaintNode) SetStyle(s style.Style)         {}
func (n clipPaintNode) TextContent() string            { return n.text }
func (n clipPaintNode) Paint(x, y int) []paint.DrawCmd { return n.cmds }

func TestPaintEngineClipsCustomDrawCommands(t *testing.T) {
	buffer := paint.NewBuffer(10, 3)
	box := &paint.PaintableBox{
		Node:   clipPaintNode{id: "cmd", cmds: []paint.DrawCmd{paint.NewTextCmd(0, 0, "0123456789", style.Style{})}},
		X:      0,
		Y:      0,
		Width:  10,
		Height: 1,
		Clip:   &paint.Rect{X: 2, Y: 0, Width: 4, Height: 1},
	}

	if err := NewPaintEngine().paintBoxOnly(box, buffer); err != nil {
		t.Fatal(err)
	}
	if got := rowText(buffer, 0); got != "  2345    " {
		t.Fatalf("row = %q, want clipped text", got)
	}
}

func TestPaintEngineClipsElementTextAndBackground(t *testing.T) {
	buffer := paint.NewBuffer(8, 2)
	box := &paint.PaintableBox{
		Node:   clipPaintNode{id: "text", text: "ABCDEFGH"},
		X:      0,
		Y:      0,
		Width:  8,
		Height: 2,
		Clip:   &paint.Rect{X: 1, Y: 0, Width: 3, Height: 1},
	}

	if err := NewPaintEngine().paintBoxOnly(box, buffer); err != nil {
		t.Fatal(err)
	}
	if got := rowText(buffer, 0); got != " BCD    " {
		t.Fatalf("row = %q, want clipped text", got)
	}
}

func rowText(buffer *paint.Buffer, y int) string {
	out := make([]rune, 0, buffer.Width)
	for x := 0; x < buffer.Width; x++ {
		cell := buffer.GetContent(x, y)
		if cell.IsContinuation {
			continue
		}
		if cell.Cluster == "" {
			out = append(out, ' ')
			continue
		}
		for _, r := range cell.Cluster {
			out = append(out, r)
			break
		}
	}
	return string(out)
}
