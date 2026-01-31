# Mint UI 性能基准测试规范

**版本**: v1.0
**日期**: 2026-01-31

---

## 目录

1. [性能目标](#一性能目标)
2. [基准测试方法](#二基准测试方法)
3. [性能指标](#三性能指标)
4. [测试场景](#四测试场景)
5. [性能分析工具](#五性能分析工具)
6. [优化策略](#六优化策略)
7. [性能回归检测](#七性能回归检测)

---

## 一、性能目标

### 1.1 核心指标

| 指标 | 目标值 | 测量方法 |
|------|--------|---------|
| **渲染帧率** | ≥ 60 FPS | 帧间隔 < 16.6ms |
| **布局计算** | < 1 ms | 基准测试 |
| **Diff 算法** | < 5 ms (1000 节点) | 基准测试 |
| **Style Diff 优化** | 输出减少 ≥ 95% | 字节数对比 |
| **组件数量** | ≥ 10,000 | 压力测试 |
| **虚拟滚动** | ≥ 100,000 项 | 列表测试 |
| **内存占用** | < 50 MB | 空闲应用 |
| **启动时间** | < 100 ms | 冷启动测试 |

### 1.2 响应时间目标

| 操作 | 目标时间 | 说明 |
|------|---------|------|
| 按键响应 | < 16 ms | 用户无感知延迟 |
| 点击响应 | < 50 ms | 即时反馈 |
| 页面切换 | < 100 ms | 流畅体验 |
| 大列表渲染 | < 200 ms | 首次渲染 |
| 表单提交 | < 50 ms | 即时处理 |

---

## 二、基准测试方法

### 2.1 测试环境

```go
// 测试配置
type BenchmarkConfig struct {
    // 终端尺寸
    Width  int  // 默认: 80
    Height int  // 默认: 24

    // 测试时长
    Duration time.Duration // 默认: 5s

    // 暖身轮数
    Warmup int // 默认: 3

    // 迭代次数
    Iterations int // 默认: 根据时长自动计算

    // GC 设置
    GC bool // 默认: true (每次迭代前 GC)
}
```

### 2.2 基准测试模板

```go
// framework/benchmark/benchmark_test.go

package benchmark

import (
    "testing"
    "time"
)

// BenchMark 基准测试辅助函数
func BenchMark(b *testing.B, fn func()) {
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        fn()
    }
}

// BenchMarkTimer 带计时器的基准测试
func BenchMarkTimer(b *testing.B, fn func(*testing.B)) {
    b.ResetTimer()
    fn(b)
}
```

### 2.3 运行基准测试

```bash
# 运行所有基准测试
go test ./framework/benchmark/... -bench=. -benchmem

# 运行特定测试
go test ./framework/benchmark/... -bench=BenchmarkRender -benchmem

# 运行多次取平均值
go test ./framework/benchmark/... -bench=. -benchmem -count=5

# 生成 CPU profile
go test ./framework/benchmark/... -bench=. -cpuprofile=cpu.prof

# 生成内存 profile
go test ./framework/benchmark/... -bench=. -memprofile=mem.prof
```

---

## 三、性能指标

### 3.1 渲染性能

```go
// framework/benchmark/render_test.go

package benchmark

import (
    "testing"
    "github.com/wwsheng009/mint/ui"
)

// BenchmarkSimpleRender 简单渲染基准测试
func BenchmarkSimpleRender(b *testing.B) {
    app := func() ui.VNode {
        return ui.Text("Hello, World!")
    }

    BenchMark(b, func() {
        // 模拟渲染
        node := app()
        _ = node.Build() // 构建 VNode 树
    })
}

// BenchmarkDeepRender 深度嵌套渲染
func BenchmarkDeepRender(b *testing.B) {
    depth := 10
    app := func() ui.VNode {
        return deepVStack(depth)
    }

    BenchMark(b, func() {
        node := app()
        _ = node.Build()
    })
}

// BenchmarkWideRender 宽度渲染
func BenchmarkWideRender(b *testing.B) {
    width := 100
    app := func() ui.VNode {
        children := make([]ui.VNode, width)
        for i := 0; i < width; i++ {
            children[i] = ui.Text("Item")
        }
        return ui.HStack(children...)
    }

    BenchMark(b, func() {
        node := app()
        _ = node.Build()
    })
}
```

### 3.2 Diff 性能

```go
// framework/benchmark/diff_test.go

package benchmark

import (
    "testing"
    "github.com/wwsheng009/mint/framework/reconciler"
)

// BenchmarkDiffSame Diff 相同节点
func BenchmarkDiffSame(b *testing.B) {
    old := ui.Text("Hello")
    new := ui.Text("Hello")

    BenchMark(b, func() {
        _ = reconciler.Diff(old, new)
    })
}

// BenchmarkDiffChangedText Diff 文本变化
func BenchmarkDiffChangedText(b *testing.B) {
    old := ui.Text("Hello")
    new := ui.Text("World")

    BenchMark(b, func() {
        _ = reconciler.Diff(old, new)
    })
}

// BenchmarkDiffList Diff 列表
func BenchmarkDiffList(b *testing.B) {
    n := 100
    oldItems := make([]ui.VNode, n)
    newItems := make([]ui.VNode, n)

    for i := 0; i < n; i++ {
        oldItems[i] = ui.Text(fmt.Sprintf("Item %d", i))
        newItems[i] = ui.Text(fmt.Sprintf("Item %d", i))
    }

    old := ui.VStack(oldItems...)
    new := ui.VStack(newItems...)

    BenchMark(b, func() {
        _ = reconciler.Diff(old, new)
    })
}

// BenchmarkDiffListWithKey 带 Key 的列表 Diff
func BenchmarkDiffListWithKey(b *testing.B) {
    n := 100
    oldItems := make([]ui.VNode, n)
    newItems := make([]ui.VNode, n)

    for i := 0; i < n; i++ {
        oldItems[i] = ui.Key(fmt.Sprintf("key-%d", i), ui.Text(fmt.Sprintf("Item %d", i)))
        newItems[i] = ui.Key(fmt.Sprintf("key-%d", i), ui.Text(fmt.Sprintf("Item %d", i)))
    }

    old := ui.VStack(oldItems...)
    new := ui.VStack(newItems...)

    BenchMark(b, func() {
        _ = reconciler.Diff(old, new)
    })
}
```

### 3.2.1 Style Diff 性能 (新增)

```go
// framework/benchmark/style_diff_test.go

package benchmark

import (
    "testing"
    "github.com/wwsheng009/mint/framework/render"
)

// BenchmarkStyleDiffNoChange 样式无变化
func BenchmarkStyleDiffNoChange(b *testing.B) {
    state := render.NewTerminalState()
    style := render.Style{
        FgColor: color.White,
        BgColor: color.Black,
        Bold:    true,
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = render.DiffStyles(style, state)
    }
}

// BenchmarkStyleDiffFullChange 样式完全变化
func BenchmarkStyleDiffFullChange(b *testing.B) {
    state := render.NewTerminalState()
    oldStyle := render.Style{
        FgColor: color.White,
        BgColor: color.Black,
    }
    newStyle := render.Style{
        FgColor: color.Red,
        BgColor: color.Blue,
        Bold:    true,
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = render.DiffStyles(newStyle, state)
        state.Apply(oldStyle)
    }
}

// BenchmarkRLEEncoding RLE 编码性能
func BenchmarkRLEEncoding(b *testing.B) {
    data := make([]byte, 1000)
    for i := range data {
        data[i] = 'A'
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = render.EncodeRLE(data)
    }
}

// BenchmarkStyleDiffReduction 测试输出缩减率
func BenchmarkStyleDiffReduction(b *testing.B) {
    // 生成大量相同样式的文本
    items := make([]ui.VNode, 100)
    for i := 0; i < 100; i++ {
        items[i] = ui.Text("Item").FgColor(color.White).Bold(true)
    }
    vnode := ui.VStack(items...)

    // 渲染到 DrawCmd
    cmds := render.VNodeToDrawCmds(vnode)

    // 优化前后对比
    before := render.Serialize(cmds)               // 原始输出
    after := render.OptimizeStyleDiff(cmds)       // 优化后

    reduction := 1.0 - float64(len(after))/float64(len(before))

    b.ReportMetric(reduction*100, "pct_reduction")
}
```

### 3.3 Hooks 性能

```go
// framework/benchmark/hooks_test.go

package benchmark

import (
    "testing"
    "github.com/wwsheng009/mint/framework/hooks"
)

// BenchmarkUseState useState 性能
func BenchmarkUseState(b *testing.B) {
    BenchMark(b, func() {
        ctx := hooks.NewContext()
        value, setValue := hooks.UseState(ctx, 0)
        _ = value
        _ = setValue
    })
}

// BenchmarkUseStateUpdate useState 更新性能
func BenchmarkUseStateUpdate(b *testing.B) {
    ctx := hooks.NewContext()
    _, setValue := hooks.UseState(ctx, 0)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        setValue(i)
    }
}

// BenchmarkUseMemo useMemo 性能
func BenchmarkUseMemo(b *testing.B) {
    ctx := hooks.NewContext()
    expensive := func() interface{} {
        return 42
    }

    BenchMark(b, func() {
        _ = hooks.UseMemo(ctx, expensive, nil)
    })
}

// BenchmarkUseMemoCached useMemo 缓存命中
func BenchmarkUseMemoCached(b *testing.B) {
    ctx := hooks.NewContext()
    expensive := func() interface{} {
        return 42
    }
    hooks.UseMemo(ctx, expensive, nil) // 首次调用

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        hooks.UseMemo(ctx, expensive, nil) // 应该从缓存返回
    }
}
```

### 3.4 布局性能

```go
// framework/benchmark/layout_test.go

package benchmark

import (
    "testing"
    "github.com/wwsheng009/mint/runtime/layout"
)

// BenchmarkMeasure 测量性能
func BenchmarkMeasure(b *testing.B) {
    node := createTestNode(100) // 100 个子节点
    constraint := layout.Constraint{
        MinWidth:  0,
        MaxWidth:  80,
        MinHeight: 0,
        MaxHeight: 24,
    }

    BenchMark(b, func() {
        _ = node.Measure(constraint)
    })
}

// BenchmarkFlexLayout Flex 布局性能
func BenchmarkFlexLayout(b *testing.B) {
    children := make([]layout.Node, 50)
    for i := 0; i < 50; i++ {
        children[i] = createTestNode(1)
    }

    container := createFlexContainer(layout.DirectionRow, children)

    BenchMark(b, func() {
        container.Layout(0, 0, 80, 24)
    })
}
```

---

## 四、测试场景

### 4.1 静态场景

```go
// 场景：静态页面，无交互
func BenchmarkStaticPage(b *testing.B) {
    app := func() ui.VNode {
        return ui.VStack(
            ui.Text("Title").Bold(true),
            ui.Separator(),
            ui.Text("Content line 1"),
            ui.Text("Content line 2"),
            ui.Text("Content line 3"),
        )
    }

    BenchMark(b, func() {
        node := app()
        _ = node.Build()
    })
}
```

### 4.2 动态更新场景

```go
// 场景：状态频繁更新
func BenchmarkFrequentUpdate(b *testing.B) {
    app := func() ui.VNode {
        count, _ := hooks.UseStateInt(0)
        return ui.Text(fmt.Sprintf("Count: %d", count))
    }

    ctx := hooks.NewContext()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, setValue := hooks.UseStateInt(ctx, i)
        setValue(i)
        node := app()
        _ = node.Build()
    }
}
```

### 4.3 大列表场景

```go
// 场景：大量数据列表
func BenchmarkLargeList(b *testing.B) {
    n := 10000
    items := make([]string, n)
    for i := 0; i < n; i++ {
        items[i] = fmt.Sprintf("Item %d", i)
    }

    app := func() ui.VNode {
        return ui.VirtualList(items).
            ItemHeight(1).
            RenderItem(func(item interface{}) ui.VNode {
                return ui.Text(item.(string))
            })
    }

    BenchMark(b, func() {
        node := app()
        _ = node.Build()
    })
}
```

### 4.4 复杂表单场景

```go
// 场景：复杂表单
func BenchmarkComplexForm(b *testing.B) {
    app := func() ui.VNode {
        return ui.VStack(
            ui.Input("Name"),
            ui.Input("Email"),
            ui.Input("Phone"),
            ui.CheckBox("Subscribe"),
            ui.Select([]string{"Option A", "Option B", "Option C"}),
            ui.TextArea("Message"),
            ui.Button("Submit"),
        )
    }

    BenchMark(b, func() {
        node := app()
        _ = node.Build()
    })
}
```

### 4.5 嵌套组件场景

```go
// 场景：深度嵌套组件
func BenchmarkNestedComponents(b *testing.B) {
    // 创建 10 层嵌套
    var nested func(int) ui.VNode
    nested = func(depth int) ui.VNode {
        if depth == 0 {
            return ui.Text("Leaf")
        }
        return ui.VStack(
            ui.Text(fmt.Sprintf("Level %d", depth)),
            nested(depth - 1),
        )
    }

    app := func() ui.VNode {
        return nested(10)
    }

    BenchMark(b, func() {
        node := app()
        _ = node.Build()
    })
}
```

---

## 五、性能分析工具

### 5.1 CPU Profile

```go
// 启用 CPU profiling
func TestWithCPUProfile(t *testing.T) {
    f, _ := os.Create("cpu.prof")
    defer f.Close()
    pprof.StartCPUProfile(f)
    defer pprof.StopCPUProfile()

    // 运行测试代码
    runTest()
}
```

```bash
# 分析 CPU profile
go tool pprof cpu.prof

# 常用命令
(pprof) top10        # 显示前 10 个热点
(pprof) list render  # 显示 render 函数的详细分析
(pprof) web          # 生成可视化图表（需要 graphviz）
```

### 5.2 Memory Profile

```go
// 启用内存 profiling
func TestWithMemProfile(t *testing.T) {
    f, _ := os.Create("mem.prof")
    defer f.Close()
    runtime.GC()
    pprof.WriteHeapProfile(f)

    // 运行测试代码
    runTest()
}
```

```bash
# 分析内存 profile
go tool pprof mem.prof

# 常用命令
(pprof) top           # 显示内存分配热点
(pprof) list render   # 显示 render 函数的内存分配
(pprof) web           # 生成可视化图表
```

### 5.3 内存分配追踪

```bash
# 开启内存分配追踪
go test ./... -bench=. -memprofile=mem.prof -alloc-object

# 查看分配对象数量
go tool pprof -alloc_objects mem.prof
```

### 5.4 逃逸分析

```bash
# 查看变量逃逸情况
go build -gcflags="-m" ./...

# 更详细的逃逸分析
go build -gcflags="-m -m" ./...
```

### 5.5 竞态检测

```bash
# 运行竞态检测
go test ./... -race

# 运行应用并检测竞态
go run -race main.go
```

---

## 六、优化策略

### 6.1 VNode 优化

```go
// ❌ 错误：每次渲染创建新切片
func BadComponent() ui.VNode {
    items := []string{"A", "B", "C"}
    children := make([]ui.VNode, len(items))
    for i, item := range items {
        children[i] = ui.Text(item)
    }
    return ui.VStack(children...)
}

// ✅ 正确：使用 For 减少分配
func GoodComponent() ui.VNode {
    items := []string{"A", "B", "C"}
    return ui.VStack(
        ui.For(items, func(i int, item string) ui.VNode {
            return ui.Text(item)
        }),
    )
}
```

### 6.2 Key 优化

```go
// ❌ 错误：使用索引作为 Key
func BadList(items []Item) ui.VNode {
    return ui.VStack(
        ui.For(items, func(i int, item Item) ui.VNode {
            return ui.Text(item.Name).Key(strconv.Itoa(i))
        }),
    )
}

// ✅ 正确：使用稳定的 ID
func GoodList(items []Item) ui.VNode {
    return ui.VStack(
        ui.For(items, func(i int, item Item) ui.VNode {
            return ui.Text(item.Name).Key(item.ID)
        }),
    )
}
```

### 6.3 useMemo 优化

```go
// ❌ 错误：每次重新计算
func BadComponent(items []Item) ui.VNode {
    filtered := filterItems(items) // 每次渲染都计算
    return ui.List(filtered)
}

// ✅ 正确：使用 useMemo 缓存
func GoodComponent(items []Item) ui.VNode {
    filtered := ui.UseMemo(func() interface{} {
        return filterItems(items)
    }, []interface{}{items})

    return ui.List(filtered.([]Item))
}
```

### 6.4 useCallback 优化

```go
// ❌ 错误：每次创建新函数
func BadComponent() ui.VNode {
    count, setCount := ui.UseStateInt(0)

    return ui.Button("Click").OnClick(func() {
        setCount(count + 1) // 新函数，导致子组件重新渲染
    })
}

// ✅ 正确：使用 useCallback
func GoodComponent() ui.VNode {
    count, setCount := ui.UseStateInt(0)

    handleClick := ui.UseCallback(func() {
        setCount(count + 1)
    }, []interface{}{count})

    return ui.Button("Click").OnClick(handleClick)
}
```

### 6.5 虚拟化优化

```go
// ❌ 错误：渲染所有项
func BadList(items []Item) ui.VNode {
    return ui.VStack(
        ui.For(items, func(i int, item Item) ui.VNode {
            return ui.Text(item.Name)
        }),
    )
}

// ✅ 正确：使用虚拟化
func GoodList(items []Item) ui.VNode {
    return ui.VirtualList(items).
        ItemHeight(1).
        RenderItem(func(item interface{}) ui.VNode {
            return ui.Text(item.(Item).Name)
        })
}
```

---

## 七、性能回归检测

### 7.1 基准线文件

```go
// framework/benchmark/baseline.go
//go:build ignore

package main

import (
    "encoding/json"
    "os"
    "testing"
)

// BenchmarkResult 基准测试结果
type BenchmarkResult struct {
    Name         string  `json:"name"`
    Iterations   int     `json:"iterations"`
    NsPerOp      int64   `json:"ns_per_op"`
    AllocedBytesPerOp uint64 `json:"alloced_bytes_per_op"`
    AllocsPerOp  uint64  `json:"allocs_per_op"`
}

// SaveBaseline 保存基准线
func SaveBaseline(results []BenchmarkResult, filename string) error {
    data, err := json.MarshalIndent(results, "", "  ")
    if err != nil {
        return err
    }
    return os.WriteFile(filename, data, 0644)
}

// LoadBaseline 加载基准线
func LoadBaseline(filename string) ([]BenchmarkResult, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, err
    }

    var results []BenchmarkResult
    err = json.Unmarshal(data, &results)
    return results, err
}
```

### 7.2 回归检测脚本

```bash
#!/bin/bash
# scripts/check_regression.sh

# 运行基准测试
go test ./framework/benchmark/... -bench=. -benchmem > current.txt

# 与基准线比较
benchstat baseline.txt current.txt

# 检查是否超过阈值
THRESHOLD=10  # 10% 阈值
if benchstat -threshold=$THRESHOLD baseline.txt current.txt | grep -q "~"; then
    echo "✅ 性能正常"
else
    echo "❌ 性能回归！"
    exit 1
fi
```

### 7.3 CI 集成

```yaml
# .github/workflows/benchmark.yml

name: Benchmark

on:
  pull_request:
    branches: [main]

jobs:
  benchmark:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
        with:
          fetch-depth: 0  # 获取历史记录用于比较

      - name: Setup Go
        uses: actions/setup-go@v3
        with:
          go-version: '1.23'

      - name: Run benchmarks
        run: |
          go test ./framework/benchmark/... -bench=. -benchmem | tee benchmark.txt

      - name: Compare with baseline
        run: |
          git checkout main
          go test ./framework/benchmark/... -bench=. -benchmem | tee baseline.txt

          git checkout -
          benchstat baseline.txt benchmark.txt

      - name: Upload results
        uses: actions/upload-artifact@v3
        with:
          name: benchmark-results
          path: benchmark.txt
```

---

## 八、性能检查清单

### 8.1 开发阶段

- [ ] 使用 Go benchmark 测试关键路径
- [ ] 检查内存分配是否合理
- [ ] 使用 pprof 分析 CPU 热点
- [ ] 检查是否有逃逸到堆的变量
- [ ] 运行竞态检测

### 8.2 发布前

- [ ] 运行完整基准测试套件
- [ ] 与上一版本比较性能
- [ ] 确认无性能回归
- [ ] 检查内存泄漏
- [ ] 验证压力测试通过

### 8.3 持续监控

- [ ] CI 中运行基准测试
- [ ] 定期更新基准线
- [ ] 监控生产环境性能
- [ ] 收集用户反馈

---

## 九、快速参考

### 9.1 常用命令

```bash
# 基准测试
go test ./... -bench=. -benchmem

# CPU profile
go test ./... -bench=. -cpuprofile=cpu.prof
go tool pprof cpu.prof

# Memory profile
go test ./... -bench=. -memprofile=mem.prof
go tool pprof mem.prof

# 竞态检测
go test ./... -race

# 逃逸分析
go build -gcflags="-m" ./...

# 比较基准
benchstat baseline.txt current.txt
```

### 9.2 性能阈值

| 指标 | 警告 | 严重 |
|------|------|------|
| 渲染帧率 | < 60 FPS | < 30 FPS |
| 布局计算 | > 1 ms | > 5 ms |
| 内存增长 | > 10% | > 20% |
| CPU 使用 | > 50% | > 80% |

---

**文档结束**

**版本历史**:
- v1.0 (2026-01-31): 初始版本
