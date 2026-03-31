package sparkline

import (
	"math"
	"strings"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	chartcanvas "github.com/wwsheng009/mint/ui/components/charts/internal/canvas"
	"github.com/wwsheng009/mint/ui/components/charts/internal/downsample"
	"github.com/wwsheng009/mint/ui/components/charts/internal/palette"
	"github.com/wwsheng009/mint/ui/components/internal/proputil"
)

var (
	blockGlyphs   = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
	brailleGlyphs = []rune{'⣀', '⣄', '⣆', '⣇', '⣧', '⣷', '⣾', '⣿'}
	asciiGlyphs   = []rune{'.', ':', '-', '=', '+', '*', '#', '@'}
)

// Instance is the runtime entity for sparkline components.
type Instance struct {
	key                string
	data               []float64
	title              string
	width              int
	height             int
	inlineLabel        string
	highlightMinMax    bool
	autoHeight         bool
	renderMode         RenderMode
	sparkStyle         style.Style
	measuredPlotHeight int
	dirty              bool
}

var (
	_ rtui.ComponentInstance = (*Instance)(nil)
	_ rtui.PaintableInstance = (*Instance)(nil)
	_ interface {
		Measure(layout.Constraints) layout.Size
	} = (*Instance)(nil)
)

// NewInstance creates a new sparkline instance.
func NewInstance(props rtui.Props) *Instance {
	return &Instance{
		key:             proputil.GetString(props, propKey, ""),
		data:            getDataProp(props, nil),
		title:           proputil.GetString(props, propTitle, ""),
		width:           proputil.GetInt(props, propWidth, 0),
		height:          proputil.GetInt(props, propHeight, 0),
		inlineLabel:     proputil.GetString(props, propInlineLabel, ""),
		highlightMinMax: proputil.GetBool(props, propHighlightMinMax, false),
		autoHeight:      proputil.GetBool(props, propAutoHeight, false),
		renderMode:      getRenderModeProp(props, RenderModeAuto),
		sparkStyle:      proputil.GetStyle(props, propStyle, style.Style{}),
		dirty:           true,
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
	oldInlineLabel := inst.inlineLabel
	oldHighlightMinMax := inst.highlightMinMax
	oldAutoHeight := inst.autoHeight
	oldMode := inst.renderMode
	oldStyle := inst.sparkStyle
	oldData := inst.data

	inst.key = proputil.GetString(props, propKey, inst.key)
	inst.data = getDataProp(props, inst.data)
	inst.title = proputil.GetString(props, propTitle, inst.title)
	inst.width = proputil.GetInt(props, propWidth, inst.width)
	inst.height = proputil.GetInt(props, propHeight, inst.height)
	inst.inlineLabel = proputil.GetString(props, propInlineLabel, inst.inlineLabel)
	inst.highlightMinMax = proputil.GetBool(props, propHighlightMinMax, inst.highlightMinMax)
	inst.autoHeight = proputil.GetBool(props, propAutoHeight, inst.autoHeight)
	inst.renderMode = getRenderModeProp(props, inst.renderMode)
	inst.sparkStyle = proputil.GetStyle(props, propStyle, inst.sparkStyle)

	changed := oldTitle != inst.title ||
		oldWidth != inst.width ||
		oldHeight != inst.height ||
		oldInlineLabel != inst.inlineLabel ||
		oldHighlightMinMax != inst.highlightMinMax ||
		oldAutoHeight != inst.autoHeight ||
		oldMode != inst.renderMode ||
		oldStyle != inst.sparkStyle ||
		!float64SlicesEqual(oldData, inst.data)
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		propKey:             inst.key,
		propData:            copyFloat64Slice(inst.data),
		propTitle:           inst.title,
		propWidth:           inst.width,
		propHeight:          inst.height,
		propInlineLabel:     inst.inlineLabel,
		propHighlightMinMax: inst.highlightMinMax,
		propAutoHeight:      inst.autoHeight,
		propRenderMode:      inst.renderMode,
		propStyle:           inst.sparkStyle,
	}
}

func (inst *Instance) MarkDirty()                         { inst.dirty = true }
func (inst *Instance) IsDirty() bool                      { return inst.dirty }
func (inst *Instance) GetContext() *rtui.ComponentContext { return nil }

func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	width, height, plotHeight := inst.contentSize(constraints)
	inst.measuredPlotHeight = plotHeight
	return layout.Size{
		Width:  constraints.ConstrainWidth(width),
		Height: constraints.ConstrainHeight(height),
	}
}

func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	plotHeight := inst.measuredPlotHeight
	if plotHeight <= 0 {
		plotHeight = inst.resolvePlotHeight(layout.UnboundedConstraints())
	}
	return chartcanvas.BufferToDrawCmds(inst.renderBuffer(plotHeight), x, y)
}

func (inst *Instance) resolveStyle() style.Style {
	s := inst.sparkStyle
	if s.FG == "" {
		s = s.Foreground(palette.SeriesColor(0))
	}
	return s
}

func (inst *Instance) minStyle() style.Style {
	return style.NewStyle().Foreground(palette.DownColor()).Merge(inst.sparkStyle)
}

func (inst *Instance) maxStyle() style.Style {
	return style.NewStyle().Foreground(palette.UpColor()).Merge(inst.sparkStyle)
}

func (inst *Instance) inlineLabelStyle() style.Style {
	return style.NewStyle().Foreground(palette.LabelColor()).Merge(inst.sparkStyle)
}

func (inst *Instance) titleStyle() style.Style {
	return style.NewStyle().Foreground(palette.LabelColor()).Merge(inst.sparkStyle)
}

func (inst *Instance) contentSize(constraints layout.Constraints) (int, int, int) {
	plotHeight := inst.resolvePlotHeight(constraints)
	width := inst.plotWidth() + inst.inlineLabelWidth()
	if titleWidth := paint.StringWidth(inst.title); titleWidth > width {
		width = titleWidth
	}
	height := plotHeight
	if strings.TrimSpace(inst.title) != "" {
		height++
	}
	return width, height, plotHeight
}

func (inst *Instance) resolvePlotHeight(constraints layout.Constraints) int {
	if inst.height > 0 {
		return inst.height
	}
	if !inst.autoHeight {
		return 1
	}

	titleRows := 0
	if strings.TrimSpace(inst.title) != "" {
		titleRows = 1
	}

	if constraints.MaxHeight > 0 && constraints.MaxHeight < layout.MaxInt {
		available := constraints.MaxHeight - titleRows
		if available < 1 {
			return 1
		}
		return available
	}
	if constraints.MinHeight > titleRows {
		return constraints.MinHeight - titleRows
	}
	return 1
}

func (inst *Instance) renderBuffer(plotHeight int) *paint.Buffer {
	width, height, _ := inst.contentSize(layout.TightConstraints(inst.plotWidth()+inst.inlineLabelWidth(), plotHeight+inst.titleRows()))
	buf := paint.NewBuffer(width, height)
	mode := inst.resolvedRenderMode(plotHeight)

	cursorY := 0
	if strings.TrimSpace(inst.title) != "" {
		buf.SetString(0, cursorY, inst.title, inst.titleStyle())
		cursorY++
	}

	if len(inst.data) == 0 {
		inst.renderInlineLabel(buf, cursorY, 0)
		return buf
	}

	sampled := inst.sampledData()
	minVal, maxVal := downsample.MinMaxFloat64(sampled)
	if plotHeight <= 1 {
		inst.renderSingleRow(buf, cursorY, sampled, minVal, maxVal, mode)
	} else {
		inst.renderMultiRow(buf, cursorY, sampled, minVal, maxVal, plotHeight, mode)
	}
	inst.renderInlineLabel(buf, cursorY+plotHeight-1, inst.plotWidth())
	return buf
}

func (inst *Instance) renderSingleRow(buf *paint.Buffer, rowY int, sampled []float64, minVal, maxVal float64, mode RenderMode) {
	degenerate := math.Abs(maxVal-minVal) < 1e-9
	for index, value := range sampled {
		glyph := inst.singleRowGlyph(value, minVal, maxVal, mode)
		cellStyle := inst.styleForSample(index, sampled, minVal, maxVal, degenerate)
		buf.SetCell(index, rowY, glyph, cellStyle)
	}
}

func (inst *Instance) renderMultiRow(buf *paint.Buffer, startY int, sampled []float64, minVal, maxVal float64, plotHeight int, mode RenderMode) {
	degenerate := math.Abs(maxVal-minVal) < 1e-9
	glyphs := glyphsForRenderMode(mode)
	for index, value := range sampled {
		totalUnits := inst.sampleUnits(value, minVal, maxVal, plotHeight, len(glyphs))
		for row := 0; row < plotHeight; row++ {
			glyph := multiRowGlyphForUnits(totalUnits, row, plotHeight, glyphs)
			if glyph == ' ' {
				continue
			}
			cellStyle := inst.styleForSample(index, sampled, minVal, maxVal, degenerate)
			buf.SetCell(index, startY+row, glyph, cellStyle)
		}
	}
}

func (inst *Instance) renderInlineLabel(buf *paint.Buffer, rowY, plotWidth int) {
	label := strings.TrimSpace(inst.inlineLabel)
	if label == "" {
		return
	}
	if plotWidth > 0 {
		buf.SetString(plotWidth+1, rowY, label, inst.inlineLabelStyle())
		return
	}
	buf.SetString(0, rowY, label, inst.inlineLabelStyle())
}

func (inst *Instance) styleForSample(index int, sampled []float64, minVal, maxVal float64, degenerate bool) style.Style {
	if !inst.highlightMinMax || degenerate {
		return inst.resolveStyle()
	}
	value := sampled[index]
	if nearlyEqual(value, maxVal) {
		return inst.maxStyle()
	}
	if nearlyEqual(value, minVal) {
		return inst.minStyle()
	}
	return inst.resolveStyle()
}

func (inst *Instance) sampledData() []float64 {
	width := inst.plotWidth()
	return downsample.ResampleNearestFloat64(inst.data, width)
}

func (inst *Instance) plotWidth() int {
	width := inst.width
	if width <= 0 {
		width = len(inst.data)
	}
	if width < 1 {
		width = 1
	}
	return width
}

func (inst *Instance) inlineLabelWidth() int {
	label := strings.TrimSpace(inst.inlineLabel)
	if label == "" {
		return 0
	}
	return 1 + paint.StringWidth(label)
}

func (inst *Instance) titleRows() int {
	if strings.TrimSpace(inst.title) == "" {
		return 0
	}
	return 1
}

func (inst *Instance) resolvedRenderMode(plotHeight int) RenderMode {
	switch inst.renderMode {
	case RenderModeBraille:
		if plotHeight > 1 {
			return RenderModeBlock
		}
		return RenderModeBraille
	case RenderModeBlock:
		return RenderModeBlock
	case RenderModeASCII:
		return RenderModeASCII
	case RenderModeAuto:
		if plotHeight > 1 {
			return RenderModeBlock
		}
		if inst.plotWidth() < 6 {
			return RenderModeASCII
		}
		return RenderModeBraille
	default:
		if plotHeight > 1 {
			return RenderModeBlock
		}
		return RenderModeBraille
	}
}

func glyphsForRenderMode(mode RenderMode) []rune {
	switch mode {
	case RenderModeBraille:
		return brailleGlyphs
	case RenderModeASCII:
		return asciiGlyphs
	case RenderModeBlock:
		return blockGlyphs
	default:
		return blockGlyphs
	}
}

func (inst *Instance) singleRowGlyph(value, minVal, maxVal float64, mode RenderMode) rune {
	glyphs := glyphsForRenderMode(mode)
	if math.Abs(maxVal-minVal) < 1e-9 {
		return glyphs[len(glyphs)/2]
	}
	level := int(math.Round(((value - minVal) / (maxVal - minVal)) * float64(len(glyphs)-1)))
	if level < 0 {
		level = 0
	}
	if level >= len(glyphs) {
		level = len(glyphs) - 1
	}
	return glyphs[level]
}

func (inst *Instance) sampleUnits(value, minVal, maxVal float64, plotHeight, glyphLevels int) int {
	if plotHeight < 1 {
		return 0
	}
	if glyphLevels < 1 {
		glyphLevels = len(blockGlyphs)
	}
	if math.Abs(maxVal-minVal) < 1e-9 {
		return int(math.Round(0.5 * float64(plotHeight*glyphLevels)))
	}
	ratio := (value - minVal) / (maxVal - minVal)
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	units := int(math.Round(ratio * float64(plotHeight*glyphLevels)))
	if units <= 0 {
		return 1
	}
	return units
}

func multiRowGlyphForUnits(totalUnits, rowIndex, plotHeight int, glyphs []rune) rune {
	if totalUnits <= 0 || plotHeight <= 0 {
		return ' '
	}
	if len(glyphs) == 0 {
		return ' '
	}
	rowBase := (plotHeight - rowIndex - 1) * len(glyphs)
	rowUnits := totalUnits - rowBase
	if rowUnits <= 0 {
		return ' '
	}
	if rowUnits >= len(glyphs) {
		return glyphs[len(glyphs)-1]
	}
	return glyphs[rowUnits-1]
}

func nearlyEqual(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}

func getDataProp(props rtui.Props, def []float64) []float64 {
	if value, ok := props[propData]; ok {
		if data, ok := value.([]float64); ok {
			return copyFloat64Slice(data)
		}
	}
	return copyFloat64Slice(def)
}

func getRenderModeProp(props rtui.Props, def RenderMode) RenderMode {
	if value, ok := props[propRenderMode]; ok {
		if mode, ok := value.(RenderMode); ok {
			return mode
		}
	}
	return def
}

func copyFloat64Slice(src []float64) []float64 {
	if len(src) == 0 {
		return nil
	}
	dst := make([]float64, len(src))
	copy(dst, src)
	return dst
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
