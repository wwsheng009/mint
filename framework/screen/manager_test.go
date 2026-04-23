package screen

import (
	"bytes"
	"strings"
	"testing"

	"github.com/wwsheng009/mint/runtime/style"
)

func TestManagerRenderDetectsMutationsWhenBufferReused(t *testing.T) {
	mgr := NewManager(1, 1)
	var out bytes.Buffer
	mgr.out = &out
	mgr.stdin = nil

	buf := NewBuffer(1, 1)
	buf.SetCell(0, 0, 'A', style.NewStyle())

	if err := mgr.Render(buf); err != nil {
		t.Fatalf("first render failed: %v", err)
	}

	out.Reset()
	buf.SetCell(0, 0, 'B', style.NewStyle())

	if err := mgr.Render(buf); err != nil {
		t.Fatalf("second render failed: %v", err)
	}

	if got := out.String(); !strings.Contains(got, "B") {
		t.Fatalf("expected reused buffer mutation to be rendered, got %q", got)
	}
}

func TestManagerRenderBatchesAdjacentCells(t *testing.T) {
	mgr := NewManager(3, 1)
	var out bytes.Buffer
	mgr.out = &out
	mgr.stdin = nil

	red := style.NewStyle().Foreground(style.Red)
	buf := NewBuffer(3, 1)
	buf.SetLine(0, "ABC", red)

	if err := mgr.Render(buf); err != nil {
		t.Fatalf("render failed: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "ABC") {
		t.Fatalf("expected merged text output, got %q", got)
	}
	if count := strings.Count(got, red.ToANSI()); count != 1 {
		t.Fatalf("expected one style transition for merged cells, got %d in %q", count, got)
	}
	if count := strings.Count(got, "\x1b[1;"); count != 1 {
		t.Fatalf("expected one absolute cursor move for merged cells, got %d in %q", count, got)
	}
}

func TestManagerSetSizeForcesFullRedrawWithNewDimensions(t *testing.T) {
	mgr := NewManager(1, 1)
	var out bytes.Buffer
	mgr.out = &out
	mgr.stdin = nil

	initial := NewBuffer(1, 1)
	initial.SetCell(0, 0, 'A', style.NewStyle())
	if err := mgr.Render(initial); err != nil {
		t.Fatalf("initial render failed: %v", err)
	}

	out.Reset()
	mgr.SetSize(2, 1)

	resized := NewBuffer(2, 1)
	resized.SetLine(0, "AB", style.NewStyle())
	if err := mgr.Render(resized); err != nil {
		t.Fatalf("resized render failed: %v", err)
	}

	if got := out.String(); !strings.Contains(got, "AB") {
		t.Fatalf("expected full redraw after resize, got %q", got)
	}
}

func TestManagerInitAndCloseWriteTerminalControlSequences(t *testing.T) {
	mgr := NewManager(2, 1)
	var out bytes.Buffer
	mgr.out = &out
	mgr.stdin = nil

	if err := mgr.Init(); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if mgr.rawState != nil {
		t.Fatalf("expected raw mode to be skipped without terminal stdin")
	}

	if err := mgr.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}

	got := out.String()
	expected := []string{
		"\x1b[?1049h",
		"\x1b[?25l",
		"\x1b[?25h",
		"\x1b[?1049l",
	}
	for _, seq := range expected {
		if !strings.Contains(got, seq) {
			t.Fatalf("expected output to contain %q, got %q", seq, got)
		}
	}

	enterAlt := strings.Index(got, "\x1b[?1049h")
	hideCursor := strings.Index(got, "\x1b[?25l")
	showCursor := strings.Index(got, "\x1b[?25h")
	exitAlt := strings.Index(got, "\x1b[?1049l")
	if !(enterAlt >= 0 && enterAlt < hideCursor && hideCursor < showCursor && showCursor < exitAlt) {
		t.Fatalf("unexpected terminal control sequence order: %q", got)
	}
}
