package feedback

import (
	"fmt"
	"unicode/utf8"

	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui"
)

// ProgressVNode represents a progress bar component
type ProgressVNode struct {
	*ui.ElementVNode
	value      int
	max        int
	showPercent bool
	label      string
	width      int
}

// NewProgress creates a new progress bar
func NewProgress() *ProgressVNode {
	return &ProgressVNode{
		ElementVNode: ui.NewElement("progress"),
		value:        0,
		max:          100,
		showPercent:  true,
		label:        "",
		width:        30,
	}
}

// Progress creates a new progress node
func Progress() ui.VNode {
	return NewProgress()
}

// ProgressBuilder creates a progress builder for chained calls
func ProgressBuilder() *ProgressBuilderType {
	return &ProgressBuilderType{
		node: NewProgress(),
	}
}

// =============================================================================
// ProgressVNode methods
// =============================================================================

// Value returns the current progress value
func (p *ProgressVNode) Value() int {
	return p.value
}

// SetValue sets the current progress value
func (p *ProgressVNode) SetValue(v int) *ProgressVNode {
	p.value = v
	if p.value < 0 {
		p.value = 0
	}
	if p.value > p.max {
		p.value = p.max
	}
	p.SetProp("value", v)
	return p
}

// Max returns the maximum value
func (p *ProgressVNode) Max() int {
	return p.max
}

// SetMax sets the maximum value
func (p *ProgressVNode) SetMax(v int) *ProgressVNode {
	p.max = v
	p.SetProp("max", v)
	return p
}

// ShowPercent returns whether percentage is shown
func (p *ProgressVNode) ShowPercent() bool {
	return p.showPercent
}

// SetShowPercent sets whether to show percentage
func (p *ProgressVNode) SetShowPercent(v bool) *ProgressVNode {
	p.showPercent = v
	p.SetProp("showPercent", v)
	return p
}

// Label returns the label text
func (p *ProgressVNode) Label() string {
	return p.label
}

// SetLabel sets the label text
func (p *ProgressVNode) SetLabel(text string) *ProgressVNode {
	p.label = text
	p.SetProp("label", text)
	return p
}

// Width returns the progress bar width
func (p *ProgressVNode) Width() int {
	return p.width
}

// SetWidth sets the progress bar width
func (p *ProgressVNode) SetWidth(w int) *ProgressVNode {
	p.width = w
	if p.width < 10 {
		p.width = 10
	}
	p.SetProp("width", w)
	return p
}

// Percent returns the percentage (0-100)
func (p *ProgressVNode) Percent() int {
	if p.max == 0 {
		return 0
	}
	return (p.value * 100) / p.max
}

// =============================================================================
// ProgressBuilderType provides fluent API for building progress bars
// =============================================================================

// ProgressBuilderType is the builder for Progress
type ProgressBuilderType struct {
	node *ProgressVNode
}

// Value sets the current value
func (b *ProgressBuilderType) Value(v int) *ProgressBuilderType {
	b.node.SetValue(v)
	return b
}

// Max sets the maximum value
func (b *ProgressBuilderType) Max(v int) *ProgressBuilderType {
	b.node.SetMax(v)
	return b
}

// ShowPercent sets whether to show percentage
func (b *ProgressBuilderType) ShowPercent(v bool) *ProgressBuilderType {
	b.node.SetShowPercent(v)
	return b
}

// Label sets the label text
func (b *ProgressBuilderType) Label(text string) *ProgressBuilderType {
	b.node.SetLabel(text)
	return b
}

// Width sets the progress bar width
func (b *ProgressBuilderType) Width(w int) *ProgressBuilderType {
	b.node.SetWidth(w)
	return b
}

// Key sets the key for diffing
func (b *ProgressBuilderType) Key(key string) *ProgressBuilderType {
	b.node.SetKey(key)
	return b
}

// Style sets the visual style
func (b *ProgressBuilderType) Style(s style.Style) *ProgressBuilderType {
	b.node.SetStyle(s)
	return b
}

// FgColor sets the foreground color
func (b *ProgressBuilderType) FgColor(c interface{}) *ProgressBuilderType {
	if colorStr, ok := c.(string); ok {
		s := b.node.Style()
		s = s.Foreground(style.Color(colorStr))
		b.node.SetStyle(s)
	} else if color, ok := c.(style.Color); ok {
		s := b.node.Style()
		s = s.Foreground(color)
		b.node.SetStyle(s)
	}
	return b
}

// BgColor sets the background color
func (b *ProgressBuilderType) BgColor(c interface{}) *ProgressBuilderType {
	if colorStr, ok := c.(string); ok {
		s := b.node.Style()
		s = s.Background(style.Color(colorStr))
		b.node.SetStyle(s)
	} else if color, ok := c.(style.Color); ok {
		s := b.node.Style()
		s = s.Background(color)
		b.node.SetStyle(s)
	}
	return b
}

// Build returns the ui.VNode
func (b *ProgressBuilderType) Build() ui.VNode {
	return b.node
}

// =============================================================================
// Spinner component
// =============================================================================

// SpinnerVNode represents a loading spinner component
type SpinnerVNode struct {
	*ui.ElementVNode
	message string
	frames  []string
}

// NewSpinner creates a new spinner
func NewSpinner() *SpinnerVNode {
	return &SpinnerVNode{
		ElementVNode: ui.NewElement("spinner"),
		message:      "Loading...",
		frames:       []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
	}
}

// Spinner creates a new spinner node
func Spinner() ui.VNode {
	return NewSpinner()
}

// SpinnerBuilder creates a spinner builder
func SpinnerBuilder() *SpinnerBuilderType {
	return &SpinnerBuilderType{
		node: NewSpinner(),
	}
}

// Message returns the spinner message
func (s *SpinnerVNode) Message() string {
	return s.message
}

// SetMessage sets the spinner message
func (s *SpinnerVNode) SetMessage(msg string) *SpinnerVNode {
	s.message = msg
	s.SetProp("message", msg)
	return s
}

// Frames returns the spinner frames
func (s *SpinnerVNode) Frames() []string {
	return s.frames
}

// SetFrames sets the spinner frames
func (s *SpinnerVNode) SetFrames(frames []string) *SpinnerVNode {
	s.frames = frames
	return s
}

// CurrentFrame returns the current frame based on time
func (s *SpinnerVNode) CurrentFrame() string {
	// Simple implementation - just return first frame
	// In a real implementation, this would cycle based on time
	if len(s.frames) > 0 {
		return s.frames[0]
	}
	return "⠋"
}

// SpinnerBuilderType is the builder for Spinner
type SpinnerBuilderType struct {
	node *SpinnerVNode
}

// Message sets the spinner message
func (b *SpinnerBuilderType) Message(msg string) *SpinnerBuilderType {
	b.node.SetMessage(msg)
	return b
}

// Frames sets the spinner frames
func (b *SpinnerBuilderType) Frames(frames []string) *SpinnerBuilderType {
	b.node.SetFrames(frames)
	return b
}

// Key sets the key for diffing
func (b *SpinnerBuilderType) Key(key string) *SpinnerBuilderType {
	b.node.SetKey(key)
	return b
}

// Style sets the visual style
func (b *SpinnerBuilderType) Style(s style.Style) *SpinnerBuilderType {
	b.node.SetStyle(s)
	return b
}

// FgColor sets the foreground color
func (b *SpinnerBuilderType) FgColor(c interface{}) *SpinnerBuilderType {
	if colorStr, ok := c.(string); ok {
		s := b.node.Style()
		s = s.Foreground(style.Color(colorStr))
		b.node.SetStyle(s)
	} else if color, ok := c.(style.Color); ok {
		s := b.node.Style()
		s = s.Foreground(color)
		b.node.SetStyle(s)
	}
	return b
}

// Build returns the ui.VNode
func (b *SpinnerBuilderType) Build() ui.VNode {
	return b.node
}

// =============================================================================
// Measurable & Paintable Interface Implementation
// =============================================================================

// Measure implements runtime.Measurable interface for ProgressVNode
func (p *ProgressVNode) Measure(constraints runtime.BoxConstraints) runtime.Size {
	if p == nil {
		return runtime.Size{Width: 0, Height: 0}
	}

	width := p.width
	if width < 10 {
		width = 10
	}

	height := 1

	// Add height for label if present
	if p.label != "" {
		height = 2
	}

	// Apply constraints
	if width < constraints.MinWidth {
		width = constraints.MinWidth
	}
	if width > constraints.MaxWidth && constraints.MaxWidth > 0 {
		width = constraints.MaxWidth
	}
	if height < constraints.MinHeight {
		height = constraints.MinHeight
	}
	if height > constraints.MaxHeight && constraints.MaxHeight > 0 {
		height = constraints.MaxHeight
	}

	return runtime.Size{Width: width, Height: height}
}

// Paint implements paint.Paintable interface for ProgressVNode
func (p *ProgressVNode) Paint(x, y int) []paint.DrawCmd {
	if p == nil {
		return nil
	}

	progressStyle := p.Style()
	var cmds []paint.DrawCmd

	measured := p.Measure(runtime.BoxConstraints{})
	width := measured.Width

	// Calculate filled width
	percent := p.Percent()

	// Build progress bar: [======>     ]
	barWidth := width - 2
	filledCount := (percent * barWidth) / 100

	bar := "["
	for i := 0; i < barWidth; i++ {
		if i < filledCount {
			bar += "="
		} else if i == filledCount {
			bar += ">"
		} else {
			bar += " "
		}
	}
	bar += "]"

	// Draw progress bar
	cmds = append(cmds, paint.NewTextCmd(x, y, bar, progressStyle))

	// Draw percentage or label below
	if p.showPercent || p.label != "" {
		labelY := y + 1
		var label string
		if p.label != "" && p.showPercent {
			label = fmt.Sprintf("%s: %d%%", p.label, percent)
		} else if p.label != "" {
			label = p.label
		} else {
			label = fmt.Sprintf("%d%%", percent)
		}
		cmds = append(cmds, paint.NewTextCmd(x, labelY, label, progressStyle))
	}

	return cmds
}

// Measure implements runtime.Measurable interface for SpinnerVNode
func (s *SpinnerVNode) Measure(constraints runtime.BoxConstraints) runtime.Size {
	if s == nil {
		return runtime.Size{Width: 0, Height: 0}
	}

	width := utf8.RuneCountInString(s.message) + 2 // +2 for spinner icon
	height := 1

	// Apply constraints
	if width < constraints.MinWidth {
		width = constraints.MinWidth
	}
	if width > constraints.MaxWidth && constraints.MaxWidth > 0 {
		width = constraints.MaxWidth
	}
	if height < constraints.MinHeight {
		height = constraints.MinHeight
	}
	if height > constraints.MaxHeight && constraints.MaxHeight > 0 {
		height = constraints.MaxHeight
	}

	return runtime.Size{Width: width, Height: height}
}

// Paint implements paint.Paintable interface for SpinnerVNode
func (s *SpinnerVNode) Paint(x, y int) []paint.DrawCmd {
	if s == nil {
		return nil
	}

	spinnerStyle := s.Style()

	// Get current frame
	frame := s.CurrentFrame()

	// Build spinner display: ⠋ Loading...
	displayText := frame + " " + s.message

	return []paint.DrawCmd{
		paint.NewTextCmd(x, y, displayText, spinnerStyle),
	}
}
