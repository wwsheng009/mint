package timeline

import (
	"strings"

	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/internal/proputil"
)

type renderLine struct {
	label       string
	marker      string
	text        string
	labelStyle  style.Style
	markerStyle style.Style
	textStyle   style.Style
}

// Instance is the runtime entity for Timeline components.
type Instance struct {
	key          string
	items        []Item
	pending      string
	reverse      bool
	width        int
	rootStyle    style.Style
	labelStyle   style.Style
	contentStyle style.Style
	pendingStyle style.Style
	lineStyle    style.Style
	bounds       [4]int
	dirty        bool
}

var (
	_ rtui.ComponentInstance = (*Instance)(nil)
	_ rtui.PaintableInstance = (*Instance)(nil)
	_ interface {
		Measure(layout.Constraints) layout.Size
	} = (*Instance)(nil)
)

// NewInstance creates a new Timeline instance.
func NewInstance(props rtui.Props) *Instance {
	inst := &Instance{
		key:          proputil.GetString(props, propKey, ""),
		items:        normalizeItems(getItemsProp(props)),
		pending:      proputil.GetString(props, propPending, ""),
		reverse:      proputil.GetBool(props, propReverse, false),
		width:        proputil.GetInt(props, propWidth, 0),
		rootStyle:    proputil.GetStyle(props, propStyle, style.Style{}),
		labelStyle:   proputil.GetStyle(props, propLabelStyle, style.Style{}),
		contentStyle: proputil.GetStyle(props, propContentStyle, style.Style{}),
		pendingStyle: proputil.GetStyle(props, propPendingStyle, style.Style{}),
		lineStyle:    proputil.GetStyle(props, propLineStyle, style.Style{}),
		dirty:        true,
	}
	return inst
}

func (inst *Instance) Key() string           { return inst.key }
func (inst *Instance) SetKey(key string)     { inst.key = key }
func (inst *Instance) Init(props rtui.Props) { inst.SetProps(props) }
func (inst *Instance) Destroy()              {}
func (inst *Instance) OnMount()              {}
func (inst *Instance) OnUnmount()            {}

func (inst *Instance) SetProps(props rtui.Props) bool {
	oldItems := cloneItems(inst.items)
	oldPending := inst.pending
	oldReverse := inst.reverse
	oldWidth := inst.width
	oldRootStyle := inst.rootStyle
	oldLabelStyle := inst.labelStyle
	oldContentStyle := inst.contentStyle
	oldPendingStyle := inst.pendingStyle
	oldLineStyle := inst.lineStyle

	inst.key = proputil.GetString(props, propKey, inst.key)
	inst.items = normalizeItems(getItemsProp(props))
	inst.pending = proputil.GetString(props, propPending, inst.pending)
	inst.reverse = proputil.GetBool(props, propReverse, inst.reverse)
	inst.width = proputil.GetInt(props, propWidth, inst.width)
	inst.rootStyle = proputil.GetStyle(props, propStyle, inst.rootStyle)
	inst.labelStyle = proputil.GetStyle(props, propLabelStyle, inst.labelStyle)
	inst.contentStyle = proputil.GetStyle(props, propContentStyle, inst.contentStyle)
	inst.pendingStyle = proputil.GetStyle(props, propPendingStyle, inst.pendingStyle)
	inst.lineStyle = proputil.GetStyle(props, propLineStyle, inst.lineStyle)

	changed := !itemsEqual(oldItems, inst.items) ||
		oldPending != inst.pending ||
		oldReverse != inst.reverse ||
		oldWidth != inst.width ||
		oldRootStyle != inst.rootStyle ||
		oldLabelStyle != inst.labelStyle ||
		oldContentStyle != inst.contentStyle ||
		oldPendingStyle != inst.pendingStyle ||
		oldLineStyle != inst.lineStyle
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		propContentStyle: inst.contentStyle,
		propItems:        cloneItems(inst.items),
		propKey:          inst.key,
		propLabelStyle:   inst.labelStyle,
		propLineStyle:    inst.lineStyle,
		propPending:      inst.pending,
		propPendingStyle: inst.pendingStyle,
		propReverse:      inst.reverse,
		propStyle:        inst.rootStyle,
		propWidth:        inst.width,
	}
}

func (inst *Instance) MarkDirty()                         { inst.dirty = true }
func (inst *Instance) IsDirty() bool                      { return inst.dirty }
func (inst *Instance) GetContext() *rtui.ComponentContext { return nil }
func (inst *Instance) SetBounds(x, y, w, h int)           { inst.bounds = [4]int{x, y, w, h} }

func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	width, height := inst.measureLayout(constraints.MaxWidth)
	return layout.Size{
		Width:  constraints.ConstrainWidth(width),
		Height: constraints.ConstrainHeight(height),
	}
}

func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	availableWidth := inst.width
	if inst.bounds[2] > 0 {
		availableWidth = inst.bounds[2]
	}
	labelWidth, lines := inst.renderLines(availableWidth)
	if len(lines) == 0 {
		return nil
	}

	cmds := make([]paint.DrawCmd, 0, len(lines)*3)
	for index, line := range lines {
		cursor := x
		if labelWidth > 0 {
			labelText := padDisplayWidth(line.label, labelWidth)
			if labelText != "" {
				cmds = append(cmds, paint.DrawCmd{
					X:     cursor,
					Y:     y + index,
					Text:  labelText,
					Style: line.labelStyle.Merge(inst.rootStyle),
				})
			}
			cursor += labelWidth + 1
		}
		cmds = append(cmds, paint.DrawCmd{
			X:     cursor,
			Y:     y + index,
			Text:  line.marker,
			Style: line.markerStyle.Merge(inst.rootStyle),
		})
		cursor += paint.StringWidth(line.marker) + 1
		if line.text != "" {
			cmds = append(cmds, paint.DrawCmd{
				X:     cursor,
				Y:     y + index,
				Text:  line.text,
				Style: line.textStyle.Merge(inst.rootStyle),
			})
		}
	}
	return cmds
}

func (inst *Instance) measureLayout(maxWidth int) (int, int) {
	labelWidth, lines := inst.renderLines(maxWidth)
	width := 0
	for _, line := range lines {
		current := 0
		if labelWidth > 0 {
			current += labelWidth + 1
		}
		current += paint.StringWidth(line.marker)
		if line.text != "" {
			current += 1 + paint.StringWidth(line.text)
		}
		if current > width {
			width = current
		}
	}
	return width, len(lines)
}

func (inst *Instance) renderLines(maxWidth int) (int, []renderLine) {
	items := inst.orderedItems()
	labelWidth := 0
	for _, item := range items {
		if w := paint.StringWidth(item.Label); w > labelWidth {
			labelWidth = w
		}
	}
	if strings.TrimSpace(inst.pending) != "" && paint.StringWidth("Pending") > labelWidth {
		labelWidth = paint.StringWidth("Pending")
	}

	contentWidth := 0
	if maxWidth > 0 {
		contentWidth = maxWidth - labelWidth - 3
		if contentWidth < 8 {
			contentWidth = 8
		}
	}

	lines := make([]renderLine, 0, len(items)*3)
	for index, item := range items {
		lines = append(lines, inst.itemLines(item, index < len(items)-1, contentWidth)...)
	}
	if pending := strings.TrimSpace(inst.pending); pending != "" {
		pendingItem := Event(pending).WithLabel("Pending").WithStatus(StatusPending)
		lines = append(lines, inst.itemLines(pendingItem, false, contentWidth)...)
	}
	return labelWidth, lines
}

func (inst *Instance) orderedItems() []Item {
	items := cloneItems(inst.items)
	if !inst.reverse {
		return items
	}
	reversed := make([]Item, 0, len(items))
	for index := len(items) - 1; index >= 0; index-- {
		reversed = append(reversed, items[index])
	}
	return reversed
}

func (inst *Instance) itemLines(item Item, hasNext bool, contentWidth int) []renderLine {
	contentStyle := inst.contentTextStyle(item)
	lines := make([]renderLine, 0, 4)
	contentLines := wrapLines(item.Content, contentWidth)
	descLines := wrapLines(item.Description, contentWidth)
	if len(contentLines) == 0 {
		contentLines = []string{""}
	}

	marker := inst.markerText(item)
	lines = append(lines, renderLine{
		label:       item.Label,
		marker:      marker,
		text:        contentLines[0],
		labelStyle:  inst.labelTextStyle(),
		markerStyle: inst.markerStyle(item),
		textStyle:   contentStyle,
	})
	connector := "│"
	connectorStyle := inst.connectorStyle()
	for _, line := range contentLines[1:] {
		lines = append(lines, renderLine{
			marker:      connector,
			text:        line,
			labelStyle:  inst.labelTextStyle(),
			markerStyle: connectorStyle,
			textStyle:   contentStyle,
		})
	}
	descriptionStyle := inst.descriptionTextStyle(item)
	for _, line := range descLines {
		lines = append(lines, renderLine{
			marker:      connector,
			text:        line,
			labelStyle:  inst.labelTextStyle(),
			markerStyle: connectorStyle,
			textStyle:   descriptionStyle,
		})
	}
	if hasNext {
		lines = append(lines, renderLine{
			marker:      connector,
			labelStyle:  inst.labelTextStyle(),
			markerStyle: connectorStyle,
			textStyle:   contentStyle,
		})
	}
	return lines
}

func (inst *Instance) markerText(item Item) string {
	if item.Dot != "" {
		return item.Dot
	}
	switch item.Status {
	case StatusSuccess:
		return "●"
	case StatusWarning:
		return "▲"
	case StatusError:
		return "✖"
	case StatusPending:
		return "○"
	default:
		return "●"
	}
}

func (inst *Instance) markerStyle(item Item) style.Style {
	base := style.NewStyle()
	if item.Color != "" {
		base = base.Foreground(item.Color)
	} else {
		switch item.Status {
		case StatusSuccess:
			base = base.Foreground(theme.Success())
		case StatusWarning:
			base = base.Foreground(theme.Warning())
		case StatusError:
			base = base.Foreground(theme.Error())
		case StatusPending:
			base = base.Foreground(theme.Muted())
		default:
			base = base.Foreground(theme.Primary())
		}
	}
	return base.Bold(true)
}

func (inst *Instance) connectorStyle() style.Style {
	base := style.NewStyle().Foreground(theme.Muted())
	if !inst.lineStyle.IsEmpty() {
		base = base.Merge(inst.lineStyle)
	}
	return base
}

func (inst *Instance) labelTextStyle() style.Style {
	base := style.NewStyle().Foreground(theme.Muted())
	if !inst.labelStyle.IsEmpty() {
		base = base.Merge(inst.labelStyle)
	}
	return base
}

func (inst *Instance) contentTextStyle(item Item) style.Style {
	base := style.NewStyle().Foreground(theme.Text())
	if item.Status == StatusPending && !inst.pendingStyle.IsEmpty() {
		return base.Merge(inst.pendingStyle)
	}
	if !inst.contentStyle.IsEmpty() {
		base = base.Merge(inst.contentStyle)
	}
	return base
}

func (inst *Instance) descriptionTextStyle(item Item) style.Style {
	base := style.NewStyle().Foreground(theme.Muted())
	if item.Status == StatusPending && !inst.pendingStyle.IsEmpty() {
		return base.Merge(inst.pendingStyle)
	}
	if !inst.contentStyle.IsEmpty() {
		base = base.Merge(inst.contentStyle)
	}
	return base
}

func getItemsProp(props rtui.Props) []Item {
	if items, ok := props[propItems].([]Item); ok {
		return cloneItems(items)
	}
	return nil
}

func itemsEqual(a, b []Item) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func wrapLines(text string, maxWidth int) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	rawLines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	lines := make([]string, 0, len(rawLines))
	for _, raw := range rawLines {
		if maxWidth <= 0 || paint.StringWidth(raw) <= maxWidth {
			lines = append(lines, raw)
			continue
		}
		lines = append(lines, wrapSingleLine(raw, maxWidth)...)
	}
	return lines
}

func wrapSingleLine(text string, maxWidth int) []string {
	if maxWidth <= 0 {
		return []string{text}
	}
	var lines []string
	var builder strings.Builder
	width := 0
	for _, r := range text {
		rw := paint.RuneWidth(r)
		if width+rw > maxWidth && builder.Len() > 0 {
			lines = append(lines, builder.String())
			builder.Reset()
			width = 0
		}
		builder.WriteRune(r)
		width += rw
	}
	if builder.Len() > 0 {
		lines = append(lines, builder.String())
	}
	return lines
}

func truncateByDisplayWidth(text string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	var builder strings.Builder
	width := 0
	for _, r := range text {
		rw := paint.RuneWidth(r)
		if width+rw > maxWidth {
			break
		}
		builder.WriteRune(r)
		width += rw
	}
	return builder.String()
}

func padDisplayWidth(content string, width int) string {
	content = truncateByDisplayWidth(content, width)
	padding := width - paint.StringWidth(content)
	if padding <= 0 {
		return content
	}
	return content + strings.Repeat(" ", padding)
}
