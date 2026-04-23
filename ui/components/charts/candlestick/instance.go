package candlestick

import (
	"math"
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
	"github.com/wwsheng009/mint/ui/components/charts/internal/scale"
	"github.com/wwsheng009/mint/ui/components/internal/proputil"
)

const (
	defaultPlotWidth    = 11
	defaultVolumeHeight = 3
	wickRune            = '│'
	upBodyRune          = '▓'
	downBodyRune        = '█'
	flatBodyRune        = '■'
	volumeRune          = '▆'
)

type trend int

const (
	trendFlat trend = iota
	trendUp
	trendDown
)

type domain struct {
	minLow  float64
	maxHigh float64
	hasData bool
}

// Instance is the runtime entity for candlestick components.
type Instance struct {
	key          string
	title        string
	candles      []Candle
	width        int
	height       int
	showAxis     bool
	showGrid     bool
	showLegend   bool
	showVolume   bool
	volumeHeight int
	upStyle      style.Style
	downStyle    style.Style
	flatStyle    style.Style
	wickStyle    style.Style
	volumeStyle  style.Style
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

// NewInstance creates a new candlestick instance.
func NewInstance(props rtui.Props) *Instance {
	return &Instance{
		key:          proputil.GetString(props, propKey, ""),
		title:        proputil.GetString(props, propTitle, ""),
		candles:      getCandlesProp(props, nil),
		width:        proputil.GetInt(props, propWidth, 0),
		height:       proputil.GetInt(props, propHeight, dimension.ChartMinHeight),
		showAxis:     proputil.GetBool(props, propShowAxis, true),
		showGrid:     proputil.GetBool(props, propShowGrid, false),
		showLegend:   proputil.GetBool(props, propShowLegend, false),
		showVolume:   proputil.GetBool(props, propShowVolume, false),
		volumeHeight: proputil.GetInt(props, propVolumeHeight, defaultVolumeHeight),
		upStyle:      proputil.GetStyle(props, propUpStyle, style.Style{}),
		downStyle:    proputil.GetStyle(props, propDownStyle, style.Style{}),
		flatStyle:    proputil.GetStyle(props, propFlatStyle, style.Style{}),
		wickStyle:    proputil.GetStyle(props, propWickStyle, style.Style{}),
		volumeStyle:  proputil.GetStyle(props, propVolumeStyle, style.Style{}),
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
	oldShowVolume := inst.showVolume
	oldVolumeHeight := inst.volumeHeight
	oldCandles := inst.candles
	oldUpStyle := inst.upStyle
	oldDownStyle := inst.downStyle
	oldFlatStyle := inst.flatStyle
	oldWickStyle := inst.wickStyle
	oldVolumeStyle := inst.volumeStyle
	oldStyle := inst.chartStyle

	inst.key = proputil.GetString(props, propKey, inst.key)
	inst.title = proputil.GetString(props, propTitle, inst.title)
	inst.candles = getCandlesProp(props, inst.candles)
	inst.width = proputil.GetInt(props, propWidth, inst.width)
	inst.height = proputil.GetInt(props, propHeight, inst.height)
	if inst.height <= 0 {
		inst.height = dimension.ChartMinHeight
	}
	inst.showAxis = proputil.GetBool(props, propShowAxis, inst.showAxis)
	inst.showGrid = proputil.GetBool(props, propShowGrid, inst.showGrid)
	inst.showLegend = proputil.GetBool(props, propShowLegend, inst.showLegend)
	inst.showVolume = proputil.GetBool(props, propShowVolume, inst.showVolume)
	inst.volumeHeight = proputil.GetInt(props, propVolumeHeight, inst.volumeHeight)
	if inst.volumeHeight <= 0 {
		inst.volumeHeight = defaultVolumeHeight
	}
	inst.upStyle = proputil.GetStyle(props, propUpStyle, inst.upStyle)
	inst.downStyle = proputil.GetStyle(props, propDownStyle, inst.downStyle)
	inst.flatStyle = proputil.GetStyle(props, propFlatStyle, inst.flatStyle)
	inst.wickStyle = proputil.GetStyle(props, propWickStyle, inst.wickStyle)
	inst.volumeStyle = proputil.GetStyle(props, propVolumeStyle, inst.volumeStyle)
	inst.chartStyle = proputil.GetStyle(props, propStyle, inst.chartStyle)

	changed := oldTitle != inst.title ||
		oldWidth != inst.width ||
		oldHeight != inst.height ||
		oldShowAxis != inst.showAxis ||
		oldShowGrid != inst.showGrid ||
		oldShowLegend != inst.showLegend ||
		oldShowVolume != inst.showVolume ||
		oldVolumeHeight != inst.volumeHeight ||
		oldUpStyle != inst.upStyle ||
		oldDownStyle != inst.downStyle ||
		oldFlatStyle != inst.flatStyle ||
		oldWickStyle != inst.wickStyle ||
		oldVolumeStyle != inst.volumeStyle ||
		oldStyle != inst.chartStyle ||
		!candleSlicesEqual(oldCandles, inst.candles)
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		propKey:          inst.key,
		propTitle:        inst.title,
		propCandles:      copyCandleSlice(inst.candles),
		propWidth:        inst.width,
		propHeight:       inst.height,
		propShowAxis:     inst.showAxis,
		propShowGrid:     inst.showGrid,
		propShowLegend:   inst.showLegend,
		propShowVolume:   inst.showVolume,
		propVolumeHeight: inst.volumeHeight,
		propUpStyle:      inst.upStyle,
		propDownStyle:    inst.downStyle,
		propFlatStyle:    inst.flatStyle,
		propWickStyle:    inst.wickStyle,
		propVolumeStyle:  inst.volumeStyle,
		propStyle:        inst.chartStyle,
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

	height := len(header.Rows()) + inst.totalPlotHeight() + len(footer.Rows())
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

	volumeBuffer := inst.renderVolumeBuffer()
	if volumeBuffer.Height > 0 {
		cmds = append(cmds, chartcanvas.BufferToDrawCmds(volumeBuffer, x, plotY+plotBuffer.Height)...)
	}

	footerY := plotY + inst.totalPlotHeight()
	cmds = append(cmds, footer.Paint(x, footerY)...)
	return cmds
}

func (inst *Instance) buildHeaderFrame() *chartlayout.Frame {
	frame := chartlayout.NewFrame()
	frame.AddIfNotEmpty(chartlayout.SectionTitle, inst.title, style.NewStyle().Foreground(palette.TitleColor()))

	if !inst.showLegend {
		return frame
	}

	frame.Add(chartlayout.SectionLegend, string(upBodyRune)+" Up", inst.resolveBodyStyle(trendUp, style.Style{}))
	frame.Add(chartlayout.SectionLegend, string(downBodyRune)+" Down", inst.resolveBodyStyle(trendDown, style.Style{}))
	frame.Add(chartlayout.SectionLegend, string(flatBodyRune)+" Flat", inst.resolveBodyStyle(trendFlat, style.Style{}))
	if inst.showVolume && inst.hasVisibleVolumeData() {
		frame.Add(chartlayout.SectionLegend, string(volumeRune)+" Volume", inst.resolveVolumeLegendStyle())
	}
	return frame
}

func (inst *Instance) buildFooterFrame() *chartlayout.Frame {
	frame := chartlayout.NewFrame()
	if !inst.showAxis {
		return frame
	}

	frame.Add(chartlayout.SectionAxis, axis.HorizontalLine(inst.plotWidth()), style.NewStyle().Foreground(palette.AxisColor()))
	if labels := inst.visibleLabels(); len(labels) > 0 {
		frame.Add(chartlayout.SectionLabels, inst.compactLabelRow(labels, inst.plotPositions(), '•'), style.NewStyle().Foreground(palette.LabelColor()))
	}
	return frame
}

func (inst *Instance) renderPlotBuffer() *paint.Buffer {
	height := inst.plotHeight()
	width := inst.plotWidth()
	buffer := paint.NewBuffer(width, height)
	candles := inst.visibleCandles()
	domain := inst.candleDomain(candles)

	if !domain.hasData {
		buffer.SetString(0, 0, "No data", style.NewStyle().Foreground(palette.LabelColor()))
		return buffer
	}

	if inst.showGrid {
		gridStyle := style.NewStyle().Foreground(palette.GridColor())
		for _, row := range axis.GridRows(height, 3) {
			buffer.SetString(0, row, strings.Repeat("┈", width), gridStyle)
		}
	}

	xBand := scale.NewBand(len(candles), 0, width-1)
	yScale := scale.NewLinear(domain.minLow, domain.maxHigh, height-1, 0)

	for index, raw := range candles {
		candle := normalizeCandle(raw)
		xPos := xBand.Position(index)
		highY := clampInt(yScale.Map(candle.High), 0, height-1)
		lowY := clampInt(yScale.Map(candle.Low), 0, height-1)
		openY := clampInt(yScale.Map(candle.Open), 0, height-1)
		closeY := clampInt(yScale.Map(candle.Close), 0, height-1)

		state := candleTrend(candle)
		wickStyle := inst.resolveWickStyle(state, candle.Style)
		bodyStyle := inst.resolveBodyStyle(state, candle.Style)
		for y := highY; y <= lowY; y++ {
			buffer.SetCell(xPos, y, wickRune, wickStyle)
		}

		bodyTop := minInt(openY, closeY)
		bodyBottom := maxInt(openY, closeY)
		bodyRune := candleBodyRune(state)
		for y := bodyTop; y <= bodyBottom; y++ {
			buffer.SetCell(xPos, y, bodyRune, bodyStyle)
		}
	}

	return buffer
}

func (inst *Instance) renderVolumeBuffer() *paint.Buffer {
	height := inst.volumePlotHeight()
	width := inst.plotWidth()
	buffer := paint.NewBuffer(width, height)
	if height <= 0 {
		return buffer
	}

	candles := inst.visibleCandles()
	maxVolume := inst.maxVolume(candles)
	if len(candles) == 0 || maxVolume <= 0 {
		return buffer
	}

	xBand := scale.NewBand(len(candles), 0, width-1)
	volumeScale := scale.NewLinear(0, maxVolume, 0, height)

	for index, raw := range candles {
		candle := normalizeCandle(raw)
		if candle.Volume <= 0 {
			continue
		}

		barHeight := volumeScale.Map(candle.Volume)
		if barHeight <= 0 {
			barHeight = 1
		}
		if barHeight > height {
			barHeight = height
		}

		xPos := xBand.Position(index)
		barStyle := inst.resolveVolumeStyle(candleTrend(candle), candle.Style)
		for offset := 0; offset < barHeight; offset++ {
			buffer.SetCell(xPos, height-1-offset, volumeRune, barStyle)
		}
	}

	return buffer
}

func (inst *Instance) visibleCandles() []Candle {
	if len(inst.candles) == 0 {
		return nil
	}
	maxVisible := inst.maxVisibleCandles()
	if maxVisible >= len(inst.candles) {
		return copyCandleSlice(inst.candles)
	}

	indices := sampleIndices(len(inst.candles), maxVisible)
	visible := make([]Candle, 0, len(indices))
	for _, index := range indices {
		visible = append(visible, inst.candles[index])
	}
	return visible
}

func (inst *Instance) visibleLabels() []string {
	candles := inst.visibleCandles()
	labels := make([]string, 0, len(candles))
	hasLabel := false
	for _, candle := range candles {
		label := strings.TrimSpace(candle.Label)
		if label != "" {
			hasLabel = true
		}
		labels = append(labels, label)
	}
	if !hasLabel {
		return nil
	}
	return labels
}

func (inst *Instance) plotPositions() []int {
	candles := inst.visibleCandles()
	if len(candles) == 0 {
		return nil
	}

	xBand := scale.NewBand(len(candles), 0, inst.plotWidth()-1)
	positions := make([]int, len(candles))
	for i := range candles {
		positions[i] = xBand.Position(i)
	}
	return positions
}

func (inst *Instance) compactLabelRow(labels []string, positions []int, fallback rune) string {
	if inst.plotWidth() <= 0 {
		return ""
	}
	if fallback == 0 {
		fallback = '•'
	}

	row := []rune(strings.Repeat(" ", inst.plotWidth()))
	limit := len(labels)
	if len(positions) < limit {
		limit = len(positions)
	}
	for i := 0; i < limit; i++ {
		left, right := compactLabelSlot(inst.plotWidth(), positions, i)
		if left < 0 || right < left || right >= len(row) {
			continue
		}

		slotWidth := right - left + 1
		maxTokenWidth := compactLabelMaxWidth(slotWidth)
		token := foldAxisLabel(labels[i], maxTokenWidth, fallback)
		if token == "" {
			continue
		}

		tokenRunes := []rune(token)
		tokenWidth := len(tokenRunes)
		if tokenWidth > slotWidth {
			tokenRunes = tokenRunes[:slotWidth]
			tokenWidth = len(tokenRunes)
		}

		start := left + maxInt(0, (slotWidth-tokenWidth)/2)
		for offset, r := range tokenRunes {
			x := start + offset
			if x < left || x > right || x < 0 || x >= len(row) {
				continue
			}
			row[x] = r
		}
	}
	return string(row)
}

func (inst *Instance) plotHeight() int {
	if inst.height > 0 {
		return inst.height
	}
	return dimension.ChartMinHeight
}

func (inst *Instance) totalPlotHeight() int {
	return inst.plotHeight() + inst.volumePlotHeight()
}

func (inst *Instance) volumePlotHeight() int {
	if !inst.showVolume || !inst.hasVisibleVolumeData() {
		return 0
	}
	if inst.volumeHeight > 0 {
		return inst.volumeHeight
	}
	return defaultVolumeHeight
}

func (inst *Instance) plotWidth() int {
	if inst.width > 0 {
		return inst.width
	}
	candles := inst.visibleCandles()
	if len(candles) == 0 {
		return 1
	}
	width := len(candles)*2 - 1
	if width < 1 {
		width = 1
	}
	if width < defaultPlotWidth {
		return width
	}
	return width
}

func (inst *Instance) maxVisibleCandles() int {
	if inst.width <= 0 {
		return 1 << 30
	}
	if inst.width < 1 {
		return 1
	}
	return inst.width
}

func (inst *Instance) resolveTrendBaseStyle(state trend) style.Style {
	base := style.NewStyle()
	switch state {
	case trendUp:
		base = base.Foreground(palette.UpColor())
	case trendDown:
		base = base.Foreground(palette.DownColor())
	default:
		base = base.Foreground(palette.FlatColor())
	}
	return base.Merge(inst.chartStyle)
}

func (inst *Instance) resolveBodyStyle(state trend, candleStyle style.Style) style.Style {
	base := inst.resolveTrendBaseStyle(state)
	switch state {
	case trendUp:
		base = base.Merge(inst.upStyle)
	case trendDown:
		base = base.Merge(inst.downStyle)
	default:
		base = base.Merge(inst.flatStyle)
	}
	return base.Merge(candleStyle)
}

func (inst *Instance) resolveWickStyle(state trend, candleStyle style.Style) style.Style {
	return inst.resolveTrendBaseStyle(state).Merge(inst.wickStyle).Merge(candleStyle)
}

func (inst *Instance) resolveVolumeStyle(state trend, candleStyle style.Style) style.Style {
	return inst.resolveTrendBaseStyle(state).Merge(inst.volumeStyle).Merge(candleStyle)
}

func (inst *Instance) resolveVolumeLegendStyle() style.Style {
	return style.NewStyle().Foreground(palette.LabelColor()).Merge(inst.chartStyle).Merge(inst.volumeStyle)
}

func (inst *Instance) hasVisibleVolumeData() bool {
	for _, candle := range inst.visibleCandles() {
		if candle.Volume > 0 {
			return true
		}
	}
	return false
}

func (inst *Instance) candleDomain(candles []Candle) domain {
	result := domain{}
	for _, raw := range candles {
		candle := normalizeCandle(raw)
		if !result.hasData {
			result.minLow = candle.Low
			result.maxHigh = candle.High
			result.hasData = true
			continue
		}
		if candle.Low < result.minLow {
			result.minLow = candle.Low
		}
		if candle.High > result.maxHigh {
			result.maxHigh = candle.High
		}
	}
	return result
}

func (inst *Instance) maxVolume(candles []Candle) float64 {
	maxVolume := 0.0
	for _, candle := range candles {
		if candle.Volume > maxVolume {
			maxVolume = candle.Volume
		}
	}
	return maxVolume
}

func getCandlesProp(props rtui.Props, def []Candle) []Candle {
	if value, ok := props[propCandles]; ok {
		if candles, ok := value.([]Candle); ok {
			return copyCandleSlice(candles)
		}
	}
	return copyCandleSlice(def)
}

func normalizeCandle(c Candle) Candle {
	high := maxFloat64(c.High, c.Open, c.Close, c.Low)
	low := minFloat64(c.Low, c.Open, c.Close, c.High)
	c.High = high
	c.Low = low
	return c
}

func candleTrend(c Candle) trend {
	switch {
	case c.Close > c.Open:
		return trendUp
	case c.Close < c.Open:
		return trendDown
	default:
		return trendFlat
	}
}

func candleBodyRune(state trend) rune {
	switch state {
	case trendUp:
		return upBodyRune
	case trendDown:
		return downBodyRune
	default:
		return flatBodyRune
	}
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

func compactLabelSlot(width int, positions []int, index int) (int, int) {
	if width <= 0 || index < 0 || index >= len(positions) {
		return -1, -1
	}
	left := 0
	if index > 0 {
		left = (positions[index-1]+positions[index])/2 + 1
	}
	right := width - 1
	if index < len(positions)-1 {
		right = (positions[index] + positions[index+1]) / 2
	}
	if left < 0 {
		left = 0
	}
	if right >= width {
		right = width - 1
	}
	return left, right
}

func compactLabelMaxWidth(slotWidth int) int {
	switch {
	case slotWidth <= 1:
		return 1
	case slotWidth == 2:
		// Reserve one visual gap when labels are dense.
		return 1
	case slotWidth >= 4:
		return 4
	default:
		return slotWidth - 1
	}
}

func foldAxisLabel(label string, maxWidth int, fallback rune) string {
	label = strings.TrimSpace(label)
	if maxWidth <= 0 {
		return ""
	}
	if fallback == 0 {
		fallback = '•'
	}
	if label == "" {
		return string(fallback)
	}
	if len([]rune(label)) <= maxWidth {
		return label
	}

	segments := splitLabelSegments(label)
	if maxWidth == 1 {
		if hasDigit(label) {
			return string(lastAlphaNumericRune(label, fallback))
		}
		return string(axis.LabelRune(label, fallback))
	}

	if initials := labelInitialism(segments); len([]rune(initials)) >= 2 && len([]rune(initials)) <= maxWidth {
		return initials
	}

	if len(segments) > 1 {
		last := segments[len(segments)-1]
		if last != "" && len([]rune(last)) <= maxWidth {
			return last
		}
	}

	runes := []rune(label)
	if len(runes) > maxWidth {
		runes = runes[:maxWidth]
	}
	return string(runes)
}

func splitLabelSegments(label string) []string {
	return strings.FieldsFunc(label, func(r rune) bool {
		return !(unicode.IsLetter(r) || unicode.IsDigit(r))
	})
}

func labelInitialism(segments []string) string {
	if len(segments) <= 1 {
		return ""
	}
	var builder strings.Builder
	for _, segment := range segments {
		for _, r := range segment {
			builder.WriteRune(unicode.ToUpper(r))
			break
		}
	}
	return builder.String()
}

func hasDigit(label string) bool {
	for _, r := range label {
		if unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

func lastAlphaNumericRune(label string, fallback rune) rune {
	runes := []rune(label)
	for i := len(runes) - 1; i >= 0; i-- {
		r := runes[i]
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return unicode.ToUpper(r)
		}
	}
	return fallback
}

func maxFloat64(values ...float64) float64 {
	if len(values) == 0 {
		return 0
	}
	maxValue := values[0]
	for _, value := range values[1:] {
		if value > maxValue {
			maxValue = value
		}
	}
	return maxValue
}

func minFloat64(values ...float64) float64 {
	if len(values) == 0 {
		return 0
	}
	minValue := values[0]
	for _, value := range values[1:] {
		if value < minValue {
			minValue = value
		}
	}
	return minValue
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
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
