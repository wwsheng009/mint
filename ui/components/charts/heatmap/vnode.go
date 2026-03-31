package heatmap

import (
	fwtheme "github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

const (
	propKey              = "key"
	propTitle            = "title"
	propRowLabels        = "rowLabels"
	propColLabels        = "colLabels"
	propValues           = "values"
	propShowAxis         = "showAxis"
	propShowLegend       = "showLegend"
	propShowSummary      = "showSummary"
	propSummaryMode      = "summaryMode"
	propLegendMode       = "legendMode"
	propColorMode        = "colorMode"
	propScaleMode        = "scaleMode"
	propViewport         = "viewport"
	propMaxRowLabelWidth = "maxRowLabelWidth"
	propStyle            = "style"
)

// LegendMode controls how the heatmap legend is rendered.
type LegendMode int

const (
	LegendModeFull LegendMode = iota
	LegendModeCompact
)

// SummaryMode controls how the optional heatmap summary row is rendered.
type SummaryMode int

const (
	SummaryModeNone SummaryMode = iota
	SummaryModeCompact
	SummaryModeDetailed
)

// ScaleMode controls which value domain the heatmap uses for color scaling.
type ScaleMode int

const (
	ScaleModeGlobal ScaleMode = iota
	ScaleModeViewport
	ScaleModeAuto
)

// VNode is the declarative description of a heatmap component.
type VNode struct {
	*rtui.ElementVNode

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
}

var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// New creates a new heatmap VNode.
func New(values [][]float64) *VNode {
	return &VNode{
		ElementVNode: rtui.NewElement("heatmap"),
		values:       copyFloat64Matrix(values),
		showAxis:     true,
		showLegend:   true,
		summaryMode:  SummaryModeNone,
		legendMode:   LegendModeFull,
		colorMode:    fwtheme.NewTerminalColorCapabilities().GetMode(),
		scaleMode:    ScaleModeGlobal,
	}
}

func (v *VNode) Key() string                                  { return v.key }
func (v *VNode) SetKey(key string) rtui.VNode                 { v.key = key; return v }
func (v *VNode) Tag() string                                  { return "heatmap" }
func (v *VNode) Style() style.Style                           { return v.heatmapStyle }
func (v *VNode) SetStyle(s style.Style) rtui.VNode            { v.heatmapStyle = s; return v }
func (v *VNode) Children() []rtui.VNode                       { return nil }
func (v *VNode) SetChildren(children []rtui.VNode) rtui.VNode { return v }
func (v *VNode) GetLayer() rtui.Layer                         { return rtui.LayerBase }
func (v *VNode) SetLayer(l rtui.Layer) rtui.VNode             { return v }

func (v *VNode) Props() rtui.Props {
	return rtui.Props{
		propKey:              v.key,
		propTitle:            v.title,
		propRowLabels:        copyStringSlice(v.rowLabels),
		propColLabels:        copyStringSlice(v.colLabels),
		propValues:           copyFloat64Matrix(v.values),
		propShowAxis:         v.showAxis,
		propShowLegend:       v.showLegend,
		propShowSummary:      v.summaryMode != SummaryModeNone,
		propSummaryMode:      v.summaryMode,
		propLegendMode:       v.legendMode,
		propColorMode:        v.colorMode,
		propScaleMode:        v.scaleMode,
		propViewport:         v.viewport,
		propMaxRowLabelWidth: v.maxRowLabelWidth,
		propStyle:            v.heatmapStyle,
	}
}

func (v *VNode) SetProps(props rtui.Props) rtui.VNode {
	if key, ok := props[propKey].(string); ok {
		v.key = key
	}
	if title, ok := props[propTitle].(string); ok {
		v.title = title
	}
	if rowLabels, ok := props[propRowLabels].([]string); ok {
		v.rowLabels = copyStringSlice(rowLabels)
	}
	if colLabels, ok := props[propColLabels].([]string); ok {
		v.colLabels = copyStringSlice(colLabels)
	}
	if values, ok := props[propValues].([][]float64); ok {
		v.values = copyFloat64Matrix(values)
	}
	if showAxis, ok := props[propShowAxis].(bool); ok {
		v.showAxis = showAxis
	}
	if showLegend, ok := props[propShowLegend].(bool); ok {
		v.showLegend = showLegend
	}
	if summaryMode, ok := props[propSummaryMode].(SummaryMode); ok {
		v.summaryMode = summaryMode
	} else if showSummary, ok := props[propShowSummary].(bool); ok {
		if showSummary {
			v.summaryMode = SummaryModeDetailed
		} else {
			v.summaryMode = SummaryModeNone
		}
	}
	if legendMode, ok := props[propLegendMode].(LegendMode); ok {
		v.legendMode = legendMode
	}
	if colorMode, ok := props[propColorMode].(fwtheme.ColorMode); ok {
		v.colorMode = colorMode
	}
	if scaleMode, ok := props[propScaleMode].(ScaleMode); ok {
		v.scaleMode = scaleMode
	}
	if viewport, ok := props[propViewport].(Viewport); ok {
		v.viewport = normalizeViewport(viewport)
	}
	if maxRowLabelWidth, ok := props[propMaxRowLabelWidth].(int); ok {
		v.maxRowLabelWidth = maxRowLabelWidth
	}
	if s, ok := props[propStyle].(style.Style); ok {
		v.heatmapStyle = s
	}
	return v
}

func (v *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(v.Props())
}

func (v *VNode) SetTitle(title string) *VNode {
	v.title = title
	return v
}

func (v *VNode) SetRowLabels(labels []string) *VNode {
	v.rowLabels = copyStringSlice(labels)
	return v
}

func (v *VNode) SetColLabels(labels []string) *VNode {
	v.colLabels = copyStringSlice(labels)
	return v
}

func (v *VNode) SetValues(values [][]float64) *VNode {
	v.values = copyFloat64Matrix(values)
	return v
}

func (v *VNode) SetShowAxis(show bool) *VNode {
	v.showAxis = show
	return v
}

func (v *VNode) SetShowLegend(show bool) *VNode {
	v.showLegend = show
	return v
}

func (v *VNode) SetShowSummary(show bool) *VNode {
	if show {
		v.summaryMode = SummaryModeDetailed
	} else {
		v.summaryMode = SummaryModeNone
	}
	return v
}

func (v *VNode) SetSummaryMode(mode SummaryMode) *VNode {
	v.summaryMode = mode
	return v
}

func (v *VNode) SetLegendMode(mode LegendMode) *VNode {
	v.legendMode = mode
	return v
}

func (v *VNode) SetColorMode(mode fwtheme.ColorMode) *VNode {
	v.colorMode = mode
	return v
}

func (v *VNode) SetScaleMode(mode ScaleMode) *VNode {
	v.scaleMode = mode
	return v
}

func (v *VNode) SetViewport(viewport Viewport) *VNode {
	v.viewport = normalizeViewport(viewport)
	return v
}

func (v *VNode) SetMaxRowLabelWidth(width int) *VNode {
	v.maxRowLabelWidth = width
	return v
}

func (v *VNode) SetHeatmapStyle(s style.Style) *VNode {
	v.heatmapStyle = s
	return v
}

func (v *VNode) Title() string       { return v.title }
func (v *VNode) RowLabels() []string { return copyStringSlice(v.rowLabels) }
func (v *VNode) ColLabels() []string { return copyStringSlice(v.colLabels) }
func (v *VNode) Values() [][]float64 { return copyFloat64Matrix(v.values) }
func (v *VNode) ShowAxis() bool      { return v.showAxis }
func (v *VNode) ShowLegend() bool    { return v.showLegend }
func (v *VNode) ShowSummary() bool   { return v.summaryMode != SummaryModeNone }
func (v *VNode) SummaryMode() SummaryMode {
	return v.summaryMode
}
func (v *VNode) LegendMode() LegendMode {
	return v.legendMode
}
func (v *VNode) ColorMode() fwtheme.ColorMode {
	return v.colorMode
}
func (v *VNode) ScaleMode() ScaleMode  { return v.scaleMode }
func (v *VNode) Viewport() Viewport    { return v.viewport }
func (v *VNode) MaxRowLabelWidth() int { return v.maxRowLabelWidth }

func copyStringSlice(src []string) []string {
	if len(src) == 0 {
		return nil
	}
	dst := make([]string, len(src))
	copy(dst, src)
	return dst
}

func copyFloat64Matrix(src [][]float64) [][]float64 {
	if len(src) == 0 {
		return nil
	}
	dst := make([][]float64, len(src))
	for i := range src {
		if len(src[i]) == 0 {
			continue
		}
		row := make([]float64, len(src[i]))
		copy(row, src[i])
		dst[i] = row
	}
	return dst
}

func float64MatrixEqual(a, b [][]float64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if len(a[i]) != len(b[i]) {
			return false
		}
		for j := range a[i] {
			if a[i][j] != b[i][j] {
				return false
			}
		}
	}
	return true
}
