# Charts BulletChart Demo

这个示例提供一个最小的独立 `bulletchart` 入口，用来直接展示当前最重要的方向语义能力：

- `HigherIsBetter()`
- `LowerIsBetter()`
- `NeutralDirection()`

示例里分别对应三类常见指标：

- `Throughput`：越高越好
- `Latency Ceiling`：越低越好
- `Error Rate`：保持中性语义，不强行推断业务方向

运行：

```bash
go test ./examples/charts_bulletchart_demo
go run ./examples/charts_bulletchart_demo
```
