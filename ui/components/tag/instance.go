package tag

import (
	"fmt"
	"unicode/utf8"

	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui/components/internal/proputil"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Instance - Runtime Entity
// =============================================================================

// Instance is the runtime entity for Tag components.
type Instance struct {
	key         string
	text        string
	color       TagColor
	closable    bool
	closeIntent interface{}
	icon        string
	tagStyle    style.Style
	bounds      [4]int
	dirty       bool
}

var (
	_ rtui.ComponentInstance                               = (*Instance)(nil)
	_ rtui.PaintableInstance                              = (*Instance)(nil)
	_ interface{ Measure(layout.Constraints) layout.Size } = (*Instance)(nil)
)

// NewInstance creates a new Tag Instance from props.
func NewInstance(props rtui.Props) *Instance {
	return &Instance{
		key:         proputil.GetString(props, propKey, ""),
		text:        proputil.GetString(props, propText, ""),
		color:       getTagColorProp(props, ColorDefault),
		closable:    proputil.GetBool(props, propClosable, false),
		closeIntent: props[propCloseIntent],
		icon:        proputil.GetString(props, propIcon, ""),
		tagStyle:    proputil.GetStyle(props, propStyle, style.Style{}),
		dirty:       true,
	}
}

// =============================================================================
// ComponentInstance Interface
// =============================================================================

func (inst *Instance) Key() string                      { return inst.key }
func (inst *Instance) SetKey(key string)                { inst.key = key }
func (inst *Instance) IsDirty() bool                    { return inst.dirty }
func (inst *Instance) MarkClean()                       { inst.dirty = false }
func (inst *Instance) MarkDirty()                       { inst.dirty = true }
func (inst *Instance) Destroy()                         {}
func (inst *Instance) OnMount()                         {}
func (inst *Instance) OnUnmount()                       {}
func (inst *Instance) GetContext() *rtui.ComponentContext { return nil }

func (inst *Instance) Init(props rtui.Props) { inst.SetProps(props) }

func (inst *Instance) SetProps(props rtui.Props) bool {
	inst.text = proputil.GetString(props, propText, "")
	inst.color = getTagColorProp(props, ColorDefault)
	inst.closable = proputil.GetBool(props, propClosable, false)
	if ci, ok := props[propCloseIntent]; ok {
		inst.closeIntent = ci
	}
	inst.icon = proputil.GetString(props, propIcon, "")
	inst.tagStyle = proputil.GetStyle(props, propStyle, style.Style{})
	inst.dirty = true
	return true
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		propKey:         inst.key,
		propText:        inst.text,
		propColor:       inst.color,
		propClosable:    inst.closable,
		propCloseIntent: inst.closeIntent,
		propIcon:        inst.icon,
		propStyle:       inst.tagStyle,
	}
}

// =============================================================================
// Measure
// =============================================================================

// Measure returns the size needed to render the tag.
func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	// width = icon + space + text + closable suffix + 2 padding spaces
	width := utf8.RuneCountInString(inst.text) + 2 // leading and trailing space
	if inst.icon != "" {
		width += utf8.RuneCountInString(inst.icon) + 1 // icon + space separator
	}
	if inst.closable {
		width += 2 // " ×"
	}
	if constraints.MaxWidth > 0 && width > constraints.MaxWidth {
		width = constraints.MaxWidth
	}
	return layout.Size{Width: width, Height: 1}
}

// =============================================================================
// PaintableInstance Interface
// =============================================================================

// SetBounds sets the render bounds.
func (inst *Instance) SetBounds(x, y, w, h int) {
	inst.bounds = [4]int{x, y, w, h}
}

// Paint renders the tag as a draw command.
func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	// Build display text: " " + [icon+" "] + text + [" ×"] + " "
	content := ""
	if inst.icon != "" {
		content = inst.icon + " "
	}
	content += inst.text
	if inst.closable {
		content += " ×"
	}
	text := fmt.Sprintf(" %s ", content)

	s := inst.resolveStyle()

	return []paint.DrawCmd{
		{X: x, Y: y, Text: text, Style: s},
	}
}

// resolveStyle returns the style for the tag based on its color.
func (inst *Instance) resolveStyle() style.Style {
	s := inst.tagStyle
	switch inst.color {
	case ColorPrimary:
		return s.Foreground(theme.BG()).Background(style.Color("blue")).Bold(true)
	case ColorSuccess:
		return s.Foreground(theme.BG()).Background(style.Color("green")).Bold(true)
	case ColorWarning:
		return s.Foreground(theme.BG()).Background(style.Color("yellow")).Bold(true)
	case ColorError:
		return s.Foreground(theme.BG()).Background(style.Color("red")).Bold(true)
	case ColorProcessing:
		return s.Foreground(theme.BG()).Background(style.Color("cyan")).Bold(true)
	default:
		return s.Foreground(theme.Foreground()).Background(theme.Surface())
	}
}

// =============================================================================
// Prop Helper
// =============================================================================

func getTagColorProp(props rtui.Props, def TagColor) TagColor {
	if v, ok := props[propColor]; ok {
		if c, ok := v.(TagColor); ok {
			return c
		}
	}
	return def
}
