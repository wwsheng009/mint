package steps

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/internal/proputil"
)

type segment struct {
	text      string
	style     style.Style
	width     int
	stepIndex int
}

// Instance is the runtime entity for Steps components.
type Instance struct {
	key                string
	componentID        string
	items              []Item
	current            int
	currentControlled  bool
	disabled           bool
	percent            int
	progressDot        bool
	direction          Direction
	stepsStyle         style.Style
	titleStyle         style.Style
	descriptionStyle   style.Style
	separatorStyle     style.Style
	waitStyle          style.Style
	processStyle       style.Style
	finishStyle        style.Style
	errorStyle         style.Style
	currentIntent      intent.Intent
	currentIntentField intent.FieldIntent
	bounds             [4]int
	dirty              bool
	focused            bool
	parent             rtui.ComponentInstance
	intentEmitter      func(intent.Intent)
}

var (
	_ rtui.ComponentInstance     = (*Instance)(nil)
	_ rtui.PaintableInstance     = (*Instance)(nil)
	_ rtui.ActionHandlerInstance = (*Instance)(nil)
	_ rtui.FocusableInstance     = (*Instance)(nil)
	_ interface {
		Measure(layout.Constraints) layout.Size
	} = (*Instance)(nil)
)

// NewInstance creates a new Steps instance from props.
func NewInstance(props rtui.Props) *Instance {
	currentControlled := proputil.GetBool(props, propCurrentControlled, false)
	current := proputil.GetInt(props, propInitialCurrent, 0)
	if currentControlled {
		current = proputil.GetInt(props, propCurrent, 0)
	}
	inst := &Instance{
		key:                proputil.GetString(props, propKey, ""),
		componentID:        proputil.GetString(props, propComponentID, ""),
		items:              getItemsProp(props),
		current:            current,
		currentControlled:  currentControlled,
		disabled:           proputil.GetBool(props, propDisabled, false),
		percent:            proputil.GetInt(props, propPercent, -1),
		progressDot:        proputil.GetBool(props, propProgressDot, false),
		direction:          getDirectionProp(props),
		stepsStyle:         proputil.GetStyle(props, propStyle, style.Style{}),
		titleStyle:         proputil.GetStyle(props, propTitleStyle, style.Style{}),
		descriptionStyle:   proputil.GetStyle(props, propDescriptionStyle, style.Style{}),
		separatorStyle:     proputil.GetStyle(props, propSeparatorStyle, style.Style{}),
		waitStyle:          proputil.GetStyle(props, propWaitStyle, style.Style{}),
		processStyle:       proputil.GetStyle(props, propProcessStyle, style.Style{}),
		finishStyle:        proputil.GetStyle(props, propFinishStyle, style.Style{}),
		errorStyle:         proputil.GetStyle(props, propErrorStyle, style.Style{}),
		currentIntent:      proputil.GetIntent(props, propCurrentIntent, nil),
		currentIntentField: getFieldIntentProp(props, propCurrentIntent),
		dirty:              true,
	}
	inst.normalize()
	return inst
}

func (inst *Instance) Key() string { return inst.key }

func (inst *Instance) SetKey(key string) { inst.key = key }

func (inst *Instance) Init(props rtui.Props) { inst.SetProps(props) }

func (inst *Instance) Destroy() {}

func (inst *Instance) OnMount() {}

func (inst *Instance) OnUnmount() {}

func (inst *Instance) SetProps(props rtui.Props) bool {
	old := *inst
	oldItems := cloneItems(inst.items)

	inst.key = proputil.GetString(props, propKey, inst.key)
	inst.componentID = proputil.GetString(props, propComponentID, inst.componentID)
	inst.items = getItemsProp(props)
	nextControlled := proputil.GetBool(props, propCurrentControlled, inst.currentControlled)
	if nextControlled {
		inst.current = proputil.GetInt(props, propCurrent, inst.current)
	} else if old.currentControlled && !nextControlled {
		inst.current = proputil.GetInt(props, propInitialCurrent, inst.current)
	}
	inst.currentControlled = nextControlled
	inst.disabled = proputil.GetBool(props, propDisabled, inst.disabled)
	inst.percent = proputil.GetInt(props, propPercent, inst.percent)
	inst.progressDot = proputil.GetBool(props, propProgressDot, inst.progressDot)
	inst.direction = getDirectionPropWithDefault(props, inst.direction)
	inst.stepsStyle = proputil.GetStyle(props, propStyle, inst.stepsStyle)
	inst.titleStyle = proputil.GetStyle(props, propTitleStyle, inst.titleStyle)
	inst.descriptionStyle = proputil.GetStyle(props, propDescriptionStyle, inst.descriptionStyle)
	inst.separatorStyle = proputil.GetStyle(props, propSeparatorStyle, inst.separatorStyle)
	inst.waitStyle = proputil.GetStyle(props, propWaitStyle, inst.waitStyle)
	inst.processStyle = proputil.GetStyle(props, propProcessStyle, inst.processStyle)
	inst.finishStyle = proputil.GetStyle(props, propFinishStyle, inst.finishStyle)
	inst.errorStyle = proputil.GetStyle(props, propErrorStyle, inst.errorStyle)
	inst.currentIntent = proputil.GetIntent(props, propCurrentIntent, inst.currentIntent)
	inst.currentIntentField = getFieldIntentProp(props, propCurrentIntent)
	inst.normalize()

	changed := old.key != inst.key ||
		old.componentID != inst.componentID ||
		!itemsEqual(oldItems, inst.items) ||
		old.current != inst.current ||
		old.currentControlled != inst.currentControlled ||
		old.disabled != inst.disabled ||
		old.percent != inst.percent ||
		old.progressDot != inst.progressDot ||
		old.direction != inst.direction ||
		old.stepsStyle != inst.stepsStyle ||
		old.titleStyle != inst.titleStyle ||
		old.descriptionStyle != inst.descriptionStyle ||
		old.separatorStyle != inst.separatorStyle ||
		old.waitStyle != inst.waitStyle ||
		old.processStyle != inst.processStyle ||
		old.finishStyle != inst.finishStyle ||
		old.errorStyle != inst.errorStyle ||
		old.currentIntent != inst.currentIntent ||
		old.currentIntentField != inst.currentIntentField
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		propKey:                inst.key,
		propComponentID:        inst.componentID,
		propItems:              cloneItems(inst.items),
		propCurrent:            inst.current,
		propCurrentControlled:  inst.currentControlled,
		propDisabled:           inst.disabled,
		propPercent:            inst.percent,
		propProgressDot:        inst.progressDot,
		propDirection:          inst.direction,
		propStyle:              inst.stepsStyle,
		propTitleStyle:         inst.titleStyle,
		propDescriptionStyle:   inst.descriptionStyle,
		propSeparatorStyle:     inst.separatorStyle,
		propWaitStyle:          inst.waitStyle,
		propProcessStyle:       inst.processStyle,
		propFinishStyle:        inst.finishStyle,
		propErrorStyle:         inst.errorStyle,
		propCurrentIntent:      inst.currentIntent,
		propCurrentIntentField: inst.currentIntentField,
	}
}

func (inst *Instance) MarkDirty() { inst.dirty = true }

func (inst *Instance) IsDirty() bool { return inst.dirty }

func (inst *Instance) MarkClean() { inst.dirty = false }

func (inst *Instance) GetContext() *rtui.ComponentContext { return nil }

func (inst *Instance) Parent() interface{} { return inst.parent }

func (inst *Instance) SetParent(parent rtui.ComponentInstance) { inst.parent = parent }

func (inst *Instance) SetIntentEmitter(fn func(intent.Intent)) { inst.intentEmitter = fn }

func (inst *Instance) SetBounds(x, y, w, h int) {
	inst.bounds = [4]int{x, y, w, h}
}

func (inst *Instance) SetFocus(focused bool) {
	if inst.focused == focused {
		return
	}
	inst.focused = focused
	inst.dirty = true
}

func (inst *Instance) HasFocus() bool { return inst.focused }

func (inst *Instance) IsDisabled() bool { return inst.disabled }

// Measure returns the size needed to render steps.
func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	if inst.direction == DirectionVertical {
		width, height := inst.verticalIntrinsicSize()
		return layout.Size{
			Width:  constraints.ConstrainWidth(width),
			Height: constraints.ConstrainHeight(height),
		}
	}

	width := inst.displayWidth(inst.horizontalSegments())
	return layout.Size{
		Width:  constraints.ConstrainWidth(width),
		Height: constraints.ConstrainHeight(1),
	}
}

// Paint renders the steps component.
func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	if inst.direction == DirectionVertical {
		return inst.paintVertical(x, y)
	}
	return inst.paintHorizontal(x, y)
}

func (inst *Instance) HandleAction(act *action.Action) bool {
	if act == nil || inst.disabled || len(inst.items) == 0 {
		return false
	}
	switch act.Type {
	case action.ActionNavigateLeft, action.ActionNavigatePrev, action.ActionNavigateUp:
		return inst.applyCurrent(inst.current - 1)
	case action.ActionNavigateRight, action.ActionNavigateNext, action.ActionNavigateDown:
		return inst.applyCurrent(inst.current + 1)
	case action.ActionNavigateHome, action.ActionNavigateFirst:
		return inst.applyCurrent(0)
	case action.ActionNavigateEnd, action.ActionNavigateLast:
		return inst.applyCurrent(len(inst.items) - 1)
	case action.ActionClick, action.ActionSelect, action.ActionEnter:
		if mouseMsg, ok := act.Payload.(*runtimemsg.MouseMsg); ok && mouseMsg != nil {
			return inst.handleClick(mouseMsg.LocalX, mouseMsg.LocalY)
		}
	}
	return false
}

func (inst *Instance) HandleIntent(i intent.Intent) bool {
	change, ok := i.(StepChangeIntent)
	if !ok {
		return false
	}
	if inst.componentID != "" && change.ComponentID != "" && change.ComponentID != inst.componentID {
		return false
	}
	return inst.applyCurrent(change.ToIndex)
}

func (inst *Instance) GetCurrent() int { return inst.current }

func (inst *Instance) paintHorizontal(x, y int) []paint.DrawCmd {
	segments := inst.horizontalSegments()
	if len(segments) == 0 {
		return nil
	}

	width := inst.bounds[2]
	if width > 0 {
		segments = fitSegments(segments, width)
	}

	cmds := make([]paint.DrawCmd, 0, len(segments))
	cursor := x
	for _, seg := range segments {
		if seg.text == "" {
			continue
		}
		cmds = append(cmds, paint.DrawCmd{
			X:     cursor,
			Y:     y,
			Text:  seg.text,
			Style: seg.style,
		})
		cursor += seg.width
	}
	return cmds
}

func (inst *Instance) paintVertical(x, y int) []paint.DrawCmd {
	lines := inst.verticalLines()
	if len(lines) == 0 {
		return nil
	}

	width := inst.bounds[2]
	height := inst.bounds[3]
	if height > 0 && len(lines) > height {
		lines = lines[:height]
	}

	cmds := make([]paint.DrawCmd, 0, len(lines)*2)
	for lineIndex, line := range lines {
		if width > 0 {
			line = fitSegments(line, width)
		}
		cursor := x
		for _, seg := range line {
			if seg.text == "" {
				continue
			}
			cmds = append(cmds, paint.DrawCmd{
				X:     cursor,
				Y:     y + lineIndex,
				Text:  seg.text,
				Style: seg.style,
			})
			cursor += seg.width
		}
	}
	return cmds
}

func (inst *Instance) horizontalSegments() []segment {
	if len(inst.items) == 0 {
		return nil
	}

	separator := segment{
		text:      " ── ",
		style:     inst.resolvedSeparatorStyle(),
		width:     paint.StringWidth(" ── "),
		stepIndex: -1,
	}
	segments := make([]segment, 0, len(inst.items)*4)
	for index, item := range inst.items {
		if index > 0 {
			segments = append(segments, separator)
		}
		status := inst.resolvedStatus(index, item)
		indicator := inst.indicatorText(index, item, status)
		title := strings.TrimSpace(item.Title)
		if title == "" {
			title = fmt.Sprintf("Step %d", index+1)
		}

		segments = append(segments,
			segment{text: indicator, style: inst.indicatorStyle(status), width: paint.StringWidth(indicator), stepIndex: index},
			segment{text: " " + title, style: inst.titleStyleFor(status), width: paint.StringWidth(" " + title), stepIndex: index},
		)
		if desc := strings.TrimSpace(item.Description); desc != "" {
			text := " — " + desc
			segments = append(segments, segment{
				text:      text,
				style:     inst.descriptionStyleFor(status),
				width:     paint.StringWidth(text),
				stepIndex: index,
			})
		}
	}
	return segments
}

func (inst *Instance) verticalLines() [][]segment {
	if len(inst.items) == 0 {
		return nil
	}

	lines := make([][]segment, 0, len(inst.items)*3)
	separatorStyle := inst.resolvedSeparatorStyle()
	for index, item := range inst.items {
		status := inst.resolvedStatus(index, item)
		indicator := inst.indicatorText(index, item, status)
		title := strings.TrimSpace(item.Title)
		if title == "" {
			title = fmt.Sprintf("Step %d", index+1)
		}
		indicatorWidth := paint.StringWidth(indicator)
		lines = append(lines, []segment{
			{text: indicator, style: inst.indicatorStyle(status), width: indicatorWidth, stepIndex: index},
			{text: " " + title, style: inst.titleStyleFor(status), width: paint.StringWidth(" " + title), stepIndex: index},
		})

		if desc := strings.TrimSpace(item.Description); desc != "" {
			prefix := strings.Repeat(" ", indicatorWidth+1)
			text := prefix + desc
			lines = append(lines, []segment{{
				text:      text,
				style:     inst.descriptionStyleFor(status),
				width:     paint.StringWidth(text),
				stepIndex: index,
			}})
		}

		if index < len(inst.items)-1 {
			prefix := strings.Repeat(" ", maxInt(1, indicatorWidth/2))
			text := prefix + "│"
			lines = append(lines, []segment{{
				text:      text,
				style:     separatorStyle,
				width:     paint.StringWidth(text),
				stepIndex: -1,
			}})
		}
	}
	return lines
}

func (inst *Instance) verticalIntrinsicSize() (int, int) {
	lines := inst.verticalLines()
	maxWidth := 0
	for _, line := range lines {
		maxWidth = maxInt(maxWidth, inst.displayWidth(line))
	}
	return maxWidth, len(lines)
}

func (inst *Instance) resolvedStatus(index int, item Item) Status {
	if item.Status != StatusAuto {
		return item.Status
	}
	if inst.current < 0 {
		return StatusWait
	}
	if index < inst.current {
		return StatusFinish
	}
	if index == inst.current {
		return StatusProcess
	}
	return StatusWait
}

func (inst *Instance) indicatorText(index int, item Item, status Status) string {
	if inst.progressDot {
		switch status {
		case StatusFinish:
			return "●"
		case StatusError:
			return "◉"
		case StatusProcess:
			return progressDotGlyph(inst.percent)
		default:
			return "○"
		}
	}
	if status == StatusProcess && inst.percent >= 0 {
		return fmt.Sprintf("[%d%%]", inst.percent)
	}
	if item.Icon != "" && status != StatusFinish && status != StatusError {
		return "[" + item.Icon + "]"
	}
	switch status {
	case StatusFinish:
		return "[✓]"
	case StatusError:
		return "[!]"
	default:
		return fmt.Sprintf("[%d]", index+1)
	}
}

func progressDotGlyph(percent int) string {
	switch {
	case percent < 0:
		return "●"
	case percent < 25:
		return "◔"
	case percent < 50:
		return "◑"
	case percent < 75:
		return "◕"
	default:
		return "●"
	}
}

func (inst *Instance) indicatorStyle(status Status) style.Style {
	return inst.titleStyleFor(status)
}

func (inst *Instance) titleStyleFor(status Status) style.Style {
	base := inst.stepsStyle.Merge(inst.titleStyle)
	if inst.focused && status == StatusProcess {
		base = base.Merge(style.NewStyle().Underline(true))
	}
	switch status {
	case StatusFinish:
		return base.Merge(style.NewStyle().Foreground(theme.Success()).Bold(true)).Merge(inst.finishStyle)
	case StatusError:
		return base.Merge(style.NewStyle().Foreground(theme.Error()).Bold(true)).Merge(inst.errorStyle)
	case StatusProcess:
		return base.Merge(style.NewStyle().Foreground(theme.Primary()).Bold(true)).Merge(inst.processStyle)
	default:
		return base.Merge(style.NewStyle().Foreground(theme.Muted())).Merge(inst.waitStyle)
	}
}

func (inst *Instance) descriptionStyleFor(status Status) style.Style {
	base := inst.stepsStyle.Merge(style.NewStyle().Foreground(theme.Muted()))
	base = base.Merge(inst.descriptionStyle)
	switch status {
	case StatusError:
		return base.Merge(style.NewStyle().Foreground(theme.Error()))
	case StatusProcess:
		return base.Merge(style.NewStyle().Foreground(theme.Foreground()))
	default:
		return base
	}
}

func (inst *Instance) resolvedSeparatorStyle() style.Style {
	return inst.stepsStyle.Merge(style.NewStyle().Foreground(theme.Muted())).Merge(inst.separatorStyle)
}

func (inst *Instance) displayWidth(segments []segment) int {
	width := 0
	for _, seg := range segments {
		width += seg.width
	}
	return width
}

func (inst *Instance) handleClick(localX, localY int) bool {
	if inst.direction == DirectionVertical {
		index, ok := inst.verticalStepIndexAt(localY)
		if !ok {
			return false
		}
		return inst.applyCurrent(index)
	}
	if localY != 0 {
		return false
	}
	segments := inst.horizontalSegments()
	width := inst.bounds[2]
	if width > 0 {
		segments = fitSegments(segments, width)
	}
	cursor := 0
	for _, seg := range segments {
		if localX >= cursor && localX < cursor+seg.width {
			if seg.stepIndex >= 0 {
				return inst.applyCurrent(seg.stepIndex)
			}
			return false
		}
		cursor += seg.width
	}
	return false
}

func (inst *Instance) verticalStepIndexAt(localY int) (int, bool) {
	if localY < 0 {
		return 0, false
	}
	row := 0
	for index, item := range inst.items {
		if localY == row {
			return index, true
		}
		row++
		if strings.TrimSpace(item.Description) != "" {
			if localY == row {
				return index, true
			}
			row++
		}
		if index < len(inst.items)-1 {
			if localY == row {
				return 0, false
			}
			row++
		}
	}
	return 0, false
}

func (inst *Instance) applyCurrent(target int) bool {
	if len(inst.items) == 0 {
		return false
	}
	clamped := clampInt(target, 0, len(inst.items)-1)
	if inst.current == clamped {
		return false
	}
	from := inst.current
	inst.current = clamped
	inst.dirty = true
	inst.emitCurrentChanged(from, clamped)
	return true
}

func (inst *Instance) emitCurrentChanged(fromIndex, toIndex int) {
	if inst.intentEmitter == nil {
		return
	}
	item := inst.items[toIndex]
	if inst.componentID != "" {
		inst.intentEmitter(StepChangeWithID(inst.componentID, fromIndex, toIndex, len(inst.items), item.Key, item.Title))
	} else {
		inst.intentEmitter(StepChange(fromIndex, toIndex, len(inst.items), item.Key, item.Title))
	}
	if inst.currentIntentField != nil {
		inst.intentEmitter(intent.FieldChangeIntent{
			Field: inst.currentIntentField.GetField(),
			Value: strconv.Itoa(toIndex),
		})
		return
	}
	if inst.currentIntent != nil {
		inst.intentEmitter(inst.currentIntent)
	}
}

func (inst *Instance) normalize() {
	if len(inst.items) == 0 {
		inst.current = 0
		return
	}
	inst.current = clampInt(inst.current, 0, len(inst.items)-1)
	if inst.percent < 0 {
		inst.percent = -1
	}
	if inst.percent > 100 {
		inst.percent = 100
	}
}

func fitSegments(segments []segment, width int) []segment {
	if width <= 0 || len(segments) == 0 {
		return segments
	}
	fitted := make([]segment, 0, len(segments))
	used := 0
	for index, seg := range segments {
		if seg.width <= 0 {
			seg.width = paint.StringWidth(seg.text)
		}
		isLast := index == len(segments)-1
		if used+seg.width < width || (used+seg.width == width && isLast) {
			fitted = append(fitted, seg)
			used += seg.width
			continue
		}
		remaining := width - used
		if remaining <= 0 {
			break
		}
		truncated := truncateWithEllipsis(seg.text, remaining)
		if !isLast {
			truncated = truncateWithContinuation(seg.text, remaining)
		}
		if truncated != "" {
			fitted = append(fitted, segment{
				text:      truncated,
				style:     seg.style,
				width:     paint.StringWidth(truncated),
				stepIndex: seg.stepIndex,
			})
		}
		break
	}
	return fitted
}

func truncateWithContinuation(content string, width int) string {
	const ellipsis = "…"
	if width <= 0 {
		return ""
	}
	ellipsisWidth := paint.StringWidth(ellipsis)
	if width <= ellipsisWidth {
		return ellipsis
	}
	trimmed := strings.TrimRight(truncateByDisplayWidth(content, width-ellipsisWidth), " ")
	if trimmed == "" {
		return ellipsis
	}
	return trimmed + ellipsis
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

func getItemsProp(props rtui.Props) []Item {
	if items, ok := props[propItems].([]Item); ok {
		return cloneItems(items)
	}
	return nil
}

func getDirectionProp(props rtui.Props) Direction {
	return getDirectionPropWithDefault(props, DirectionHorizontal)
}

func getDirectionPropWithDefault(props rtui.Props, def Direction) Direction {
	if direction, ok := props[propDirection].(Direction); ok {
		return direction
	}
	return def
}

func getFieldIntentProp(props rtui.Props, key string) intent.FieldIntent {
	if value, ok := props[key]; ok {
		if result, ok := value.(intent.FieldIntent); ok {
			return result
		}
	}
	if value, ok := props[key+"Field"]; ok {
		if result, ok := value.(intent.FieldIntent); ok {
			return result
		}
	}
	return nil
}

func truncateWithEllipsis(content string, width int) string {
	const ellipsis = "…"
	if width <= 0 {
		return ""
	}
	if paint.StringWidth(content) <= width {
		return content
	}
	ellipsisWidth := paint.StringWidth(ellipsis)
	if width <= ellipsisWidth {
		return ellipsis
	}
	trimmed := strings.TrimRight(truncateByDisplayWidth(content, width-ellipsisWidth), " ")
	if trimmed == "" {
		return ellipsis
	}
	return trimmed + ellipsis
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

func clampInt(v, minValue, maxValue int) int {
	if v < minValue {
		return minValue
	}
	if v > maxValue {
		return maxValue
	}
	return v
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
