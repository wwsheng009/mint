# OptionGroup 组件文档

> **最后更新**: 2026-03-06
> **版本**: 1.0
> **状态**: ✅ 分析完成，🚧 准备实施

---

## 📖 文档概览

本目录包含 OptionGroup 组件的完整分析和实施文档：

| 文档 | 描述 | 目标读者 |
|------|------|---------|
| [ARCHITECTURE_ANALYSIS_REPORT.md](./ARCHITECTURE_ANALYSIS_REPORT.md) | 深度架构分析和方案对比 | 架构师、高级开发者 |
| [CURRENT_STATUS.md](./CURRENT_STATUS.md) | 当前系统现状和问题清单 | 开发者、测试人员 |
| [IMPLEMENTATION_GUIDE.md](./IMPLEMENTATION_GUIDE.md) | 详细实施指南和步骤 | 实施开发者 |
| [README.md](./README.md) | 文档入口和快速导航 | （本文件） |

---

## 🎯 快速导航

### 我想了解...

#### 🔍 问题是什么？

→ 阅读 [ARCHITECTURE_ANALYSIS_REPORT.md - 问题分析](./ARCHITECTURE_ANALYSIS_REPORT.md#-问题分析)

#### 💡 为什么会出现这个问题？

→ 阅读 [ARCHITECTURE_ANALYSIS_REPORT.md - 时序问题详解](./ARCHITECTURE_ANALYSIS_REPORT.md#-根本原因时序问题)

#### 🌐 React 是如何解决的？

→ 阅读 [ARCHITECTURE_ANALYSIS_REPORT.md - React 解决方案对比](./ARCHITECTURE_ANALYSIS_REPORT.md#-react-如何处理类似问题)

#### 🏗️ 当前架构是什么样子的？

→ 阅读 [ARCHITECTURE_ANALYSIS_REPORT.md - Mint Fiber-first 架构特点](./ARCHITECTURE_ANALYSIS_REPORT.md#-mint-fiber-first-架构特点)

#### 📊 有哪些可用的解决方案？

→ 阅读 [ARCHITECTURE_ANALYSIS_REPORT.md - Mint 解决方案对比](./ARCHITECTURE_ANALYSIS_REPORT.md#-mint-解决方案对比)

#### ✅ 哪个方案最推荐？

→ 阅读 [ARCHITECTURE_ANALYSIS_REPORT.md - 推荐方案](./ARCHITECTURE_ANALYSIS_REPORT.md#-推荐方案)

#### 📋 当前系统状态如何？

→ 阅读 [CURRENT_STATUS.md - 已实现功能清单](./CURRENT_STATUS.md#-已实现功能清单)

#### 🔴 已知有哪些问题？

→ 阅读 [CURRENT_STATUS.md - 已知问题](./CURRENT_STATUS.md#-已知问题)

#### 📁 有哪些相关文件？

→ 阅读 [CURRENT_STATUS.md - 文件清单](./CURRENT_STATUS.md#-文件清单)

#### 🔧 如何实施修复？

→ 阅读 [IMPLEMENTATION_GUIDE.md - 整体规划](./IMPLEMENTATION_GUIDE.md#-整体规划)

#### 🚀 现在就可以开始实施吗？

→ 阅读 [IMPLEMENTATION_GUIDE.md - 准备工作](./IMPLEMENTATION_GUIDE.md#-准备工作)

#### 🐛 实施中遇到问题怎么办？

→ 阅读 [IMPLEMENTATION_GUIDE.md - 问题排查](./IMPLEMENTATION_GUIDE.md#-phase-4-问题排查)

---

## 📚 文档结构

```
docs/ui/optiongroup/
├── README.md                          # 文档入口（本文件）
├── ARCHITECTURE_ANALYSIS_REPORT.md    # 架构分析报告
├── CURRENT_STATUS.md                  # 当前系统现状
├── IMPLEMENTATION_GUIDE.md            # 实施指南
└── (未来可能添加的文档)
    ├── CHANGELOG.md                   # 变更日志
    ├── API_REFERENCE.md               # API 参考
    └── TROUBLESHOOTING.md             # 故障排除
```

---

## 🎯 OptionGroup 组件简介

### 核心功能

OptionGroup 是 Mint UI 的多选组件，支持：

- ✅ **单选模式**（类似 Radio按钮）
- ✅ **多选模式**（类似 Checkbox）
- ✅ **Store + Reducer 集成**
- ✅ **垂直/水平排列**
- ✅ **键盘/鼠标交互**
- ✅ **自定义间距和样式**

### 架构特点

```
Fiber-first Architecture

Parent: OptionGroup.VNode
    ├── optionSelectFunc  ← 父回调（闭包包装）
    ├── Children()
        ├── Child: OptionVNode (opt1)
        │   ├── selectFunc  ← 闭包捕获父引用
        │   └── Instance: OptionInstance
        ├── Child: OptionVNode (opt2)
        │   └── Instance: OptionInstance
        └── Child: OptionVNode (opt3)
            └── Instance: OptionInstance
```

### 使用示例

```go
import (
    "github.com/wwsheng009/mint/runtime/intent"
    "github.com/wwsheng009/mint/ui/components/optiongroup"
)

// 定义选项
options := []optiongroup.Option{
    {Value: "fire", Label: "Fire 🔥"},
    {Value: "ice", Label: "Ice ❄️"},
    {Value: "thunder", Label: "Thunder ⚡"},
}

// 创建多选组
group := optiongroup.NewBuilder(options).
    Key("weapon-selector").
    Label("Weapons:").
    Mode(optiongroup.ModeMultiple).
    ForField(intent.BindField("weapons")).  // Store 集成
    Vertical().
    Spacing(1).
    Build()
```

更多示例请查看：
- [`examples/multiselect_demo/main.go`](../../examples/multiselect_demo/)
- [`examples/typed_intent_demo/main.go`](../../examples/typed_intent_demo/)

---

## 🔴 当前问题

### 主要问题：子选项无法选中 ⚠️

**现象**：
- 鼠标点击选项无响应
- 键盘 Enter/Space 无响应
- 状态不会更新

**根本原因**：
Fiber-first 架构的 Fiber 创建时序导致父-子回调无法正确传递。

**状态分析**：
- 架构重已完成 ✅
- 代码已编译通过 ✅
- 单元测试全部通过 ✅
- 但存在运行时问题 🔴

**详细分析** → [ARCHITECTURE_ANALYSIS_REPORT.md](./ARCHITECTURE_ANALYSIS_REPORT.md)

---

## 💡 推荐方案

### 首选：闭包包装（方案E）⭐

**原理**：使用闭包延迟查找父回调

**优势**：
- 改动最小（仅修改 `vnode.go`）
- 无需全局状态或框架变更
- 性能最优
- 完全向后兼容

**实施指南** → [IMPLEMENTATION_GUIDE.md - Phase 1](./IMPLEMENTATION_GUIDE.md#-phase-1-实施闭包方案)

**其他方案**：
- 方案B: OnMount 注册（全局注册表）- 可用作回退方案
- 方案D: Intent 系统 - 长期方案，适合大型应用

**方案对比** → [ARCHITECTURE_ANALYSIS_REPORT.md - 方案对比表](./ARCHITECTURE_ANALYSIS_REPORT.md#-方案对比表)

---

## 🚀 快速开始

### 对于想了解问题的读者

1. 阅读 [ARCHITECTURE_ANALYSIS_REPORT.md](./ARCHITECTURE_ANALYSIS_REPORT.md)
   - 第1章：问题分析
   - 第2章：时序问题详解
   - 第3章：React 解决方案对比

2. 阅读 [CURRENT_STATUS.md](./CURRENT_STATUS.md)
   - 第1章：当前架构状态
   - 第2章：已实现功能清单
   - 第3章：已知问题

### 对于想实施修复的读者

1. 阅读 [IMPLEMENTATION_GUIDE.md](./IMPLEMENTATION_GUIDE.md)
   - **准备工作**：环境验证
   - **Phase 1**: 实施闭包方案
   - **Phase 2**: 清理冗余代码
   - **Phase 3**: 功能验证

2. 按照步骤执行：
   ```bash
   cd E:\projects\yao\wwsheng009\mint
   # 按照 IMPLEMENTATION_GUIDE.md 的步骤执行
   ```

3. 验证修复：
   ```bash
   go test ./ui/components/optiongroup/...
   go run ./examples/multiselect_demo/
   ```

### 对于想深入了解的读者

1. 阅读 [ARCHITECTURE_ANALYSIS_REPORT.md](./ARCHITECTURE_ANALYSIS_REPORT.md)
   - 全部章节

2. 研究相关源码：
   - `ui/components/optiongroup/vnode.go`
   - `ui/components/optiongroup/instance.go`
   - `ui/components/optiongroup/option.go`

3. 查看示例代码：
   - `examples/multiselect_demo/main.go`
   - `examples/typed_intent_demo/main.go`

---

## 📊 文档统计

| 文档 | 章节 | 字数 | 代码示例 |
|------|------|------|----------|
| ARCHITECTURE_ANALYSIS_REPORT.md | 8 | ~8,000 | 25+ |
| CURRENT_STATUS.md | 8 | ~4,000 | 15+ |
| IMPLEMENTATION_GUIDE.md | 7 | ~6,000 | 30+ |
| README.md | 4 | ~1,000 | 5+ |

---

## 📞 获取帮助

### 文档相关

- 问题描述不清晰？→ 阅读 [ARCHITECTURE_ANALYSIS_REPORT.md](./ARCHITECTURE_ANALYSIS_REPORT.md)
- 代码位置不清楚？→ 阅读 [CURRENT_STATUS.md - 文件清单](./CURRENT_STATUS.md#-文件清单)
- 不知道如何实施？→ 阅读 [IMPLEMENTATION_GUIDE.md](./IMPLEMENTATION_GUIDE.md)

### 问题排查

- 实施中遇到编译错误 → [IMPLEMENTATION_GUIDE.md - 问题排查](./IMPLEMENTATION_GUIDE.md#-problem-编译错误)
- 功能测试不通过 → [IMPLEMENTATION_GUIDE.md - 问题排查](./IMPLEMENTATION_GUIDE.md#-problem-点击选项无响应)

### 其他资源

- **主仓库**：`https://github.com/wwsheng009/mint`
- **Issue 追踪**：`https://github.com/wwsheng009/mint/issues`
- **讨论区**：`https://github.com/wwsheng009/mint/discussions`

---

## 🔗 相关链接

### Mint 文档

- [系统现状分析](../../current_system_analysis.md)
- [组件迁移计划](../../COMPONENT_MIGRATION_PLAN.md)
- [主 README](../../README.md)

### 示例程序

- [Multi-Select Demo](../../examples/multiselect_demo/)
- [Type-Safe Form Demo](../../examples/typesafe_form_demo_runapp/)

### 组件源码

- [VNode 实现](../../ui/components/optiongroup/vnode.go)
- [Instance 实现](../../ui/components/optiongroup/instance.go)
- [Option 子组件](../../ui/components/optiongroup/option.go)
- [Builder API](../../ui/components/optiongroup/builder.go)

### 架构相关

- [Fiber-first 架构](../../runtime/ui/fiber.go)
- [Store + Reducer](../../runtime/statemachine/runtime.go)
- [Intent 系统](../../runtime/intent/)

---

## 📅 更新历史

| 日期 | 版本 | 说明 |
|------|------|------|
| 2026-03-06 | 1.0 | 初始版本，完整文档发布 |

---

## 🙏 贡献

欢迎对本文档提出改进建议：

1. 提交 Issue 描述问题
2. 提交 Pull Request 改进文档
3. 在讨论区交流想法

---

**祝您使用愉快！** 🚀
