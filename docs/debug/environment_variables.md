# TUI Debug 环境变量参考

所有 Mint TUI 的调试环境变量都使用 `TUI_UI_DEBUG_*` 前缀，便于快速定位和统一管理。

## 📖 使用说明

**如果您想要快速使用示例**，请查看 [quick_start.md](quick_start.md)

**本文档用途**:
- ✅ 完整的环境变量参考
- ✅ 每个变量的详细说明
- ✅ 所有参数和选项
- ✅ 优先级和组合规则

**快速参考**: [README.md#环境变量速查](README.md#环境变量速查)

## 🎛️ 环境变量列表

### 布局调试

#### `TUI_UI_DEBUG_LAYOUT`

**类型**: Boolean
**默认**: false
**启用**: `true`

**说明**: 启用布局结构调试，输出完整的组件树、位置、尺寸、flex 等信息。

**示例**:
```bash
TUI_UI_DEBUG_LAYOUT=true ./demo2.exe
```

**输出内容**:
- 📋 完整布局树（层级结构）
- 🔘 所有组件的详细信息
- 📦 布局容器配置
- 📊 组件类型统计
- 🔍 按钮分布分析

---

### 渲染调试

#### `TUI_UI_DEBUG_PAINT`

**类型**: Boolean
**默认**: false
**启用**: `true`

**说明**: 启用渲染过程调试，输出 Paint 方法调用、文本对齐计算、DrawCmd 详情。

**示例**:
```bash
TUI_UI_DEBUG_PAINT=true ./demo2.exe
```

**输出内容**:
- `[DEBUG-PAINT]` - 按钮渲染详情
- `[DEBUG-ALIGN]` - 文本对齐计算
- `[DEBUG-RETURN]` - DrawCmd 返回值
- naturalWidth vs layoutWidth 对比
- 文本填充结果

---

### 引擎调试

#### `TUI_UI_DEBUG_ENGINE`

**类型**: Boolean
**默认**: false
**启用**: `true`

**说明**: 启用布局引擎调试，输出测量、计算、缓存等内部细节。

**示例**:
```bash
TUI_UI_DEBUG_ENGINE=true ./demo2.exe
```

**输出内容**:
- `[buildComputedBox]` - 节点构建
- `[Layout.Measure]` - 尺寸测量
- `[Layout.CacheSet/Get]` - 缓存操作
- `[Layout.Position]` - 位置计算
- 单次/多次测量选择逻辑

---

### 组合调试

#### `TUI_UI_DEBUG_ALL`

**类型**: Boolean
**默认**: false
**启用**: `true`

**说明**: 启用所有调试模式，等同于同时设置所有其他调试变量。

**示例**:
```bash
TUI_UI_DEBUG_ALL=true ./demo2.exe
```

**等价于**:
```bash
TUI_UI_DEBUG_LAYOUT=true \
TUI_UI_DEBUG_PAINT=true \
TUI_UI_DEBUG_ENGINE=true \
./demo2.exe
```

---

### 调试级别控制

#### `TUI_UI_DEBUG_LEVEL`

**类型**: String
**默认**: "info"
**可选值**: "error", "warn", "info", "debug", "trace"

**说明**: 控制调试输出的详细程度。

**示例**:
```bash
# 只显示错误
TUI_UI_DEBUG_LEVEL=error ./demo2.exe

# 显示详细信息
TUI_UI_DEBUG_LEVEL=debug ./demo2.exe

# 显示所有 trace 信息
TUI_UI_DEBUG_LEVEL=trace ./demo2.exe
```

---

### 输出控制

#### `TUI_UI_DEBUG_FILE`

**类型**: String (文件路径)
**默认**: "" (输出到 stderr)

**说明**: 将调试输出重定向到指定文件，而不是 stderr。

**示例**:
```bash
TUI_UI_DEBUG_FILE=debug.log ./demo2.exe

# 保存到特定目录
TUI_UI_DEBUG_FILE=/tmp/mint_debug.log ./demo2.exe
```

---

#### `TUI_UI_DEBUG_QUIET`

**类型**: Boolean
**默认**: false
**启用**: `true`

**说明**: 静默模式，只输出错误信息，抑制所有调试输出。

**示例**:
```bash
TUI_UI_DEBUG_QUIET=true ./demo2.exe
```

---

#### `TUI_UI_DEBUG_COLOR`

**类型**: Boolean
**默认**: true
**禁用**: `false`

**说明**: 控制调试输出是否包含 ANSI 颜色代码。

**示例**:
```bash
# 禁用颜色（便于日志处理）
TUI_UI_DEBUG_COLOR=false ./demo2.exe

# 启用颜色（默认）
TUI_UI_DEBUG_COLOR=true ./demo2.exe
```

---

## 🔧 使用示例

### 基本调试

```bash
# 启用布局调试
TUI_UI_DEBUG_LAYOUT=true ./demo2.exe

# 启用渲染调试
TUI_UI_DEBUG_PAINT=true ./demo2.exe

# 启用引擎调试
TUI_UI_DEBUG_ENGINE=true ./demo2.exe
```

### 组合调试

```bash
# 启用多个调试模式
TUI_UI_DEBUG_LAYOUT=true TUI_UI_DEBUG_PAINT=true ./demo2.exe

# 启用所有调试
TUI_UI_DEBUG_ALL=true ./demo2.exe
```

### 输出到文件

```bash
# 方法 1: 使用 shell 重定向
TUI_UI_DEBUG_LAYOUT=true ./demo2.exe > debug.txt 2>&1

# 方法 2: 使用 DEBUG_FILE 环境变量
TUI_UI_DEBUG_LAYOUT=true TUI_UI_DEBUG_FILE=debug.txt ./demo2.exe

# 同时查看和保存
TUI_UI_DEBUG_LAYOUT=true ./demo2.exe 2>&1 | tee debug.txt
```

### 过滤和格式化

```bash
# 禁用颜色（便于 grep）
TUI_UI_DEBUG_COLOR=false TUI_UI_DEBUG_LAYOUT=true ./demo2.exe

# 只看按钮信息
TUI_UI_DEBUG_LAYOUT=true ./demo2.exe 2>&1 | grep "按钮"

# 只看 Paint debug
TUI_UI_DEBUG_PAINT=true ./demo2.exe 2>&1 | grep "DEBUG-PAINT"
```

### 特定场景

```bash
# 调试按钮不均匀分布
TUI_UI_DEBUG_LAYOUT=true ./demo2.exe 2>&1 | grep -A 10 "按钮分布"

# 调试文本对齐问题
TUI_UI_DEBUG_PAINT=true ./demo2.exe 2>&1 | grep "DEBUG-ALIGN"

# 调试 Wrap 换行问题
TUI_UI_DEBUG_LAYOUT=true ./demo2.exe 2>&1 | grep "找到.*个 HStack"

# 性能分析
TUI_UI_DEBUG_ENGINE=true ./demo2.exe 2>&1 | grep "Cache"
```

---

## 📊 环境变量优先级

当多个环境变量同时设置时，优先级从高到低：

1. `TUI_UI_DEBUG_QUIET=true` - 最高优先级，会抑制所有其他调试输出
2. `TUI_UI_DEBUG_ALL=true` - 启用所有调试
3. `TUI_UI_DEBUG_LAYOUT/PAINT/ENGINE` - 单独控制
4. `TUI_UI_DEBUG_LEVEL` - 控制输出详细程度

**示例**:
```bash
# QUIET 会覆盖其他设置
TUI_UI_DEBUG_ALL=true TUI_UI_DEBUG_QUIET=true ./demo2.exe
# 结果：无调试输出

# ALL 会覆盖单独的设置
TUI_UI_DEBUG_LAYOUT=false TUI_UI_DEBUG_ALL=true ./demo2.exe
# 结果：启用布局调试（ALL 覆盖了 false）

# LEVEL 控制详细程度
TUI_UI_DEBUG_LAYOUT=true TUI_UI_DEBUG_LEVEL=error ./demo2.exe
# 结果：只输出布局相关的错误
```

---

## 🎯 快速参考表

| 环境变量 | 作用 | 输出位置 | 典型用途 |
|---------|------|---------|---------|
| `TUI_UI_DEBUG_LAYOUT` | 布局结构 | stderr | 调试组件位置、尺寸、flex |
| `TUI_UI_DEBUG_PAINT` | 渲染过程 | stderr | 调试文本对齐、渲染命令 |
| `TUI_UI_DEBUG_ENGINE` | 引擎计算 | stderr | 调试布局算法、缓存 |
| `TUI_UI_DEBUG_ALL` | 启用所有 | stderr | 快速启用完整调试 |
| `TUI_UI_DEBUG_LEVEL` | 输出级别 | stderr | 控制输出详细程度 |
| `TUI_UI_DEBUG_FILE` | 输出文件 | 文件 | 保存调试日志 |
| `TUI_UI_DEBUG_QUIET` | 静默模式 | - | 禁用所有调试输出 |
| `TUI_UI_DEBUG_COLOR` | 颜色控制 | stderr | 控制是否显示颜色 |

---

## 💡 最佳实践

### 开发时

```bash
# 启用完整调试，输出到文件
TUI_UI_DEBUG_ALL=true TUI_UI_DEBUG_FILE=debug.log ./demo2.exe
```

### 测试时

```bash
# 只启用布局调试，禁用颜色
TUI_UI_DEBUG_LAYOUT=true TUI_UI_DEBUG_COLOR=false ./demo2.exe
```

### 生产环境

```bash
# 禁用所有调试（默认）
./demo2.exe

# 或者明确禁用
TUI_UI_DEBUG_QUIET=true ./demo2.exe
```

### 调试特定问题

```bash
# 按钮布局问题
TUI_UI_DEBUG_LAYOUT=true ./demo2.exe 2>&1 | grep -A 5 "按钮"

# 文本对齐问题
TUI_UI_DEBUG_PAINT=true ./demo2.exe 2>&1 | grep "DEBUG-ALIGN"

# 性能问题
TUI_UI_DEBUG_ENGINE=true ./demo2.exe 2>&1 | grep "Cache"
```

---

## 🔄 从旧变量迁移

如果您之前使用了旧的环境变量名称，这里是迁移指南：

| 旧变量名 | 新变量名 | 状态 |
|---------|---------|------|
| `TUI_LAYOUT_DEBUG` | `TUI_UI_DEBUG_LAYOUT` | ⚠️ 旧名称，已废弃 |
| `TUI_PAINT_DEBUG` | `TUI_UI_DEBUG_PAINT` | ⚠️ 旧名称，已废弃 |
| `TUI_LAYOUT_ENGINE_DEBUG` | `TUI_UI_DEBUG_ENGINE` | ⚠️ 旧名称，已废弃 |
| `TUI_WRAP_DEBUG` | `TUI_UI_DEBUG_ENGINE` | ⚠️ 旧名称，已废弃 |

**迁移建议**: 旧变量仍然可用，但建议更新为新的 `TUI_UI_DEBUG_*` 前缀以保持一致性。

---

**版本**: v1.0
**最后更新**: 2025-02-08
**维护者**: Claude Sonnet 4.5
