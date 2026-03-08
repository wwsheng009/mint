package textarea

import (
	"time"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/cursor"
)

// =============================================================================
// VNode - Description Only (No State, No Closures, No Paint)
// =============================================================================

// VNode is the textarea description.
type VNode struct {
	*rtui.ElementVNode

	// === Identification ===
	key string

	// === Visual Props ===
	placeholder string
	style       style.Style

	// === Layout Props ===
	rows                   int
	cols                   int
	scrollOffset           int
	scrollOffsetControlled bool
	showScrollbar          bool
	scrollbarStyle         style.Style

	// === Intent Props (no closures!) ===
	changeIntent intent.Intent
	submitIntent intent.Intent

	// === State Props (declarative) ===
	value        string
	maxLen       int
	disabled     bool
	formID       string // Form ID for Form integration (Phase 6)
	cursorConfig cursor.Config

	// === Box Model ===
	rtui.BoxModelMixin
}

// Ensure VNode implements required interfaces
var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
	_ rtui.BoxModel        = (*VNode)(nil)
)

// New creates a new Textarea VNode.
func New() *VNode {
	return &VNode{
		ElementVNode:  rtui.NewElement("textarea"),
		rows:          3,
		cols:          40,
		showScrollbar: true,
		cursorConfig:  cursor.DefaultConfig(),
	}
}

// =============================================================================
// rtui.VNode Interface Implementation
// =============================================================================

func (t *VNode) Key() string                                  { return t.key }
func (t *VNode) SetKey(key string) rtui.VNode                 { t.key = key; return t }
func (t *VNode) Tag() string                                  { return "textarea" }
func (t *VNode) Style() style.Style                           { return t.style }
func (t *VNode) SetStyle(s style.Style) rtui.VNode            { t.style = s; return t }
func (t *VNode) Children() []rtui.VNode                       { return nil }
func (t *VNode) SetChildren(children []rtui.VNode) rtui.VNode { return t }
func (t *VNode) GetLayer() rtui.Layer                         { return rtui.LayerBase }
func (t *VNode) SetLayer(l rtui.Layer) rtui.VNode             { return t }

func (t *VNode) Props() rtui.Props {
	props := rtui.Props{
		"key":                    t.key,
		"placeholder":            t.placeholder,
		"style":                  t.style,
		"rows":                   t.rows,
		"cols":                   t.cols,
		"scrollOffsetControlled": t.scrollOffsetControlled,
		"showScrollbar":          t.showScrollbar,
		"scrollbarStyle":         t.scrollbarStyle,
		"changeIntent":           t.changeIntent,
		"submitIntent":           t.submitIntent,
		"value":                  t.value,
		"maxLen":                 t.maxLen,
		"disabled":               t.disabled,
		"formID":                 t.formID,
		"cursorConfig":           t.cursorConfig,
	}
	if t.scrollOffsetControlled {
		props["scrollOffset"] = t.scrollOffset
	}
	return props
}

func (t *VNode) SetProps(p rtui.Props) rtui.VNode {
	if v, ok := p["key"].(string); ok {
		t.key = v
	}
	if v, ok := p["placeholder"].(string); ok {
		t.placeholder = v
	}
	if v, ok := p["style"].(style.Style); ok {
		t.style = v
	}
	if v, ok := p["rows"].(int); ok {
		t.rows = v
	}
	if v, ok := p["cols"].(int); ok {
		t.cols = v
	}
	if v, ok := p["scrollOffset"].(int); ok {
		t.scrollOffset = v
		t.scrollOffsetControlled = true
	}
	if v, ok := p["scrollOffsetControlled"].(bool); ok {
		t.scrollOffsetControlled = v
	}
	if v, ok := p["showScrollbar"].(bool); ok {
		t.showScrollbar = v
	}
	if v, ok := p["scrollbarStyle"].(style.Style); ok {
		t.scrollbarStyle = v
	}
	if v, ok := p["changeIntent"].(intent.Intent); ok {
		t.changeIntent = v
	}
	if v, ok := p["submitIntent"].(intent.Intent); ok {
		t.submitIntent = v
	}
	if v, ok := p["value"].(string); ok {
		t.value = v
	}
	if v, ok := p["maxLen"].(int); ok {
		t.maxLen = v
	}
	if v, ok := p["disabled"].(bool); ok {
		t.disabled = v
	}
	if v, ok := p["formID"].(string); ok {
		t.formID = v
	}
	if v, ok := p["cursorConfig"].(cursor.Config); ok {
		t.cursorConfig = cursor.NormalizeConfig(v)
	}
	return t
}

// =============================================================================
// InstanceFactory Implementation
// =============================================================================

func (t *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(t.Props())
}

// =============================================================================
// Builder Methods
// =============================================================================

func (t *VNode) SetPlaceholder(text string) *VNode { t.placeholder = text; return t }
func (t *VNode) SetValue(value string) *VNode      { t.value = value; return t }
func (t *VNode) SetRows(rows int) *VNode           { t.rows = rows; return t }
func (t *VNode) SetCols(cols int) *VNode           { t.cols = cols; return t }
func (t *VNode) SetScrollOffset(offset int) *VNode {
	t.scrollOffset = offset
	t.scrollOffsetControlled = true
	return t
}
func (t *VNode) SetShowScrollbar(show bool) *VNode      { t.showScrollbar = show; return t }
func (t *VNode) SetScrollbarStyle(s style.Style) *VNode { t.scrollbarStyle = s; return t }
func (t *VNode) SetMaxLen(len int) *VNode               { t.maxLen = len; return t }
func (t *VNode) SetDisabled(disabled bool) *VNode       { t.disabled = disabled; return t }
func (t *VNode) SetChangeIntent(i intent.Intent) *VNode { t.changeIntent = i; return t }
func (t *VNode) SetSubmitIntent(i intent.Intent) *VNode { t.submitIntent = i; return t }
func (t *VNode) SetCursorConfig(cfg cursor.Config) *VNode {
	t.cursorConfig = cursor.NormalizeConfig(cfg)
	return t
}

// SetCursorShape sets the embedded caret shape.
func (t *VNode) SetCursorShape(shape cursor.Shape) *VNode {
	t.cursorConfig.Shape = shape
	return t
}

// SetInsertCursor configures a thin insertion caret.
func (t *VNode) SetInsertCursor() *VNode {
	t.cursorConfig.Shape = cursor.ShapeBar
	t.cursorConfig.Glyph = "|"
	return t
}

// SetBlockCursor configures a block caret.
func (t *VNode) SetBlockCursor() *VNode {
	t.cursorConfig.Shape = cursor.ShapeBlock
	t.cursorConfig.Glyph = ""
	return t
}

// SetUnderlineCursor configures an underline caret.
func (t *VNode) SetUnderlineCursor() *VNode {
	t.cursorConfig.Shape = cursor.ShapeUnderline
	t.cursorConfig.Glyph = ""
	return t
}

// SetCursorBlink enables or disables caret blink.
func (t *VNode) SetCursorBlink(enabled bool) *VNode {
	t.cursorConfig.Blink = enabled
	if t.cursorConfig.BlinkInterval <= 0 {
		t.cursorConfig.BlinkInterval = cursor.NormalBlinkInterval
	}
	return t
}

// SetCursorBlinkInterval sets caret blink interval.
func (t *VNode) SetCursorBlinkInterval(interval time.Duration) *VNode {
	t.cursorConfig.BlinkInterval = interval
	return t
}

// =============================================================================
// Props Accessors
// =============================================================================

func (t *VNode) Placeholder() string         { return t.placeholder }
func (t *VNode) Value() string               { return t.value }
func (t *VNode) Rows() int                   { return t.rows }
func (t *VNode) Cols() int                   { return t.cols }
func (t *VNode) ScrollOffset() int           { return t.scrollOffset }
func (t *VNode) ShowScrollbar() bool         { return t.showScrollbar }
func (t *VNode) MaxLen() int                 { return t.maxLen }
func (t *VNode) Disabled() bool              { return t.disabled }
func (t *VNode) ChangeIntent() intent.Intent { return t.changeIntent }
func (t *VNode) SubmitIntent() intent.Intent { return t.submitIntent }

// SetFormID sets the form ID for Form integration (Phase 6).
func (t *VNode) SetFormID(formID string) *VNode {
	t.formID = formID
	return t
}

// =============================================================================
// layout.BoxModelProvider Implementation
// =============================================================================

// GetBoxModel returns the box model for the TextArea VNode.
// Implements layout.BoxModelProvider for unified padding/border handling.
// Note: TextArea uses BoxModelMixin for padding/margin, and has no border.
func (t *VNode) GetBoxModel() layout.BoxModel {
	return layout.BoxModel{
		Padding: layout.Padding{
			Left:   t.BoxModelMixin.Padding()[3],
			Right:  t.BoxModelMixin.Padding()[1],
			Top:    t.BoxModelMixin.Padding()[0],
			Bottom: t.BoxModelMixin.Padding()[2],
		},
		Margin: layout.Margin{
			Left:   t.BoxModelMixin.Margin()[3],
			Right:  t.BoxModelMixin.Margin()[1],
			Top:    t.BoxModelMixin.Margin()[0],
			Bottom: t.BoxModelMixin.Margin()[2],
		},
		Border: layout.Border{Style: layout.BorderNone},
	}
}
