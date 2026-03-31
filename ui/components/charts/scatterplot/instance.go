package scatterplot

import (
	"fmt"
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
	defaultPlotWidth = 12
	defaultPointRune = '●'
	maxCollisionRune = '+'
	bandRuneLight    = '░'
	bandRuneMedium   = '▒'
	bandRuneHeavy    = '▓'
)

var defaultSeriesGlyphs = []rune{'●', '◆', '■', '▲', '✚', '○'}

type domain struct {
	minX    float64
	maxX    float64
	minY    float64
	maxY    float64
	hasData bool
}

// Instance is the runtime entity for scatter plot components.
type Instance struct {
	key          string
	title        string
	points       []Point
	series       []Series
	seriesName   string
	width        int
	height       int
	domainSpec   Domain
	viewportSpec Viewport
	xRefs        []ReferenceLine
	yRefs        []ReferenceLine
	xBands       []ReferenceBand
	yBands       []ReferenceBand
	showAxis     bool
	showGrid     bool
	showLegend   bool
	chartStyle   style.Style
	dirty        bool
}

var (
	_ rtui.ComponentInstance = (*Instance)(nil)
	_ rtui.PaintableInstance = (*Instance)(nil)
	_ interface {
		Measure(layout.Constraints) layout.Size
	} = (*Instance)(nil)
)

// NewInstance creates a new scatter plot instance.
func NewInstance(props rtui.Props) *Instance {
	return &Instance{
		key:          proputil.GetString(props, propKey, ""),
		title:        proputil.GetString(props, propTitle, ""),
		points:       getPointsProp(props, nil),
		series:       getSeriesProp(props, nil),
		seriesName:   proputil.GetString(props, propSeriesName, ""),
		width:        proputil.GetInt(props, propWidth, 0),
		height:       proputil.GetInt(props, propHeight, dimension.ChartMinHeight),
		domainSpec:   getDomainProp(props, Domain{}),
		viewportSpec: getViewportProp(props, Viewport{}),
		xRefs:        getReferenceLinesProp(props, propXRefDefs, propXRefs, nil),
		yRefs:        getReferenceLinesProp(props, propYRefDefs, propYRefs, nil),
		xBands:       getReferenceBandsProp(props, propXBands, nil),
		yBands:       getReferenceBandsProp(props, propYBands, nil),
		showAxis:     proputil.GetBool(props, propShowAxis, true),
		showGrid:     proputil.GetBool(props, propShowGrid, false),
		showLegend:   proputil.GetBool(props, propShowLegend, false),
		chartStyle:   proputil.GetStyle(props, propStyle, style.Style{}),
		dirty:        true,
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
	oldSeriesName := inst.seriesName
	oldDomainSpec := inst.domainSpec
	oldViewportSpec := inst.viewportSpec
	oldXRefs := inst.xRefs
	oldYRefs := inst.yRefs
	oldXBands := inst.xBands
	oldYBands := inst.yBands
	oldPoints := inst.points
	oldSeries := inst.series
	oldStyle := inst.chartStyle

	inst.key = proputil.GetString(props, propKey, inst.key)
	inst.title = proputil.GetString(props, propTitle, inst.title)
	inst.points = getPointsProp(props, inst.points)
	inst.series = getSeriesProp(props, inst.series)
	inst.seriesName = proputil.GetString(props, propSeriesName, inst.seriesName)
	inst.width = proputil.GetInt(props, propWidth, inst.width)
	inst.height = proputil.GetInt(props, propHeight, inst.height)
	inst.domainSpec = getDomainProp(props, inst.domainSpec)
	inst.viewportSpec = getViewportProp(props, inst.viewportSpec)
	inst.xRefs = getReferenceLinesProp(props, propXRefDefs, propXRefs, inst.xRefs)
	inst.yRefs = getReferenceLinesProp(props, propYRefDefs, propYRefs, inst.yRefs)
	inst.xBands = getReferenceBandsProp(props, propXBands, inst.xBands)
	inst.yBands = getReferenceBandsProp(props, propYBands, inst.yBands)
	if inst.height <= 0 {
		inst.height = dimension.ChartMinHeight
	}
	inst.showAxis = proputil.GetBool(props, propShowAxis, inst.showAxis)
	inst.showGrid = proputil.GetBool(props, propShowGrid, inst.showGrid)
	inst.showLegend = proputil.GetBool(props, propShowLegend, inst.showLegend)
	inst.chartStyle = proputil.GetStyle(props, propStyle, inst.chartStyle)

	changed := oldTitle != inst.title ||
		oldWidth != inst.width ||
		oldHeight != inst.height ||
		oldShowAxis != inst.showAxis ||
		oldShowGrid != inst.showGrid ||
		oldShowLegend != inst.showLegend ||
		oldSeriesName != inst.seriesName ||
		oldDomainSpec != inst.domainSpec ||
		oldViewportSpec != inst.viewportSpec ||
		!referenceLineSlicesEqual(oldXRefs, inst.xRefs) ||
		!referenceLineSlicesEqual(oldYRefs, inst.yRefs) ||
		!referenceBandSlicesEqual(oldXBands, inst.xBands) ||
		!referenceBandSlicesEqual(oldYBands, inst.yBands) ||
		oldStyle != inst.chartStyle ||
		!pointSlicesEqual(oldPoints, inst.points) ||
		!seriesSlicesEqual(oldSeries, inst.series)
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		propKey:        inst.key,
		propTitle:      inst.title,
		propPoints:     copyPointSlice(inst.points),
		propSeries:     copySeriesSlice(inst.series),
		propSeriesName: inst.seriesName,
		propWidth:      inst.width,
		propHeight:     inst.height,
		propDomain:     inst.domainSpec,
		propViewport:   inst.viewportSpec,
		propXRefs:      referenceLineValues(inst.xRefs),
		propYRefs:      referenceLineValues(inst.yRefs),
		propXRefDefs:   copyReferenceLineSlice(inst.xRefs),
		propYRefDefs:   copyReferenceLineSlice(inst.yRefs),
		propXBands:     copyReferenceBandSlice(inst.xBands),
		propYBands:     copyReferenceBandSlice(inst.yBands),
		propShowAxis:   inst.showAxis,
		propShowGrid:   inst.showGrid,
		propShowLegend: inst.showLegend,
		propStyle:      inst.chartStyle,
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

	if !inst.showLegend {
		return frame
	}

	seriesList := inst.resolvedSeries()
	for index, series := range seriesList {
		frame.Add(
			chartlayout.SectionLegend,
			inst.legendText(index, series),
			inst.resolveSeriesStyle(index, len(seriesList), series),
		)
	}
	for _, line := range inst.xRefs {
		if text := inst.referenceLineLegendText("x", '│', line); text != "" {
			frame.Add(chartlayout.SectionLegend, text, inst.referenceLineStyle())
		}
	}
	for _, line := range inst.yRefs {
		if text := inst.referenceLineLegendText("y", '─', line); text != "" {
			frame.Add(chartlayout.SectionLegend, text, inst.referenceLineStyle())
		}
	}
	for _, band := range inst.xBands {
		if text := inst.referenceBandLegendText("x", band); text != "" {
			frame.Add(chartlayout.SectionLegend, text, inst.referenceBandStyle())
		}
	}
	for _, band := range inst.yBands {
		if text := inst.referenceBandLegendText("y", band); text != "" {
			frame.Add(chartlayout.SectionLegend, text, inst.referenceBandStyle())
		}
	}

	return frame
}

func (inst *Instance) buildFooterFrame() *chartlayout.Frame {
	frame := chartlayout.NewFrame()
	if !inst.showAxis {
		return frame
	}

	frame.Add(chartlayout.SectionAxis, axis.HorizontalLine(inst.plotWidth()), style.NewStyle().Foreground(palette.AxisColor()))
	if summary := inst.domainSummary(); summary != "" {
		frame.Add(chartlayout.SectionLabels, summary, style.NewStyle().Foreground(palette.LabelColor()))
	}
	return frame
}

func (inst *Instance) renderPlotBuffer() *paint.Buffer {
	height := inst.plotHeight()
	width := inst.plotWidth()
	buffer := paint.NewBuffer(width, height)
	seriesList := inst.resolvedSeries()
	visibleDomain := inst.resolvedViewport(seriesList)

	if !visibleDomain.hasData {
		buffer.SetString(0, 0, "No data", style.NewStyle().Foreground(palette.LabelColor()))
		return buffer
	}

	if inst.showGrid {
		inst.renderGrid(buffer)
	}

	xScale := scale.NewLinear(visibleDomain.minX, visibleDomain.maxX, 0, width-1)
	yScale := scale.NewLinear(visibleDomain.minY, visibleDomain.maxY, height-1, 0)
	inst.renderReferenceBands(buffer, xScale, yScale, visibleDomain)
	inst.renderReferenceLines(buffer, xScale, yScale, visibleDomain)
	pointHits := make([][]int, height)
	for y := range pointHits {
		pointHits[y] = make([]int, width)
	}

	for index, series := range seriesList {
		seriesStyle := inst.resolveSeriesStyle(index, len(seriesList), series)
		glyph := inst.resolveSeriesGlyph(index, series)
		for _, point := range series.Points {
			if !pointInDomain(point, visibleDomain) {
				continue
			}
			xPos := xScale.Map(point.X)
			yPos := clampInt(yScale.Map(point.Y), 0, height-1)
			pointHits[yPos][xPos]++
			if pointHits[yPos][xPos] == 1 {
				buffer.SetCell(xPos, yPos, glyph, seriesStyle)
				continue
			}
			buffer.SetCell(xPos, yPos, collisionRuneForCount(pointHits[yPos][xPos]), inst.collisionStyle())
		}
	}

	return buffer
}

func (inst *Instance) renderGrid(buffer *paint.Buffer) {
	gridStyle := style.NewStyle().Foreground(palette.GridColor())
	for _, col := range gridCols(buffer.Width, 3) {
		for y := 0; y < buffer.Height; y++ {
			buffer.SetCell(col, y, '┆', gridStyle)
		}
	}
	for _, row := range axis.GridRows(buffer.Height, 3) {
		buffer.SetString(0, row, strings.Repeat("┈", buffer.Width), gridStyle)
	}
}

func (inst *Instance) renderReferenceBands(buffer *paint.Buffer, xScale, yScale scale.Linear, visibleDomain domain) {
	bandStyle := style.NewStyle().Foreground(palette.ReferenceBandColor()).Merge(inst.chartStyle)

	for _, band := range inst.xBands {
		clipped, ok := clipReferenceBand(band, visibleDomain.minX, visibleDomain.maxX)
		if !ok {
			continue
		}
		start := clampInt(xScale.Map(clipped.Min), 0, buffer.Width-1)
		end := clampInt(xScale.Map(clipped.Max), 0, buffer.Width-1)
		if start > end {
			start, end = end, start
		}
		for x := start; x <= end; x++ {
			for y := 0; y < buffer.Height; y++ {
				buffer.SetCell(x, y, mergeReferenceBandRune(buffer.Cells[y][x].Cluster), bandStyle)
			}
		}
	}

	for _, band := range inst.yBands {
		clipped, ok := clipReferenceBand(band, visibleDomain.minY, visibleDomain.maxY)
		if !ok {
			continue
		}
		start := clampInt(yScale.Map(clipped.Max), 0, buffer.Height-1)
		end := clampInt(yScale.Map(clipped.Min), 0, buffer.Height-1)
		if start > end {
			start, end = end, start
		}
		for y := start; y <= end; y++ {
			for x := 0; x < buffer.Width; x++ {
				buffer.SetCell(x, y, mergeReferenceBandRune(buffer.Cells[y][x].Cluster), bandStyle)
			}
		}
	}
}

func (inst *Instance) renderReferenceLines(buffer *paint.Buffer, xScale, yScale scale.Linear, visibleDomain domain) {
	refStyle := style.NewStyle().Foreground(palette.ReferenceLineColor()).Merge(inst.chartStyle)

	for _, line := range inst.xRefs {
		xValue := line.Value
		if xValue < visibleDomain.minX || xValue > visibleDomain.maxX {
			continue
		}
		xPos := clampInt(xScale.Map(xValue), 0, buffer.Width-1)
		for y := 0; y < buffer.Height; y++ {
			buffer.SetCell(xPos, y, mergeReferenceRune(buffer.Cells[y][xPos].Cluster, '│'), refStyle)
		}
	}

	for _, line := range inst.yRefs {
		yValue := line.Value
		if yValue < visibleDomain.minY || yValue > visibleDomain.maxY {
			continue
		}
		yPos := clampInt(yScale.Map(yValue), 0, buffer.Height-1)
		for x := 0; x < buffer.Width; x++ {
			buffer.SetCell(x, yPos, mergeReferenceRune(buffer.Cells[yPos][x].Cluster, '─'), refStyle)
		}
	}
}

func (inst *Instance) legendText(index int, series Series) string {
	label := strings.TrimSpace(series.Name)
	if label == "" {
		label = "Series " + strconv.Itoa(index+1)
	}
	return string(inst.resolveSeriesGlyph(index, series)) + " " + label
}

func (inst *Instance) domainSummary() string {
	domain := inst.resolvedViewport(inst.resolvedSeries())
	if !domain.hasData {
		return ""
	}
	return fmt.Sprintf("x:%s..%s y:%s..%s",
		formatFloat(domain.minX),
		formatFloat(domain.maxX),
		formatFloat(domain.minY),
		formatFloat(domain.maxY),
	)
}

func (inst *Instance) plotHeight() int {
	if inst.height > 0 {
		return inst.height
	}
	return dimension.ChartMinHeight
}

func (inst *Instance) plotWidth() int {
	if inst.width > 0 {
		return inst.width
	}
	return defaultPlotWidth
}

func (inst *Instance) resolvedSeries() []Series {
	if len(inst.series) > 0 {
		return copySeriesSlice(inst.series)
	}
	if len(inst.points) == 0 {
		return nil
	}
	return []Series{{
		Name:   inst.seriesName,
		Points: copyPointSlice(inst.points),
	}}
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

func (inst *Instance) resolveSeriesGlyph(index int, series Series) rune {
	if series.Glyph != 0 {
		return series.Glyph
	}
	if len(defaultSeriesGlyphs) == 0 {
		return defaultPointRune
	}
	return defaultSeriesGlyphs[index%len(defaultSeriesGlyphs)]
}

func (inst *Instance) collisionStyle() style.Style {
	return style.NewStyle().
		Foreground(palette.CollisionColor()).
		Merge(inst.chartStyle)
}

func (inst *Instance) referenceBandStyle() style.Style {
	return style.NewStyle().
		Foreground(palette.ReferenceBandColor()).
		Merge(inst.chartStyle)
}

func (inst *Instance) referenceLineStyle() style.Style {
	return style.NewStyle().
		Foreground(palette.ReferenceLineColor()).
		Merge(inst.chartStyle)
}

func (inst *Instance) referenceLineLegendText(axisName string, glyph rune, line ReferenceLine) string {
	label := strings.TrimSpace(line.Label)
	if label == "" {
		return ""
	}
	return string(glyph) + " " + axisName + ": " + label
}

func (inst *Instance) referenceBandLegendText(axisName string, band ReferenceBand) string {
	label := strings.TrimSpace(band.Label)
	if label == "" {
		return ""
	}
	return string(bandRuneLight) + " " + axisName + ": " + label
}

func collisionRuneForCount(count int) rune {
	if count <= 1 {
		return defaultPointRune
	}
	if count >= 2 && count <= 9 {
		return rune('0' + count)
	}
	return maxCollisionRune
}

func (inst *Instance) seriesDomain(seriesList []Series) domain {
	result := domain{}
	for _, series := range seriesList {
		for _, point := range series.Points {
			if !result.hasData {
				result.minX = point.X
				result.maxX = point.X
				result.minY = point.Y
				result.maxY = point.Y
				result.hasData = true
				continue
			}
			if point.X < result.minX {
				result.minX = point.X
			}
			if point.X > result.maxX {
				result.maxX = point.X
			}
			if point.Y < result.minY {
				result.minY = point.Y
			}
			if point.Y > result.maxY {
				result.maxY = point.Y
			}
		}
	}
	return result
}

func (inst *Instance) resolvedDomain(seriesList []Series) domain {
	result := inst.seriesDomain(seriesList)
	if !result.hasData {
		return result
	}

	spec := normalizeDomainSpec(inst.domainSpec)
	if spec.HasX {
		result.minX = spec.MinX
		result.maxX = spec.MaxX
	}
	if spec.HasY {
		result.minY = spec.MinY
		result.maxY = spec.MaxY
	}
	return result
}

func (inst *Instance) resolvedViewport(seriesList []Series) domain {
	result := inst.resolvedDomain(seriesList)
	if !result.hasData {
		return result
	}

	spec := normalizeViewportSpec(inst.viewportSpec)
	if spec.HasX {
		result.minX = spec.MinX
		result.maxX = spec.MaxX
	}
	if spec.HasY {
		result.minY = spec.MinY
		result.maxY = spec.MaxY
	}
	return result
}

func getPointsProp(props rtui.Props, def []Point) []Point {
	if value, ok := props[propPoints]; ok {
		if points, ok := value.([]Point); ok {
			return copyPointSlice(points)
		}
	}
	return copyPointSlice(def)
}

func getSeriesProp(props rtui.Props, def []Series) []Series {
	if value, ok := props[propSeries]; ok {
		if series, ok := value.([]Series); ok {
			return copySeriesSlice(series)
		}
	}
	return copySeriesSlice(def)
}

func getDomainProp(props rtui.Props, def Domain) Domain {
	if value, ok := props[propDomain]; ok {
		if domain, ok := value.(Domain); ok {
			return normalizeDomainSpec(domain)
		}
	}
	return normalizeDomainSpec(def)
}

func getViewportProp(props rtui.Props, def Viewport) Viewport {
	if value, ok := props[propViewport]; ok {
		if viewport, ok := value.(Viewport); ok {
			return normalizeViewportSpec(viewport)
		}
	}
	return normalizeViewportSpec(def)
}

func getFloat64SliceProp(props rtui.Props, key string, def []float64) []float64 {
	if value, ok := props[key]; ok {
		if values, ok := value.([]float64); ok {
			return copyFloat64Slice(values)
		}
	}
	return copyFloat64Slice(def)
}

func getReferenceLinesProp(props rtui.Props, defsKey, valuesKey string, def []ReferenceLine) []ReferenceLine {
	if value, ok := props[valuesKey]; ok {
		if values, ok := value.([]float64); ok {
			def = referenceLinesFromValues(values)
		}
	}
	if value, ok := props[defsKey]; ok {
		if lines, ok := value.([]ReferenceLine); ok {
			return copyReferenceLineSlice(lines)
		}
	}
	return copyReferenceLineSlice(def)
}

func getReferenceBandsProp(props rtui.Props, key string, def []ReferenceBand) []ReferenceBand {
	if value, ok := props[key]; ok {
		if bands, ok := value.([]ReferenceBand); ok {
			return copyReferenceBandSlice(bands)
		}
	}
	return copyReferenceBandSlice(def)
}

func gridCols(width, count int) []int {
	if width <= 2 || count <= 0 {
		return nil
	}

	cols := make([]int, 0, count)
	last := -1
	for i := 1; i <= count; i++ {
		col := int(float64(i) * float64(width-1) / float64(count+1))
		if col <= 0 || col >= width-1 || col == last {
			continue
		}
		cols = append(cols, col)
		last = col
	}
	return cols
}

func formatFloat(value float64) string {
	if math.Abs(value-math.Round(value)) < 1e-9 {
		return strconv.FormatInt(int64(math.Round(value)), 10)
	}
	return strconv.FormatFloat(value, 'f', -1, 64)
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

func pointInDomain(point Point, visible domain) bool {
	return point.X >= visible.minX &&
		point.X <= visible.maxX &&
		point.Y >= visible.minY &&
		point.Y <= visible.maxY
}

func clipReferenceBand(band ReferenceBand, minVal, maxVal float64) (ReferenceBand, bool) {
	band = normalizeReferenceBand(band)
	if band.Max < minVal || band.Min > maxVal {
		return ReferenceBand{}, false
	}
	if band.Min < minVal {
		band.Min = minVal
	}
	if band.Max > maxVal {
		band.Max = maxVal
	}
	return band, true
}

func referenceLinesFromValues(values []float64) []ReferenceLine {
	if len(values) == 0 {
		return nil
	}
	lines := make([]ReferenceLine, len(values))
	for i, value := range values {
		lines[i] = NewReferenceLine(value)
	}
	return lines
}

func mergeReferenceBandRune(existing string) rune {
	switch existing {
	case "", " ":
		return bandRuneLight
	case string(bandRuneLight):
		return bandRuneMedium
	case string(bandRuneMedium), string(bandRuneHeavy):
		return bandRuneHeavy
	default:
		return bandRuneLight
	}
}

func mergeReferenceRune(existing string, incoming rune) rune {
	if existing == "" || existing == " " {
		return incoming
	}

	switch []rune(existing)[0] {
	case '│', '┆':
		if incoming == '─' {
			return '┼'
		}
		return '│'
	case '─', '┈':
		if incoming == '│' {
			return '┼'
		}
		return '─'
	case '┼':
		return '┼'
	default:
		return incoming
	}
}
