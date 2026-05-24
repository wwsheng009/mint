package ui

import (
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/metricrow"
)

type MetricItem = metricrow.Item
type MetricRowOption = metricrow.Option
type MetricRowConfig = metricrow.Config

func MetricRow(items []MetricItem, opts ...MetricRowOption) rtui.VNode {
	return metricrow.New(items, opts...)
}

func MetricRowItemWidth(width int) MetricRowOption {
	return metricrow.ItemWidth(width)
}

func MetricRowGap(gap int) MetricRowOption {
	return metricrow.Gap(gap)
}

func MetricRowBorder(borderStyle layout.BorderStyle, color style.Color) MetricRowOption {
	return metricrow.Border(borderStyle, color)
}

func MetricRowValueStyle(valueStyle style.Style) MetricRowOption {
	return metricrow.ValueStyle(valueStyle)
}

func MetricRowPanelStyle(panelStyle style.Style) MetricRowOption {
	return metricrow.PanelStyle(panelStyle)
}

func MetricRowFormatter(formatter func(interface{}) string) MetricRowOption {
	return metricrow.Formatter(formatter)
}
