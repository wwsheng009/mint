package progress

import (
	"fmt"
	"strings"
	"time"

	"github.com/wwsheng009/mint/framework/theme"
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
	key           string
	label         string
	progressStyle style.Style
	width         int
	value         int
	max           int
	progressType  Type
	status        Status
	showPercent   bool
	activeFrame   int
	lastTick      time.Time
	bounds        [4]int
	dirty         bool
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

	return &Instance{
		key:           proputil.GetString(props, propKey, ""),
		label:         proputil.GetString(props, propLabel, ""),
		progressStyle: proputil.GetStyle(props, propStyle, style.Style{}),
		width:         proputil.GetInt(props, propWidth, 30),
		value:         value,
		max:           max,
		progressType:  getTypeProp(props, TypeLine),
		status:        getStatusProp(props, StatusNormal),
		showPercent:   proputil.GetBool(props, propShowPercent, true),
		dirty:         true,
	}
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
	oldType := inst.progressType
	oldStatus := inst.status
	oldShowPercent := inst.showPercent

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
	inst.progressType = getTypeProp(props, inst.progressType)
	inst.status = getStatusProp(props, inst.status)
	inst.showPercent = proputil.GetBool(props, propShowPercent, inst.showPercent)

	if oldType != inst.progressType || oldStatus != inst.status || oldValue != inst.value || oldMax != inst.max {
		inst.activeFrame = 0
		inst.lastTick = time.Time{}
	}

	changed := oldKey != inst.key ||
		oldLabel != inst.label ||
		oldStyle != inst.progressStyle ||
		oldWidth != inst.width ||
		oldValue != inst.value ||
		oldMax != inst.max ||
		oldType != inst.progressType ||
		oldStatus != inst.status ||
		oldShowPercent != inst.showPercent
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		propKey:         inst.key,
		propLabel:       inst.label,
		propStyle:       inst.progressStyle,
		propWidth:       inst.width,
		propValue:       inst.value,
		propMax:         inst.max,
		propType:        inst.progressType,
		propStatus:      inst.status,
		propShowPercent: inst.showPercent,
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
	return inst.status == StatusActive && inst.value > 0 && inst.value < inst.max
}

func (inst *Instance) Tick(now time.Time) bool {
	if !inst.WantsTick() {
		return false
	}
	if !inst.lastTick.IsZero() && now.Sub(inst.lastTick) < activeTickInterval {
		return false
	}

	inst.lastTick = now
	inst.activeFrame++
	inst.dirty = true
	return true
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
	default:
		return theme.Primary()
	}
}

func (inst *Instance) visualRows() []string {
	switch inst.progressType {
	case TypeCircle:
		return inst.circleRows()
	case TypeDashboard:
		return inst.dashboardRows()
	default:
		return []string{inst.lineRow()}
	}
}

func (inst *Instance) lineRow() string {
	width := inst.visualWidth()
	innerWidth := width - 2
	if innerWidth < 1 {
		innerWidth = 1
	}

	percent := inst.Percent()
	filledCount := (percent * innerWidth) / 100
	if filledCount > innerWidth {
		filledCount = innerWidth
	}

	cells := make([]rune, innerWidth)
	for i := range cells {
		cells[i] = '-'
	}
	for i := 0; i < filledCount; i++ {
		cells[i] = '='
	}

	if inst.status == StatusActive && percent < 100 && filledCount > 0 {
		head := inst.activeFrame % filledCount
		cells[head] = '>'
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
	filledCount := segmentFillCount(inst.Percent(), len(positions))
	trackRune := 'o'
	activeIndex := -1
	if inst.status == StatusActive && filledCount > 0 {
		activeIndex = inst.activeFrame % filledCount
	}

	for idx, pos := range positions {
		if idx < filledCount {
			grid[pos.row][pos.col] = '#'
			if idx == activeIndex {
				grid[pos.row][pos.col] = '*'
			}
		} else {
			grid[pos.row][pos.col] = trackRune
		}
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

func segmentFillCount(percent, total int) int {
	if total <= 0 || percent <= 0 {
		return 0
	}
	count := (percent*total + 99) / 100
	if count > total {
		return total
	}
	return count
}

func (inst *Instance) labelText() string {
	percent := inst.Percent()
	switch {
	case inst.label != "" && inst.showPercent:
		return fmt.Sprintf("%s: %d%%", inst.label, percent)
	case inst.label != "":
		return inst.label
	case inst.showPercent:
		return fmt.Sprintf("%d%%", percent)
	default:
		return ""
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
		inst.value = value
		inst.dirty = true
	}
}

func (inst *Instance) GetValue() int { return inst.value }
func (inst *Instance) GetMax() int   { return inst.max }

func (inst *Instance) Percent() int {
	if inst.max <= 0 {
		return 0
	}

	value, max := normalizeProgressRange(inst.value, inst.max)
	if max == 0 {
		return 0
	}
	return (value * 100) / max
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
