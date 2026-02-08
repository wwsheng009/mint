# TUI Debug 工具使用手册

Mint TUI 提供了一套强大的调试工具，帮助开发者快速定位和解决 UI 渲染问题。

## 📚 文档导航

| 文档 | 说明 | 适用场景 |
|------|------|---------|
| [quick_start.md](quick_start.md) | 快速入门指南 | 第一次使用、常用命令 |
| [environment_variables.md](environment_variables.md) | 环境变量参考 | 查询所有调试选项 |
| [layout_api.md](layout_api.md) | 布局调试 API | 编程方式使用调试功能 |
| [paint_debug.md](paint_debug.md) | 渲染调试详解 | 深入理解 Paint 过程 |

## 🎯 三种调试模式

### 1. Layout Debug - 布局结构调试

**用途**: 查看组件树、位置、尺寸、flex 属性

**启用**:
```bash
TUI_UI_DEBUG_LAYOUT=true ./demo2.exe
```

**查看详细用法**: [quick_start.md#1-布局调试](quick_start.md)

### 2. Paint Debug - 渲染过程调试

**用途**: 查看 Paint 方法、文本对齐、DrawCmd

**启用**:
```bash
TUI_UI_DEBUG_PAINT=true ./demo2.exe
```

**查看详细用法**: [quick_start.md#2-渲染调试](quick_start.md)

### 3. Engine Debug - 布局引擎调试

**用途**: 查看测量、计算、缓存等内部细节

**启用**:
```bash
TUI_UI_DEBUG_ENGINE=true ./demo2.exe
```

**查看详细用法**: [quick_start.md#3-引擎调试](quick_start.md)

## 🚀 快速开始

### 启用所有调试

```bash
TUI_UI_DEBUG_ALL=true ./demo2.exe
```

### 保存调试输出

```bash
# 保存到文件
TUI_UI_DEBUG_ALL=true ./demo2.exe > debug.txt 2>&1

# 同时查看和保存
TUI_UI_DEBUG_ALL=true ./demo2.exe 2>&1 | tee debug.txt
```

### 过滤特定信息

```bash
# 只看按钮信息
TUI_UI_DEBUG_LAYOUT=true ./demo2.exe 2>&1 | grep "按钮"

# 只看 Paint debug
TUI_UI_DEBUG_PAINT=true ./demo2.exe 2>&1 | grep "DEBUG-PAINT"
```

## 🔍 常见问题快速定位

### 问题: 按钮没有均匀分布

```bash
# Step 1: 检查布局结构
TUI_UI_DEBUG_LAYOUT=true ./demo2.exe 2>&1 | grep -A 5 "按钮分布"

# Step 2: 检查 Flex 值
TUI_UI_DEBUG_LAYOUT=true ./demo2.exe 2>&1 | grep "Flex:"

# Step 3: 检查实际宽度
TUI_UI_DEBUG_LAYOUT=true ./demo2.exe 2>&1 | grep "尺寸:"
```

**详细调试方法**: [quick_start.md#用例-1-验证-flex-布局](quick_start.md)

### 问题: 文本没有居中

```bash
# 检查文本对齐计算
TUI_UI_DEBUG_PAINT=true ./demo2.exe 2>&1 | grep "DEBUG-ALIGN"

# 查看 buttonText 最终长度
TUI_UI_DEBUG_PAINT=true ./demo2.exe 2>&1 | grep "buttonText len="
```

**详细调试方法**: [paint_debug.md#实际调试案例](paint_debug.md)

### 问题: Wrap 没有换行

```bash
# 查看 HStack 数量（每个 HStack 是一行）
TUI_UI_DEBUG_LAYOUT=true ./demo2.exe 2>&1 | grep "找到.*个 HStack"

# 查看每个 HStack 的子组件数量
TUI_UI_DEBUG_LAYOUT=true ./demo2.exe 2>&1 | grep -A 3 "子组件:"
```

**详细调试方法**: [quick_start.md#用例-2-调试-wrap-组件换行](quick_start.md)

## 📊 环境变量速查

| 变量 | 作用 |
|------|------|
| `TUI_UI_DEBUG_LAYOUT` | 布局结构调试 |
| `TUI_UI_DEBUG_PAINT` | 渲染过程调试 |
| `TUI_UI_DEBUG_ENGINE` | 引擎计算调试 |
| `TUI_UI_DEBUG_ALL` | 启用所有调试 |
| `TUI_UI_DEBUG_COLOR` | 控制颜色输出 |

**完整参考**: [environment_variables.md](environment_variables.md)

## 💡 使用技巧

### 保存输出便于分析

```bash
TUI_UI_DEBUG_LAYOUT=true ./demo2.exe > output.txt 2>&1
```

### 清理 ANSI 颜色代码

```bash
TUI_UI_DEBUG_COLOR=false TUI_UI_DEBUG_LAYOUT=true ./demo2.exe
```

### 只看关键信息

```bash
TUI_UI_DEBUG_LAYOUT=true ./demo2.exe 2>&1 | grep "✅"
```

**更多技巧**: [quick_start.md#过滤技巧](quick_start.md)

## 🎓 学习路径

### 新手入门

1. 阅读 [quick_start.md](quick_start.md) - 了解基本用法
2. 尝试常见场景的调试示例
3. 学会使用 grep 过滤输出

### 进阶使用

1. 阅读 [paint_debug.md](paint_debug.md) - 理解渲染原理
2. 学习 [layout_api.md](layout_api.md) - 编程方式使用
3. 查看 [environment_variables.md](environment_variables.md) - 所有选项

### 专家模式

1. 添加自定义 debug 输出
2. 编写调试脚本
3. 性能分析和优化

## 🔄 环境变量命名规则

所有调试环境变量都使用 `TUI_UI_DEBUG_*` 前缀：

- ✅ `TUI_UI_DEBUG_LAYOUT` - 新命名（推荐）
- ⚠️ `TUI_LAYOUT_DEBUG` - 旧命名（仍可用）

**迁移指南**: [MIGRATION.md](MIGRATION.md)

## 📖 扩展阅读

- [渲染流程说明](../report/rendering_flow_explained.md) - 理解两阶段渲染
- [Demo2 按钮布局分析](../report/demo2_button_layout_analysis.md) - 实战案例分析
- [布局调试 API 详解](layout_api.md) - 完整 API 参考

---

**需要帮助?**
- 快速问题: 查看 [quick_start.md](quick_start.md)
- 参数查询: 查看 [environment_variables.md](environment_variables.md)
- 深入学习: 查看 [paint_debug.md](paint_debug.md)

**版本**: v1.0
**最后更新**: 2025-02-08
**维护者**: Claude Sonnet 4.5
