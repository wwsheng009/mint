package raster

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	chartcanvas "github.com/wwsheng009/mint/ui/components/charts/internal/canvas"
)

func TestDrawPolylineHorizontal(t *testing.T) {
	surface := chartcanvas.NewRuneCanvas(5, 1, ' ')
	DrawPolyline(surface, []Point{{X: 0, Y: 0}, {X: 4, Y: 0}}, PolylineOptions{
		ShowPoints: false,
		Glyphs:     DefaultLineGlyphs(),
	})

	rows := surface.Rows()
	if len(rows) != 1 {
		t.Fatalf("Rows() len = %d, want 1", len(rows))
	}
	if rows[0] != "─────" {
		t.Fatalf("Rows()[0] = %q, want %q", rows[0], "─────")
	}
}

func TestDrawPolylineWithPoints(t *testing.T) {
	surface := chartcanvas.NewRuneCanvas(3, 1, ' ')
	DrawPolyline(surface, []Point{{X: 0, Y: 0}, {X: 2, Y: 0}}, PolylineOptions{
		ShowPoints: true,
		PointRune:  '●',
		Glyphs:     DefaultLineGlyphs(),
	})

	rows := surface.Rows()
	if rows[0] != "●─●" {
		t.Fatalf("Rows()[0] = %q, want %q", rows[0], "●─●")
	}
}

func TestDrawPolylineDiagonal(t *testing.T) {
	surface := chartcanvas.NewRuneCanvas(3, 3, ' ')
	DrawPolyline(surface, []Point{{X: 0, Y: 2}, {X: 2, Y: 0}}, PolylineOptions{
		ShowPoints: false,
		Glyphs:     DefaultLineGlyphs(),
	})

	rows := surface.Rows()
	if rows[0] != "  ╱" {
		t.Fatalf("Rows()[0] = %q, want %q", rows[0], "  ╱")
	}
	if rows[1] != " ╱ " {
		t.Fatalf("Rows()[1] = %q, want %q", rows[1], " ╱ ")
	}
	if rows[2] != "╱  " {
		t.Fatalf("Rows()[2] = %q, want %q", rows[2], "╱  ")
	}
}

func TestDrawPolylineToBuffer(t *testing.T) {
	buf := paint.NewBuffer(3, 1)
	DrawPolylineToBuffer(buf, []Point{{X: 0, Y: 0}, {X: 2, Y: 0}}, style.NewStyle().Foreground(style.Red), PolylineOptions{
		ShowPoints: true,
		PointRune:  '●',
		Glyphs:     DefaultLineGlyphs(),
	})

	if got := buf.Cells[0][0].Cluster; got != "●" {
		t.Fatalf("buf.Cells[0][0].Cluster = %q, want ●", got)
	}
	if got := buf.Cells[0][1].Cluster; got != "─" {
		t.Fatalf("buf.Cells[0][1].Cluster = %q, want ─", got)
	}
	if got := buf.Cells[0][0].Style.FG; got != style.Red {
		t.Fatalf("buf.Cells[0][0].Style.FG = %q, want red", got)
	}
}
