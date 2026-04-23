package bulletchart

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	chartcanvas "github.com/wwsheng009/mint/ui/components/charts/internal/canvas"
	"github.com/wwsheng009/mint/ui/components/charts/internal/palette"
	"github.com/wwsheng009/mint/ui/components/internal/proputil"
)

const minVisualWidth = 10

// Instance is the runtime entity for bullet chart components.
type Instance struct {
	key               string
	label             string
	value             int
	target            int
	max               int
	width             int
	showTarget        bool
	showValueText     bool
	valueLabelMode    ValueLabelMode
	direction         Direction
	qualitativeRanges []QualitativeRange
	targetMarkerRune  rune
	targetMarkerStyle style.Style
	chartStyle        style.Style
	dirty             bool
}

var (
	_ rtui.ComponentInstance = (*Instance)(nil)
	_ rtui.PaintableInstance = (*Instance)(nil)
	_ interface {
		Measure(layout.Constraints) layout.Size
	} = (*Instance)(nil)
)

// NewInstance creates a new bullet chart instance.
func NewInstance(props rtui.Props) *Instance {
	maxValue := proputil.GetInt(props, propMax, 100)
	if maxValue <= 0 {
		maxValue = 100
	}
	return &Instance{
		key:               proputil.GetString(props, propKey, ""),
		label:             proputil.GetString(props, propLabel, ""),
		value:             clampInt(proputil.GetInt(props, propValue, 0), 0, maxValue),
		target:            clampInt(proputil.GetInt(props, propTarget, 0), 0, maxValue),
		max:               maxValue,
		width:             proputil.GetInt(props, propWidth, 20),
		showTarget:        proputil.GetBool(props, propShowTarget, true),
		showValueText:     proputil.GetBool(props, propShowValueText, true),
		valueLabelMode:    getValueLabelMode(props, ValueLabelModeAuto),
		direction:         getDirection(props, DirectionNeutral),
		qualitativeRanges: getQualitativeRanges(props, nil, maxValue),
		targetMarkerRune:  getTargetMarkerRune(props, '│'),
		targetMarkerStyle: proputil.GetStyle(props, propTargetMarkerStyle, style.Style{}),
		chartStyle:        proputil.GetStyle(props, propStyle, style.Style{}),
		dirty:             true,
	}
}

func (inst *Instance) Key() string           { return inst.key }
func (inst *Instance) SetKey(key string)     { inst.key = key }
func (inst *Instance) Init(props rtui.Props) { inst.SetProps(props) }
func (inst *Instance) Destroy()              {}
func (inst *Instance) OnMount()              {}
func (inst *Instance) OnUnmount()            {}

func (inst *Instance) SetProps(props rtui.Props) bool {
	old := *inst

	maxValue := proputil.GetInt(props, propMax, inst.max)
	if maxValue <= 0 {
		maxValue = 100
	}
	inst.key = proputil.GetString(props, propKey, inst.key)
	inst.label = proputil.GetString(props, propLabel, inst.label)
	inst.max = maxValue
	inst.value = clampInt(proputil.GetInt(props, propValue, inst.value), 0, maxValue)
	inst.target = clampInt(proputil.GetInt(props, propTarget, inst.target), 0, maxValue)
	inst.width = proputil.GetInt(props, propWidth, inst.width)
	inst.showTarget = proputil.GetBool(props, propShowTarget, inst.showTarget)
	inst.showValueText = proputil.GetBool(props, propShowValueText, inst.showValueText)
	inst.valueLabelMode = getValueLabelMode(props, inst.valueLabelMode)
	inst.direction = getDirection(props, inst.direction)
	inst.qualitativeRanges = getQualitativeRanges(props, inst.qualitativeRanges, maxValue)
	inst.targetMarkerRune = getTargetMarkerRune(props, inst.targetMarkerRune)
	inst.targetMarkerStyle = proputil.GetStyle(props, propTargetMarkerStyle, inst.targetMarkerStyle)
	inst.chartStyle = proputil.GetStyle(props, propStyle, inst.chartStyle)

	changed := old.key != inst.key ||
		old.label != inst.label ||
		old.value != inst.value ||
		old.target != inst.target ||
		old.max != inst.max ||
		old.width != inst.width ||
		old.showTarget != inst.showTarget ||
		old.showValueText != inst.showValueText ||
		old.valueLabelMode != inst.valueLabelMode ||
		old.direction != inst.direction ||
		old.targetMarkerRune != inst.targetMarkerRune ||
		old.targetMarkerStyle != inst.targetMarkerStyle ||
		!qualitativeRangesEqual(old.qualitativeRanges, inst.qualitativeRanges, inst.max) ||
		old.chartStyle != inst.chartStyle
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		propKey:               inst.key,
		propLabel:             inst.label,
		propValue:             inst.value,
		propTarget:            inst.target,
		propMax:               inst.max,
		propWidth:             inst.width,
		propShowTarget:        inst.showTarget,
		propShowValueText:     inst.showValueText,
		propValueLabelMode:    inst.valueLabelMode,
		propDirection:         inst.direction,
		propQualitativeRanges: copyQualitativeRangeSlice(inst.qualitativeRanges, inst.max),
		propTargetMarkerRune:  inst.targetMarkerRune,
		propTargetMarkerStyle: inst.targetMarkerStyle,
		propStyle:             inst.chartStyle,
	}
}

func (inst *Instance) MarkDirty()                         { inst.dirty = true }
func (inst *Instance) IsDirty() bool                      { return inst.dirty }
func (inst *Instance) GetContext() *rtui.ComponentContext { return nil }

func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	width := inst.chartWidth()
	labelWidth := paint.StringWidth(inst.labelText())
	if labelWidth > width {
		width = labelWidth
	}
	height := 1
	if inst.labelText() != "" {
		height++
	}
	return layout.Size{
		Width:  constraints.ConstrainWidth(width),
		Height: constraints.ConstrainHeight(height),
	}
}

func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	cmds := chartcanvas.BufferToDrawCmds(inst.chartBuffer(), x, y)
	if label := inst.labelText(); label != "" {
		cmds = append(cmds, paint.DrawCmd{
			X:     x,
			Y:     y + 1,
			Text:  label,
			Style: style.NewStyle().Foreground(palette.LabelColor()),
		})
	}
	return cmds
}

func (inst *Instance) resolveChartStyle() style.Style {
	s := inst.chartStyle
	if s.FG == "" {
		s = s.Foreground(palette.SeriesColor(0))
	}
	return s
}

func (inst *Instance) chartRow() string {
	width := inst.width
	if width < minVisualWidth {
		width = minVisualWidth
	}
	innerWidth := width - 2
	if innerWidth < 1 {
		innerWidth = 1
	}

	cells := make([]rune, innerWidth)
	inst.fillQualitativeRanges(cells)
	inst.fillValue(cells)
	inst.placeTarget(cells)

	row := "[" + string(cells) + "]"
	if inst.showValueText && inst.shouldInlineValueLabel() {
		row += " " + inst.valueSummary()
	}
	return row
}

func (inst *Instance) labelText() string {
	label := strings.TrimSpace(inst.label)
	if !inst.showValueText {
		return label
	}

	valueText := inst.valueSummary()
	if inst.shouldInlineValueLabel() {
		if label == "" {
			return ""
		}
		return label
	}

	if label == "" {
		return valueText
	}
	return fmt.Sprintf("%s: %s", label, valueText)
}

func (inst *Instance) chartWidth() int {
	return paint.StringWidth(inst.chartRow())
}

func (inst *Instance) chartBuffer() *paint.Buffer {
	row := inst.chartRow()
	buf := paint.NewBuffer(paint.StringWidth(row), 1)
	for x := 0; x < buf.Width; x++ {
		buf.SetCell(x, 0, ' ', style.Style{})
	}

	cellsWidth := inst.visualWidth()
	buf.SetCell(0, 0, '[', style.NewStyle().Foreground(palette.LabelColor()))
	buf.SetCell(cellsWidth+1, 0, ']', style.NewStyle().Foreground(palette.LabelColor()))

	cells := inst.chartCells()
	for i, cell := range cells {
		buf.SetCell(i+1, 0, cell.r, cell.s)
	}

	if inst.showValueText && inst.shouldInlineValueLabel() {
		summary := inst.valueSummary()
		if summary != "" {
			buf.SetString(cellsWidth+3, 0, summary, style.NewStyle().Foreground(palette.LabelColor()))
		}
	}
	return buf
}

type chartCell struct {
	r rune
	s style.Style
}

func (inst *Instance) chartCells() []chartCell {
	innerWidth := inst.visualWidth()
	cells := make([]chartCell, innerWidth)
	ranges := normalizeQualitativeRanges(inst.qualitativeRanges, inst.max)
	for i := range cells {
		cells[i] = chartCell{r: '░', s: inst.resolveRangeStyle(0, len(ranges), style.Style{})}
	}

	inst.fillQualitativeCellRanges(cells, ranges)
	inst.fillValueCells(cells)
	inst.placeTargetCell(cells)
	return cells
}

func (inst *Instance) visualWidth() int {
	width := inst.width
	if width < minVisualWidth {
		width = minVisualWidth
	}
	innerWidth := width - 2
	if innerWidth < 1 {
		return 1
	}
	return innerWidth
}

func (inst *Instance) fillQualitativeRanges(cells []rune) {
	ranges := normalizeQualitativeRanges(inst.qualitativeRanges, inst.max)
	last := 0
	for i, qr := range ranges {
		end := inst.limitToCellCount(qr.Limit, len(cells))
		if i == len(ranges)-1 {
			end = len(cells)
		}
		if end < last {
			end = last
		}
		for j := last; j < end && j < len(cells); j++ {
			cells[j] = qr.Glyph
		}
		last = end
	}
}

func (inst *Instance) fillQualitativeCellRanges(cells []chartCell, ranges []QualitativeRange) {
	last := 0
	for i, qr := range ranges {
		end := inst.limitToCellCount(qr.Limit, len(cells))
		if i == len(ranges)-1 {
			end = len(cells)
		}
		if end < last {
			end = last
		}
		for j := last; j < end && j < len(cells); j++ {
			cells[j].r = qr.Glyph
			cells[j].s = inst.resolveRangeStyle(i, len(ranges), qr.Style)
		}
		last = end
	}
}

func (inst *Instance) fillValue(cells []rune) {
	filled := inst.limitToCellCount(inst.value, len(cells))
	for i := 0; i < filled; i++ {
		cells[i] = '█'
	}
}

func (inst *Instance) fillValueCells(cells []chartCell) {
	filled := inst.limitToCellCount(inst.value, len(cells))
	valueStyle := inst.resolveChartStyle()
	for i := 0; i < filled; i++ {
		cells[i].r = '█'
		cells[i].s = valueStyle
	}
}

func (inst *Instance) placeTarget(cells []rune) {
	if !inst.showTarget || inst.max <= 0 || len(cells) == 0 {
		return
	}
	cells[inst.targetIndex(len(cells))] = inst.resolvedTargetMarkerRune()
}

func (inst *Instance) placeTargetCell(cells []chartCell) {
	if !inst.showTarget || inst.max <= 0 || len(cells) == 0 {
		return
	}
	targetIndex := inst.targetIndex(len(cells))
	cells[targetIndex].r = inst.resolvedTargetMarkerRune()
	cells[targetIndex].s = inst.resolveTargetMarkerStyle()
}

func (inst *Instance) resolveRangeStyle(index, total int, extra style.Style) style.Style {
	base := inst.defaultQualitativeRangeStyle(index, total)
	return base.
		Merge(inheritNonForegroundStyle(inst.chartStyle)).
		Merge(extra)
}

func (inst *Instance) resolveTargetMarkerStyle() style.Style {
	base := style.NewStyle().
		Foreground(inst.defaultTargetMarkerColor()).
		Bold(true)
	return base.
		Merge(inheritNonForegroundStyle(inst.chartStyle)).
		Merge(inst.targetMarkerStyle)
}

func (inst *Instance) targetIndex(innerWidth int) int {
	targetIndex := (inst.target * (innerWidth - 1)) / inst.max
	if targetIndex < 0 {
		return 0
	}
	if targetIndex >= innerWidth {
		return innerWidth - 1
	}
	return targetIndex
}

func (inst *Instance) resolvedTargetMarkerRune() rune {
	if inst.targetMarkerRune == 0 {
		return '│'
	}
	return inst.targetMarkerRune
}

func (inst *Instance) limitToCellCount(limit, width int) int {
	if inst.max <= 0 || width <= 0 {
		return 0
	}
	count := (limit * width) / inst.max
	if count < 0 {
		return 0
	}
	if count > width {
		return width
	}
	return count
}

func (inst *Instance) shouldInlineValueLabel() bool {
	if !inst.showValueText {
		return false
	}
	switch inst.valueLabelMode {
	case ValueLabelModeInline:
		return true
	case ValueLabelModeBelow:
		return false
	default:
		return inst.shouldAutoInlineValueLabel()
	}
}

func (inst *Instance) shouldAutoInlineValueLabel() bool {
	summaryWidth := paint.StringWidth(inst.valueSummary())
	if summaryWidth == 0 {
		return false
	}

	chartSlots := inst.visualWidth()
	if chartSlots < 12 {
		return false
	}
	if summaryWidth > chartSlots {
		return false
	}

	label := strings.TrimSpace(inst.label)
	if label != "" && summaryWidth > maxInt(8, chartSlots/2) {
		return false
	}
	return true
}

func (inst *Instance) defaultQualitativeRangeStyle(index, total int) style.Style {
	if total <= 1 {
		return style.NewStyle().Foreground(palette.ReferenceBandColor())
	}

	color := palette.LabelColor()
	switch inst.direction {
	case DirectionHigherBetter:
		switch {
		case index >= total-1:
			color = palette.UpColor()
		case index > 0:
			color = palette.ReferenceLineColor()
		default:
			color = palette.DownColor()
		}
	case DirectionLowerBetter:
		switch {
		case index >= total-1:
			color = palette.DownColor()
		case index > 0:
			color = palette.ReferenceLineColor()
		default:
			color = palette.UpColor()
		}
	default:
		switch {
		case index >= total-1:
			color = palette.TitleColor()
		case index > 0:
			color = palette.ReferenceBandColor()
		default:
			color = palette.LabelColor()
		}
	}
	return style.NewStyle().Foreground(color)
}

func (inst *Instance) defaultTargetMarkerColor() style.Color {
	switch inst.direction {
	case DirectionHigherBetter:
		return palette.UpColor()
	case DirectionLowerBetter:
		return palette.DownColor()
	default:
		return palette.ReferenceLineColor()
	}
}

func (inst *Instance) valueSummary() string {
	parts := []string{fmt.Sprintf("%d/%d", inst.value, inst.max)}
	if inst.showTarget {
		parts = append(parts, fmt.Sprintf("target %d", inst.target))
	}
	return strings.Join(parts, " ")
}

func clampInt(value, minVal, maxVal int) int {
	if value < minVal {
		return minVal
	}
	if value > maxVal {
		return maxVal
	}
	return value
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func inheritNonForegroundStyle(src style.Style) style.Style {
	dst := style.NewStyle()
	if src.BG != "" {
		dst = dst.Background(src.BG)
	}
	if src.IsBold() {
		dst = dst.Bold(true)
	}
	if src.IsItalic() {
		dst = dst.Italic(true)
	}
	if src.IsUnderline() {
		dst = dst.Underline(true)
	}
	if src.IsStrikethrough() {
		dst = dst.Strikethrough(true)
	}
	if src.IsReverse() {
		dst = dst.Reverse(true)
	}
	if src.IsBlink() {
		dst = dst.Blink(true)
	}
	dst.Width = src.Width
	dst.Height = src.Height
	return dst
}

func getValueLabelMode(props rtui.Props, def ValueLabelMode) ValueLabelMode {
	if value, ok := props[propValueLabelMode]; ok {
		if mode, ok := value.(ValueLabelMode); ok {
			return mode
		}
	}
	return def
}

func getQualitativeRanges(props rtui.Props, def []QualitativeRange, max int) []QualitativeRange {
	if value, ok := props[propQualitativeRanges]; ok {
		if ranges, ok := value.([]QualitativeRange); ok {
			return copyQualitativeRangeSlice(ranges, max)
		}
	}
	return copyQualitativeRangeSlice(def, max)
}

func getTargetMarkerRune(props rtui.Props, def rune) rune {
	if value, ok := props[propTargetMarkerRune]; ok {
		if marker, ok := value.(rune); ok {
			if marker != 0 {
				return marker
			}
		}
	}
	if def == 0 {
		return '│'
	}
	return def
}

func getDirection(props rtui.Props, def Direction) Direction {
	if value, ok := props[propDirection]; ok {
		if direction, ok := value.(Direction); ok {
			return direction
		}
	}
	return def
}
