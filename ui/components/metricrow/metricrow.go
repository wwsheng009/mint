package metricrow

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/panel"
	"github.com/wwsheng009/mint/ui/components/text"
)

// Item describes a compact key/value metric.
type Item struct {
	Title string
	Value interface{}
}

// Option configures MetricRow.
type Option func(*Config)

// Config stores MetricRow configuration.
type Config struct {
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

// New builds a horizontal row of compact metric panels.
func New(items []Item, opts ...Option) rtui.VNode {
	cfg := Config{
		ItemWidth:     20,
		Gap:           1,
		BorderStyle:   layout.BorderSingle,
		BorderColor:   style.Color("blue"),
		ValueStyle:    style.NewStyle().Foreground(style.Color("cyan")).Bold(true),
		HasValueStyle: true,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}

	nodes := make([]rtui.VNode, 0, len(items))
	for _, item := range items {
		valueBuilder := text.NewBuilder(formatValue(item.Value, cfg.FormatValue))
		if cfg.HasValueStyle {
			valueBuilder.Style(cfg.ValueStyle)
		}
		panelBuilder := panel.NewBuilder().
			Title(item.Title).
			BorderStyle(cfg.BorderStyle).
			BorderColor(cfg.BorderColor).
			Width(cfg.ItemWidth).
			Content(valueBuilder.Build())
		if cfg.HasPanelStyle {
			panelBuilder.Style(cfg.PanelStyle)
		}
		nodes = append(nodes, panelBuilder.Build())
	}
	return rtui.HStackBuilder(nodes...).Gap(cfg.Gap).Build()
}

func formatValue(value interface{}, formatter func(interface{}) string) string {
	if formatter != nil {
		return formatter(value)
	}
	if value == nil {
		return "-"
	}
	return fmt.Sprint(value)
}

func ItemWidth(width int) Option {
	return func(cfg *Config) {
		cfg.ItemWidth = width
	}
}

func Gap(gap int) Option {
	return func(cfg *Config) {
		cfg.Gap = gap
	}
}

func Border(borderStyle layout.BorderStyle, color style.Color) Option {
	return func(cfg *Config) {
		cfg.BorderStyle = borderStyle
		cfg.BorderColor = color
	}
}

func ValueStyle(valueStyle style.Style) Option {
	return func(cfg *Config) {
		cfg.ValueStyle = valueStyle
		cfg.HasValueStyle = true
	}
}

func PanelStyle(panelStyle style.Style) Option {
	return func(cfg *Config) {
		cfg.PanelStyle = panelStyle
		cfg.HasPanelStyle = true
	}
}

func Formatter(formatter func(interface{}) string) Option {
	return func(cfg *Config) {
		cfg.FormatValue = formatter
	}
}
