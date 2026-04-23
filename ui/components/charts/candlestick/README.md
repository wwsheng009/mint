# Candlestick

`candlestick` 是 charts 组件族里的 OHLC 蜡烛图组件。

当前第一版范围：

- 单序列 OHLC 数据渲染
- 基础 `title / legend / grid / axis`
- 可选成交量子图
- 密集时间轴下的标签折叠
- 上涨 / 下跌 / 平盘三种语义样式
- 支持 `up / down / flat / wick / volume` 五类样式覆盖
- 基于宽度预算的离散采样

当前不包含：

- 多面板联动
- 交互缩放与窗口滚动
- 技术指标叠加

## 最小示例

```go
package main

import (
	"github.com/wwsheng009/mint/ui"
	"github.com/wwsheng009/mint/ui/components/charts/candlestick"
)

func main() {
	_ = ui.Run(func() ui.VNode {
		return candlestick.NewBuilder([]candlestick.Candle{
			{Label: "M", Open: 100, High: 110, Low: 96, Close: 107, Volume: 1800},
			{Label: "T", Open: 107, High: 112, Low: 101, Close: 103, Volume: 1200},
			{Label: "W", Open: 103, High: 116, Low: 99, Close: 111, Volume: 1500},
			{Label: "T", Open: 111, High: 118, Low: 108, Close: 109, Volume: 900},
		}).
			Title("Daily Tape").
			Width(9).
			Height(6).
			ShowLegend(true).
			ShowGrid(true).
			ShowAxis(true).
			ShowVolume(true).
			VolumeHeight(3).
			Build()
	})
}
```

完整可运行示例见 [examples/charts_candlestick_demo](/E:/projects/yao/wwsheng009/mint/examples/charts_candlestick_demo)。

## 设计说明

- `candlestick` 适合终端里的紧凑 OHLC 走势展示，不追求 Web K 线图那种高密度交互，而是优先保证在固定宽度字符栅格里可读。
- 蜡烛实体和影线分开处理，可以稳定表达上涨、下跌和平盘状态，同时保留 `wick` 的独立样式覆盖。
- 成交量子图是可选的第二面板，适合在不引入复杂布局系统的前提下补充交易活跃度。
- 时间轴标签已经支持密集场景下的折叠，因此更适合日线、周线或少量分钟级窗口，不适合特别长的连续序列。

## 推荐用法

- 用在紧凑金融面板、交易概览卡片或监控页里的价格带与成交量摘要。
- 打开 `ShowVolume(true)` 时，建议同时设置 `VolumeHeight(...)`，避免价格区被压得过薄。
- 当面板宽度较窄时，优先减少 candle 数量，而不是把过多标签硬塞进同一行。
- 如果业务上需要更强的语义强调，优先使用 `UpStyle(...) / DownStyle(...) / FlatStyle(...) / WickStyle(...) / VolumeStyle(...)`，不要直接依赖终端默认色。
