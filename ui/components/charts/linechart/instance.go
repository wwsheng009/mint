package linechart

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/wwsheng009/mint/runtime/dimension"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/charts/internal/axis"
	chartcanvas "github.com/wwsheng009/mint/ui/components/charts/internal/canvas"
	chartlayout "github.com/wwsheng009/mint/ui/components/charts/internal/layout"
	"github.com/wwsheng009/mint/ui/components/charts/internal/palette"
	"github.com/wwsheng009/mint/ui/components/charts/internal/raster"
	"github.com/wwsheng009/mint/ui/components/charts/internal/scale"
	"github.com/wwsheng009/mint/ui/components/internal/proputil"
)

const pointGlyph = '●'

// Instance is the runtime entity for line chart components.
type Instance struct {
	key           string
	title         string
	data          []float64
	series        []Series
	seriesName    string
	labels        []string
	axisLabelMode AxisLabelMode
	renderBackend RenderBackend
	width         int
	height        int
	showAxis      bool
	showGrid      bool
	showLegend    bool
	showPoints    bool
	chartStyle    style.Style
	bounds        [4]int
	dirty         bool
}

var (
	_ rtui.ComponentInstance      = (*Instance)(nil)
	_ rtui.PaintableInstance      = (*Instance)(nil)
	_ rtui.ScenePaintableInstance = (*Instance)(nil)
	_ interface {
		Measure(layout.Constraints) layout.Size
	} = (*Instance)(nil)
)

// NewInstance creates a new line chart instance.
func NewInstance(props rtui.Props) *Instance {
	return &Instance{
		key:           proputil.GetString(props, propKey, ""),
		title:         proputil.GetString(props, propTitle, ""),
		data:          getDataProp(props, nil),
		series:        getSeriesProp(props, nil),
		seriesName:    proputil.GetString(props, propSeriesName, ""),
		labels:        getLabelsProp(props, nil),
		axisLabelMode: getAxisLabelModeProp(props, AxisLabelModeAuto),
		renderBackend: getRenderBackendProp(props, RenderBackendText),
		width:         proputil.GetInt(props, propWidth, 0),
		height:        proputil.GetInt(props, propHeight, dimension.ChartMinHeight),
		showAxis:      proputil.GetBool(props, propShowAxis, true),
		showGrid:      proputil.GetBool(props, propShowGrid, false),
		showLegend:    proputil.GetBool(props, propShowLegend, false),
		showPoints:    proputil.GetBool(props, propShowPoints, true),
		chartStyle:    proputil.GetStyle(props, propStyle, style.Style{}),
		dirty:         true,
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
	oldShowGrid := inst.showGrid
	oldShowLegend := inst.showLegend
	oldShowPoints := inst.showPoints
	oldSeriesName := inst.seriesName
	oldLabels := inst.labels
	oldAxisLabelMode := inst.axisLabelMode
	oldRenderBackend := inst.renderBackend
	oldSeries := inst.series
	oldStyle := inst.chartStyle
	oldData := inst.data

	inst.key = proputil.GetString(props, propKey, inst.key)
	inst.title = proputil.GetString(props, propTitle, inst.title)
	inst.data = getDataProp(props, inst.data)
	inst.series = getSeriesProp(props, inst.series)
	inst.seriesName = proputil.GetString(props, propSeriesName, inst.seriesName)
	inst.labels = getLabelsProp(props, inst.labels)
	inst.axisLabelMode = getAxisLabelModeProp(props, inst.axisLabelMode)
	inst.renderBackend = getRenderBackendProp(props, inst.renderBackend)
	inst.width = proputil.GetInt(props, propWidth, inst.width)
	inst.height = proputil.GetInt(props, propHeight, inst.height)
	if inst.height <= 0 {
		inst.height = dimension.ChartMinHeight
	}
	inst.showAxis = proputil.GetBool(props, propShowAxis, inst.showAxis)
	inst.showGrid = proputil.GetBool(props, propShowGrid, inst.showGrid)
	inst.showLegend = proputil.GetBool(props, propShowLegend, inst.showLegend)
	inst.showPoints = proputil.GetBool(props, propShowPoints, inst.showPoints)
	inst.chartStyle = proputil.GetStyle(props, propStyle, inst.chartStyle)

	changed := oldTitle != inst.title ||
		!seriesSlicesEqual(oldSeries, inst.series) ||
		oldSeriesName != inst.seriesName ||
		!stringSlicesEqual(oldLabels, inst.labels) ||
		oldAxisLabelMode != inst.axisLabelMode ||
		oldRenderBackend != inst.renderBackend ||
		oldWidth != inst.width ||
		oldHeight != inst.height ||
		oldShowAxis != inst.showAxis ||
		oldShowGrid != inst.showGrid ||
		oldShowLegend != inst.showLegend ||
		oldShowPoints != inst.showPoints ||
		oldStyle != inst.chartStyle ||
		!float64SlicesEqual(oldData, inst.data)
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		propKey:           inst.key,
		propTitle:         inst.title,
		propData:          copyFloat64Slice(inst.data),
		propSeries:        copySeriesSlice(inst.series),
		propSeriesName:    inst.seriesName,
		propLabels:        copyStringSlice(inst.labels),
		propAxisLabelMode: inst.axisLabelMode,
		propRenderBackend: inst.renderBackend,
		propWidth:         inst.width,
		propHeight:        inst.height,
		propShowAxis:      inst.showAxis,
		propShowGrid:      inst.showGrid,
		propShowLegend:    inst.showLegend,
		propShowPoints:    inst.showPoints,
		propStyle:         inst.chartStyle,
	}
}

func (inst *Instance) MarkDirty()                         { inst.dirty = true }
func (inst *Instance) IsDirty() bool                      { return inst.dirty }
func (inst *Instance) GetContext() *rtui.ComponentContext { return nil }
func (inst *Instance) SetBounds(x, y, w, h int)           { inst.bounds = [4]int{x, y, w, h} }

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
	for index, series := range seriesList {
		legend := inst.legendText(index, series)
		if legend == "" {
			continue
		}
		frame.Add(chartlayout.SectionLegend, legend, inst.resolveSeriesStyle(index, len(seriesList), series))
	}

	return frame
}

func (inst *Instance) buildFooterFrame() *chartlayout.Frame {
	frame := chartlayout.NewFrame()
	if inst.showAxis {
		frame.Add(chartlayout.SectionAxis, axis.HorizontalLine(inst.plotWidth()), style.NewStyle().Foreground(palette.AxisColor()))
		if labelRow := inst.axisLabelRow(); labelRow != "" {
			frame.Add(chartlayout.SectionAxis, labelRow, style.NewStyle().Foreground(palette.LabelColor()))
		}
	}
	return frame
}

func (inst *Instance) renderPlotBuffer() *paint.Buffer {
	height := inst.plotHeight()
	width := inst.plotWidth()
	buffer := paint.NewBuffer(width, height)
	seriesList := inst.resolvedSeries()

	if len(seriesList) == 0 {
		return buffer
	}

	if inst.showGrid {
		gridStyle := style.NewStyle().Foreground(palette.GridColor())
		gridRows := axis.GridRows(height, 3)
		if inst.hasAxisLabels() && len(gridRows) > 1 {
			gridRows = gridRows[:len(gridRows)-1]
		}
		for _, row := range gridRows {
			buffer.SetString(0, row, strings.Repeat("┈", width), gridStyle)
		}
	}

	minVal, maxVal := inst.seriesDomain(seriesList)
	yScale := scale.NewLinear(minVal, maxVal, height-1, 0)

	for index, series := range seriesList {
		if len(series.Data) == 0 {
			continue
		}

		sampled := resampleForContinuity(series.Data, inst.sampleCount())
		ys := inst.seriesRows(sampled, yScale, height)
		xBand := scale.NewBand(len(ys), 0, width-1)
		points := make([]raster.Point, 0, len(ys))
		for pointIndex, row := range ys {
			xPos := xBand.Position(pointIndex)
			if xPos >= width {
				break
			}
			points = append(points, raster.Point{X: xPos, Y: row})
		}

		raster.DrawPolylineToBuffer(buffer, points, inst.resolveSeriesStyle(index, len(seriesList), series), raster.PolylineOptions{
			ShowPoints: inst.showPoints,
			PointRune:  pointGlyph,
			Glyphs:     raster.DefaultLineGlyphs(),
		})
	}

	return buffer
}

func (inst *Instance) legendText(index int, series Series) string {
	if !inst.showLegend {
		return ""
	}

	label := strings.TrimSpace(series.Name)
	if label == "" {
		label = fmt.Sprintf("Series %d", index+1)
	}
	return string(pointGlyph) + " " + label
}

func (inst *Instance) seriesRows(data []float64, yScale scale.Linear, height int) []int {
	rows := make([]int, len(data))
	if len(data) == 0 {
		return rows
	}

	for i, value := range data {
		rows[i] = clampInt(yScale.Map(value), 0, height-1)
	}
	return rows
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

func (inst *Instance) plotHeight() int {
	if inst.height > 0 {
		return inst.height
	}
	return dimension.ChartMinHeight
}

func (inst *Instance) sampleCount() int {
	if inst.width > 0 {
		count := (inst.width + 1) / 2
		if count < 1 {
			return 1
		}
		return count
	}
	if maxLen := inst.maxSeriesLength(); maxLen > 0 {
		return maxLen
	}
	return 1
}

func (inst *Instance) plotWidth() int {
	if inst.width > 0 {
		return inst.width
	}
	width := inst.sampleCount()*2 - 1
	if width < 1 {
		return 1
	}
	return width
}

func getDataProp(props rtui.Props, def []float64) []float64 {
	if value, ok := props[propData]; ok {
		if data, ok := value.([]float64); ok {
			return copyFloat64Slice(data)
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

func getLabelsProp(props rtui.Props, def []string) []string {
	if value, ok := props[propLabels]; ok {
		if labels, ok := value.([]string); ok {
			return copyStringSlice(labels)
		}
	}
	return copyStringSlice(def)
}

func getAxisLabelModeProp(props rtui.Props, def AxisLabelMode) AxisLabelMode {
	if value, ok := props[propAxisLabelMode]; ok {
		if mode, ok := value.(AxisLabelMode); ok {
			return mode
		}
	}
	return def
}

func (inst *Instance) resolvedSeries() []Series {
	if len(inst.series) > 0 {
		return copySeriesSlice(inst.series)
	}
	if len(inst.data) == 0 {
		return nil
	}
	return []Series{{
		Name: inst.seriesName,
		Data: copyFloat64Slice(inst.data),
	}}
}

func (inst *Instance) maxSeriesLength() int {
	maxLen := 0
	for _, series := range inst.resolvedSeries() {
		if len(series.Data) > maxLen {
			maxLen = len(series.Data)
		}
	}
	return maxLen
}

func (inst *Instance) seriesDomain(seriesList []Series) (float64, float64) {
	hasValue := false
	minVal := 0.0
	maxVal := 0.0
	for _, series := range seriesList {
		for _, value := range series.Data {
			if !hasValue {
				minVal = value
				maxVal = value
				hasValue = true
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
	return minVal, maxVal
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

func (inst *Instance) hasAxisLabels() bool {
	return inst.showAxis && len(inst.labels) > 0 && inst.plotWidth() > 0
}

func (inst *Instance) axisLabelRow() string {
	if !inst.hasAxisLabels() {
		return ""
	}

	positions, labels := inst.axisLabelSlots()
	if len(positions) == 0 || len(labels) == 0 {
		return ""
	}
	return axis.LabelRow(inst.plotWidth(), labels, positions, '•')
}

func (inst *Instance) axisLabelSlots() ([]int, []string) {
	width := inst.plotWidth()
	labels := inst.visibleAxisLabels()
	if width <= 0 || len(labels) == 0 {
		return nil, nil
	}

	xBand := scale.NewBand(len(labels), 0, width-1)
	switch inst.axisLabelMode {
	case AxisLabelModeDense:
		return inst.denseAxisLabelSlots(labels, xBand)
	case AxisLabelModeSparse:
		return inst.sparseAxisLabelSlots(labels, xBand)
	}

	return inst.autoAxisLabelSlots(labels, xBand)
}

func (inst *Instance) denseAxisLabelSlots(labels []string, xBand scale.Band) ([]int, []string) {
	positions := make([]int, 0, len(labels))
	visible := make([]string, 0, len(labels))
	lastX := -1
	for i, label := range labels {
		xPos := xBand.Position(i)
		if xPos == lastX {
			continue
		}
		positions = append(positions, xPos)
		visible = append(visible, foldAxisLabel(label))
		lastX = xPos
	}
	return positions, visible
}

func (inst *Instance) sparseAxisLabelSlots(labels []string, xBand scale.Band) ([]int, []string) {
	if len(labels) <= 2 {
		return inst.denseAxisLabelSlots(labels, xBand)
	}

	targetCount := maxInt(2, (inst.plotWidth()+1)/4)
	if targetCount >= len(labels) {
		return inst.denseAxisLabelSlots(labels, xBand)
	}

	positions := make([]int, 0, targetCount)
	visible := make([]string, 0, targetCount)
	lastIndex := -1
	lastX := -1
	for i := 0; i < targetCount; i++ {
		index := 0
		if targetCount > 1 {
			index = int(float64(i) * float64(len(labels)-1) / float64(targetCount-1))
		}
		if index <= lastIndex {
			index = lastIndex + 1
		}
		if index >= len(labels) {
			index = len(labels) - 1
		}
		xPos := xBand.Position(index)
		if xPos == lastX {
			continue
		}
		positions = append(positions, xPos)
		visible = append(visible, foldAxisLabel(labels[index]))
		lastIndex = index
		lastX = xPos
	}
	return positions, visible
}

func (inst *Instance) autoAxisLabelSlots(labels []string, xBand scale.Band) ([]int, []string) {
	positions := make([]int, 0, len(labels))
	visible := make([]string, 0, len(labels))
	width := inst.plotWidth()
	minSpacing := 1
	if len(labels) > maxInt(3, width/2) {
		minSpacing = 2
	}

	lastX := -minSpacing - 1
	for i, label := range labels {
		xPos := xBand.Position(i)
		isLast := i == len(labels)-1
		if !isLast && xPos-lastX < minSpacing {
			continue
		}
		if isLast && len(positions) > 0 && xPos == positions[len(positions)-1] {
			continue
		}

		positions = append(positions, xPos)
		visible = append(visible, foldAxisLabel(label))
		lastX = xPos
	}
	return positions, visible
}

func (inst *Instance) visibleAxisLabels() []string {
	if len(inst.labels) == 0 {
		return nil
	}

	count := inst.sampleCount()
	if count <= 0 {
		return nil
	}
	if len(inst.labels) <= count {
		return copyStringSlice(inst.labels)
	}

	labels := make([]string, 0, count)
	lastIndex := -1
	for i := 0; i < count; i++ {
		index := 0
		if count > 1 {
			index = int(float64(i) * float64(len(inst.labels)-1) / float64(count-1))
		}
		if index <= lastIndex {
			index = lastIndex + 1
		}
		if index >= len(inst.labels) {
			index = len(inst.labels) - 1
		}
		labels = append(labels, inst.labels[index])
		lastIndex = index
	}
	return labels
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

func foldAxisLabel(label string) string {
	label = strings.TrimSpace(label)
	if label == "" {
		return ""
	}

	for i := len(label) - 1; i >= 0; i-- {
		r := rune(label[i])
		if unicode.IsDigit(r) {
			return string(r)
		}
	}

	for _, r := range label {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return string(unicode.ToUpper(r))
		}
	}

	return axis.LabelRow(1, []string{label}, []int{0}, '•')
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
