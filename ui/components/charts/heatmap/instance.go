package heatmap

import (
	"fmt"
	"math"
	"strings"
	"unicode/utf8"

	fwtheme "github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	chartcanvas "github.com/wwsheng009/mint/ui/components/charts/internal/canvas"
	"github.com/wwsheng009/mint/ui/components/charts/internal/palette"
	"github.com/wwsheng009/mint/ui/components/internal/proputil"
)

var heatmapGlyphs = []rune{'░', '▒', '▓', '█'}

const autoViewportScaleAreaThreshold = 0.75

// Instance is the runtime entity for heatmap components.
type Instance struct {
	key              string
	title            string
	rowLabels        []string
	colLabels        []string
	values           [][]float64
	showAxis         bool
	showLegend       bool
	summaryMode      SummaryMode
	legendMode       LegendMode
	colorMode        fwtheme.ColorMode
	scaleMode        ScaleMode
	viewport         Viewport
	maxRowLabelWidth int
	heatmapStyle     style.Style
	dirty            bool
}

var (
	_ rtui.ComponentInstance = (*Instance)(nil)
	_ rtui.PaintableInstance = (*Instance)(nil)
	_ interface {
		Measure(layout.Constraints) layout.Size
	} = (*Instance)(nil)
)

// NewInstance creates a new heatmap instance.
func NewInstance(props rtui.Props) *Instance {
	return &Instance{
		key:              proputil.GetString(props, propKey, ""),
		title:            proputil.GetString(props, propTitle, ""),
		rowLabels:        getStringSliceProp(props, propRowLabels, nil),
		colLabels:        getStringSliceProp(props, propColLabels, nil),
		values:           getMatrixProp(props, propValues, nil),
		showAxis:         proputil.GetBool(props, propShowAxis, true),
		showLegend:       proputil.GetBool(props, propShowLegend, true),
		summaryMode:      getSummaryModeProp(props, SummaryModeNone),
		legendMode:       getLegendModeProp(props, LegendModeFull),
		colorMode:        getColorModeProp(props, fwtheme.NewTerminalColorCapabilities().GetMode()),
		scaleMode:        getScaleModeProp(props, ScaleModeGlobal),
		viewport:         getViewportProp(props, Viewport{}),
		maxRowLabelWidth: proputil.GetInt(props, propMaxRowLabelWidth, 0),
		heatmapStyle:     proputil.GetStyle(props, propStyle, style.Style{}),
		dirty:            true,
	}
}

func (inst *Instance) Key() string           { return inst.key }
func (inst *Instance) SetKey(key string)     { inst.key = key }
func (inst *Instance) Init(props rtui.Props) { inst.SetProps(props) }
func (inst *Instance) Destroy()              {}
func (inst *Instance) OnMount()              {}
func (inst *Instance) OnUnmount()            {}

func (inst *Instance) SetProps(props rtui.Props) bool {
	oldTitle := inst.title
	oldRowLabels := inst.rowLabels
	oldColLabels := inst.colLabels
	oldValues := inst.values
	oldShowAxis := inst.showAxis
	oldShowLegend := inst.showLegend
	oldSummaryMode := inst.summaryMode
	oldLegendMode := inst.legendMode
	oldColorMode := inst.colorMode
	oldScaleMode := inst.scaleMode
	oldViewport := inst.viewport
	oldMaxRowLabelWidth := inst.maxRowLabelWidth
	oldStyle := inst.heatmapStyle

	inst.key = proputil.GetString(props, propKey, inst.key)
	inst.title = proputil.GetString(props, propTitle, inst.title)
	inst.rowLabels = getStringSliceProp(props, propRowLabels, inst.rowLabels)
	inst.colLabels = getStringSliceProp(props, propColLabels, inst.colLabels)
	inst.values = getMatrixProp(props, propValues, inst.values)
	inst.showAxis = proputil.GetBool(props, propShowAxis, inst.showAxis)
	inst.showLegend = proputil.GetBool(props, propShowLegend, inst.showLegend)
	inst.summaryMode = getSummaryModeProp(props, inst.summaryMode)
	inst.legendMode = getLegendModeProp(props, inst.legendMode)
	inst.colorMode = getColorModeProp(props, inst.colorMode)
	inst.scaleMode = getScaleModeProp(props, inst.scaleMode)
	inst.viewport = getViewportProp(props, inst.viewport)
	inst.maxRowLabelWidth = proputil.GetInt(props, propMaxRowLabelWidth, inst.maxRowLabelWidth)
	inst.heatmapStyle = proputil.GetStyle(props, propStyle, inst.heatmapStyle)

	changed := oldTitle != inst.title ||
		!stringSlicesEqual(oldRowLabels, inst.rowLabels) ||
		!stringSlicesEqual(oldColLabels, inst.colLabels) ||
		!float64MatrixEqual(oldValues, inst.values) ||
		oldShowAxis != inst.showAxis ||
		oldShowLegend != inst.showLegend ||
		oldSummaryMode != inst.summaryMode ||
		oldLegendMode != inst.legendMode ||
		oldColorMode != inst.colorMode ||
		oldScaleMode != inst.scaleMode ||
		oldViewport != inst.viewport ||
		oldMaxRowLabelWidth != inst.maxRowLabelWidth ||
		oldStyle != inst.heatmapStyle
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		propKey:              inst.key,
		propTitle:            inst.title,
		propRowLabels:        copyStringSlice(inst.rowLabels),
		propColLabels:        copyStringSlice(inst.colLabels),
		propValues:           copyFloat64Matrix(inst.values),
		propShowAxis:         inst.showAxis,
		propShowLegend:       inst.showLegend,
		propShowSummary:      inst.summaryMode != SummaryModeNone,
		propSummaryMode:      inst.summaryMode,
		propLegendMode:       inst.legendMode,
		propColorMode:        inst.colorMode,
		propScaleMode:        inst.scaleMode,
		propViewport:         inst.viewport,
		propMaxRowLabelWidth: inst.maxRowLabelWidth,
		propStyle:            inst.heatmapStyle,
	}
}

func (inst *Instance) MarkDirty()                         { inst.dirty = true }
func (inst *Instance) IsDirty() bool                      { return inst.dirty }
func (inst *Instance) GetContext() *rtui.ComponentContext { return nil }

func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	width, height := inst.contentSize()
	return layout.Size{
		Width:  constraints.ConstrainWidth(width),
		Height: constraints.ConstrainHeight(height),
	}
}

func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	buf := inst.renderBuffer()
	return chartcanvas.BufferToDrawCmds(buf, x, y)
}

func (inst *Instance) contentSize() (int, int) {
	width := 0
	height := 0

	if strings.TrimSpace(inst.title) != "" {
		width = max(width, paint.StringWidth(inst.title))
		height++
	}
	if inst.showLegend {
		width = max(width, paint.StringWidth(inst.legendRow()))
		height++
	}
	if summaryRow := inst.summaryRow(); summaryRow != "" {
		width = max(width, paint.StringWidth(summaryRow))
		height++
	}
	if axisRow := inst.columnLabelRow(inst.rowLabelWidth()); axisRow != "" {
		width = max(width, paint.StringWidth(axisRow))
		height++
	}

	rows := inst.plotRows()
	if len(rows) == 0 {
		width = max(width, paint.StringWidth("No data"))
		height++
		return width, height
	}

	gutter := inst.rowLabelWidth()
	for i, row := range rows {
		rowWidth := paint.StringWidth(row)
		if gutter > 0 {
			rowWidth += gutter + 1
			if labelWidth := paint.StringWidth(inst.rowLabel(i)); labelWidth > gutter {
				rowWidth = labelWidth + 1 + paint.StringWidth(row)
			}
		}
		width = max(width, rowWidth)
		height++
	}
	return width, height
}

func (inst *Instance) rowCount() int {
	start, end := inst.visibleRowRange()
	return end - start
}

func (inst *Instance) columnCount() int {
	start, end := inst.visibleColumnRange()
	return end - start
}

func (inst *Instance) totalColumnCount() int {
	count := len(inst.colLabels)
	for _, row := range inst.values {
		if len(row) > count {
			count = len(row)
		}
	}
	return count
}

func (inst *Instance) rowLabelWidth() int {
	if !inst.showAxis {
		return 0
	}
	width := 0
	for i := 0; i < inst.rowCount(); i++ {
		width = max(width, paint.StringWidth(inst.rowLabel(i)))
	}
	if inst.maxRowLabelWidth > 0 && width > inst.maxRowLabelWidth {
		return inst.maxRowLabelWidth
	}
	return width
}

func (inst *Instance) rowLabel(index int) string {
	if !inst.showAxis {
		return ""
	}
	sourceIndex := inst.visibleRowStart() + index
	if sourceIndex >= 0 && sourceIndex < len(inst.rowLabels) && inst.rowLabels[sourceIndex] != "" {
		return clipLabel(inst.rowLabels[sourceIndex], inst.maxRowLabelWidth)
	}
	return clipLabel(fmt.Sprintf("R%d", sourceIndex+1), inst.maxRowLabelWidth)
}

func (inst *Instance) columnLabelRow(gutter int) string {
	if !inst.showAxis || inst.columnCount() == 0 {
		return ""
	}

	colStart := inst.visibleColStart()
	cells := make([]string, inst.columnCount())
	for i := range cells {
		sourceIndex := colStart + i
		label := ""
		if sourceIndex < len(inst.colLabels) {
			label = inst.colLabels[sourceIndex]
		}
		if label == "" {
			label = fmt.Sprintf("%d", sourceIndex+1)
		}
		r, _ := utf8.DecodeRuneInString(label)
		if r == utf8.RuneError {
			r = '?'
		}
		cells[i] = string(r)
	}

	row := strings.Join(cells, " ")
	if gutter > 0 {
		return strings.Repeat(" ", gutter) + " " + row
	}
	return row
}

func (inst *Instance) legendRow() string {
	switch inst.legendMode {
	case LegendModeCompact:
		return "L ░▒▓█ H"
	default:
		return "Low ░ ▒ ▓ █ High"
	}
}

func (inst *Instance) legendRatios() []float64 {
	return []float64{0.00, 0.33, 0.66, 1.00}
}

func (inst *Instance) summaryRow() string {
	if inst.summaryMode == SummaryModeNone {
		return ""
	}
	stats, ok := inst.visibleStats()
	if !ok {
		return ""
	}
	switch inst.summaryMode {
	case SummaryModeCompact:
		return fmt.Sprintf(
			"%s..%s avg %s",
			formatHeatmapNumber(stats.Min),
			formatHeatmapNumber(stats.Max),
			formatHeatmapNumber(stats.Avg),
		)
	default:
		return fmt.Sprintf(
			"range %s..%s avg %s",
			formatHeatmapNumber(stats.Min),
			formatHeatmapNumber(stats.Max),
			formatHeatmapNumber(stats.Avg),
		)
	}
}

func (inst *Instance) plotRows() []string {
	if len(inst.values) == 0 || inst.columnCount() == 0 {
		return nil
	}

	minVal, maxVal, ok := inst.resolvedMinMax()
	if !ok {
		return nil
	}

	rowStart, rowEnd := inst.visibleRowRange()
	colStart, colEnd := inst.visibleColumnRange()
	rows := make([]string, 0, rowEnd-rowStart)
	for rowIndex := rowStart; rowIndex < rowEnd; rowIndex++ {
		values := inst.values[rowIndex]
		cells := make([]string, 0, inst.columnCount())
		for col := colStart; col < colEnd; col++ {
			if col >= len(values) {
				cells = append(cells, " ")
				continue
			}
			cells = append(cells, string(inst.valueGlyph(values[col], minVal, maxVal)))
		}
		rows = append(rows, strings.Join(cells, " "))
	}
	return rows
}

func (inst *Instance) globalMinMax() (float64, float64, bool) {
	var (
		minVal float64
		maxVal float64
		ok     bool
	)
	for _, row := range inst.values {
		for _, value := range row {
			if !ok {
				minVal = value
				maxVal = value
				ok = true
				continue
			}
			if value < minVal {
				minVal = value
			}
			if value > maxVal {
				maxVal = value
			}
		}
	}
	return minVal, maxVal, ok
}

func (inst *Instance) viewportMinMax() (float64, float64, bool) {
	var (
		minVal float64
		maxVal float64
		ok     bool
	)
	rowStart, rowEnd := inst.visibleRowRange()
	colStart, colEnd := inst.visibleColumnRange()
	for rowIndex := rowStart; rowIndex < rowEnd; rowIndex++ {
		row := inst.values[rowIndex]
		for colIndex := colStart; colIndex < colEnd; colIndex++ {
			if colIndex >= len(row) {
				continue
			}
			value := row[colIndex]
			if !ok {
				minVal = value
				maxVal = value
				ok = true
				continue
			}
			if value < minVal {
				minVal = value
			}
			if value > maxVal {
				maxVal = value
			}
		}
	}
	return minVal, maxVal, ok
}

func (inst *Instance) resolvedMinMax() (float64, float64, bool) {
	if inst.shouldUseViewportScale() {
		return inst.viewportMinMax()
	}
	return inst.globalMinMax()
}

type windowStats struct {
	Min   float64
	Max   float64
	Avg   float64
	Count int
}

func (inst *Instance) visibleStats() (windowStats, bool) {
	var (
		stats windowStats
		sum   float64
		ok    bool
	)
	rowStart, rowEnd := inst.visibleRowRange()
	colStart, colEnd := inst.visibleColumnRange()
	for rowIndex := rowStart; rowIndex < rowEnd; rowIndex++ {
		row := inst.values[rowIndex]
		for colIndex := colStart; colIndex < colEnd; colIndex++ {
			if colIndex >= len(row) {
				continue
			}
			value := row[colIndex]
			if !ok {
				stats.Min = value
				stats.Max = value
				ok = true
			} else {
				if value < stats.Min {
					stats.Min = value
				}
				if value > stats.Max {
					stats.Max = value
				}
			}
			sum += value
			stats.Count++
		}
	}
	if !ok || stats.Count == 0 {
		return windowStats{}, false
	}
	stats.Avg = sum / float64(stats.Count)
	return stats, true
}

func (inst *Instance) shouldUseViewportScale() bool {
	switch inst.scaleMode {
	case ScaleModeViewport:
		return true
	case ScaleModeAuto:
		return inst.shouldAutoUseViewportScale()
	default:
		return false
	}
}

func (inst *Instance) hasPartialViewport() bool {
	rowStart, rowEnd := inst.visibleRowRange()
	colStart, colEnd := inst.visibleColumnRange()
	return rowStart > 0 ||
		colStart > 0 ||
		rowEnd < len(inst.values) ||
		colEnd < inst.totalColumnCount()
}

func (inst *Instance) shouldAutoUseViewportScale() bool {
	if !inst.hasPartialViewport() {
		return false
	}

	totalRows := len(inst.values)
	totalCols := inst.totalColumnCount()
	if totalRows <= 0 || totalCols <= 0 {
		return false
	}

	visibleRows := inst.rowCount()
	visibleCols := inst.columnCount()
	if visibleRows <= 0 || visibleCols <= 0 {
		return false
	}

	visibleAreaRatio := float64(visibleRows*visibleCols) / float64(totalRows*totalCols)
	return visibleAreaRatio < autoViewportScaleAreaThreshold
}

func (inst *Instance) valueGlyph(value, minVal, maxVal float64) rune {
	if math.Abs(maxVal-minVal) < 1e-9 {
		return heatmapGlyphs[len(heatmapGlyphs)/2]
	}
	level := int(math.Round(((value - minVal) / (maxVal - minVal)) * float64(len(heatmapGlyphs)-1)))
	if level < 0 {
		level = 0
	}
	if level >= len(heatmapGlyphs) {
		level = len(heatmapGlyphs) - 1
	}
	return heatmapGlyphs[level]
}

func (inst *Instance) resolvePlotStyle() style.Style {
	return inst.heatmapStyle
}

func (inst *Instance) renderBuffer() *paint.Buffer {
	width, height := inst.contentSize()
	buf := paint.NewBuffer(width, height)

	titleStyle := style.NewStyle().Foreground(palette.TitleColor())
	labelStyle := style.NewStyle().Foreground(palette.LabelColor())

	cursorY := 0
	if strings.TrimSpace(inst.title) != "" {
		buf.SetString(0, cursorY, inst.title, titleStyle)
		cursorY++
	}
	if inst.showLegend {
		inst.renderLegend(buf, 0, cursorY, labelStyle)
		cursorY++
	}
	if summaryRow := inst.summaryRow(); summaryRow != "" {
		buf.SetString(0, cursorY, summaryRow, labelStyle)
		cursorY++
	}

	gutter := inst.rowLabelWidth()
	if axisRow := inst.columnLabelRow(gutter); axisRow != "" {
		buf.SetString(0, cursorY, axisRow, labelStyle)
		cursorY++
	}

	if len(inst.values) == 0 || inst.columnCount() == 0 {
		buf.SetString(0, cursorY, "No data", labelStyle)
		return buf
	}

	minVal, maxVal, ok := inst.resolvedMinMax()
	if !ok {
		buf.SetString(0, cursorY, "No data", labelStyle)
		return buf
	}

	plotStartX := 0
	if gutter > 0 {
		plotStartX = gutter + 1
	}

	rowStart, rowEnd := inst.visibleRowRange()
	colStart, colEnd := inst.visibleColumnRange()
	for rowIndex := rowStart; rowIndex < rowEnd; rowIndex++ {
		rowY := cursorY + (rowIndex - rowStart)
		rowValues := inst.values[rowIndex]
		if gutter > 0 {
			buf.SetString(0, rowY, fmt.Sprintf("%-*s", gutter, inst.rowLabel(rowIndex-rowStart)), labelStyle)
		}
		for colIndex := colStart; colIndex < colEnd; colIndex++ {
			cellX := plotStartX + (colIndex-colStart)*2
			if colIndex >= len(rowValues) {
				continue
			}
			glyph, ratio := inst.valueGlyphAndRatio(rowValues[colIndex], minVal, maxVal)
			buf.SetCell(cellX, rowY, glyph, inst.resolveCellStyle(ratio))
		}
	}

	return buf
}

func (inst *Instance) visibleRowStart() int {
	return normalizeViewport(inst.viewport).RowStart
}

func (inst *Instance) visibleColStart() int {
	return normalizeViewport(inst.viewport).ColStart
}

func (inst *Instance) visibleRowRange() (int, int) {
	viewport := normalizeViewport(inst.viewport)
	start := viewport.RowStart
	total := len(inst.values)
	if start > total {
		start = total
	}
	end := total
	if viewport.RowCount > 0 && start+viewport.RowCount < end {
		end = start + viewport.RowCount
	}
	return start, end
}

func (inst *Instance) visibleColumnRange() (int, int) {
	viewport := normalizeViewport(inst.viewport)
	start := viewport.ColStart
	total := inst.totalColumnCount()
	if start > total {
		start = total
	}
	end := total
	if viewport.ColCount > 0 && start+viewport.ColCount < end {
		end = start + viewport.ColCount
	}
	return start, end
}

func clipLabel(label string, width int) string {
	label = strings.TrimSpace(label)
	if width <= 0 || paint.StringWidth(label) <= width {
		return label
	}
	if width == 1 {
		for _, r := range label {
			return string(r)
		}
		return ""
	}
	var builder strings.Builder
	currentWidth := 0
	for _, r := range label {
		runeWidth := paint.StringWidth(string(r))
		if currentWidth+runeWidth >= width {
			break
		}
		builder.WriteRune(r)
		currentWidth += runeWidth
	}
	clipped := strings.TrimSpace(builder.String())
	if clipped == "" {
		for _, r := range label {
			clipped = string(r)
			break
		}
	}
	if paint.StringWidth(clipped) >= width {
		runes := []rune(clipped)
		if len(runes) > 0 {
			clipped = string(runes[:len(runes)-1])
		}
	}
	return clipped + "~"
}

func formatHeatmapNumber(value float64) string {
	if math.Abs(value-math.Round(value)) < 1e-9 {
		return fmt.Sprintf("%.0f", value)
	}
	return fmt.Sprintf("%.1f", value)
}

func (inst *Instance) renderLegend(buf *paint.Buffer, x, y int, labelStyle style.Style) {
	ratios := inst.legendRatios()
	switch inst.legendMode {
	case LegendModeCompact:
		buf.SetString(x, y, "L ", labelStyle)
		cursorX := x + paint.StringWidth("L ")
		for i, ratio := range ratios {
			buf.SetCell(cursorX, y, heatmapGlyphs[i], inst.resolveCellStyle(ratio))
			cursorX++
		}
		buf.SetString(cursorX, y, " H", labelStyle)
	default:
		buf.SetString(x, y, "Low ", labelStyle)
		cursorX := x + paint.StringWidth("Low ")
		for i, ratio := range ratios {
			buf.SetCell(cursorX, y, heatmapGlyphs[i], inst.resolveCellStyle(ratio))
			cursorX += 2
		}
		buf.SetString(cursorX, y, "High", labelStyle)
	}
}

func (inst *Instance) resolveCellStyle(ratio float64) style.Style {
	s := inst.resolvePlotStyle()
	if color := palette.HeatmapColor(ratio, inst.colorMode); color != style.NoColor {
		s = s.Foreground(color)
	}
	return s
}

func (inst *Instance) valueGlyphAndRatio(value, minVal, maxVal float64) (rune, float64) {
	if math.Abs(maxVal-minVal) < 1e-9 {
		return heatmapGlyphs[len(heatmapGlyphs)/2], 0.5
	}
	ratio := (value - minVal) / (maxVal - minVal)
	level := int(math.Round(ratio * float64(len(heatmapGlyphs)-1)))
	if level < 0 {
		level = 0
	}
	if level >= len(heatmapGlyphs) {
		level = len(heatmapGlyphs) - 1
	}
	return heatmapGlyphs[level], ratio
}

func getStringSliceProp(props rtui.Props, key string, def []string) []string {
	if value, ok := props[key]; ok {
		if labels, ok := value.([]string); ok {
			return copyStringSlice(labels)
		}
	}
	return copyStringSlice(def)
}

func getMatrixProp(props rtui.Props, key string, def [][]float64) [][]float64 {
	if value, ok := props[key]; ok {
		if matrix, ok := value.([][]float64); ok {
			return copyFloat64Matrix(matrix)
		}
	}
	return copyFloat64Matrix(def)
}

func getColorModeProp(props rtui.Props, def fwtheme.ColorMode) fwtheme.ColorMode {
	if value, ok := props[propColorMode]; ok {
		if mode, ok := value.(fwtheme.ColorMode); ok {
			return mode
		}
	}
	return def
}

func getLegendModeProp(props rtui.Props, def LegendMode) LegendMode {
	if value, ok := props[propLegendMode]; ok {
		if mode, ok := value.(LegendMode); ok {
			return mode
		}
	}
	return def
}

func getSummaryModeProp(props rtui.Props, def SummaryMode) SummaryMode {
	if value, ok := props[propSummaryMode]; ok {
		if mode, ok := value.(SummaryMode); ok {
			return mode
		}
	}
	if value, ok := props[propShowSummary]; ok {
		if show, ok := value.(bool); ok {
			if show {
				return SummaryModeDetailed
			}
			return SummaryModeNone
		}
	}
	return def
}

func getScaleModeProp(props rtui.Props, def ScaleMode) ScaleMode {
	if value, ok := props[propScaleMode]; ok {
		if mode, ok := value.(ScaleMode); ok {
			return mode
		}
	}
	return def
}

func getViewportProp(props rtui.Props, def Viewport) Viewport {
	if value, ok := props[propViewport]; ok {
		if viewport, ok := value.(Viewport); ok {
			return normalizeViewport(viewport)
		}
	}
	return normalizeViewport(def)
}

func stringSlicesEqual(a, b []string) bool {
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
