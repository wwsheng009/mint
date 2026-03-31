package layout

import (
	"testing"

	rtlayout "github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/style"
)

func TestFrameMeasure(t *testing.T) {
	frame := NewFrame().
		Add(SectionTitle, "Trend", style.Style{}).
		Add(SectionPlot, "●─●", style.Style{})

	size := frame.Measure(rtlayout.UnboundedConstraints())
	if size.Width != 5 {
		t.Fatalf("Measure().Width = %d, want 5", size.Width)
	}
	if size.Height != 2 {
		t.Fatalf("Measure().Height = %d, want 2", size.Height)
	}
}

func TestFramePaintOrder(t *testing.T) {
	frame := NewFrame().
		Add(SectionTitle, "Bars", style.Style{}).
		Add(SectionAxis, "─────", style.Style{})

	cmds := frame.Paint(3, 4)
	if len(cmds) != 2 {
		t.Fatalf("Paint() len = %d, want 2", len(cmds))
	}
	if cmds[0].X != 3 || cmds[0].Y != 4 || cmds[0].Text != "Bars" {
		t.Fatalf("first draw cmd = %#v, want x=3 y=4 text=Bars", cmds[0])
	}
	if cmds[1].Y != 5 || cmds[1].Text != "─────" {
		t.Fatalf("second draw cmd = %#v, want y=5 text=─────", cmds[1])
	}
}

func TestFrameAddIfNotEmpty(t *testing.T) {
	frame := NewFrame().
		AddIfNotEmpty(SectionTitle, "   ", style.Style{}).
		AddIfNotEmpty(SectionPlot, "●", style.Style{})

	rows := frame.Rows()
	if len(rows) != 1 {
		t.Fatalf("Rows() len = %d, want 1", len(rows))
	}
	if rows[0].Section != SectionPlot {
		t.Fatalf("Rows()[0].Section = %q, want %q", rows[0].Section, SectionPlot)
	}
}
