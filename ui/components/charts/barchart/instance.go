package barchart

import (
	"math"
	"strconv"
	"strings"

	"github.com/wwsheng009/mint/runtime/dimension"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/charts/internal/axis"
	chartcanvas "github.com/wwsheng009/mint/ui/components/charts/internal/canvas"
	chartlayout "github.com/wwsheng009/mint/ui/components/charts/internal/layout"
	"github.com/wwsheng009/mint/ui/components/charts/internal/palette"
	"github.com/wwsheng009/mint/ui/components/charts/internal/scale"
	"github.com/wwsheng009/mint/ui/components/internal/proputil"
)

const (
	maxHorizontalLabelWidth   = 12
	minHorizontalBarAreaWidth = 4
	verticalLabelFallbackRune = '•'
)

// Instance is the runtime entity for bar chart components.
type Instance struct {
	key         string
	title       string
	labels      []string
	values      []float64
	series      []Series
	mode        Mode
	orientation Orientation
	width       int
	height      int
	showAxis    bool
	showLegend  bool
	showValue   bool
	renderMode  RenderMode
	chartStyle  style.Style
	dirty       bool
}

var (
	_ rtui.ComponentInstance = (*Instance)(nil)
	_ rtui.PaintableInstance = (*Instance)(nil)
	_ interface {
		Measure(layout.Constraints) layout.Size
	} = (*Instance)(nil)
)

// NewInstance creates a new bar chart instance.
func NewInstance(props rtui.Props) *Instance {
	return &Instance{
		key:         proputil.GetString(props, propKey, ""),
		title:       proputil.GetString(props, propTitle, ""),
		labels:      getLabelsProp(props, nil),
		values:      getValuesProp(props, nil),
		series:      getSeriesProp(props, nil),
		mode:        getModeProp(props, ModeGrouped),
		orientation: getOrientationProp(props, OrientationVertical),
		width:       proputil.GetInt(props, propWidth, 0),
		height:      proputil.GetInt(props, propHeight, dimension.ChartMinHeight),
		showAxis:    proputil.GetBool(props, propShowAxis, true),
		showLegend:  proputil.GetBool(props, propShowLegend, false),
		showValue:   proputil.GetBool(props, propShowValue, false),
		renderMode:  getRenderModeProp(props, RenderModeBlock),
		chartStyle:  proputil.GetStyle(props, propStyle, style.Style{}),
		dirty:       true,
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
	oldWidth := inst.width
	oldHeight := inst.height
	oldShowAxis := inst.showAxis
	oldShowLegend := inst.showLegend
	oldShowValue := inst.showValue
	oldRenderMode := inst.renderMode
	oldMode := inst.mode
	oldOrientation := inst.orientation
	oldStyle := inst.chartStyle
	oldLabels := inst.labels
	oldValues := inst.values
	oldSeries := inst.series

	inst.key = proputil.GetString(props, propKey, inst.key)
	inst.title = proputil.GetString(props, propTitle, inst.title)
	inst.labels = getLabelsProp(props, inst.labels)
	inst.values = getValuesProp(props, inst.values)
	inst.series = getSeriesProp(props, inst.series)
	inst.mode = getModeProp(props, inst.mode)
	inst.orientation = getOrientationProp(props, inst.orientation)
	inst.width = proputil.GetInt(props, propWidth, inst.width)
	inst.height = proputil.GetInt(props, propHeight, inst.height)
	if inst.height <= 0 {
		inst.height = dimension.ChartMinHeight
	}
	inst.showAxis = proputil.GetBool(props, propShowAxis, inst.showAxis)
	inst.showLegend = proputil.GetBool(props, propShowLegend, inst.showLegend)
	inst.showValue = proputil.GetBool(props, propShowValue, inst.showValue)
	inst.renderMode = getRenderModeProp(props, inst.renderMode)
	inst.chartStyle = proputil.GetStyle(props, propStyle, inst.chartStyle)

	changed := oldTitle != inst.title ||
		oldWidth != inst.width ||
		oldHeight != inst.height ||
		oldShowAxis != inst.showAxis ||
		oldShowLegend != inst.showLegend ||
		oldShowValue != inst.showValue ||
		oldRenderMode != inst.renderMode ||
		oldMode != inst.mode ||
		oldOrientation != inst.orientation ||
		oldStyle != inst.chartStyle ||
		!stringSlicesEqual(oldLabels, inst.labels) ||
		!float64SlicesEqual(oldValues, inst.values) ||
		!seriesSlicesEqual(oldSeries, inst.series)
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		propKey:         inst.key,
		propTitle:       inst.title,
		propLabels:      copyStringSlice(inst.labels),
		propValues:      copyFloat64Slice(inst.values),
		propSeries:      copySeriesSlice(inst.series),
		propMode:        inst.mode,
		propOrientation: inst.orientation,
		propWidth:       inst.width,
		propHeight:      inst.height,
		propShowAxis:    inst.showAxis,
		propShowLegend:  inst.showLegend,
		propShowValue:   inst.showValue,
		propRenderMode:  inst.renderMode,
		propStyle:       inst.chartStyle,
	}
}

func (inst *Instance) MarkDirty()                         { inst.dirty = true }
func (inst *Instance) IsDirty() bool                      { return inst.dirty }
func (inst *Instance) GetContext() *rtui.ComponentContext { return nil }

func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	header := inst.buildHeaderFrame()
	footer := inst.buildFooterFrame()
	headerSize := header.Measure(layout.UnboundedConstraints())
	footerSize := footer.Measure(layout.UnboundedConstraints())

	width := inst.plotWidth()
	if headerSize.Width > width {
		width = headerSize.Width
	}
	if footerSize.Width > width {
		width = footerSize.Width
	}

	height := len(header.Rows()) + inst.plotHeight() + len(footer.Rows())
	width = constraints.ConstrainWidth(width)
	height = constraints.ConstrainHeight(height)
	return layout.Size{Width: width, Height: height}
}

func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	header := inst.buildHeaderFrame()
	footer := inst.buildFooterFrame()
	plotBuffer := inst.renderPlotBuffer()

	cmds := make([]paint.DrawCmd, 0, len(header.Rows())+len(footer.Rows())+plotBuffer.Height*2)
	cmds = append(cmds, header.Paint(x, y)...)

	plotY := y + len(header.Rows())
	cmds = append(cmds, chartcanvas.BufferToDrawCmds(plotBuffer, x, plotY)...)

	footerY := plotY + inst.plotHeight()
	cmds = append(cmds, footer.Paint(x, footerY)...)
	return cmds
}

func (inst *Instance) buildHeaderFrame() *chartlayout.Frame {
	frame := chartlayout.NewFrame()
	frame.AddIfNotEmpty(chartlayout.SectionTitle, inst.title, style.NewStyle().Foreground(palette.TitleColor()))

	seriesList := inst.resolvedSeries()
	if inst.showLegend {
		for index, series := range seriesList {
			frame.Add(chartlayout.SectionLegend, inst.legendText(index, series), inst.resolveSeriesStyle(index, len(seriesList), series))
		}
	}
	return frame
}

func (inst *Instance) buildFooterFrame() *chartlayout.Frame {
	frame := chartlayout.NewFrame()
	seriesList := inst.visibleSeries()
	if len(seriesList) == 0 {
		return frame
	}

	if inst.orientation == OrientationHorizontal {
		if inst.showAxis {
			frame.Add(
				chartlayout.SectionAxis,
				strings.Repeat(" ", inst.horizontalBarStart())+inst.axisLine(inst.horizontalBarAreaWidth()),
				style.NewStyle().Foreground(palette.AxisColor()),
			)
		}
		return frame
	}

	plotWidth := inst.plotWidth()
	groupCenters := inst.groupCenters(inst.visibleCategoryCount(seriesList))

	if inst.showAxis {
		frame.Add(chartlayout.SectionAxis, inst.axisLine(plotWidth), style.NewStyle().Foreground(palette.AxisColor()))
	}
	if labels := inst.visibleLabels(inst.visibleCategoryCount(seriesList)); len(labels) > 0 {
		frame.Add(chartlayout.SectionLabels, inst.verticalLabelRow(plotWidth, labels, groupCenters), style.NewStyle().Foreground(palette.LabelColor()))
	}
	inst.appendVerticalValueRows(frame, seriesList)
	return frame
}

func (inst *Instance) renderPlotBuffer() *paint.Buffer {
	height := inst.plotHeight()
	width := inst.plotWidth()
	buffer := paint.NewBuffer(width, height)
	seriesList := inst.visibleSeries()
	if len(seriesList) == 0 {
		return buffer
	}

	maxValue := inst.maxVisibleValue(seriesList)
	if maxValue <= 0 {
		maxValue = 1
	}
	if inst.orientation == OrientationHorizontal {
		inst.renderHorizontalBars(buffer, seriesList, maxValue)
		return buffer
	}

	heightScale := scale.NewLinear(0, maxValue, 0, height)
	categoryCount := inst.visibleCategoryCount(seriesList)

	for categoryIndex := 0; categoryIndex < categoryCount; categoryIndex++ {
		groupStart := categoryIndex * (inst.groupWidth(len(seriesList)) + 1)
		if inst.mode == ModeStacked {
			cumulative := 0.0
			for seriesIndex, series := range seriesList {
				if categoryIndex >= len(series.Values) {
					continue
				}
				value := series.Values[categoryIndex]
				if value <= 0 {
					continue
				}
				lower := heightScale.Map(cumulative)
				cumulative += value
				upper := heightScale.Map(cumulative)
				if upper <= lower {
					continue
				}
				barStyle := inst.resolveSeriesStyle(seriesIndex, len(seriesList), series)
				for y := height - upper; y < height-lower; y++ {
					buffer.SetCell(groupStart, y, inst.barRune(), barStyle)
				}
			}
			continue
		}

		for seriesIndex, series := range seriesList {
			if categoryIndex >= len(series.Values) {
				continue
			}
			barHeight := heightScale.Map(series.Values[categoryIndex])
			if barHeight <= 0 {
				continue
			}
			xPos := groupStart + seriesIndex
			if xPos < 0 || xPos >= width {
				continue
			}
			barStyle := inst.resolveSeriesStyle(seriesIndex, len(seriesList), series)
			for y := height - 1; y >= height-barHeight; y-- {
				buffer.SetCell(xPos, y, inst.barRune(), barStyle)
			}
		}
	}

	return buffer
}

func (inst *Instance) resolvedSeries() []Series {
	if len(inst.series) > 0 {
		return copySeriesSlice(inst.series)
	}
	if len(inst.values) == 0 {
		return nil
	}
	return []Series{{
		Values: copyFloat64Slice(inst.values),
	}}
}

func (inst *Instance) visibleSeries() []Series {
	seriesList := inst.resolvedSeries()
	if len(seriesList) == 0 {
		return nil
	}

	categoryCount := inst.maxCategoryCount(seriesList)
	if categoryCount == 0 {
		return nil
	}

	maxCategories := categoryCount
	maxByBudget := inst.maxVisibleCategories(len(seriesList))
	if maxByBudget < maxCategories {
		maxCategories = maxByBudget
	}

	if maxCategories >= categoryCount {
		return copySeriesSlice(seriesList)
	}

	indices := sampleIndices(categoryCount, maxCategories)
	visible := make([]Series, len(seriesList))
	for i, series := range seriesList {
		values := make([]float64, 0, len(indices))
		for _, index := range indices {
			if index < len(series.Values) {
				values = append(values, series.Values[index])
				continue
			}
			values = append(values, 0)
		}
		visible[i] = Series{
			Name:   series.Name,
			Values: values,
			Style:  series.Style,
		}
	}
	return visible
}

func (inst *Instance) visibleLabels(size int) []string {
	labels := paddedLabels(inst.labels, inst.maxCategoryCount(inst.resolvedSeries()))
	if size >= len(labels) {
		return labels
	}
	indices := sampleIndices(len(labels), size)
	result := make([]string, 0, len(indices))
	for _, index := range indices {
		result = append(result, labels[index])
	}
	return result
}

func (inst *Instance) groupCenters(count int) []int {
	if count <= 0 {
		return nil
	}
	centers := make([]int, count)
	groupWidth := inst.groupWidth(len(inst.visibleSeries()))
	for i := 0; i < count; i++ {
		start := i * (groupWidth + 1)
		centers[i] = start + groupWidth/2
	}
	return centers
}

func (inst *Instance) groupWidth(seriesCount int) int {
	if inst.mode == ModeStacked {
		return 1
	}
	if seriesCount <= 0 {
		return 1
	}
	return seriesCount
}

func (inst *Instance) visibleCategoryCount(seriesList []Series) int {
	maxCount := 0
	for _, series := range seriesList {
		if len(series.Values) > maxCount {
			maxCount = len(series.Values)
		}
	}
	return maxCount
}

func (inst *Instance) maxCategoryCount(seriesList []Series) int {
	return inst.visibleCategoryCount(seriesList)
}

func (inst *Instance) plotHeight() int {
	if inst.orientation == OrientationHorizontal {
		seriesList := inst.visibleSeries()
		categoryCount := inst.visibleCategoryCount(seriesList)
		if categoryCount == 0 {
			return 1
		}
		groupHeight := inst.groupWidth(len(seriesList))
		return categoryCount*groupHeight + (categoryCount - 1)
	}
	if inst.height > 0 {
		return inst.height
	}
	return dimension.ChartMinHeight
}

func (inst *Instance) plotWidth() int {
	if inst.orientation == OrientationHorizontal {
		labelWidth := inst.horizontalLabelWidth()
		barWidth := inst.horizontalBarAreaWidth()
		if labelWidth == 0 {
			return barWidth
		}
		return labelWidth + 1 + barWidth
	}

	seriesList := inst.visibleSeries()
	categoryCount := inst.visibleCategoryCount(seriesList)
	if categoryCount == 0 {
		return 1
	}
	groupWidth := inst.groupWidth(len(seriesList))
	width := categoryCount*groupWidth + (categoryCount - 1)
	if inst.width > 0 && inst.width < width {
		return inst.width
	}
	return width
}

func (inst *Instance) maxVisibleCategories(seriesCount int) int {
	span := inst.groupWidth(seriesCount) + 1
	if span <= 1 {
		span = 2
	}

	if inst.orientation == OrientationHorizontal {
		heightBudget := inst.height
		if heightBudget <= 0 {
			heightBudget = dimension.ChartMinHeight
		}
		count := (heightBudget + 1) / span
		if count < 1 {
			return 1
		}
		return count
	}

	if inst.width <= 0 {
		return 1 << 30
	}
	count := (inst.width + 1) / span
	if count < 1 {
		return 1
	}
	return count
}

func (inst *Instance) horizontalLabelWidth() int {
	labels := inst.visibleLabels(inst.visibleCategoryCount(inst.visibleSeries()))
	maxWidth := 0
	for _, label := range labels {
		if width := paint.StringWidth(strings.TrimSpace(label)); width > maxWidth {
			maxWidth = width
		}
	}
	if maxWidth <= 0 {
		return 0
	}
	if maxWidth > maxHorizontalLabelWidth {
		maxWidth = maxHorizontalLabelWidth
	}
	if inst.width > 0 {
		maxBudget := inst.width - 1 - minHorizontalBarAreaWidth
		if maxBudget < 0 {
			maxBudget = 0
		}
		if maxWidth > maxBudget {
			maxWidth = maxBudget
		}
	}
	return maxWidth
}

func (inst *Instance) horizontalBarStart() int {
	labelWidth := inst.horizontalLabelWidth()
	if labelWidth == 0 {
		return 0
	}
	return labelWidth + 1
}

func (inst *Instance) horizontalBarAreaWidth() int {
	totalWidth := inst.width
	if totalWidth <= 0 {
		totalWidth = inst.horizontalLabelWidth() + 1 + dimension.ProgressBarMinWidth
	}
	barWidth := totalWidth - inst.horizontalBarStart()
	if barWidth < 1 {
		return 1
	}
	return barWidth
}

func (inst *Instance) resolveSeriesStyle(index, total int, series Series) style.Style {
	base := style.NewStyle().Foreground(palette.SeriesColor(index))
	common := inst.chartStyle
	if total > 1 {
		common.FG = ""
	}
	base = base.Merge(common)
	return base.Merge(series.Style)
}

func (inst *Instance) appendVerticalValueRows(frame *chartlayout.Frame, seriesList []Series) {
	if !inst.showValue || inst.orientation == OrientationHorizontal || len(seriesList) == 0 {
		return
	}

	if len(seriesList) == 1 {
		frame.Add(
			chartlayout.SectionMeta,
			"Values: "+joinValueLabels(seriesList[0].Values),
			inst.resolveSeriesStyle(0, 1, seriesList[0]),
		)
		return
	}

	for index, series := range seriesList {
		frame.Add(
			chartlayout.SectionMeta,
			inst.seriesSummaryPrefix(index, series)+": "+joinValueLabels(series.Values),
			inst.resolveSeriesStyle(index, len(seriesList), series),
		)
	}

	if inst.mode == ModeStacked {
		frame.Add(
			chartlayout.SectionMeta,
			"Total: "+joinValueLabels(inst.categoryTotals(seriesList)),
			style.NewStyle().Foreground(palette.LabelColor()),
		)
	}
}

func (inst *Instance) renderHorizontalBars(buffer *paint.Buffer, seriesList []Series, maxValue float64) {
	barStart := inst.horizontalBarStart()
	barWidth := inst.horizontalBarAreaWidth()
	valueScale := scale.NewLinear(0, maxValue, 0, barWidth)
	categoryCount := inst.visibleCategoryCount(seriesList)
	labels := inst.visibleLabels(categoryCount)

	for categoryIndex := 0; categoryIndex < categoryCount; categoryIndex++ {
		groupStart := categoryIndex * (inst.groupWidth(len(seriesList)) + 1)
		inst.renderHorizontalLabel(buffer, groupStart, labels, categoryIndex)

		if inst.mode == ModeStacked {
			cumulative := 0.0
			for seriesIndex, series := range seriesList {
				if categoryIndex >= len(series.Values) {
					continue
				}
				value := series.Values[categoryIndex]
				if value <= 0 {
					continue
				}
				lower := valueScale.Map(cumulative)
				cumulative += value
				upper := valueScale.Map(cumulative)
				if upper <= lower {
					continue
				}
				row := groupStart
				barStyle := inst.resolveSeriesStyle(seriesIndex, len(seriesList), series)
				for x := barStart + lower; x < barStart+upper && x < buffer.Width; x++ {
					buffer.SetCell(x, row, inst.barRune(), barStyle)
				}
			}
			if inst.showValue && cumulative > 0 {
				inst.renderInlineValueLabel(
					buffer,
					groupStart,
					barStart+valueScale.Map(cumulative)+1,
					cumulative,
					style.NewStyle().Foreground(palette.LabelColor()),
				)
			}
			continue
		}

		for seriesIndex, series := range seriesList {
			if categoryIndex >= len(series.Values) {
				continue
			}
			barLength := valueScale.Map(series.Values[categoryIndex])
			if barLength <= 0 {
				continue
			}
			row := groupStart + seriesIndex
			if row < 0 || row >= buffer.Height {
				continue
			}
			barStyle := inst.resolveSeriesStyle(seriesIndex, len(seriesList), series)
			for x := barStart; x < barStart+barLength && x < buffer.Width; x++ {
				buffer.SetCell(x, row, inst.barRune(), barStyle)
			}
			if inst.showValue {
				inst.renderInlineValueLabel(buffer, row, barStart+barLength+1, series.Values[categoryIndex], barStyle)
			}
		}
	}
}

func (inst *Instance) renderHorizontalLabel(buffer *paint.Buffer, row int, labels []string, index int) {
	if row < 0 || row >= buffer.Height || index < 0 || index >= len(labels) {
		return
	}
	labelWidth := inst.horizontalLabelWidth()
	if labelWidth <= 0 {
		return
	}
	labelText := inst.fitLabel(labels[index], labelWidth)
	if strings.TrimSpace(labelText) == "" {
		labelText = strings.Repeat(" ", labelWidth)
	}
	buffer.SetStringAligned(0, row, labelText, style.NewStyle().Foreground(palette.LabelColor()), labelWidth)
}

func (inst *Instance) verticalLabelRow(width int, labels []string, centers []int) string {
	if width <= 0 {
		return ""
	}
	row := []rune(strings.Repeat(" ", width))
	limit := len(labels)
	if len(centers) < limit {
		limit = len(centers)
	}
	for index := 0; index < limit; index++ {
		left, right := inst.verticalLabelBounds(width, centers, index)
		if left > right {
			continue
		}
		slotWidth := right - left + 1
		labelText := inst.fitLabel(labels[index], slotWidth)
		if slotWidth <= 2 {
			labelText = compactLabel(labels[index], 1)
		}
		if strings.TrimSpace(labelText) == "" {
			labelText = string(inst.labelFallbackRune())
		}
		labelWidth := paint.StringWidth(labelText)
		if labelWidth <= 0 {
			continue
		}
		start := left
		if labelWidth == 1 && index < len(centers) {
			start = centers[index]
		} else if slotWidth > labelWidth {
			start = left + (slotWidth-labelWidth)/2
		}
		writeLabelRunes(row, start, right+1, labelText)
	}
	return string(row)
}

func (inst *Instance) verticalLabelBounds(width int, centers []int, index int) (int, int) {
	left := 0
	if index > 0 {
		left = (centers[index-1]+centers[index])/2 + 1
	}
	right := width - 1
	if index+1 < len(centers) {
		right = (centers[index] + centers[index+1]) / 2
	}
	if left < 0 {
		left = 0
	}
	if right >= width {
		right = width - 1
	}
	return left, right
}

func writeLabelRunes(row []rune, start, limit int, label string) {
	cursor := start
	for _, r := range label {
		runeWidth := paint.StringWidth(string(r))
		if runeWidth <= 0 {
			continue
		}
		if cursor+runeWidth > limit || cursor >= len(row) {
			break
		}
		row[cursor] = r
		for i := 1; i < runeWidth && cursor+i < limit && cursor+i < len(row); i++ {
			row[cursor+i] = ' '
		}
		cursor += runeWidth
	}
}

func (inst *Instance) renderInlineValueLabel(buffer *paint.Buffer, row, x int, value float64, textStyle style.Style) {
	if row < 0 || row >= buffer.Height {
		return
	}
	label := formatValueLabel(value)
	if label == "" {
		return
	}
	labelWidth := paint.StringWidth(label)
	if labelWidth <= 0 || buffer.Width <= 0 {
		return
	}
	if x < 0 {
		x = 0
	}
	if x+labelWidth > buffer.Width {
		x = buffer.Width - labelWidth
	}
	if x < 0 || x >= buffer.Width {
		return
	}
	buffer.SetString(x, row, label, textStyle)
}

func (inst *Instance) legendText(index int, series Series) string {
	label := strings.TrimSpace(series.Name)
	if label == "" {
		label = "Series " + strconv.Itoa(index+1)
	}
	return string(inst.barRune()) + " " + label
}

func (inst *Instance) maxVisibleValue(seriesList []Series) float64 {
	if inst.mode == ModeStacked {
		maxValue := 0.0
		categoryCount := inst.visibleCategoryCount(seriesList)
		for categoryIndex := 0; categoryIndex < categoryCount; categoryIndex++ {
			sum := 0.0
			for _, series := range seriesList {
				if categoryIndex < len(series.Values) {
					sum += series.Values[categoryIndex]
				}
			}
			if sum > maxValue {
				maxValue = sum
			}
		}
		return maxValue
	}

	maxValue := 0.0
	for _, series := range seriesList {
		for _, value := range series.Values {
			if value > maxValue {
				maxValue = value
			}
		}
	}
	return maxValue
}

func (inst *Instance) categoryTotals(seriesList []Series) []float64 {
	categoryCount := inst.visibleCategoryCount(seriesList)
	if categoryCount <= 0 {
		return nil
	}
	totals := make([]float64, categoryCount)
	for _, series := range seriesList {
		for index := 0; index < categoryCount; index++ {
			if index < len(series.Values) {
				totals[index] += series.Values[index]
			}
		}
	}
	return totals
}

func getLabelsProp(props rtui.Props, def []string) []string {
	if value, ok := props[propLabels]; ok {
		if labels, ok := value.([]string); ok {
			return copyStringSlice(labels)
		}
	}
	return copyStringSlice(def)
}

func getValuesProp(props rtui.Props, def []float64) []float64 {
	if value, ok := props[propValues]; ok {
		if values, ok := value.([]float64); ok {
			return copyFloat64Slice(values)
		}
	}
	return copyFloat64Slice(def)
}

func getSeriesProp(props rtui.Props, def []Series) []Series {
	if value, ok := props[propSeries]; ok {
		if series, ok := value.([]Series); ok {
			return copySeriesSlice(series)
		}
	}
	return copySeriesSlice(def)
}

func getModeProp(props rtui.Props, def Mode) Mode {
	if value, ok := props[propMode]; ok {
		if mode, ok := value.(Mode); ok {
			return mode
		}
	}
	return def
}

func getOrientationProp(props rtui.Props, def Orientation) Orientation {
	if value, ok := props[propOrientation]; ok {
		if orientation, ok := value.(Orientation); ok {
			return orientation
		}
	}
	return def
}

func getRenderModeProp(props rtui.Props, def RenderMode) RenderMode {
	if value, ok := props[propRenderMode]; ok {
		if mode, ok := value.(RenderMode); ok {
			return mode
		}
	}
	return def
}

func sampleIndices(length, count int) []int {
	if length <= 0 || count <= 0 {
		return nil
	}
	if count >= length {
		indices := make([]int, length)
		for i := range indices {
			indices[i] = i
		}
		return indices
	}
	if count == 1 {
		return []int{length - 1}
	}

	indices := make([]int, count)
	last := float64(length - 1)
	for i := 0; i < count; i++ {
		index := int(math.Round(float64(i) * last / float64(count-1)))
		if index < 0 {
			index = 0
		}
		if index >= length {
			index = length - 1
		}
		indices[i] = index
	}
	return indices
}

func paddedLabels(labels []string, size int) []string {
	if size <= 0 {
		return nil
	}
	result := make([]string, size)
	for i := 0; i < size; i++ {
		if i < len(labels) {
			result[i] = labels[i]
			continue
		}
		result[i] = ""
	}
	return result
}

func float64SlicesEqual(a, b []float64) bool {
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

func (inst *Instance) barRune() rune {
	if inst.renderMode == RenderModeASCII {
		return '#'
	}
	return '█'
}

func (inst *Instance) axisLine(width int) string {
	if width <= 0 {
		return ""
	}
	if inst.renderMode == RenderModeASCII {
		return strings.Repeat("-", width)
	}
	return axis.HorizontalLine(width)
}

func (inst *Instance) labelFallbackRune() rune {
	if inst.renderMode == RenderModeASCII {
		return '.'
	}
	return verticalLabelFallbackRune
}

func (inst *Instance) fitLabel(label string, width int) string {
	return fitLabelWithFallback(label, width, inst.labelFallbackRune())
}

func fitLabel(label string, width int) string {
	return fitLabelWithFallback(label, width, verticalLabelFallbackRune)
}

func fitLabelWithFallback(label string, width int, fallback rune) string {
	label = strings.TrimSpace(label)
	if width <= 0 {
		return ""
	}
	if paint.StringWidth(label) <= width {
		return label
	}
	if compact := compactLabel(label, width); compact != "" {
		return compact
	}
	if width == 1 {
		return string(axis.LabelRune(label, fallback))
	}

	var builder strings.Builder
	currentWidth := 0
	for _, r := range label {
		runeWidth := paint.StringWidth(string(r))
		if currentWidth+runeWidth > width-1 {
			break
		}
		builder.WriteRune(r)
		currentWidth += runeWidth
	}
	if builder.Len() == 0 {
		return string(axis.LabelRune(label, fallback))
	}
	return builder.String() + "."
}

func compactLabel(label string, width int) string {
	label = strings.TrimSpace(label)
	if width <= 0 || label == "" {
		return ""
	}

	tokens := splitLabelTokens(label)
	if width == 1 {
		if digits := trailingDigits(label); digits != "" {
			return rightRunes(digits, 1)
		}
		if initials := tokenInitials(tokens); initials != "" {
			return firstRunes(initials, 1)
		}
		return firstRunes(label, 1)
	}
	if initials := tokenInitials(tokens); initials != "" && paint.StringWidth(initials) <= width {
		return initials
	}
	if digits := trailingDigits(label); digits != "" && paint.StringWidth(digits) <= width {
		return digits
	}
	if width == 2 {
		if initials := tokenInitials(tokens); initials != "" {
			return firstRunes(initials, 2)
		}
		first := firstRunes(label, 1)
		last := rightRunes(label, 1)
		if first != "" && last != "" && first != last {
			return first + last
		}
		return firstRunes(label, 2)
	}
	return ""
}

func splitLabelTokens(label string) []string {
	return strings.FieldsFunc(label, func(r rune) bool {
		switch r {
		case ' ', '-', '_', '/', '\\', '.', ':':
			return true
		default:
			return false
		}
	})
}

func tokenInitials(tokens []string) string {
	if len(tokens) <= 1 {
		return ""
	}
	var builder strings.Builder
	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		builder.WriteString(firstRunes(token, 1))
	}
	return builder.String()
}

func trailingDigits(label string) string {
	runes := []rune(strings.TrimSpace(label))
	end := len(runes)
	start := end
	for start > 0 {
		r := runes[start-1]
		if r < '0' || r > '9' {
			break
		}
		start--
	}
	if start == end {
		return ""
	}
	return string(runes[start:end])
}

func firstRunes(text string, count int) string {
	if count <= 0 {
		return ""
	}
	var builder strings.Builder
	current := 0
	for _, r := range text {
		if current >= count {
			break
		}
		builder.WriteRune(r)
		current++
	}
	return builder.String()
}

func rightRunes(text string, count int) string {
	if count <= 0 {
		return ""
	}
	runes := []rune(text)
	if count >= len(runes) {
		return text
	}
	return string(runes[len(runes)-count:])
}

func formatValueLabel(value float64) string {
	if math.Abs(value-math.Round(value)) < 1e-9 {
		return strconv.FormatInt(int64(math.Round(value)), 10)
	}
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func joinValueLabels(values []float64) string {
	if len(values) == 0 {
		return ""
	}
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, formatValueLabel(value))
	}
	return strings.Join(parts, " ")
}

func (inst *Instance) seriesSummaryPrefix(index int, series Series) string {
	label := strings.TrimSpace(series.Name)
	if label == "" {
		label = "Series " + strconv.Itoa(index+1)
	}
	return label
}
