package badge

import (
	"fmt"
	"strconv"

	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/internal/proputil"
)

type badgeSegment struct {
	text  string
	style style.Style
	width int
}

// Instance is the runtime entity for Badge components.
type Instance struct {
	key           string
	label         string
	count         int
	text          string
	dot           bool
	showZero      bool
	overflowCount int
	status        Status
	baseStyle     style.Style
	labelStyle    style.Style
	badgeStyle    style.Style
	bounds        [4]int
	dirty         bool
}

var (
	_ rtui.ComponentInstance = (*Instance)(nil)
	_ rtui.PaintableInstance = (*Instance)(nil)
	_ interface {
		Measure(layout.Constraints) layout.Size
	} = (*Instance)(nil)
)

// NewInstance creates a new Badge instance from props.
func NewInstance(props rtui.Props) *Instance {
	inst := &Instance{
		key:           proputil.GetString(props, propKey, ""),
		label:         proputil.GetString(props, propLabel, ""),
		count:         proputil.GetInt(props, propCount, 0),
		text:          proputil.GetString(props, propText, ""),
		dot:           proputil.GetBool(props, propDot, false),
		showZero:      proputil.GetBool(props, propShowZero, false),
		overflowCount: proputil.GetInt(props, propOverflowCount, 99),
		status:        getStatusProp(props, StatusError),
		baseStyle:     proputil.GetStyle(props, propStyle, style.Style{}),
		labelStyle:    proputil.GetStyle(props, propLabelStyle, style.Style{}),
		badgeStyle:    proputil.GetStyle(props, propBadgeStyle, style.Style{}),
		dirty:         true,
	}
	inst.normalize()
	return inst
}

func (inst *Instance) Key() string                        { return inst.key }
func (inst *Instance) SetKey(key string)                  { inst.key = key }
func (inst *Instance) IsDirty() bool                      { return inst.dirty }
func (inst *Instance) MarkClean()                         { inst.dirty = false }
func (inst *Instance) MarkDirty()                         { inst.dirty = true }
func (inst *Instance) Destroy()                           {}
func (inst *Instance) OnMount()                           {}
func (inst *Instance) OnUnmount()                         {}
func (inst *Instance) GetContext() *rtui.ComponentContext { return nil }
func (inst *Instance) Init(props rtui.Props)              { inst.SetProps(props) }

func (inst *Instance) SetProps(props rtui.Props) bool {
	old := *inst
	inst.key = proputil.GetString(props, propKey, inst.key)
	inst.label = proputil.GetString(props, propLabel, inst.label)
	inst.count = proputil.GetInt(props, propCount, inst.count)
	inst.text = proputil.GetString(props, propText, inst.text)
	inst.dot = proputil.GetBool(props, propDot, inst.dot)
	inst.showZero = proputil.GetBool(props, propShowZero, inst.showZero)
	inst.overflowCount = proputil.GetInt(props, propOverflowCount, inst.overflowCount)
	inst.status = getStatusProp(props, inst.status)
	inst.baseStyle = proputil.GetStyle(props, propStyle, inst.baseStyle)
	inst.labelStyle = proputil.GetStyle(props, propLabelStyle, inst.labelStyle)
	inst.badgeStyle = proputil.GetStyle(props, propBadgeStyle, inst.badgeStyle)
	inst.normalize()

	changed := old.key != inst.key ||
		old.label != inst.label ||
		old.count != inst.count ||
		old.text != inst.text ||
		old.dot != inst.dot ||
		old.showZero != inst.showZero ||
		old.overflowCount != inst.overflowCount ||
		old.status != inst.status ||
		old.baseStyle != inst.baseStyle ||
		old.labelStyle != inst.labelStyle ||
		old.badgeStyle != inst.badgeStyle
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		propBadgeStyle:    inst.badgeStyle,
		propCount:         inst.count,
		propDot:           inst.dot,
		propKey:           inst.key,
		propLabel:         inst.label,
		propLabelStyle:    inst.labelStyle,
		propOverflowCount: inst.overflowCount,
		propShowZero:      inst.showZero,
		propStatus:        inst.status,
		propStyle:         inst.baseStyle,
		propText:          inst.text,
	}
}

func (inst *Instance) SetBounds(x, y, w, h int) {
	inst.bounds = [4]int{x, y, w, h}
}

// Measure returns the size needed to render the badge.
func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	width := 0
	for _, seg := range inst.segments() {
		width += seg.width
	}
	if constraints.MaxWidth > 0 && width > constraints.MaxWidth {
		width = constraints.MaxWidth
	}
	return layout.Size{Width: width, Height: 1}
}

// Paint renders the badge as draw commands.
func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	segments := inst.segments()
	if len(segments) == 0 {
		return nil
	}
	cmds := make([]paint.DrawCmd, 0, len(segments))
	cursor := x
	maxWidth := inst.bounds[2]
	for _, seg := range segments {
		if seg.text == "" {
			continue
		}
		if maxWidth > 0 && cursor-x >= maxWidth {
			break
		}
		cmdText := seg.text
		if maxWidth > 0 {
			remaining := maxWidth - (cursor - x)
			if remaining <= 0 {
				break
			}
			if seg.width > remaining {
				cmdText = truncateByDisplayWidth(seg.text, remaining)
			}
		}
		if cmdText == "" {
			break
		}
		cmds = append(cmds, paint.DrawCmd{
			X:     cursor,
			Y:     y,
			Text:  cmdText,
			Style: seg.style,
		})
		cursor += paint.StringWidth(cmdText)
	}
	return cmds
}

func (inst *Instance) segments() []badgeSegment {
	segments := make([]badgeSegment, 0, 3)
	if inst.label != "" {
		segments = append(segments, badgeSegment{
			text:  inst.label,
			style: inst.baseStyle.Merge(inst.labelStyle),
			width: paint.StringWidth(inst.label),
		})
	}
	if badgeText, ok := inst.badgeText(); ok {
		if len(segments) > 0 {
			segments = append(segments, badgeSegment{text: " ", style: inst.baseStyle, width: 1})
		}
		segments = append(segments, badgeSegment{
			text:  badgeText,
			style: inst.resolveBadgeStyle(),
			width: paint.StringWidth(badgeText),
		})
	}
	return segments
}

func (inst *Instance) badgeText() (string, bool) {
	if inst.dot {
		return "●", true
	}
	if inst.text != "" {
		return "[" + inst.text + "]", true
	}
	if inst.count < 0 {
		return "", false
	}
	if inst.count == 0 && !inst.showZero {
		return "", false
	}
	value := inst.count
	if inst.overflowCount > 0 && value > inst.overflowCount {
		return "[" + strconv.Itoa(inst.overflowCount) + "+]", true
	}
	return fmt.Sprintf("[%d]", value), true
}

func (inst *Instance) resolveBadgeStyle() style.Style {
	base := inst.baseStyle.Merge(inst.badgeStyle)
	if inst.dot {
		switch inst.status {
		case StatusPrimary:
			return base.Foreground(theme.Primary()).Bold(true)
		case StatusSuccess:
			return base.Foreground(theme.Success()).Bold(true)
		case StatusWarning:
			return base.Foreground(theme.Warning()).Bold(true)
		case StatusProcessing:
			return base.Foreground(theme.Primary()).Bold(true)
		case StatusDefault:
			return base.Foreground(theme.Muted()).Bold(true)
		default:
			return base.Foreground(theme.Error()).Bold(true)
		}
	}
	switch inst.status {
	case StatusPrimary:
		return base.Foreground(theme.BG()).Background(theme.Primary()).Bold(true)
	case StatusSuccess:
		return base.Foreground(theme.BG()).Background(theme.Success()).Bold(true)
	case StatusWarning:
		return base.Foreground(theme.BG()).Background(theme.Warning()).Bold(true)
	case StatusProcessing:
		return base.Foreground(theme.BG()).Background(theme.Primary()).Bold(true)
	case StatusDefault:
		return base.Foreground(theme.Foreground()).Background(theme.Surface())
	default:
		return base.Foreground(theme.BG()).Background(theme.Error()).Bold(true)
	}
}

func (inst *Instance) normalize() {
	if inst.overflowCount < 0 {
		inst.overflowCount = 0
	}
}

func getStatusProp(props rtui.Props, def Status) Status {
	if value, ok := props[propStatus].(Status); ok {
		return value
	}
	return def
}

func truncateByDisplayWidth(content string, width int) string {
	if width <= 0 {
		return ""
	}
	runes := []rune(content)
	result := make([]rune, 0, len(runes))
	currentWidth := 0
	for _, r := range runes {
		runeWidth := paint.RuneWidth(r)
		if currentWidth+runeWidth > width {
			break
		}
		result = append(result, r)
		currentWidth += runeWidth
	}
	return string(result)
}
