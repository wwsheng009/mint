package metricrow

import (
	"fmt"
	"strings"

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
	Width int
}

// WithWidth overrides the row default width for this metric item.
func (i Item) WithWidth(width int) Item {
	i.Width = width
	return i
}

// Value creates a metric item with the standard blank-value fallback.
func Value(title string, value interface{}) Item {
	return FallbackValue(title, value, "-")
}

// FallbackValue creates a metric item and replaces nil or blank values with fallback.
func FallbackValue(title string, value interface{}, fallback string) Item {
	return Item{Title: title, Value: metricValueText(value, fallback)}
}

// CompactValue creates a display-width-bounded metric item with "-" fallback.
func CompactValue(title string, value interface{}, maxWidth int) Item {
	return CompactFallbackValue(title, value, "-", maxWidth)
}

// CompactFallbackValue creates a display-width-bounded metric item with fallback.
func CompactFallbackValue(title string, value interface{}, fallback string, maxWidth int) Item {
	return Item{Title: title, Value: text.CompactFallbackText(metricValueText(value, fallback), fallback, maxWidth)}
}

// Count creates a non-negative integer metric item.
func Count(title string, count int) Item {
	if count < 0 {
		count = 0
	}
	return Item{Title: title, Value: count}
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
		itemWidth := cfg.ItemWidth
		if item.Width > 0 {
			itemWidth = item.Width
		}
		valueBuilder := text.NewBuilder(formatValue(item.Value, cfg.FormatValue))
		if cfg.HasValueStyle {
			valueBuilder.Style(cfg.ValueStyle)
		}
		panelBuilder := panel.NewBuilder().
			Title(item.Title).
			BorderStyle(cfg.BorderStyle).
			BorderColor(cfg.BorderColor).
			Width(itemWidth).
			Content(valueBuilder.Build())
		if cfg.HasPanelStyle {
			panelBuilder.Style(cfg.PanelStyle)
		}
		nodes = append(nodes, panelBuilder.Build())
	}
	return rtui.HStackBuilder(nodes...).Gap(cfg.Gap).Build()
}

// Operational builds a standard metrics row for operations dashboards.
func Operational(items []Item) rtui.VNode {
	return New(items, ItemWidth(20), Gap(1))
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

func metricValueText(value interface{}, fallback string) string {
	fallback = strings.TrimSpace(fallback)
	if fallback == "" {
		fallback = "-"
	}
	if value == nil {
		return fallback
	}
	text := strings.TrimSpace(fmt.Sprint(value))
	if text == "" {
		return fallback
	}
	return text
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
