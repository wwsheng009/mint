package ui

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// MetricItem describes a compact key/value metric.
type MetricItem struct {
	Title string
	Value interface{}
}

// MetricRowOption configures MetricRow.
type MetricRowOption func(*MetricRowConfig)

// MetricRowConfig stores MetricRow configuration.
type MetricRowConfig struct {
	ItemWidth     int
	Gap           int
	BorderStyle   layout.BorderStyle
	BorderColor   style.Color
	ValueStyle    style.Style
	HasValueStyle bool
	PanelStyle    style.Style
	HasPanelStyle bool
	FormatValue   func(interface{}) string
}

// MetricRow builds a horizontal row of compact metric panels.
func MetricRow(items []MetricItem, opts ...MetricRowOption) rtui.VNode {
	cfg := MetricRowConfig{
		ItemWidth:   20,
		Gap:         1,
		BorderStyle: layout.BorderSingle,
		BorderColor: style.Color("blue"),
		ValueStyle:  style.NewStyle().Foreground(style.Color("cyan")).Bold(true),
	}
	cfg.HasValueStyle = true
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}

	nodes := make([]rtui.VNode, 0, len(items))
	for _, item := range items {
		valueText := formatMetricValue(item.Value, cfg.FormatValue)
		valueBuilder := NewTextBuilder(valueText)
		if cfg.HasValueStyle {
			valueBuilder.Style(cfg.ValueStyle)
		}
		content := valueBuilder.Build()
		panel := NewPanelBuilder().
			Title(item.Title).
			BorderStyle(cfg.BorderStyle).
			BorderColor(cfg.BorderColor).
			Width(cfg.ItemWidth).
			Content(content)
		if cfg.HasPanelStyle {
			panel.Style(cfg.PanelStyle)
		}
		nodes = append(nodes, panel.Build())
	}
	return rtui.HStackBuilder(nodes...).Gap(cfg.Gap).Build()
}

func formatMetricValue(value interface{}, formatter func(interface{}) string) string {
	if formatter != nil {
		return formatter(value)
	}
	if value == nil {
		return "-"
	}
	return fmt.Sprint(value)
}

// MetricRowItemWidth sets each metric panel width.
func MetricRowItemWidth(width int) MetricRowOption {
	return func(cfg *MetricRowConfig) {
		cfg.ItemWidth = width
	}
}

// MetricRowGap sets spacing between metric panels.
func MetricRowGap(gap int) MetricRowOption {
	return func(cfg *MetricRowConfig) {
		cfg.Gap = gap
	}
}

// MetricRowBorder sets metric panel border style and color.
func MetricRowBorder(borderStyle layout.BorderStyle, color style.Color) MetricRowOption {
	return func(cfg *MetricRowConfig) {
		cfg.BorderStyle = borderStyle
		cfg.BorderColor = color
	}
}

// MetricRowValueStyle sets the value text style.
func MetricRowValueStyle(valueStyle style.Style) MetricRowOption {
	return func(cfg *MetricRowConfig) {
		cfg.ValueStyle = valueStyle
		cfg.HasValueStyle = true
	}
}

// MetricRowPanelStyle sets the panel style.
func MetricRowPanelStyle(panelStyle style.Style) MetricRowOption {
	return func(cfg *MetricRowConfig) {
		cfg.PanelStyle = panelStyle
		cfg.HasPanelStyle = true
	}
}

// MetricRowFormatter sets a custom value formatter.
func MetricRowFormatter(formatter func(interface{}) string) MetricRowOption {
	return func(cfg *MetricRowConfig) {
		cfg.FormatValue = formatter
	}
}
