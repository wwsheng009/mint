# Mint 框架稳定性优化实施方案

**日期**: 2026-01-31  
**关联文档**: [2026-01-31-style-state-fix.md](./2026-01-31-style-state-fix.md)  
**目标**: 将复盘报告中的优化方向转化为具体的可执行任务，分阶段提升系统健壮性。

---

## 第一阶段：防御性加固与测试覆盖 (Immediate / P0)

本阶段目标是巩固现有的修复成果，防止同类问题回归，并增加运行时的自我检测能力。

### 1.1 完善 StyleStateMachine 测试套件 ✅ 已完成
*   **背景**: 此次 Bug 的核心是背景色重置逻辑缺失。
*   **任务**:
    *   ✅ 创建 `runtime/paint/style_state_test.go`
    *   ✅ 编写表驱动测试，覆盖场景：
        *   `Default -> Color`: 验证生成颜色代码
        *   `Color -> Default`: **关键验证**，确保生成 `\x1b[0m`
        *   `Color A -> Color B`: 验证只生成颜色切换代码
        *   `Bold -> No Bold`: 验证生成重置代码
        *   `Complex -> Complex`: 多属性同时变化的最小化输出验证
*   **测试结果**:
    ```
    PASS: TestStyleStateMachine_BuildDiffCodes (20个子测试)
    PASS: TestStyleStateMachine_Reset
    PASS: TestStyleStateMachine_NeedsUpdate
    PASS: TestStyleStateMachine_ComplexScenario
    PASS: TestStyleStateMachine_EdgeCases
    PASS: TestStyleStateMachine_Performance
    PASS: TestStyleStateMachine_ManyChangesOptimization
    PASS: TestStyleStateMachine_ColorToDefault_Critical
    
    覆盖率:
    - NewStyleStateMachine: 100%
    - Reset: 100%
    - NeedsUpdate: 100%
    - Update: 100%
    - buildDiffCodes: 85.7%
    - fullStyle: 100%
    ```
*   **交付物**: `runtime/paint/style_state_test.go`

### 1.2 增强渲染器状态重置机制
*   **背景**: `Renderer` 的内部状态与终端实际状态不同步导致了 Diff 失效。
*   **任务**:
    *   在 `Renderer` 中引入 `ResetState()` 方法，显式重置 `styleState` 和 `cursor`。
    *   在 `Render()` 方法入口处添加断言（Debug 模式），检查 `styleState` 是否已归零。
    *   在 `Render()` 出口处，确保始终输出 `\x1b[0m`，并同步重置内部状态。
*   **交付物**: 修改后的 `runtime/paint/renderer.go`。

### 1.3 事件循环看门狗 (Event Loop Watchdog)
*   **背景**: `Throttler` 可能导致主循环在 `dirty=true` 时无限期阻塞。
*   **任务**:
    *   在 `framework/app.go` 的 `Run` 方法中，保留并优化已添加的 `timer` 唤醒机制。
    *   添加调试日志：如果 `dirty` 状态持续超过 100ms 未被清除，输出 `[WARN] Render stall detected`。
*   **交付物**: 增强版的主循环逻辑。

---

## 第二阶段：核心算法验证 (Medium Term / P1)

本阶段目标是通过更高级的测试手段，保证核心 Diff 算法的绝对正确性。

### 2.1 Diff 算法模糊测试 (Fuzz Testing)
*   **背景**: 手动编写测试用例无法覆盖所有边缘情况（如本次的"跳过区域"问题）。
*   **任务**:
    *   创建 `framework/output_diff_fuzz_test.go`。
    *   实现 Fuzzing 逻辑：
        1.  随机生成 `Buffer A` (Old) 和 `Buffer B` (New)。
        2.  运行 `CompareBuffers(B, A)` 获取 `Changes`。
        3.  运行 `FormatChangesAsANSI` 生成输出序列。
        4.  **模拟终端执行**: 编写一个简单的 ANSI 解析器，在一个虚拟网格上执行输出序列。
        5.  **验证**: 比较虚拟网格的状态是否严格等于 `Buffer B`。
*   **交付物**: 一个能自动发现 Diff 逻辑漏洞的 Fuzz Test 工具。

### 2.2 渲染器快照测试 (Snapshot Testing)
*   **背景**: 防止视觉回归。
*   **任务**:
    *   选取典型组件（Button, Input, List）构建 Golden Master 测试集。
    *   保存其渲染后的 ANSI 字符串为 `.golden` 文件。
    *   在 CI 流程中运行测试，对比当前输出与 `.golden` 文件是否一致。
*   **交付物**: 集成在 `go test` 中的快照测试框架。

---

## 第三阶段：架构重构 (Long Term / P2)

本阶段目标是从根本上消除"状态不同步"和"索引匹配错误"的架构根因。

### 3.1 引入组件 ID 系统
*   **背景**: 依赖数组索引或指针匹配组件在动态布局中非常脆弱。
*   **任务**:
    *   定义 `ComponentID` 类型（可以是字符串或哈希）。
    *   在 `NewComponent` 时自动生成或允许手动指定 ID。
    *   重构 `collectInteractiveElements`，构建 `ID -> Component` 的映射表。
    *   重构焦点逻辑，使用 `focusedID` 替代 `focusedIndex`。
*   **预期效果**: 无论渲染顺序如何，只要 ID 不变，焦点状态就永久保持。

### 3.2 渲染管线分离
*   **背景**: 目前的 `renderVNode` 混合了布局计算和绘制。
*   **任务**:
    *   将渲染过程拆分为显式的 `Layout()` 和 `Paint()` 两个 pass。
    *   `Layout()`: 计算 VNode 树中每个节点的 `Rect {x, y, w, h}`。
    *   `Paint()`: 接收 `Rect` 信息进行绘制，不再进行布局计算。
*   **预期效果**: 布局计算独立后，可以更容易地实现复杂的布局（如 Flexbox），且绘制逻辑更纯粹，不易出错。

### 3.3 绝对定位渲染模式
*   **背景**: 相对移动（`\x1b[C`）虽然节省字节，但容易受宽字符和状态误差影响。
*   **任务**:
    *   将本次修复中"强制使用绝对定位"的改动标准化。
    *   评估其对性能的影响（通常极小）。
    *   如果性能允许，默认启用"绝对定位模式"，仅在极长连续文本输出时使用相对定位。

---

## 执行路线图

| 阶段 | 任务 | 预计耗时 | 负责人 |
| :--- | :--- | :--- | :--- |
| **Phase 1** | StyleState测试, Renderer重置, Watchdog | 1周 | @Antigravity |
| **Phase 2** | Diff Fuzzing, Snapshot Testing | 2周 | @Antigravity |
| **Phase 3** | 组件ID重构, 渲染管线分离 | 1个月+ | @ArchTeam |

---

## 风险评估

*   **Diff Fuzzing 的复杂性**: 编写一个准确的 ANSI 模拟器（Virtual Terminal）本身很有挑战性。可以使用现有的开源库（如 `go-term-emulator` 类的逻辑）简化实现。
*   **架构重构的兼容性**: 引入 ID 系统可能会破坏现有的组件 API，需要考虑向后兼容或提供迁移工具。

---

## 当前状态更新 (2026-01-31)

### 已完成 ✅

| 任务 | 状态 | 说明 |
|------|------|------|
| StyleState 属性关闭检测 | ✅ | `runtime/paint/style_state.go` 第57-87行 |
| Renderer ResetState | ✅ | `runtime/paint/renderer.go` 已添加 |
| Component ID 系统 (Phase 3简化版) | ✅ | 使用 `focusIndex` 替代指针比较 |
| 递归收集 Bug 修复 | ✅ | 分离 `resetInteractiveElements()` |
| StyleStateMachine 测试套件 | ✅ | 覆盖率: Update/Reset/NeedsUpdate 100%, buildDiffCodes 85.7% |

### 待完成 ⏳

| 任务 | 优先级 | 说明 |
|------|--------|------|
| ~~style_state_test.go 覆盖~~ | ✅ | ~~表驱动测试~~ 已完成 (关键方法100%) |
| Event Loop Watchdog | P0 | 100ms stall 检测 |
| Diff Fuzz Testing | P1 | 随机输入验证 |
| Snapshot Testing | P1 | Golden Master |
| ComponentID 完整实现 | P2 | 字符串ID替代索引 |
| Layout/Paint 分离 | P2 | 渲染管线重构 |
