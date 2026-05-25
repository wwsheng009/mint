package progress

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/animation"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/internal/proputil"
)

const (
	lineMinWidth          = 10
	circleVisualWidth     = 5
	circleVisualHeight    = 3
	dashboardVisualWidth  = 7
	dashboardVisualHeight = 2
	activeTickInterval    = 120 * time.Millisecond
	valueTickInterval     = time.Second / 60
	valueTweenDuration    = 180 * time.Millisecond
)

type gridPoint struct {
	row int
	col int
}

// =============================================================================
// Instance - Runtime Entity
// =============================================================================

// Instance is the runtime entity for Progress components.
type Instance struct {
	key            string
	label          string
	progressStyle  style.Style
	width          int
	value          int
	max            int
	indeterminate  bool
	progressType   Type
	status         Status
	showPercent    bool
	showValue      bool
	unit           string
	displayPercent float64
	percentTween   *animation.TweenDriver
	activeFrame    int
	activeLoop     *animation.LoopDriver
	bounds         [4]int
	dirty          bool
}

var (
	_ rtui.ComponentInstance = (*Instance)(nil)
	_ rtui.PaintableInstance = (*Instance)(nil)
	_ rtui.TickableInstance  = (*Instance)(nil)
	_ interface {
		Measure(layout.Constraints) layout.Size
	} = (*Instance)(nil)
)

// NewInstance creates a new ProgressInstance from props.
func NewInstance(props rtui.Props) *Instance {
	value, max := normalizeProgressRange(
		proputil.GetInt(props, propValue, 0),
		proputil.GetInt(props, propMax, 100),
	)

	inst := &Instance{
		key:            proputil.GetString(props, propKey, ""),
		label:          proputil.GetString(props, propLabel, ""),
		progressStyle:  proputil.GetStyle(props, propStyle, style.Style{}),
		width:          proputil.GetInt(props, propWidth, 30),
		value:          value,
		max:            max,
		indeterminate:  proputil.GetBool(props, propIndeterminate, false),
		progressType:   getTypeProp(props, TypeLine),
		status:         getStatusProp(props, StatusNormal),
		showPercent:    proputil.GetBool(props, propShowPercent, true),
		showValue:      proputil.GetBool(props, propShowValue, false),
		unit:           normalizeUnit(proputil.GetString(props, propUnit, "")),
		displayPercent: float64(progressPercent(value, max)),
		dirty:          true,
	}
	inst.resetActiveLoop()
	return inst
}

// =============================================================================
// ComponentInstance Interface
// =============================================================================

func (inst *Instance) Key() string       { return inst.key }
func (inst *Instance) SetKey(key string) { inst.key = key }
func (inst *Instance) Init(props rtui.Props) {
	inst.SetProps(props)
}
func (inst *Instance) Destroy()   {}
func (inst *Instance) OnMount()   {}
func (inst *Instance) OnUnmount() {}

func (inst *Instance) SetProps(props rtui.Props) bool {
	oldKey := inst.key
	oldLabel := inst.label
	oldStyle := inst.progressStyle
	oldWidth := inst.width
	oldValue := inst.value
	oldMax := inst.max
	oldIndeterminate := inst.indeterminate
	oldType := inst.progressType
	oldStatus := inst.status
	oldShowPercent := inst.showPercent
	oldShowValue := inst.showValue
	oldUnit := inst.unit
	currentPercent := inst.currentPercentFloat()

	value, max := normalizeProgressRange(
		proputil.GetInt(props, propValue, inst.value),
		proputil.GetInt(props, propMax, inst.max),
	)

	inst.key = proputil.GetString(props, propKey, inst.key)
	inst.label = proputil.GetString(props, propLabel, inst.label)
	inst.progressStyle = proputil.GetStyle(props, propStyle, inst.progressStyle)
	inst.width = proputil.GetInt(props, propWidth, inst.width)
	inst.value = value
	inst.max = max
	inst.indeterminate = proputil.GetBool(props, propIndeterminate, inst.indeterminate)
	inst.progressType = getTypeProp(props, inst.progressType)
	inst.status = getStatusProp(props, inst.status)
	inst.showPercent = proputil.GetBool(props, propShowPercent, inst.showPercent)
	inst.showValue = proputil.GetBool(props, propShowValue, inst.showValue)
	inst.unit = normalizeUnit(proputil.GetString(props, propUnit, inst.unit))

	if oldValue != inst.value || oldMax != inst.max {
		inst.startPercentTween(currentPercent, float64(inst.targetPercent()))
	}

	if oldType != inst.progressType || oldStatus != inst.status || oldIndeterminate != inst.indeterminate {
		inst.resetActiveLoop()
	} else {
		inst.syncActiveLoop(false)
	}

	changed := oldKey != inst.key ||
		oldLabel != inst.label ||
		oldStyle != inst.progressStyle ||
		oldWidth != inst.width ||
		oldValue != inst.value ||
		oldMax != inst.max ||
		oldIndeterminate != inst.indeterminate ||
		oldType != inst.progressType ||
		oldStatus != inst.status ||
		oldShowPercent != inst.showPercent ||
		oldShowValue != inst.showValue ||
		oldUnit != inst.unit
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		propKey:           inst.key,
		propLabel:         inst.label,
		propStyle:         inst.progressStyle,
		propWidth:         inst.width,
		propValue:         inst.value,
		propMax:           inst.max,
		propIndeterminate: inst.indeterminate,
		propType:          inst.progressType,
		propStatus:        inst.status,
		propShowPercent:   inst.showPercent,
		propShowValue:     inst.showValue,
		propUnit:          inst.unit,
	}
}

func (inst *Instance) MarkDirty() { inst.dirty = true }
func (inst *Instance) IsDirty() bool {
	return inst.dirty
}
func (inst *Instance) GetContext() *rtui.ComponentContext { return nil }

// =============================================================================
// TickableInstance Interface
// =============================================================================

func (inst *Instance) WantsTick() bool {
	return (inst.percentTween != nil && inst.percentTween.WantsTick()) || inst.wantsActiveLoop()
}

func (inst *Instance) Tick(now time.Time) bool {
	if !inst.WantsTick() {
		return false
	}

	changed := false

	if inst.percentTween != nil {
		if !inst.percentTween.Primed() {
			inst.percentTween.Prime(now.Add(-valueTickInterval))
		}
		if inst.percentTween.Tick(now) {
			nextPercent := normalizePercent(inst.percentTween.Value())
			if inst.displayPercent != nextPercent {
				inst.displayPercent = nextPercent
				changed = true
			}
		}
		if inst.percentTween.Done() {
			finalPercent := float64(inst.targetPercent())
			if inst.displayPercent != finalPercent {
				inst.displayPercent = finalPercent
				changed = true
			}
			inst.percentTween = nil
		}
	}

	inst.syncActiveLoop(false)
	if inst.activeLoop == nil {
		if changed {
			inst.dirty = true
		}
		return changed
	}

	if !inst.activeLoop.Primed() {
		inst.activeLoop.Prime(now.Add(-activeTickInterval))
	}
	prevFrame := inst.activeFrame
	if inst.activeLoop.Tick(now) {
		inst.activeFrame = inst.activeLoop.Cycle()
		if inst.activeFrame != prevFrame {
			changed = true
		}
	}

	if changed {
		inst.dirty = true
	}
	return changed
}

// =============================================================================
// PaintableInstance Interface
// =============================================================================

func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	visualStyle := inst.resolveStyle()
	rows := inst.visualRows()
	cmds := make([]paint.DrawCmd, 0, len(rows)+1)

	for rowIndex, row := range rows {
		cmds = append(cmds, paint.DrawCmd{
			X:     x,
			Y:     y + rowIndex,
			Text:  row,
			Style: visualStyle,
		})
	}

	if labelText := inst.labelText(); labelText != "" {
		cmds = append(cmds, paint.DrawCmd{
			X:     x,
			Y:     y + len(rows),
			Text:  labelText,
			Style: visualStyle,
		})
	}

	return cmds
}

func (inst *Instance) resolveStyle() style.Style {
	s := inst.progressStyle
	if s.FG == "" {
		s = s.Foreground(inst.statusColor())
	}
	if s.BG == "" {
		s = s.Background(theme.Surface())
	}
	if inst.status == StatusActive {
		s = s.Bold(true).Blink(true)
	}
	return s
}

func (inst *Instance) statusColor() style.Color {
	switch inst.status {
	case StatusSuccess:
		return theme.Success()
	case StatusException:
		return theme.Error()
	case StatusActive:
		return theme.Focus()
	case StatusWarning:
		return theme.Warning()
	default:
		return theme.Primary()
	}
}

func (inst *Instance) visualRows() []string {
	switch inst.progressType {
	case TypeBlock:
		return []string{inst.blockRow()}
	case TypeCircle:
		return inst.circleRows()
	case TypeDashboard:
		return inst.dashboardRows()
	default:
		return []string{inst.lineRow()}
	}
}

func (inst *Instance) lineRow() string {
	return inst.linearRow('=', '-', '>')
}

func (inst *Instance) blockRow() string {
	return inst.linearRow('█', '░', '▓')
}

func (inst *Instance) linearRow(filledRune, emptyRune, activeRune rune) string {
	width := inst.visualWidth()
	innerWidth := width - 2
	if innerWidth < 1 {
		innerWidth = 1
	}
	if inst.indeterminate {
		return inst.indeterminateLinearRow(innerWidth, emptyRune, activeRune)
	}

	percent := inst.Percent()
	filledCount := (percent * innerWidth) / 100
	if filledCount > innerWidth {
		filledCount = innerWidth
	}

	cells := make([]rune, innerWidth)
	for i := range cells {
		cells[i] = emptyRune
	}
	for i := 0; i < filledCount; i++ {
		cells[i] = filledRune
	}

	if inst.status == StatusActive && percent < 100 && filledCount > 0 {
		head := inst.activeFrame % filledCount
		cells[head] = activeRune
	}

	return "[" + string(cells) + "]"
}

func (inst *Instance) indeterminateLinearRow(innerWidth int, emptyRune, activeRune rune) string {
	cells := make([]rune, innerWidth)
	for i := range cells {
		cells[i] = emptyRune
	}
	window := innerWidth / 3
	if window < 1 {
		window = 1
	}
	if window > innerWidth {
		window = innerWidth
	}
	start := inst.activeFrame % innerWidth
	for i := 0; i < window; i++ {
		cells[(start+i)%innerWidth] = activeRune
	}
	return "[" + string(cells) + "]"
}

func (inst *Instance) circleRows() []string {
	grid := make([][]rune, circleVisualHeight)
	for row := range grid {
		grid[row] = []rune(strings.Repeat(" ", circleVisualWidth))
	}

	positions := []gridPoint{
		{0, 1}, {0, 2}, {0, 3}, {1, 4},
		{2, 3}, {2, 2}, {2, 1}, {1, 0},
	}
	inst.fillSegments(grid, positions)
	return inst.padRows(gridToStrings(grid))
}

func (inst *Instance) dashboardRows() []string {
	grid := make([][]rune, dashboardVisualHeight)
	for row := range grid {
		grid[row] = []rune(strings.Repeat(" ", dashboardVisualWidth))
	}

	positions := []gridPoint{
		{1, 0},
		{0, 1}, {0, 2}, {0, 3}, {0, 4}, {0, 5},
		{1, 6},
	}
	inst.fillSegments(grid, positions)
	return inst.padRows(gridToStrings(grid))
}

func (inst *Instance) fillSegments(grid [][]rune, positions []gridPoint) {
	if inst.indeterminate {
		inst.fillIndeterminateSegments(grid, positions)
		return
	}

	percent := inst.Percent()
	scaled := float64(percent) * float64(len(positions)) / 100
	trackRune := 'o'
	activeIndex := -1
	activeSpan := int(math.Ceil(scaled))
	if inst.status == StatusActive && percent < 100 && activeSpan > 0 {
		activeIndex = inst.activeFrame % activeSpan
	}

	for idx, pos := range positions {
		fill := scaled - float64(idx)
		if fill <= 0 {
			grid[pos.row][pos.col] = trackRune
			continue
		}

		glyph := segmentFillRune(fill)
		if idx == activeIndex {
			glyph = '@'
		}
		grid[pos.row][pos.col] = glyph
	}
}

func (inst *Instance) fillIndeterminateSegments(grid [][]rune, positions []gridPoint) {
	if len(positions) == 0 {
		return
	}
	for _, pos := range positions {
		grid[pos.row][pos.col] = 'o'
	}
	window := 2
	if window > len(positions) {
		window = len(positions)
	}
	start := inst.activeFrame % len(positions)
	for i := 0; i < window; i++ {
		pos := positions[(start+i)%len(positions)]
		grid[pos.row][pos.col] = '@'
	}
}

func (inst *Instance) padRows(rows []string) []string {
	width := inst.visualWidth()
	padded := make([]string, len(rows))
	for i, row := range rows {
		rowWidth := paint.StringWidth(row)
		if rowWidth < width {
			row += strings.Repeat(" ", width-rowWidth)
		}
		padded[i] = row
	}
	return padded
}

func gridToStrings(grid [][]rune) []string {
	rows := make([]string, len(grid))
	for i, row := range grid {
		rows[i] = string(row)
	}
	return rows
}

func segmentFillRune(fill float64) rune {
	switch {
	case fill >= 1:
		return '#'
	case fill >= 0.67:
		return '▓'
	case fill >= 0.34:
		return '▒'
	default:
		return '░'
	}
}

func (inst *Instance) labelText() string {
	percent := inst.Percent()
	switch {
	case inst.indeterminate && inst.label != "" && inst.showPercent:
		return fmt.Sprintf("%s: ...", inst.label)
	case inst.indeterminate && inst.showPercent:
		return "..."
	}

	detail := ""
	percentText := fmt.Sprintf("%d%%", percent)
	valueText := progressValueText(inst.value, inst.max, inst.unit)
	switch {
	case inst.showValue && inst.showPercent:
		detail = fmt.Sprintf("%s (%s)", valueText, percentText)
	case inst.showValue:
		detail = valueText
	case inst.showPercent:
		detail = percentText
	}

	switch {
	case inst.label != "" && detail != "":
		return fmt.Sprintf("%s: %s", inst.label, detail)
	case inst.label != "":
		return inst.label
	case detail != "":
		return detail
	default:
		return ""
	}
}

func progressValueText(value, max int, unit string) string {
	unit = normalizeUnit(unit)
	if unit == "" {
		return fmt.Sprintf("%d/%d", value, max)
	}
	if compactProgressUnit(unit) {
		return fmt.Sprintf("%d%s/%d%s", value, unit, max, unit)
	}
	return fmt.Sprintf("%d/%d %s", value, max, unit)
}

func compactProgressUnit(unit string) bool {
	switch strings.ToLower(unit) {
	case "%", "ms", "s", "m", "h", "b", "kb", "mb", "gb", "tb", "kib", "mib", "gib", "tib":
		return true
	default:
		return false
	}
}

func (inst *Instance) visualWidth() int {
	width := inst.width
	switch inst.progressType {
	case TypeCircle:
		if width < circleVisualWidth {
			width = circleVisualWidth
		}
	case TypeDashboard:
		if width < dashboardVisualWidth {
			width = dashboardVisualWidth
		}
	default:
		if width < lineMinWidth {
			width = lineMinWidth
		}
	}
	return width
}

func (inst *Instance) visualHeight() int {
	switch inst.progressType {
	case TypeCircle:
		return circleVisualHeight
	case TypeDashboard:
		return dashboardVisualHeight
	default:
		return 1
	}
}

// =============================================================================
// Progress-specific Methods
// =============================================================================

func (inst *Instance) SetValue(value int) {
	value, _ = normalizeProgressRange(value, inst.max)
	if inst.value != value {
		startPercent := inst.currentPercentFloat()
		inst.value = value
		inst.startPercentTween(startPercent, float64(inst.targetPercent()))
		inst.syncActiveLoop(false)
		inst.dirty = true
	}
}

func (inst *Instance) GetValue() int { return inst.value }
func (inst *Instance) GetMax() int   { return inst.max }

func (inst *Instance) Percent() int {
	return int(math.Round(inst.currentPercentFloat()))
}

func (inst *Instance) resetActiveLoop() {
	inst.activeFrame = 0
	if !inst.wantsActiveLoop() {
		inst.activeLoop = nil
		return
	}
	inst.activeLoop = animation.NewLoopDriver(animation.LoopDriverConfig{
		Duration:  activeTickInterval,
		Cycles:    0,
		AutoStart: true,
	})
}

func (inst *Instance) startPercentTween(from, to float64) {
	from = normalizePercent(from)
	to = normalizePercent(to)
	if from == to {
		inst.displayPercent = to
		inst.percentTween = nil
		return
	}

	inst.displayPercent = from
	inst.percentTween = animation.NewTweenDriver(animation.TweenDriverConfig{
		From:      from,
		To:        to,
		Duration:  valueTweenDuration,
		Easing:    animation.EaseOutCubic,
		AutoStart: true,
	})
}

func (inst *Instance) syncActiveLoop(reset bool) {
	if !inst.wantsActiveLoop() {
		if inst.activeLoop != nil {
			inst.activeLoop.Stop()
			inst.activeLoop = nil
		}
		inst.activeFrame = 0
		return
	}
	if reset || inst.activeLoop == nil {
		inst.resetActiveLoop()
	}
}

func (inst *Instance) wantsActiveLoop() bool {
	if inst.indeterminate {
		return true
	}
	percent := inst.Percent()
	return inst.status == StatusActive && percent > 0 && percent < 100
}

func (inst *Instance) currentPercentFloat() float64 {
	return normalizePercent(inst.displayPercent)
}

func (inst *Instance) targetPercent() int {
	return progressPercent(inst.value, inst.max)
}

// =============================================================================
// Measurable Interface
// =============================================================================

func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	if inst == nil {
		return layout.Size{}
	}

	width := inst.visualWidth()
	if labelText := inst.labelText(); labelText != "" {
		labelWidth := paint.StringWidth(labelText)
		if labelWidth > width {
			width = labelWidth
		}
	}

	height := inst.visualHeight()
	if inst.labelText() != "" {
		height++
	}

	width = constraints.ConstrainWidth(width)
	height = constraints.ConstrainHeight(height)

	return layout.Size{Width: width, Height: height}
}

// =============================================================================
// Prop Extraction Helpers
// =============================================================================

func getTypeProp(props rtui.Props, def Type) Type {
	if v, ok := props[propType]; ok {
		if t, ok := v.(Type); ok {
			return t
		}
	}
	return def
}

func getStatusProp(props rtui.Props, def Status) Status {
	if v, ok := props[propStatus]; ok {
		if status, ok := v.(Status); ok {
			return status
		}
	}
	return def
}

func normalizeProgressRange(value, max int) (int, int) {
	if max < 0 {
		max = 0
	}
	if value < 0 {
		value = 0
	}
	if max > 0 && value > max {
		value = max
	}
	return value, max
}

func progressPercent(value, max int) int {
	if max <= 0 {
		return 0
	}

	value, max = normalizeProgressRange(value, max)
	if max == 0 {
		return 0
	}
	return (value * 100) / max
}

func normalizePercent(percent float64) float64 {
	if percent < 0 {
		return 0
	}
	if percent > 100 {
		return 100
	}
	return percent
}
