package layout

import (
	"strings"

	rtlayout "github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
)

// Section identifies a logical chart layout area.
type Section string

const (
	SectionTitle  Section = "title"
	SectionLegend Section = "legend"
	SectionPlot   Section = "plot"
	SectionAxis   Section = "axis"
	SectionLabels Section = "labels"
	SectionMeta   Section = "meta"
)

// Row is a rendered row in the chart layout frame.
type Row struct {
	Section Section
	Text    string
	Style   style.Style
}

// Frame is a minimal vertical chart layout container.
type Frame struct {
	rows []Row
}

// NewFrame creates an empty chart layout frame.
func NewFrame() *Frame {
	return &Frame{}
}

// Add appends a row to the frame.
func (f *Frame) Add(section Section, text string, rowStyle style.Style) *Frame {
	f.rows = append(f.rows, Row{
		Section: section,
		Text:    text,
		Style:   rowStyle,
	})
	return f
}

// AddRows appends multiple rows with the same section and style.
func (f *Frame) AddRows(section Section, texts []string, rowStyle style.Style) *Frame {
	for _, text := range texts {
		f.Add(section, text, rowStyle)
	}
	return f
}

// AddIfNotEmpty appends a row only when the text contains non-whitespace content.
func (f *Frame) AddIfNotEmpty(section Section, text string, rowStyle style.Style) *Frame {
	if strings.TrimSpace(text) == "" {
		return f
	}
	return f.Add(section, text, rowStyle)
}

// Rows returns a shallow copy of the frame rows.
func (f *Frame) Rows() []Row {
	rows := make([]Row, len(f.rows))
	copy(rows, f.rows)
	return rows
}

// Measure calculates the size required by the current rows.
func (f *Frame) Measure(constraints rtlayout.Constraints) rtlayout.Size {
	width := 0
	for _, row := range f.rows {
		if rowWidth := paint.StringWidth(row.Text); rowWidth > width {
			width = rowWidth
		}
	}

	height := len(f.rows)
	width = constraints.ConstrainWidth(width)
	height = constraints.ConstrainHeight(height)
	return rtlayout.Size{Width: width, Height: height}
}

// Paint converts the frame rows to draw commands.
func (f *Frame) Paint(x, y int) []paint.DrawCmd {
	cmds := make([]paint.DrawCmd, 0, len(f.rows))
	for i, row := range f.rows {
		cmds = append(cmds, paint.DrawCmd{
			X:     x,
			Y:     y + i,
			Text:  row.Text,
			Style: row.Style,
		})
	}
	return cmds
}
