# Debug 工具快速使用指南

本文档专注于 Mint TUI debug 工具的实战使用和快速参考。

## 📚 前置知识

**如果您是第一次使用 debug 工具**，建议先阅读 [README.md](README.md) 了解：
- 三种调试模式的概览
- 基本使用方法
- 常见问题快速定位

**本文档适合**：
- ✅ 已经了解基本概念，需要快速查命令
- ✅ 遇到具体问题，需要调试示例
- ✅ 想要学习高级技巧和过滤方法

## 📖 相关文档

- [README.md](README.md) - Debug 工具总览和导航
- [environment_variables.md](environment_variables.md) - 完整环境变量参考
- [paint_debug.md](/docsArchive/paint_debug.md) - Paint 调试深入分析（已归档）
- [layout_api.md](/docsArchive/layout_api.md) - 布局调试 API 详解（已归档）

## 🚀 快速开始

### 1. 布局调试 (Layout Debug)

查看组件的布局信息（位置、尺寸、flex属性等）

```bash
# 启用布局调试
TUI_UI_DEBUG_LAYOUT=true ./demo2.exe

# 保存到文件
TUI_UI_DEBUG_LAYOUT=true ./demo2.exe > layout_output.txt 2>&1
```

**输出内容包括：**
- 📋 完整的布局树结构
- 🔘 所有按钮的详细信息（位置、尺寸、flex值）
- 📦 布局容器（HStack/VStack）信息
- 📊 组件类型统计
- 🔍 按钮分布分析（宽度是否均匀）

### 2. 渲染调试 (Paint Debug)

查看组件的渲染过程和文本对齐详情

```bash
# 启用 paint debug 模式
TUI_UI_DEBUG_PAINT=true ./demo2.exe
```

**输出内容包括：**
- 每个节点的渲染位置
- 组件是否实现了 Paintable 接口
- 按钮文本的宽度计算详情
- Padding 和对齐方式的计算过程

**示例输出：**
```
[DEBUG-PAINT] label="[1] Event", bounds=[3 12 19 1], x=3, y=12
[DEBUG-PAINT]   buttonText=">[ [1] Event ]", contentWidth=14, naturalWidth=14, layoutWidth=19
[DEBUG-PAINT]   willStretch=true
[DEBUG-PAINT]   final buttonText length=19, text=">[ [1] Event ]     "
```

### 3. 引擎调试 (Engine Debug)

查看布局引擎的详细计算过程

```bash
# 启用 engine debug 模式
TUI_UI_DEBUG_ENGINE=true ./demo2.exe
```

**输出内容包括：**
- 布局容器的单次/多次测量选择
- 子组件的尺寸测量结果
- Flex 布局的空间分配计算
- 缓存命中/未命中信息

**示例输出：**
```
[buildComputedBox] tag=hstack, childConstraints=4, using single-pass=true
[Layout.Measure] Element: constraints={19 19 0 1073741823}, size={19 1}
```

### 4. 启用所有调试

```bash
# 一次性启用所有调试模式
TUI_UI_DEBUG_ALL=true ./demo2.exe

# 或单独启用多个
TUI_UI_DEBUG_LAYOUT=true TUI_UI_DEBUG_PAINT=true TUI_UI_DEBUG_ENGINE=true ./demo2.exe
```

## 💡 实际使用示例

### 示例 1: 检查按钮是否均匀分布

```bash
TUI_UI_DEBUG_LAYOUT=true ./demo2.exe > layout_output.txt 2>&1
```

在输出中查找：
```
🔍 ControlPanel 按钮分布分析:
  ✅ 按钮均匀分布 (宽度差异 ≤ 1)
```

### 示例 2: 调试按钮文本对齐问题

```bash
TUI_UI_DEBUG_PAINT=true ./demo2.exe 2>&1 | grep "DEBUG-PAINT"
```

查看每个按钮的：
- `contentWidth`: 文本实际宽度
- `naturalWidth`: 不受约束时的宽度
- `layoutWidth`: 布局分配的宽度（flex拉伸后）
- 最终渲染的文本

### 示例 3: 查看 Wrap 组件换行情况

```bash
TUI_UI_DEBUG_LAYOUT=true ./demo2.exe 2>&1 | grep "找到.*个 HStack"
```

**示例输出：**
```
找到 2 个 HStack:  # 2行
  1. 路径=.0.0 位置: (1, 1) 尺寸: 98x1  (第一行)
  2. 路径=.0.1 位置: (1, 3) 尺寸: 98x1  (第二行)
```

### 示例 4: 使用 Debug API（编程方式）

```go
package main

import (
    "fmt"
    "github.com/wwsheng009/mint/runtime"
    "github.com/wwsheng009/mint/runtime/compute"
    "github.com/wwsheng009/mint/runtime/debug"
)

func main() {
    // 1. 创建 UI
    root := buildSimpleUI()

    // 2. 创建布局引擎
    engine := compute.NewEngine()

    // 3. 设置约束
    constraints := runtime.NewBoxConstraints(0, 100, 0, 35)

    // 4. 执行布局计算
    layout, err := engine.Layout(root, constraints)
    if err != nil {
        panic(err)
    }

    // 5. 提取布局信息
    tree := debug.GetLayoutTree(layout)

    // 6. 查找所有按钮
    buttons := debug.FindComponentsByType(tree, "button")
    for _, btn := range buttons {
        fmt.Printf("按钮: %s\n", btn.Label)
        fmt.Printf("  位置: (%d, %d)\n", btn.X, btn.Y)
        fmt.Printf("  尺寸: %dx%d\n", btn.Width, btn.Height)
        fmt.Printf("  Flex: %d\n", btn.Flex)
        fmt.Println()
    }
}
```

## 🛠️ 常见用例

### 用例 1: 验证 Flex 布局

**问题**: 按钮没有均匀分布

**调试步骤**:
```bash
# 1. 检查布局树
TUI_UI_DEBUG_LAYOUT=true ./demo2.exe

# 2. 查找按钮的 Flex 值
# 在输出中查找 "Flex: 1 ✅" 或 "Flex: 无 ❌"

# 3. 检查宽度计算
TUI_UI_DEBUG_PAINT=true ./demo2.exe 2>&1 | grep "layoutWidth"
```

**预期结果**:
- 所有按钮应该有 `Flex: 1`
- `layoutWidth` 应该相同（或差异 ≤ 1）
- 总宽度应该匹配容器宽度

### 用例 2: 调试 Wrap 组件换行

**问题**: Wrap 没有按预期换行

**调试步骤**:
```bash
# 查看布局树
TUI_UI_DEBUG_LAYOUT=true ./demo2.exe

# 查找 HStack 容器数量
# 每个 HStack 代表一行
# 例如：2 个 HStack = 2 行
```

**预期结果**:
```
找到 2 个 HStack:
  1. 路径=.0.0 位置: (1, 1) 尺寸: 98x1  (第一行)
  2. 路径=.0.1 位置: (1, 3) 尺寸: 98x1  (第二行)
```

### 用例 3: 检查组件重叠

**问题**: 组件相互重叠

**调试步骤**:
```go
// 查找特定位置的组件
x, y := 50, 10
if comp, found := debug.GetComponentAtPoint(tree, x, y); found {
    fmt.Printf("位置 (%d, %d) 的组件: %s\n", x, y, comp.Type)
    fmt.Printf("  尺寸: %dx%d\n", comp.Width, comp.Height)
}
```

## 📝 输出文件保存

> 💡 **提示**: 完整的环境变量列表请参考 [environment_variables.md](environment_variables.md)

## 📝 输出文件保存

```bash
# 保存到文件
TUI_UI_DEBUG_LAYOUT=true ./demo2.exe > layout.txt 2>&1

# 只保存 debug 输出（过滤掉 ANSI 转义码）
TUI_UI_DEBUG_LAYOUT=true ./demo2.exe 2>&1 | sed 's/\x1b\[[0-9;]*m//g' > clean_layout.txt

# 查看特定部分
TUI_UI_DEBUG_LAYOUT=true ./demo2.exe 2>&1 | grep -A 20 "🔘 所有按钮"
```

## 🎯 最佳实践

1. **先启用 layout debug** 查看整体结构
2. **再启用 paint debug** 查看具体渲染细节
3. **保存输出到文件** 便于分析
4. **使用 grep 过滤** 只关注相关组件
5. **对比预期和实际** 找出差异

## 🔍 过滤技巧

```bash
# 只看按钮信息
TUI_UI_DEBUG_LAYOUT=true ./demo2.exe 2>&1 | grep "按钮"

# 只看 Flex 值
TUI_UI_DEBUG_LAYOUT=true ./demo2.exe 2>&1 | grep "Flex:"

# 只看 Paint debug
TUI_UI_DEBUG_PAINT=true ./demo2.exe 2>&1 | grep "DEBUG-PAINT"

# 只看对齐信息
TUI_UI_DEBUG_PAINT=true ./demo2.exe 2>&1 | grep "DEBUG-ALIGN"

# 只看缓存信息
TUI_UI_DEBUG_ENGINE=true ./demo2.exe 2>&1 | grep "Cache"
```

## 🆘 调试提示

**如果输出过多**:
```bash
# 只查看关键信息
TUI_UI_DEBUG_LAYOUT=true ./demo2.exe 2>&1 | grep "✅"
```

**如果输出乱码（ANSI 转义码）**:
```bash
# 使用 sed 或 strings 清理
TUI_UI_DEBUG_LAYOUT=true ./demo2.exe 2>&1 | sed 's/\x1b\[[0-9;]*m//g'
```

**如果需要查看实时输出**:
```bash
# 使用 tee 同时显示和保存
TUI_UI_DEBUG_LAYOUT=true ./demo2.exe 2>&1 | tee output.txt
```

**如果需要禁用颜色**:
```bash
TUI_UI_DEBUG_COLOR=false TUI_UI_DEBUG_LAYOUT=true ./demo2.exe
```

## 📖 相关文档

- [README.md](README.md) - Debug 工具总览
- [environment_variables.md](environment_variables.md) - 完整环境变量参考
- [layout_api.md](/docsArchive/layout_api.md) - 布局调试 API 详细说明（已归档）
- [paint_debug.md](/docsArchive/paint_debug.md) - Paint 调试深入分析（已归档）

## 🔗 从旧变量迁移

| 旧变量名 | 新变量名 | 状态 |
|---------|---------|------|
| `TUI_LAYOUT_DEBUG` | `TUI_UI_DEBUG_LAYOUT` | ⚠️ 旧名称，仍可用 |
| `TUI_PAINT_DEBUG` | `TUI_UI_DEBUG_PAINT` | ⚠️ 旧名称，仍可用 |
| `TUI_LAYOUT_ENGINE_DEBUG` | `TUI_UI_DEBUG_ENGINE` | ⚠️ 旧名称，仍可用 |
| `TUI_WRAP_DEBUG` | `TUI_UI_DEBUG_ENGINE` | ⚠️ 旧名称，仍可用 |

**建议**: 逐步迁移到新的 `TUI_UI_DEBUG_*` 前缀以保持一致性。

---

**版本**: v1.0
**最后更新**: 2025-02-08
**维护者**: Claude Sonnet 4.5
