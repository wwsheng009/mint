# Runtime 模块测试 TODO 清单

**创建日期**: 2026-01-31
**基于文档**: `TEST_AUDIT_REPORT.md`
**目标**: 建立完整的测试覆盖，达到 80%+ 总体覆盖率

---

## 📋 目录

1. [总体进度追踪](#总体进度追踪)
2. [P0 紧急任务](#p0-紧急任务)
3. [P1 高优先级任务](#p1-高优先级任务)
4. [P2 中优先级任务](#p2-中优先级任务)
5. [P3 低优先级任务](#p3-低优先级任务)
6. [每周执行计划](#每周执行计划)
7. [质量检查清单](#质量检查清单)

---

## 总体进度追踪

### 进度概览

```
┌─────────────────────────────────────────────────────────────┐
│                    测试完成进度                              │
├─────────────────────────────────────────────────────────────┤
│  P0 紧急任务    [████████████████████████████████████████] 100% │
│  P1 高优先级    [░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░]  0%   │
│  P2 中优先级    [░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░]  0%   │
│  P3 低优先级    [░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░]  0%   │
│  总体进度       [██████████░░░░░░░░░░░░░░░░░░░░░░░░]  38%  │
└─────────────────────────────────────────────────────────────┘
```

### 统计数据

| 模块 | 总任务数 | 已完成 | 进行中 | 待开始 | 完成率 |
|------|---------|--------|--------|--------|--------|
| **Layout 系统** | 23 | 0 | 0 | 23 | 0% |
| **Focus 系统** | 23 | 0 | 0 | 23 | 0% |
| **Input 系统** | 17 | 0 | 0 | 17 | 0% |
| **Paint 系统** | 19 | 0 | 0 | 19 | 0% |
| **Event 系统** | 12 | 0 | 0 | 12 | 0% |
| **Platform 层** | 10 | 0 | 0 | 10 | 0% |
| **Render 系统** | 6 | 0 | 0 | 6 | 0% |
| **其他模块** | 20 | 0 | 0 | 20 | 0% |
| **总计** | 130 | 0 | 0 | 130 | 0% |

---

## P0 紧急任务

### Layout 系统（50 小时，23 个任务）

#### 任务 1.1: 基础约束测试

**目标**: 测试 BoxConstraints 的核心功能
**预计时间**: 4 小时
**文件**: `layout/constraints_test.go`

**子任务**:

- [ ] [ ] 创建 `layout/constraints_test.go` 文件
  - [ ] 包结构：`package layout`
  - [ ] 导入：`testing`, `github.com/stretchr/testify/assert`
- [ ] [ ] 实现 `TestBoxConstraints_Basic`
  - [ ] 测试正常约束创建
  - [ ] 测试宽高属性
  - [ ] 测试 MinWidth、MaxWidth、MinHeight、MaxHeight
- [ ] [ ] 实现 `TestBoxConstraints_IsTight`
  - [ ] 测试紧约束判断（MinWidth == MaxWidth）
  - [ ] 测试紧约束判断（MinHeight == MaxHeight）
  - [ ] 测试非紧约束
- [ ] [ ] 实现 `TestBoxConstraints_Loosen`
  - [ ] 测试约束放宽逻辑
  - [ ] 测试紧约束变宽松
  - [ ] 测试无限约束
- [ ] [ ] 实现 `TestBoxConstraints_Constrain_Boundaries`
  - [ ] 测试宽度约束（在范围内）
  - [ ] 测试高度约束（在范围内）
  - [ ] 测试超出范围的处理
- [ ] [ ] 实现 `TestBoxConstraints_Constrain_EdgeCases`
  - [ ] 测试零约束
  - [ ] 测试负约束
  - [ ] 测试极大值约束
- [ ] [ ] 实现 `TestBoxConstraints_Infinity`
  - [ ] 测试无限宽度
  - [ ] 测试无限高度
  - [ ] 测试双向无限
- [ ] [ ] 实现 `TestBoxConstraints_Zero`
  - [ ] 测试零宽度
  - [ ] 测试零高度
  - [ ] 测试双向零
- [ ] [ ] 实现 `TestBoxConstraints_Negative`
  - [ ] 测试负宽度处理
  - [ ] 测试负高度处理
  - [ ] 验证负约束被修正为零

**验收标准**:
- [ ] 所有测试用例通过
- [ ] 覆盖率 ≥ 90%
- [ ] 通过 `go test -race`

**创建测试文件**:
```bash
# 运行测试
go test ./runtime/layout/... -run TestBoxConstraints -v

# 运行覆盖率
go test ./runtime/layout/... -coverprofile=constraints_coverage.out
go tool cover -html=constraints_coverage.out
```

---

#### 任务 1.2: 测量功能测试

**目标**: 测试 Measure 功能的正确性
**预计时间**: 4 小时
**文件**: `layout/measure_test.go`

**子任务**:

- [ ] [ ] 创建 `layout/measure_test.go` 文件
- [ ] [ ] 实现 `TestMeasure_Text`
  - [ ] 测试单行文本测量
  - [ ] 测试多行文本测量
  - [ ] 测试宽字符测量
  - [ ] 测试制表符测量
- [ ] [ ] 实现 `TestMeasure_Box`
  - [ ] 测试固定大小盒子
  - [ ] 测试弹性盒子
  - [ ] 测试最小尺寸约束
  - [ ] 测试最大尺寸约束
- [ ] [ ] 实现 `TestMeasure_Nested`
  - [ ] 测试两层嵌套
  - [ ] 测试多层嵌套（5 层）
  - [ ] 测试混合布局嵌套
- [ ] [ ] 实现 `TestMeasure_FlexGrow`
  - [ ] 测试 flexGrow = 0
  - [ ] 测试 flexGrow = 1
  - [ ] 测试多个子节点 flexGrow
  - [ ] 测试非均匀 flexGrow
- [ ] [ ] 实现 `TestMeasure_FlexShrink`
  - [ ] 测试 flexShrink = 0
  - [ ] 测试 flexShrink = 1
  - [ ] 测试多个子节点 flexShrink
- [ ] [ ] 实现 `TestMeasure_MaxWidth`
  - [ ] 测试 maxWidth 约束生效
  - [ ] 测试 maxWidth 与 flexGrow 冲突
  - [ ] 测试 maxWidth 超出父约束
- [ ] [ ] 实现 `TestMeasure_MaxHeight`
  - [ ] 测试 maxHeight 约束生效
  - [ ] 测试 maxHeight 与 flexGrow 冲突
- [ ] [ ] 实现 `TestMeasure_MinConstraints`
  - [ ] 测试 minWidth 约束生效
  - [ ] 测试 minHeight 约束生效
  - [ ] 测试 min 与 max 矛盾处理

**验收标准**:
- [ ] 所有测试用例通过
- [ ] 覆盖率 ≥ 85%
- [ ] 性能：100 次测量 < 10ms

---

#### 任务 1.3: Flexbox 布局测试

**目标**: 测试 Flexbox 布局的完整功能
**预计时间**: 6 小时
**文件**: `layout/flex_test.go`

**子任务**:

- [ ] [ ] 创建 `layout/flex_test.go` 文件
- [ ] [ ] 实现 `TestFlex_Row_Basic`
  - [ ] 测试水平行布局
  - [ ] 测试子节点顺序
  - [ ] 测试主轴方向
- [ ] [ ] 实现 `TestFlex_Column_Basic`
  - [ ] 测试垂直列布局
  - [ ] 测试子节点顺序
  - [ ] 测试主轴方向
- [ ] [ ] 实现 `TestFlex_AlignStart`
  - [ ] 测试 Row - AlignStart
  - [ ] 测试 Column - AlignStart
  - [ ] 测试交叉轴对齐
- [ ] [ ] 实现 `TestFlex_AlignCenter`
  - [ ] 测试 Row - AlignCenter
  - [ ] 测试 Column - AlignCenter
  - [ ] 测试交叉轴居中
- [ ] [ ] 实现 `TestFlex_AlignEnd`
  - [ ] 测试 Row - AlignEnd
  - [ ] 测试 Column - AlignEnd
  - [ ] 测试交叉轴对齐
- [ ] [ ] 实现 `TestFlex_SpaceBetween`
  - [ ] 测试 Row - SpaceBetween
  - [ ] 测试 Column - SpaceBetween
  - [ ] 测试间距分布
- [ ] [ ] 实现 `TestFlex_SpaceAround`
  - [ ] 测试 Row - SpaceAround
  - [ ] 测试 Column - SpaceAround
  - [ ] 测试间距分布
- [ ] [ ] 实现 `TestFlex_Gap`
  - [ ] 测试 Row - Gap
  - [ ] 测试 Column - Gap
  - [ ] 测试 Grid - Gap
  - [ ] 测试 Gap 与冲突处理
- [ ] [ ] 实现 `TestFlex_Wrap`
  - [ ] 测试 Row - Wrap
  - [ ] 测试 Column - Wrap
  - [ ] 测试 Wrap 后布局
- [ ] [ ] 实现 `TestFlex_NestedFlex`
  - [ ] 测试 Row 内嵌 Row
  - [ ] 测试 Row 内嵌 Column
  - [ ] 测试三层嵌套
- [ ] [ ] 实现 `TestFlex_MixedFlex`
  - [ ] 测试 flexGrow + flexShrink
  - [ ] 测试固定 + 弹性子节点
  - [ ] 测试复杂布局

**验收标准**:
- [ ] 所有测试用例通过
- [ ] 覆盖率 ≥ 85%
- [ ] 性能：布局 100 个节点 < 50ms

---

#### 任务 1.4: 缓存机制测试

**目标**: 测试布局缓存机制
**预计时间**: 3 小时
**文件**: `layout/cache_test.go`

**子任务**:

- [ ] [ ] 创建 `layout/cache_test.go` 文件
- [ ] [ ] 实现 `TestCache_Hit`
  - [ ] 测试缓存命中
  - [ ] 测试缓存返回值正确
  - [ ] 测试缓存避免重复计算
- [ ] [ ] 实现 `TestCache_Miss`
  - [ ] 测试缓存未命中
  - [ ] 测试缓存填充
  - [ ] 测试缓存键不存在
- [ ] [ ] 实现 `TestCache_Invalidate`
  - [ ] 测试节点失效
  - [ ] 测试子树失效
  - [ ] 测试父节点失效
  - [ ] 测试多节点失效
- [ ] [ ] 实现 `TestCache_Clear`
  - [ ] 测试清空缓存
  - [ ] 测试缓存大小归零
  - [ ] 测试清空后重新填充
- [ ] [ ] 实现 `TestCache_ThreadSafety`
  - [ ] 测试并发读
  - [ ] 测试并发写
  - [ ] 测试读写混合
  - [ ] 测试无竞态（`go test -race`）
- [ ] [ ] 实现 `TestCache_Eviction`
  - [ ] 测试 LRU 淘汰
  - [ ] 测试容量限制
  - [ ] 测试淘汰后重新计算

**验收标准**:
- [ ] 所有测试用例通过
- [ ] 覆盖率 ≥ 85%
- [ ] 竞态检测通过

---

#### 任务 1.5: 复杂场景测试

**目标**: 测试复杂布局场景和性能
**预计时间**: 3 小时
**文件**: `layout/complex_test.go`

**子任务**:

- [ ] [ ] 创建 `layout/complex_test.go` 文件
- [ ] [ ] 实现 `TestLayout_DeepNesting`
  - [ ] 测试 10 层嵌套
  - [ ] 测试 50 层嵌套
  - [ ] 测试 100 层嵌套
  - [ ] 验证不崩溃
- [ ] [ ] 实现 `TestLayout_ConflictingConstraints`
  - [ ] 测试 minWidth > maxWidth
  - [ ] 测试 minHeight > maxHeight
  - [ ] 测试子节点超出父节点
  - [ ] 验证正确处理
- [ ] [ ] 实现 `TestLayout_Performance_Large`
  - [ ] 基准测试：布局 100 节点
  - [ ] 基准测试：布局 500 节点
  - [ ] 基准测试：布局 1000 节点
  - [ ] 验证 1000 节点 < 100ms
- [ ] [ ] 实现 `TestLayout_Performance_Deep`
  - [ ] 基准测试：深度 10
  - [ ] 基准测试：深度 50
  - [ ] 基准测试：深度 100
  - [ ] 验证深度 100 < 50ms
- [ ] [ ] 实现 `TestLayout_EdgeCases`
  - [ ] 测试空节点
  - [ ] 测试单一节点
  - [ ] 测试零尺寸节点
  - [ ] 测试无限约束
  - [ ] 测试负约束

**验收标准**:
- [ ] 所有测试用例通过
- [ ] 覆盖率 ≥ 80%
- [ ] 性能达标：1000 节点 < 100ms
- [ ] 竞态检测通过

---

### Focus 系统（38 小时，23 个任务）

#### 任务 2.1: 焦点管理器测试

**目标**: 测试 FocusManager 的核心功能
**预计时间**: 4 小时
**文件**: `focus/manager_test.go`

**子任务**:

- [ ] [ ] 创建 `focus/manager_test.go` 文件
- [ ] [ ] 实现 `TestFocusManager_SetFocus`
  - [ ] 测试设置焦点到有效节点
  - [ ] 测试设置焦点到无效节点
  - [ ] 测试重复设置焦点
  - [ ] 测试焦点变更事件
- [ ] [ ] 实现 `TestFocusManager_RemoveFocus`
  - [ ] 测试移除当前焦点
  - [ ] 测试移除非焦点节点
  - [ ] 测试移除后焦点恢复
- [ ] [ ] 实现 `TestFocusManager_GetCurrentFocus`
  - [ ] 测试获取当前焦点
  - [ ] 测试无焦点时返回 nil
  - [ ] 测试焦点变更后更新
- [ ] [ ] 实现 `TestFocusManager_GetFocusHistory`
  - [ ] 测试历史记录长度
  - [ ] 测试历史记录顺序
  - [ ] 测试最大历史限制
- [ ] [ ] 实现 `TestFocusManager_Clear`
  - [ ] 测试清除所有焦点
  - [ ] 测试清除后历史清空
  - [ ] 测试清除后可重新设置
- [ ] [ ] 实现 `TestFocusManager_Next`
  - [ ] 测试 Tab 键向前导航
  - [ ] 测试循环导航
  - [ ] 测试边界处理
- [ ] [ ] 实现 `TestFocusManager_Prev`
  - [ ] 测试 Shift+Tab 向后导航
  - [ ] 测试循环导航
  - [ ] 测试边界处理
- [ ] [ ] 实现 `TestFocusManager_TabOrder`
  - [ ] 测试默认 Tab 顺序
  - [ ] 测试自定义 Tab 顺序
  - [ ] 测试隐藏元素跳过

**验收标准**:
- [ ] 所有测试用例通过
- [ ] 覆盖率 ≥ 85%

---

#### 任务 2.2: 焦点导航测试

**目标**: 测试键盘焦点导航
**预计时间**: 4 小时
**文件**: `focus/navigation_test.go`

**子任务**:

- [ ] [ ] 创建 `focus/navigation_test.go` 文件
- [ ] [ ] 实现 `TestNavigate_Up`
  - [ ] 测试向上导航
  - [ ] 测试最上方元素处理
  - [ ] 测试跨越边界
- [ ] [ ] 实现 `TestNavigate_Down`
  - [ ] 测试向下导航
  - [ ] 测试最下方元素处理
  - [ ] 测试跨越边界
- [ ] [ ] 实现 `TestNavigate_Left`
  - [ ] 测试向左导航
  - [ ] 测试最左侧元素处理
  - [ ] 测试跨越边界
- [ ] [ ] 实现 `TestNavigate_Right`
  - [ ] 测试向右导航
  - [ ] 测试最右侧元素处理
  - [ ] 测试跨越边界
- [ ] [ ] 实现 `TestNavigate_PageUp`
  - [ ] 测试 PageUp 跳页
  - [ ] 测试超出范围处理
- [ ] [ ] 实现 `TestNavigate_PageDown`
  - [ ] 测试 PageDown 跳页
  - [ ] 测试超出范围处理
- [ ] [ ] 实现 `TestNavigate_Home`
  - [ ] 测试 Home 键到首元素
  - [ ] 测试无元素处理
- [ ] [ ] 实现 `TestNavigate_End`
  - [ ] 测试 End 键到尾元素
  - [ ] 测试无元素处理
- [ ] [ ] 实现 `TestNavigate_WrapAround`
  - [ ] 测试循环导航开启
  - [ ] 测试循环导航关闭
  - [ ] 测试单元素循环

**验收标准**:
- [ ] 所有测试用例通过
- [ ] 覆盖率 ≥ 85%

---

#### 任务 2.3: 几何导航测试

**目标**: 测试几何感知的焦点导航
**预计时间**: 3 小时
**文件**: `focus/geometric_test.go`

**子任务**:

- [ ] [ ] 创建 `focus/geometric_test.go` 文件
- [ ] [ ] 实现 `TestGeometric_Nearest`
  - [ ] 测试查找最近元素（上）
  - [ ] 测试查找最近元素（下）
  - [ ] 测试查找最近元素（左）
  - [ ] 测试查找最近元素（右）
  - [ ] 测试无最近元素
- [ ] [ ] 实现 `TestGeometric_Distance`
  - [ ] 测试欧几里得距离计算
  - [ ] 测试曼哈顿距离计算
  - [ ] 测试距离排序
- [ ] [ ] 实现 `TestGeometric_BoundingBox`
  - [ ] 测试边界框计算
  - [ ] 测试重叠检测
  - [ ] 测试包含检测
- [ ] [ ] 实现 `TestGeometric_Visibility`
  - [ ] 测试可见性检查
  - [ ] 测试隐藏元素跳过
  - [ ] 测试遮挡元素跳过
- [ ] [ ] 实现 `TestGeometric_Overlapping`
  - [ ] 测试重叠元素优先级
  - [ ] 测试部分重叠处理
  - [ ] 测试完全重叠处理

**验收标准**:
- [ ] 所有测试用例通过
- [ ] 覆盖率 ≥ 80%

---

#### 任务 2.4: 焦点陷阱测试

**目标**: 测试 FocusTrap 功能
**预计时间**: 2 小时
**文件**: `focus/trap_test.go`

**子任务**:

- [ ] [ ] 创建 `focus/trap_test.go` 文件
- [ ] [ ] 实现 `TestTrap_Enter`
  - [ ] 测试进入陷阱
  - [ ] 测试陷阱激活
  - [ ] 测试保存前焦点
- [ ] [ ] 实现 `TestTrap_Exit`
  - [ ] 测试退出陷阱
  - [ ] 测试恢复前焦点
  - [ ] 测试陷阱停用
- [ ] [ ] 实现 `TestTrap_EscapeKey`
  - [ ] 测试 ESC 键退出陷阱
  - [ ] 测试 ESC 在陷阱外无效
  - [ ] 测试嵌套陷阱 ESC
- [ ] [ ] 实现 `TestTrap_Nested`
  - [ ] 测试嵌套陷阱进入
  - [ ] 测试嵌套陷阱退出
  - [ ] 测试陷阱栈管理
- [ ] [ ] 实现 `TestTrap_RestoreFocus`
  - [ ] 测试焦点恢复到入口元素
  - [ ] 测试元素不存在时恢复
  - [ ] 测试多层陷阱恢复

**验收标准**:
- [ ] 所有测试用例通过
- [ ] 覆盖率 ≥ 80%

---

#### 任务 2.5: V3 焦点管理测试

**目标**: 测试 V3 版本的焦点管理
**预计时间**: 2 小时
**文件**: `focus/v3_test.go`

**子任务**:

- [ ] [ ] 创建 `focus/v3_test.go` 文件
- [ ] [ ] 实现 `TestV3_Scope`
  - [ ] 测试焦点域定义
  - [ ] 测试域内焦点限制
  - [ ] 测试域外跳过
- [ ] [ ] 实现 `TestV3_ScopeStack`
  - [ ] 测试域入栈
  - [ ] 测试域出栈
  - [ ] 测试域嵌套
- [ ] [ ] 实现 `TestV3_DefaultFocus`
  - [ ] 测试默认焦点设置
  - [ ] 测试域变更后默认焦点
  - [ ] 测试默认焦点不存在
- [ ] [ ] 实现 `TestV3_Migration`
  - [ ] 测试从 V2 迁移
  - [ ] 测试 API 兼容性
  - [ ] 测试数据转换
- [ ] [ ] 实现 `TestV3_Persistence`
  - [ ] 测试焦点状态持久化
  - [ ] 测试状态恢复
  - [ ] 测试异常状态处理

**验收标准**:
- [ ] 所有测试用例通过
- [ ] 覆盖率 ≥ 75%

---

### Input 系统（38 小时，17 个任务）

#### 任务 3.1: 键盘映射测试

**目标**: 测试键盘输入映射
**预计时间**: 4 小时
**文件**: `input/keymap_test.go`

**子任务**:

- [ ] [ ] 创建 `input/keymap_test.go` 文件
- [ ] [ ] 实现 `TestKeyMap_MapKey`
  - [ ] 测试基础按键映射
  - [ ] 测试字母键映射
  - [ ] 测试数字键映射
  - [ ] 测试功能键映射
- [ ] [ ] 实现 `TestKeyMap_MapModifiers`
  - [ ] 测试 Ctrl 组合
  - [ ] 测试 Alt 组合
  - [ ] 测试 Shift 组合
  - [ ] 测试组合修饰键（Ctrl+Shift 等）
- [ ] [ ] 实现 `TestKeyMap_MapCombinations`
  - [ ] 测试 Ctrl+C
  - [ ] 测试 Ctrl+V
  - [ ] 测试 Alt+Enter
  - [ ] 测试复杂组合键
- [ ] [ ] 实现 `TestKeyMap_Unmap`
  - [ ] 测试取消映射
  - [ ] 测试重复取消
  - [ ] 测试不存在的映射
- [ ] [ ] 实现 `TestKeyMap_Lookup`
  - [ ] 测试查找映射
  - [ ] 测试未映射键返回
  - [ ] 测试大小写敏感
- [ ] [ ] 实现 `TestKeyMap_Windows`
  - [ ] 测试 Windows 特殊键
  - [ ] 测试 Windows 虚拟键码
  - [ ] 测试 Windows 转义序列
- [ ] [ ] 实现 `TestKeyMap_Unix`
  - [ ] 测试 Unix 特殊键
  - [ ] 测试 Unix 转义序列
  - [ ] 测试 Termcap 兼容

**验收标准**:
- [ ] 所有测试用例通过
- [ ] 覆盖率 ≥ 80%
- [ ] 跨平台兼容性验证

---

#### 任务 3.2: 鼠标追踪测试

**目标**: 测试鼠标输入追踪
**预计时间**: 4 小时
**文件**: `input/mouse_test.go`

**子任务**:

- [ ] [ ] 创建 `input/mouse_test.go` 文件
- [ ] [ ] 实现 `TestMouse_Move`
  - [ ] 测试鼠标移动追踪
  - [ ] 测试坐标更新
  - [ ] 测试连续移动
- [ ] [ ] 实现 `TestMouse_Click`
  - [ ] 测试左键单击
  - [ ] 测试右键单击
  - [ ] 测试中键单击
  - [ ] 测试点击位置
- [ ] [ ] 实现 `TestMouse_Drag`
  - [ ] 测试拖拽开始
  - [ ] 测试拖拽过程
  - [ ] 测试拖拽结束
  - [ ] 测试拖拽路径
- [ ] [ ] 实现 `TestMouse_Scroll`
  - [ ] 测试向上滚动
  - [ ] 测试向下滚动
  - [ ] 测试横向滚动
  - [ ] 测试滚动增量
- [ ] [ ] 实现 `TestMouse_DoubleClick`
  - [ ] 测试双击检测
  - [ ] 测试双击时间窗口
  - [ ] 测试双击与单击区别
- [ ] [ ] 实现 `TestMouse_Throttle`
  - [ ] 测试鼠标移动节流
  - [ ] 测试节流频率
  - [ ] 测试节流准确性
- [ ] [ ] 实现 `TestMouse_Position`
  - [ ] 测试绝对坐标
  - [ ] 测试相对坐标
  - [ ] 测试边界坐标

**验收标准**:
- [ ] 所有测试用例通过
- [ ] 覆盖率 ≥ 80%
- [ ] 性能：1000 次移动 < 10ms

---

#### 任务 3.3: 输入队列测试

**目标**: 测试输入事件队列管理
**预计时间**: 3 小时
**文件**: `input/queue_test.go`

**子任务**:

- [ ] [ ] 创建 `input/queue_test.go` 文件
- [ ] [ ] 实现 `TestQueue_Push`
  - [ ] 测试入队操作
  - [ ] 测试队列容量
  - [ ] 测试队列满处理
- [ ] [ ] 实现 `TestQueue_Pop`
  - [ ] 测试出队操作
  - [ ] 测试空队列处理
  - [ ] 测试返回值正确
- [ ] [ ] 实现 `TestQueue_Peek`
  - [ ] 测试查看队首
  - [ ] 测试不弹出
  - [ ] 测试空队列返回
- [ ] [ ] 实现 `TestQueue_Flush`
  - [ ] 测试清空队列
  - [ ] 测试清空后状态
  - [ ] 测试清空回调
- [ ] [ ] 实现 `TestQueue_Order`
  - [ ] 测试 FIFO 顺序
  - [ ] 测试优先级顺序
  - [ ] 测试复杂顺序
- [ ] [ ] 实现 `TestQueue_Concurrency`
  - [ ] 测试并发入队
  - [ ] 测试并发出队
  - [ ] 测试竞态检测
  - [ ] 验证无数据丢失

**验收标准**:
- [ ] 所有测试用例通过
- [ ] 覆盖率 ≥ 80%
- [ ] 竞态检测通过

---

#### 任务 3.4: 平台输入测试

**目标**: 测试跨平台输入处理
**预计时间**: 4 小时
**文件**: `input/platform_test.go`

**子任务**:

- [ ] [ ] 创建 `input/platform_test.go` 文件
- [ ] [ ] 实现 `TestPlatform_Unix_Read`
  - [ ] 测试 Unix 输入读取
  - [ ] 测试标准输入
  - [ ] 测试 TTY 模式
  - [ ] 测试非阻塞读取
- [ ] [ ] 实现 `TestPlatform_Windows_Read`
  - [ ] 测试 Windows 输入读取
  - [ ] 测试控制台输入
  - [ ] 测试虚拟终端
  - [ ] 测试非阻塞读取
- [ ] [ ] 实现 `TestPlatform_Unix_Sequence`
  - [ ] 测试 ANSI 转义序列
  - [ ] 测试 CSI 序列
  - [ ] 测试特殊键序列
- [ ] [ ] 实现 `TestPlatform_Windows_Sequence`
  - [ ] 测试 Windows 转义序列
  - [ ] 测试 ENHANCED_KEY
  - [ ] 测试 KEY_EVENT_RECORD
- [ ] [ ] 实现 `TestPlatform_EscapeSequences`
  - [ ] 测试方向键
  - [ ] 测试功能键
  - [ ] 测试组合键
  - [ ] 测试鼠标序列

**验收标准**:
- [ ] 所有测试用例通过
- [ ] 覆盖率 ≥ 75%
- [ ] 跨平台兼容性验证

---

## P1 高优先级任务

### Paint 系统（49 小时，19 个任务）

#### 任务 4.1: Compositor 测试

**目标**: 测试层合成功能
**预计时间**: 4 小时
**文件**: `paint/compositor_test.go`

**子任务**:

- [ ] [ ] 创建 `paint/compositor_test.go` 文件
- [ ] [ ] 实现 `TestComposite_LayerOrder`
  - [ ] 测试层顺序
  - [ ] 测试底层到顶层
  - [ ] 测试 Z-Index 排序
- [ ] [ ] 实现 `TestComposite_ZIndex`
  - [ ] 测试正 Z-Index
  - [ ] 测试负 Z-Index
  - [ ] 测试 Z-Index 相同时顺序
- [ ] [ ] 实现 `TestComposite_Opacity`
  - [ ] 测试透明度应用
  - [ ] 测试半透明层
  - [ ] 测试全透明层
- [ ] [ ] 实现 `TestComposite_Blend`
  - [ ] 测试混合模式
  - [ ] 测试颜色混合
  - [ ] 测试背景混合
- [ ] [ ] 实现 `TestComposite_Clip`
  - [ ] 测试裁剪区域
  - [ ] 测试边界裁剪
  - [ ] 测试嵌套裁剪
- [ ] [ ] 实现 `TestComposite_MultipleLayers`
  - [ ] 测试多层合成
  - [ ] 测试复杂场景
  - [ ] 验证正确性

**验收标准**:
- [ ] 所有测试用例通过
- [ ] 覆盖率 ≥ 80%

---

#### 任务 4.2: Context 测试

**目标**: 测试绘制上下文
**预计时间**: 3 小时
**文件**: `paint/context_test.go`

**子任务**:

- [ ] [ ] 创建 `paint/context_test.go` 文件
- [ ] [ ] 实现 `TestContext_Save`
  - [ ] 测试保存状态
  - [ ] 测试保存的属性
  - [ ] 测试多次保存
- [ ] [ ] 实现 `TestContext_Restore`
  - [ ] 测试恢复状态
  - [ ] 测试恢复属性正确
  - [ ] 测试多次恢复
- [ ] [ ] 实现 `TestContext_Nesting`
  - [ ] 测试嵌套保存/恢复
  - [ ] 测试深度嵌套
  - [ ] 测试不匹配恢复
- [ ] [ ] 实现 `TestContext_StyleStack`
  - [ ] 测试样式栈
  - [ ] 测试样式继承
  - [ ] 测试样式覆盖
- [ ] [ ] 实现 `TestContext_Clipping`
  - [ ] 测试裁剪栈
  - [ ] 测试裁剪合并
  - [ ] 测试裁剪恢复

**验收标准**:
- [ ] 所有测试用例通过
- [ ] 覆盖率 ≥ 80%

---

#### 任务 4.3: Style State 测试

**目标**: 测试样式状态管理
**预计时间**: 2 小时
**文件**: `paint/style_state_test.go`

**子任务**:

- [ ] [ ] 创建 `paint/style_state_test.go` 文件
- [ ] [ ] 实现 `TestStyleState_Apply`
  - [ ] 测试应用样式
  - [ ] 测试多个样式应用
  - [ ] 测试样式优先级
- [ ] [ ] 实现 `TestStyleState_Revert`
  - [ ] 测试恢复默认样式
  - [ ] 测试部分恢复
  - [ ] 测试完全恢复
- [ ] [ ] 实现 `TestStyleState_Inherit`
  - [ ] 测试样式继承
  - [ ] 测试继承链
  - [ ] 测试继承覆盖
- [ ] [ ] 实现 `TestStyleState_Merge`
  - [ ] 测试样式合并
  - [ ] 测试冲突合并
  - [ ] 测试空合并

**验收标准**:
- [ ] 所有测试用例通过
- [ ] 覆盖率 ≥ 80%

---

#### 任务 4.4: Batch 测试

**目标**: 测试批量绘制
**预计时间**: 3 小时
**文件**: `paint/batch_test.go`

**子任务**:

- [ ] [ ] 创建 `paint/batch_test.go` 文件
- [ ] [ ] 实现 `TestBatch_Add`
  - [ ] 测试添加绘制命令
  - [ ] 测试批量添加
  - [ ] 测试满批量处理
- [ ] [ ] 实现 `TestBatch_Flush`
  - [ ] 测试刷新批量
  - [ ] 测试刷新后清空
  - [ ] 测试刷新回调
- [ ] [ ] 实现 `TestBatch_Clear`
  - [ ] 测试清空批量
  - [ ] 测试清空后状态
  - [ ] 测试多次清空
- [ ] [ ] 实现 `TestBatch_Optimization`
  - [ ] 测试命令合并
  - [ ] 测试冗余消除
  - [ ] 测试优化效果

**验收标准**:
- [ ] 所有测试用例通过
- [ ] 覆盖率 ≥ 80%
- [ ] 性能：批量处理比单独快 50%

---

#### 任务 4.5: 集成测试

**目标**: 测试 Paint 系统集成
**预计时间**: 6 小时
**文件**: `paint/integration_test.go`

**子任务**:

- [ ] [ ] 创建 `paint/integration_test.go` 文件
- [ ] [ ] 实现 `TestIntegration_ComplexScene`
  - [ ] 测试复杂场景绘制
  - [ ] 测试多层、多样式
  - [ ] 测试最终输出正确
- [ ] [ ] 实现 `TestIntegration_Performance`
  - [ ] 基准测试：10000 cells
  - [ ] 基准测试：复杂场景
  - [ ] 验证 < 50ms
- [ ] [ ] 实现 `TestIntegration_WideCharacters`
  - [ ] 测试宽字符显示
  - [ ] 测试宽字符换行
  - [ ] 测试宽字符边界
- [ ] [ ] 实现 `TestIntegration_LayerTransitions`
  - [ ] 测试层切换
  - [ ] 测试层动画
  - [ ] 测试层过渡效果

**验收标准**:
- [ ] 所有测试用例通过
- [ ] 覆盖率 ≥ 80%
- [ ] 性能达标

---

### Event 系统（12 小时，12 个任务）

#### 任务 5.1: 事件分发补充测试

**目标**: 补充 Event 系统测试
**预计时间**: 4 小时
**文件**: `event/event_test.go`

**子任务**:

- [ ] [ ] 创建 `event/event_test.go` 文件
- [ ] [ ] 实现 `TestEvent_Propagation`
  - [ ] 测试事件传播路径
  - [ ] 测试停止传播
  - [ ] 测试阻止默认
- [ ] [ ] 实现 `TestEvent_Bubbling`
  - [ ] 测试冒泡阶段
  - [ ] 测试捕获阶段
  - [ ] 测试目标阶段
- [ ] [ ] 实现 `TestEvent_Listeners`
  - [ ] 测试监听器注册
  - [ ] 测试监听器移除
  - [ ] 测试监听器触发
- [ ] [ ] 实现 `TestEvent_Pool`
  - [ ] 测试事件池复用
  - [ ] 测试池容量
  - [ ] 测试池清理

**验收标准**:
- [ ] 所有测试用例通过
- [ ] 覆盖率 ≥ 80%

---

#### 任务 5.2: 命中测试补充

**目标**: 补充命中测试
**预计时间**: 4 小时
**文件**: `event/hittest_test.go`

**子任务**:

- [ ] [ ] 创建 `event/hittest_test.go` 文件
- [ ] [ ] 实现 `TestHitTest_Point`
  - [ ] 测试点命中
  - [ ] 测试边界点
  - [ ] 测试精确命中
- [ ] [ ] 实现 `TestHitTest_Rect`
  - [ ] 测试矩形命中
  - [ ] 测试重叠矩形
  - [ ] 测试部分命中
- [ ] [ ] 实现 `TestHitTest_ComplexShape`
  - [ ] 测试复杂形状
  - [ ] 测试不规则形状
  - [ ] 测试透明区域
- [ ] [ ] 实现 `TestHitTest_Optimization`
  - [ ] 测试空间索引
  - [ ] 测试四叉树
  - [ ] 测试性能

**验收标准**:
- [ ] 所有测试用例通过
- [ ] 覆盖率 ≥ 80%
- [ ] 性能：1000 次命中 < 10ms

---

#### 任务 5.3: 阶段测试

**目标**: 测试事件阶段
**预计时间**: 4 小时
**文件**: `event/phase_test.go`

**子任务**:

- [ ] [ ] 创建 `event/phase_test.go` 文件
- [ ] [ ] 实现 `TestPhase_Capture`
  - [ ] 测试捕获阶段
  - [ ] 测试捕获顺序
  - [ ] 测试捕获停止
- [ ] [ ] 实现 `TestPhase_Target`
  - [ ] 测试目标阶段
  - [ ] 测试目标处理
  - [ ] 测试目标阻止
- [ ] [ ] 实现 `TestPhase_Bubble`
  - [ ] 测试冒泡阶段
  - [ ] 测试冒泡顺序
  - [ ] 测试冒泡停止
- [ ] [ ] 实现 `TestPhase_Transition`
  - [ ] 测试阶段转换
  - [ ] 测试状态保持
  - [ ] 测试数据传递

**验收标准**:
- [ ] 所有测试用例通过
- [ ] 覆盖率 ≥ 80%

---

## P2 中优先级任务

### Platform 层（10 小时，10 个任务）

#### 任务 6.1: 平台抽象测试

**目标**: 测试平台抽象接口
**预计时间**: 2 小时
**文件**: `platform/platform_test.go`

**子任务**:

- [ ] [ ] 创建 `platform/platform_test.go` 文件
- [ ] [ ] 实现 `TestPlatform_Initialization`
  - [ ] 测试平台初始化
  - [ ] 测试清理
  - [ ] 测试重复初始化
- [ ] [ ] 实现 `TestPlatform_Terminal`
  - [ ] 测试终端类型检测
  - [ ] 测试终端能力
  - [ ] 测试终端设置

**验收标准**:
- [ ] 所有测试用例通过
- [ ] 覆盖率 ≥ 75%

---

#### 任务 6.2: 屏幕操作测试

**目标**: 测试屏幕操作
**预计时间**: 2 小时
**文件**: `platform/screen_test.go`

**子任务**:

- [ ] [ ] 创建 `platform/screen_test.go` 文件
- [ ] [ ] 实现 `TestScreen_Size`
  - [ ] 测试获取屏幕大小
  - [ ] 测试屏幕变化
  - [ ] 测试大小限制
- [ ] [ ] 实现 `TestScreen_Refresh`
  - [ ] 测试屏幕刷新
  - [ ] 测试刷新率
  - [ ] 测试双缓冲

**验收标准**:
- [ ] 所有测试用例通过
- [ ] 覆盖率 ≥ 75%

---

#### 任务 6.3: 信号处理测试

**目标**: 测试信号处理
**预计时间**: 2 小时
**文件**: `platform/signal_test.go`

**子任务**:

- [ ] [ ] 创建 `platform/signal_test.go` 文件
- [ ] [ ] 实现 `TestSignal_SIGINT`
  - [ ] 测试 SIGINT 处理
  - [ ] 测试优雅退出
  - [ ] 测试清理回调
- [ ] [ ] 实现 `TestSignal_SIGWINCH`
  - [ ] 测试 SIGWINCH 处理
  - [ ] 测试屏幕调整
  - [ ] 测试重布局

**验收标准**:
- [ ] 所有测试用例通过
- [ ] 覆盖率 ≥ 75%

---

#### 任务 6.4: 光标控制测试

**目标**: 测试光标控制
**预计时间**: 2 小时
**文件**: `platform/cursor_test.go`

**子任务**:

- [ ] [ ] 创建 `platform/cursor_test.go` 文件
- [ ] [ ] 实现 `TestCursor_Move`
  - [ ] 测试光标移动
  - [ ] 测试绝对位置
  - [ ] 测试相对位置
- [ ] [ ] 实现 `TestCursor_Visible`
  - [ ] 测试光标显示
  - [ ] 测试光标隐藏
  - [ ] 测试光标闪烁

**验收标准**:
- [ ] 所有测试用例通过
- [ ] 覆盖率 ≥ 75%

---

#### 任务 6.5: 跨平台兼容性测试

**目标**: 测试跨平台兼容性
**预计时间**: 2 小时
**文件**: `platform/cross_platform_test.go`

**子任务**:

- [ ] [ ] 创建 `platform/cross_platform_test.go` 文件
- [ ] [ ] 实现 `TestCrossPlatform_Input`
  - [ ] 测试 Windows 输入
  - [ ] 测试 Unix 输入
  - [ ] 测试行为一致性
- [ ] [ ] 实现 `TestCrossPlatform_Output`
  - [ ] 测试 Windows 输出
  - [ ] 测试 Unix 输出
  - [ ] 测试显示一致性

**验收标准**:
- [ ] 所有测试用例通过
- [ ] 跨平台行为一致

---

### Render 系统（8 小时，6 个任务）

#### 任务 7.1: 渲染流程测试

**目标**: 测试渲染主流程
**预计时间**: 4 小时
**文件**: `render/render_test.go`

**子任务**:

- [ ] [ ] 创建 `render/render_test.go` 文件
- [ [ ] 实现 `TestRender_FullPipeline`
  - [ ] 测试完整渲染管线
  - [ ] 测试 Layout -> Paint -> Output
  - [ ] 测试数据流
- [ [ ] 实现 `TestRender_MultiLayer`
  - [ ] 测试多层渲染
  - [ ] 测试层顺序
  - [ ] 测试层合成
- [ [ ] 实现 `TestRender_Incremental`
  - [ ] 测试增量渲染
  - [ ] 测试脏区域
  - [ ] 测试性能提升

**验收标准**:
- [ ] 所有测试用例通过
- [ ] 覆盖率 ≥ 80%

---

#### 任务 7.2: 视口管理补充测试

**目标**: 补充视口管理测试
**预计时间**: 4 小时
**文件**: `render/viewport_test.go`

**子任务**:

- [ ] [ ] 读取现有 `viewport_test.go`
- [ [ ] 补充 `TestViewport_Scroll`
  - [ ] 测试滚动到位置
  - [ ] 测试滚动边界
  - [ ] 测试滚动动画
- [ [ ] 补充 `TestViewport_Zoom`
  - [ ] 测试缩放（如支持）
  - [ ] 测试缩放限制
  - [ ] 测试缩放中心

**验收标准**:
- [ ] 所有测试用例通过
- [ ] 覆盖率 ≥ 80%

---

## P3 低优先级任务

### Animation 系统（5 小时）

#### 任务 8.1: 动画系统测试

**目标**: 测试动画系统
**预计时间**: 5 小时
**文件**: `animation/animation_test.go`

**子任务**:

- [ ] [ ] 创建 `animation/animation_test.go` 文件
- [ ] [ ] 实现 `TestAnimation_Start`
  - [ ] 测试动画启动
  - [ ] 测试动画暂停
  - [ ] 测试动画停止
- [ ] [ ] 实现 `TestAnimation_Easing`
  - [ ] 测试缓动函数
  - [ ] 测试缓动精度
  - [ ] 测试缓动曲线
- [ ] [ ] 实现 `TestAnimation_Timing`
  - [ ] 测试时间控制
  - [ ] 测试帧率
  - [ ] 测试时长
- [ ] [ ] 实现 `TestAnimation_Composite`
  - [ ] 测试复合动画
  - [ ] 测试动画序列
  - [ ] 测试动画同步

**验收标准**:
- [ ] 所有测试用例通过
- [ ] 覆盖率 ≥ 70%

---

### AI 集成（5 小时）

#### 任务 9.1: AI 集成测试

**目标**: 测试 AI 集成功能
**预计时间**: 5 小时
**文件**: `ai/ai_test.go`

**子任务**:

- [ ] [ ] 创建 `ai/ai_test.go` 文件
- [ ] [ ] 实现 `TestAI_Query`
  - [ ] 测试 AI 查询
  - [ ] 测试响应解析
  - [ ] 测试错误处理
- [ ] [ ] 实现 `TestAI_Completion`
  - [ ] 测试代码补全
  - [ ] 测试补全质量
  - [ ] 测试补全延迟
- [ ] [ ] 实现 `TestAI_Cache`
  - [ ] 测试缓存命中
  - [ ] 测试缓存失效
  - [ ] 测试缓存性能

**验收标准**:
- [ ] 所有测试用例通过
- [ ] 覆盖率 ≥ 70%

---

### DSL 系统（5 小时）

#### 任务 10.1: DSL 解析测试

**目标**: 测试 DSL 解析器
**预计时间**: 5 小时
**文件**: `dsl/parser_test.go`

**子任务**:

- [ ] [ ] 创建 `dsl/parser_test.go` 文件
- [ ] [ ] 实现 `TestParser_Syntax`
  - [ ] 测试语法正确
  - [ ] 测试语法错误
  - [ ] 测试错误恢复
- [ ] [ ] 实现 `TestParser_AST`
  - [ ] 测试 AST 生成
  - [ ] 测试 AST 遍历
  - [ ] 测试 AST 转换
- [ ] [ ] 实现 `TestParser_Execution`
  - [ ] 测试 DSL 执行
  - [ ] 测试执行结果
  - [ ] 测试执行性能

**验收标准**:
- [ ] 所有测试用例通过
- [ ] 覆盖率 ≥ 70%

---

### 其他辅助模块（5 小时）

#### 任务 11.1: Adapter 系统测试

**目标**: 测试适配器系统
**预计时间**: 2 小时
**文件**: `adapter/adapter_test.go`

**子任务**:

- [ ] [ ] 创建 `adapter/adapter_test.go` 文件
- [ ] [ ] 实现 `TestAdapter_Conversion`
  - [ ] 测试数据转换
  - [ ] 测试类型适配
  - [ ] 测试错误处理

---

#### 任务 11.2: Selection 系统测试

**目标**: 测试选择系统
**预计时间**: 3 小时
**文件**: `selection/selection_test.go`

**子任务**:

- [ ] [ ] 创建 `selection/selection_test.go` 文件
- [ ] [ ] 实现 `TestSelection_Select`
  - [ ] 测试文本选择
  - [ ] 测试区域选择
  - [ ] 测试多选
- [ ] [ ] 实现 `TestSelection_Navigate`
  - [ ] 测试选择导航
  - [ ] 测试边界处理
  - [ ] 测试键盘快捷键

---

## 每周执行计划

### Week 1: Layout 系统（P0）

**目标**: 完成所有 Layout 系统测试

| 天 | 任务 | 工时 | 状态 |
|----|------|------|------|
| Day 1 | 任务 1.1: 基础约束测试 | 4h | ⏳ |
| Day 2 | 任务 1.2: 测量功能测试 | 4h | ⏳ |
| Day 3 | 任务 1.3: Flexbox 布局测试（第 1 部分） | 3h | ⏳ |
| Day 4 | 任务 1.3: Flexbox 布局测试（第 2 部分） | 3h | ⏳ |
| Day 5 | 任务 1.4 + 1.5: 缓存 + 复杂场景 | 6h | ⏳ |

**验收标准**:
- [ ] Layout 系统覆盖率 ≥ 85%
- [ ] 所有测试通过
- [ ] 性能达标

---

### Week 2: Focus 系统（P0）

**目标**: 完成所有 Focus 系统测试

| 天 | 任务 | 工时 | 状态 |
|----|------|------|------|
| Day 1 | 任务 2.1: 焦点管理器测试 | 4h | ⏳ |
| Day 2 | 任务 2.2: 焦点导航测试 | 4h | ⏳ |
| Day 3 | 任务 2.3: 几何导航测试 | 3h | ⏳ |
| Day 4 | 任务 2.4: 焦点陷阱测试 | 2h | ⏳ |
| Day 5 | 任务 2.5: V3 焦点管理测试 + 集成 | 4h | ⏳ |

**验收标准**:
- [ ] Focus 系统覆盖率 ≥ 85%
- [ ] 所有测试通过

---

### Week 3: Input 系统（P0）

**目标**: 完成所有 Input 系统测试

| 天 | 任务 | 工时 | 状态 |
|----|------|------|------|
| Day 1 | 任务 3.1: 键盘映射测试 | 4h | ⏳ |
| Day 2 | 任务 3.2: 鼠标追踪测试 | 4h | ⏳ |
| Day 3 | 任务 3.3: 输入队列测试 | 3h | ⏳ |
| Day 4 | 任务 3.4: 平台输入测试 | 4h | ⏳ |
| Day 5 | 集成测试 + 性能测试 | 3h | ⏳ |

**验收标准**:
- [ ] Input 系统覆盖率 ≥ 80%
- [ ] 所有测试通过
- [ ] 跨平台兼容性验证

---

### Week 4-5: Paint 系统（P1）

**目标**: 完成 Paint 系统测试

| 天 | 任务 | 工时 | 状态 |
|----|------|------|------|
| Week 4 Day 1 | 任务 4.1: Compositor 测试 | 4h | ⏳ |
| Week 4 Day 2 | 任务 4.2: Context 测试 | 3h | ⏳ |
| Week 4 Day 3 | 任务 4.3: Style State 测试 | 2h | ⏳ |
| Week 4 Day 4 | 任务 4.4: Batch 测试 | 3h | ⏳ |
| Week 4 Day 5 | 任务 4.5: 集成测试（第 1 部分） | 3h | ⏳ |
| Week 5 Day 1 | 任务 4.5: 集成测试（第 2 部分） | 3h | ⏳ |
| Week 5 Day 2 | 性能测试 + 优化 | 3h | ⏳ |
| Week 5 Day 3 | 跨平台测试 | 2h | ⏳ |
| Week 5 Day 4-5 | 缓冲和修复 | 6h | ⏳ |

**验收标准**:
- [ ] Paint 系统覆盖率 ≥ 85%
- [ ] 性能：10000 cells < 50ms

---

### Week 6: Event 系统（P1）

**目标**: 完成 Event 系统测试

| 天 | 任务 | 工时 | 状态 |
|----|------|------|------|
| Day 1 | 任务 5.1: 事件分发补充测试 | 4h | ⏳ |
| Day 2 | 任务 5.2: 命中测试补充 | 4h | ⏳ |
| Day 3 | 任务 5.3: 阶段测试 | 4h | ⏳ |
| Day 4 | 集成测试 | 3h | ⏳ |
| Day 5 | 性能测试 + 缓冲 | 3h | ⏳ |

**验收标准**:
- [ ] Event 系统覆盖率 ≥ 80%
- [ ] 性能：1000 事件 < 10ms

---

### Week 7: Platform + Render 系统（P2）

**目标**: 完成 Platform 和 Render 测试

| 天 | 任务 | 工时 | 状态 |
|----|------|------|------|
| Day 1 | 任务 6.1-6.2: Platform 基础测试 | 4h | ⏳ |
| Day 2 | 任务 6.3-6.4: Platform 高级测试 | 4h | ⏳ |
| Day 3 | 任务 6.5: 跨平台兼容性测试 | 2h | ⏳ |
| Day 4 | 任务 7.1: 渲染流程测试 | 4h | ⏳ |
| Day 5 | 任务 7.2: 视口管理补充 | 4h | ⏳ |

**验收标准**:
- [ ] Platform 覆盖率 ≥ 75%
- [ ] Render 覆盖率 ≥ 80%

---

### Week 8: 其他模块（P3）

**目标**: 完成辅助模块测试

| 天 | 任务 | 工时 | 状态 |
|----|------|------|------|
| Day 1 | 任务 8.1: Animation 测试 | 5h | ⏳ |
| Day 2 | 任务 9.1: AI 集成测试 | 5h | ⏳ |
| Day 3 | 任务 10.1: DSL 测试 | 5h | ⏳ |
| Day 4 | 任务 11.1-11.2: Adapter + Selection | 5h | ⏳ |
| Day 5 | 集成测试 + 文档 | 5h | ⏳ |

**验收标准**:
- [ ] 辅助模块覆盖率 ≥ 70%

---

## 质量检查清单

### 代码质量

在提交测试代码前，检查：

- [ ] 测试用例描述清晰
- [ ] 使用 table-driven tests
- [ ] 使用 `assert.Equal`、`assert.NoError` 等辅助函数
- [ ] 测试独立性（无依赖）
- [ ] 测试覆盖率增加
- [ ] 代码风格一致（使用 `golangci-lint`）
- [ ] 无编译警告
- [ ] 无 `TODO` 或 `FIXME`

### 测试质量

- [ ] 覆盖正常场景
- [ ] 覆盖边界情况
- [ ] 覆盖异常情况
- [ ] 包含错误路径测试
- [ ] 包含性能基准测试
- [ ] 包含并发测试（如适用）
- [ ] 验证竞态条件（`go test -race`）

### 文档质量

- [ ] 测试文件有注释
- [ ] 复杂逻辑有解释
- [ ] 测试用例命名清晰
- [ ] 验收标准明确

### 性能标准

| 模块 | 指标 | 目标 | 状态 |
|------|------|------|------|
| **Layout** | 布局 1000 节点 | < 100ms | ⏳ |
| **Paint** | 渲染 10000 cells | < 50ms | ⏳ |
| **Event** | 分发 1000 事件 | < 10ms | ⏳ |
| **Input** | 处理 1000 输入 | < 5ms | ⏳ |
| **Focus** | 导航 1000 组件 | < 10ms | ⏳ |

---

## 测试工具链

### 运行测试

```bash
# 运行所有测试
go test ./runtime/...

# 运行带覆盖率的测试
go test ./runtime/... -coverprofile=coverage.out

# 查看覆盖率报告
go tool cover -html=coverage.out

# 运行竞态检测
go test ./runtime/... -race

# 运行性能基准
go test ./runtime/... -bench=. -benchmem

# 运行特定模块的测试
go test ./runtime/layout/...

# 运行特定测试
go test ./runtime/layout/... -run TestFlex

# 详细输出
go test ./runtime/... -v
```

### CI/CD 集成

建议的 GitHub Actions 工作流：

```yaml
name: Runtime Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [ubuntu-latest, windows-latest, macos-latest]
        go-version: ['1.21', '1.22']

    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v4
      with:
        go-version: ${{ matrix.go-version }}

    - name: Run tests
      run: go test ./runtime/... -race -coverprofile=coverage.out

    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out
```

---

## 进度报告模板

### 每周进度报告

```markdown
## Week N 进度报告

**日期**: YYYY-MM-DD
**模块**: [Module Name]
**计划任务**: X 个
**完成任务**: Y 个
**完成率**: Z%

### 完成的任务
- [x] Task 1.1: ...
- [x] Task 1.2: ...

### 遇到的问题
1. 问题描述
   - 影响
   - 解决方案

### 测试覆盖率
- 当前: X%
- 目标: Y%
- 差距: +/-%

### 下周计划
- Task 2.1: ...
- Task 2.2: ...
```

---

## 附录

### 测试模板

```go
package layout

import (
    "testing"

    "github.com/stretchr/testify/assert"
)

func Test<FeatureName>(t *testing.T) {
    tests := []struct {
        name     string
        input    InputType
        expected ExpectedType
        wantErr  bool
    }{
        {
            name:     "basic case",
            input:    ...,
            expected: ...,
            wantErr:  false,
        },
        // 更多测试用例
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Given: 准备测试数据

            // When: 执行测试操作
            result, err := SomeFunction(tt.input)

            // Then: 验证结果
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.expected, result)
            }
        })
    }
}

func Benchmark<FeatureName>(b *testing.B) {
    for i := 0; i < b.N; i++ {
        SomeFunction(testData)
    }
}
```

### 测试检查清单（提交前）

- [ ] 测试用例描述清晰
- [ ] 覆盖正常场景
- [ ] 覆盖边界情况
- [ ] 覆盖异常情况
- [ ] 使用 assert.NoError 等辅助函数
- [ ] 测试独立性（无依赖）
- [ ] 性能基准测试通过
- [ ] 竞态检测通过
- [ ] 代码覆盖率增加
- [ ] 代码风格检查通过
- [ ] 无编译警告

---

**文档结束**

**开始日期**: 2026-01-31
**预计完成**: 2026-03-31（8 周）
**总任务数**: 130
**预计工时**: 118 小时

**下一步**: 从 Week 1 开始，执行 Layout 系统测试。
