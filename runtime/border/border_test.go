package border

import (
	"fmt"
	"strings"
	"testing"

	"github.com/wwsheng009/mint/runtime/style"
)

func TestRenderer_Render(t *testing.T) {
	tests := []struct {
		name          string
		style         Style
		label         string
		content       string
		wantContains  []string
	}{
		{
			name:    "single style no label",
			style:   StyleSingle,
			label:   "",
			content: "Hello",
			wantContains: []string{
				"┌─────┐",
				"│Hello│",
				"└─────┘",
			},
		},
		{
			name:    "double style",
			style:   StyleDouble,
			label:   "",
			content: "Hi",
			wantContains: []string{
				"╔══╗",
				"║Hi║",
				"╚══╝",
			},
		},
		{
			name:    "rounded style",
			style:   StyleRounded,
			label:   "",
			content: "Test",
			wantContains: []string{
				"╭────╮",
				"│Test│",
				"╰────╯",
			},
		},
		{
			name:    "dashed style",
			style:   StyleDashed,
			label:   "",
			content: "OK",
			wantContains: []string{
				"+--+",
				"|OK|",
				"+--+",
			},
		},
		{
			name:    "single style with label",
			style:   StyleSingle,
			label:   "Title",
			content: "Hello",
			wantContains: []string{
				"┌─ Title ┐",
				"│Hello│",
			},
		},
		{
			name:    "multi-line content",
			style:   StyleSingle,
			label:   "",
			content: "Line 1\nLine 2",
			wantContains: []string{
				"┌──────┐",
				"│Line 1│",
				"│Line 2│",
				"└──────┘",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New().WithStyle(tt.style).WithLabel(tt.label)
			result := r.String(tt.content)

			for _, want := range tt.wantContains {
				if !strings.Contains(result, want) {
					t.Errorf("Result missing expected string:\nGot:\n%s\n\nWant to contain: %q", result, want)
				}
			}
		})
	}
}

func TestRenderer_Render_Cells(t *testing.T) {
	r := New().WithStyle(StyleSingle)

	cells := r.Render(0, 0, 5, 1)

	// Expected: 3 (top) + 2 (middle) + 3 (bottom) = 8 cells
	// But if there's a label, it would be different
	minExpected := 8
	if len(cells) < minExpected {
		t.Errorf("Expected at least %d cells, got %d", minExpected, len(cells))
	}

	// Check that corners exist
	hasCorners := func() (tl, tr, bl, br bool) {
		for _, c := range cells {
			if c.Ch == '┌' && c.X == 0 && c.Y == 0 {
				tl = true
			}
			if c.Ch == '┐' && c.X == 6 && c.Y == 0 {
				tr = true
			}
			if c.Ch == '└' && c.X == 0 && c.Y == 2 {
				bl = true
			}
			if c.Ch == '┘' && c.X == 6 && c.Y == 2 {
				br = true
			}
		}
		return
	}

	tl, tr, bl, br := hasCorners()
	if !tl {
		t.Error("Missing top-left corner")
	}
	if !tr {
		t.Error("Missing top-right corner")
	}
	if !bl {
		t.Error("Missing bottom-left corner")
	}
	if !br {
		t.Error("Missing bottom-right corner")
	}
}

func TestRenderer_GetTotalSize(t *testing.T) {
	tests := []struct {
		name           string
		style          Style
		contentWidth   int
		contentHeight  int
		wantWidth      int
		wantHeight     int
	}{
		{
			name:          "single style adds 2",
			style:         StyleSingle,
			contentWidth:  5,
			contentHeight: 1,
			wantWidth:     7,
			wantHeight:    3,
		},
		{
			name:          "no style adds nothing",
			style:         StyleNone,
			contentWidth:  5,
			contentHeight: 1,
			wantWidth:     5,
			wantHeight:    1,
		},
		{
			name:          "zero content",
			style:         StyleSingle,
			contentWidth:  0,
			contentHeight: 0,
			wantWidth:     2,
			wantHeight:    2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New().WithStyle(tt.style)
			w, h := r.GetTotalSize(tt.contentWidth, tt.contentHeight)
			if w != tt.wantWidth || h != tt.wantHeight {
				t.Errorf("GetTotalSize(%d, %d) = (%d, %d), want (%d, %d)",
					tt.contentWidth, tt.contentHeight, w, h, tt.wantWidth, tt.wantHeight)
			}
		})
	}
}

func TestRenderer_GetContentOffset(t *testing.T) {
	tests := []struct {
		name    string
		style   Style
		wantX   int
		wantY   int
	}{
		{
			name:  "single style offset is 1,1",
			style: StyleSingle,
			wantX: 1,
			wantY: 1,
		},
		{
			name:  "no style offset is 0,0",
			style: StyleNone,
			wantX: 0,
			wantY: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New().WithStyle(tt.style)
			x, y := r.GetContentOffset()
			if x != tt.wantX || y != tt.wantY {
				t.Errorf("GetContentOffset() = (%d, %d), want (%d, %d)", x, y, tt.wantX, tt.wantY)
			}
		})
	}
}

func TestRenderer_Paint(t *testing.T) {
	r := New().WithStyle(StyleSingle)

	var painted []string
	paintFunc := func(x, y int, ch rune, s style.Style) {
		_ = s // Ignore style for test
		painted = append(painted, fmt.Sprintf("%d,%d:%c", x, y, ch))
	}

	r.Paint(0, 0, 3, 1, paintFunc)

	// Should paint border cells
	if len(painted) < 8 {
		t.Errorf("Expected at least 8 paint calls, got %d", len(painted))
	}
}

func TestExample(t *testing.T) {
	// This example demonstrates the border rendering
	r := New().WithStyle(StyleSingle)

	output := r.String("Hello")
	expected := "┌─────┐\n│Hello│\n└─────┘"

	if output != expected {
		t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, output)
	}
}

// Benchmark_Render-8    	  100000	     12154 ns/op
func Benchmark_Render(b *testing.B) {
	r := New().WithStyle(StyleSingle)
	for i := 0; i < b.N; i++ {
		r.Render(0, 0, 20, 10)
	}
}
