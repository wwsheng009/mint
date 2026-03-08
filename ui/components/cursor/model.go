package cursor

import (
	"time"

	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
)

// Model is the shared runtime cursor implementation used by standalone and embedded cursors.
type Model struct {
	config       Config
	visible      bool
	phaseVisible bool
	lastBlink    time.Time
}

// NewModel creates a new runtime cursor model.
func NewModel(cfg Config) *Model {
	cfg = NormalizeConfig(cfg)
	return &Model{
		config:       cfg,
		visible:      true,
		phaseVisible: true,
		lastBlink:    time.Now(),
	}
}

// Config returns the current cursor config.
func (m *Model) Config() Config {
	return m.config
}

// SetConfig updates the cursor config.
func (m *Model) SetConfig(cfg Config) bool {
	cfg = NormalizeConfig(cfg)
	if m.config == cfg {
		return false
	}
	m.config = cfg
	m.ResetBlink()
	return true
}

// SetVisible changes whether the cursor can be painted.
func (m *Model) SetVisible(visible bool) bool {
	if m.visible == visible {
		return false
	}
	m.visible = visible
	if visible {
		m.ResetBlink()
	} else {
		m.phaseVisible = false
	}
	return true
}

// Visible reports whether the cursor is currently enabled for painting.
func (m *Model) Visible() bool {
	return m.visible
}

// ResetBlink makes the cursor immediately visible and restarts the blink cycle.
func (m *Model) ResetBlink() {
	m.phaseVisible = true
	m.lastBlink = time.Now()
}

// WantsTick reports whether the cursor needs periodic updates.
func (m *Model) WantsTick() bool {
	return m.visible && m.config.Blink && m.config.BlinkInterval > 0
}

// Tick advances the blink cycle.
func (m *Model) Tick(now time.Time) bool {
	if !m.WantsTick() {
		return false
	}
	if m.lastBlink.IsZero() {
		m.lastBlink = now
		return false
	}
	if now.Sub(m.lastBlink) < m.config.BlinkInterval {
		return false
	}
	m.phaseVisible = !m.phaseVisible
	m.lastBlink = now
	return true
}

// ShouldPaint reports whether the cursor should be visible in the current frame.
func (m *Model) ShouldPaint() bool {
	if !m.visible {
		return false
	}
	if !m.config.Blink {
		return true
	}
	return m.phaseVisible
}

// DrawCmd builds the cursor draw command at the given absolute cell.
func (m *Model) DrawCmd(x, y int, hostGlyph string, hostStyle style.Style) (paint.DrawCmd, bool) {
	if !m.ShouldPaint() {
		return paint.DrawCmd{}, false
	}

	cursorGlyph := m.resolveGlyph(hostGlyph)
	cursorStyle := m.resolveStyle(hostStyle)

	return paint.DrawCmd{
		X:     x,
		Y:     y,
		Text:  cursorGlyph,
		Style: cursorStyle,
	}, true
}

func (m *Model) resolveGlyph(hostGlyph string) string {
	switch m.config.Shape {
	case ShapeBar:
		if m.config.Glyph != "" {
			return firstGlyph(m.config.Glyph)
		}
		return "|"
	case ShapeUnderline:
		if hostGlyph != "" {
			return firstGlyph(hostGlyph)
		}
		if m.config.Glyph != "" {
			return firstGlyph(m.config.Glyph)
		}
		return " "
	default:
		if hostGlyph != "" {
			return firstGlyph(hostGlyph)
		}
		if m.config.Glyph != "" {
			return firstGlyph(m.config.Glyph)
		}
		return " "
	}
}

func (m *Model) resolveStyle(base style.Style) style.Style {
	s := base
	themeFG := m.resolveThemeColor()

	if m.config.Style.FG != "" {
		s = s.Foreground(m.config.Style.FG)
	} else {
		s = s.Foreground(themeFG)
	}
	if m.config.Style.BG != "" {
		s = s.Background(m.config.Style.BG)
	}
	s = s.Merge(m.copyDecorations(m.config.Style))

	switch m.config.Shape {
	case ShapeUnderline:
		return s.Underline(true)
	case ShapeBar:
		return s
	default:
		return s.Reverse(true)
	}
}

func (m *Model) resolveThemeColor() style.Color {
	switch m.config.Theme {
	case ThemeFocus:
		return theme.Focus()
	case ThemeText:
		return theme.Text()
	case ThemeMuted:
		return theme.Placeholder()
	case ThemeAccent:
		return theme.Primary()
	default:
		return theme.Caret()
	}
}

func (m *Model) copyDecorations(src style.Style) style.Style {
	var s style.Style
	if src.IsBold() {
		s = s.Bold(true)
	}
	if src.IsItalic() {
		s = s.Italic(true)
	}
	if src.IsUnderline() {
		s = s.Underline(true)
	}
	if src.IsStrikethrough() {
		s = s.Strikethrough(true)
	}
	if src.IsReverse() {
		s = s.Reverse(true)
	}
	if src.IsBlink() {
		s = s.Blink(true)
	}
	return s
}

func firstGlyph(text string) string {
	for _, r := range text {
		return string(r)
	}
	return " "
}
