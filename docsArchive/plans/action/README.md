# Action 统一方案 - 快速参考

## 文件索引

| 文件 | 内容 |
|------|------|
| `00_overview.md` | 总览与目标 |
| `01_architecture.md` | 架构设计与接口定义 |
| `02_phase1_integration.md` | 核心系统集成 |
| `03_phase2_migration.md` | 组件迁移 |
| `04_phase3_cleanup.md` | 清理旧代码 |
| `05_phase4_enhancements.md` | 增强功能 |

## 迁移时间线

```
Week 1-2: Phase 0 - 架构设计
    │
    ├── 定义接口
    ├── 编写适配器
    └── 单元测试

Week 3-4: Phase 1 - 核心集成
    │
    ├── 修改 App 结构
    ├── 集成 ActionRouter
    └── 集成测试

Week 5-8: Phase 2 - 组件迁移
    │
    ├── P0 组件 (Button, TextInput, List)
    ├── P1 组件 (Checkbox, Select, Table)
    └── P2/P3 组件 (Tree, Form, Inspector)

Week 9-10: Phase 3 - 清理
    │
    ├── 删除旧代码
    ├── 更新文档
    └── 发布 v1.0

Week 11+: Phase 4 - 增强功能
    │
    ├── 异步 Action
    ├── 中间件
    └── 撤销/重做
```

## 关键接口速查

### ActionTarget（新统一接口）

```go
type ActionTarget interface {
    HandleAction(action *Action) bool
    GetSupportedActions() []ActionType
    CanHandleAction(action *Action) bool
}
```

### Action 结构

```go
type Action struct {
    Type     ActionType      // 语义化类型
    Payload  interface{}     // 数据
    Source   string          // 来源
    TargetID uint64          // 目标
    Timestamp time.Time      // 时间戳
    Meta      map[string]interface{} // 元数据
}
```

### ActionRouter 分发

```go
result := router.Dispatch(action)
// result.Handled - 是否被处理
// result.Stopped - 是否停止传播
// result.Phase  - 处理阶段 (Capture/Target/Bubble)
```

## 常用 ActionType

| 类型 | 用途 |
|------|------|
| `ActionClick` | 鼠标点击 |
| `ActionInputText` | 文本输入 |
| `ActionNavigateUp/Down/Left/Right` | 方向导航 |
| `ActionSelect` | 选择项 |
| `ActionEnter` | 回车确认 |
| `ActionBackspace` | 退格删除 |
| `ActionFocus/Blur` | 焦点控制 |
| `ActionScroll` | 滚动 |
| `ActionSubmit/Cancel` | 表单提交/取消 |

## 迁移检查清单

### 组件迁移

- [ ] 添加 `nodeID` 字段
- [ ] 实现 `HandleAction()`
- [ ] 实现 `GetSupportedActions()`
- [ ] 实现 `CanHandleAction()`
- [ ] 删除旧接口实现
- [ ] 更新测试
- [ ] 更新文档

### App 迁移

- [ ] 添加 `actionRouter`
- [ ] 添加 `inputProcessor`
- [ ] 添加 `actionRegistry`
- [ ] 修改主循环
- [ ] 实现注册表构建
- [ ] 集成测试

## 风险与缓解

| 风险 | 缓解措施 |
|------|---------|
| 组件改动量大 | 使用适配器过渡 |
| 行为不一致 | 保留 legacyMode 开关 |
| 性能下降 | Action 对象池复用 |
| 学习成本 | 提供迁移指南和示例 |

## 联系方式

有问题请联系架构团队或在 GitHub Issues 中提出。
