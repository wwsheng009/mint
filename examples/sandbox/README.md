# Sandbox 高级功能示例

本目录包含 Mint TUI 框架 Sandbox 高级功能的完整示例代码。

## 目录结构

```
examples/sandbox/
├── demo/                    # 原有基础示例
├── 01_event_recording/      # 事件录制与回放
├── 02_snapshot/             # 快照系统
├── 03_test_helper/          # TestHelper 链式 API
├── 04_queue_stats/          # 队列统计与监控
├── 05_injection_strategy/   # 事件注入策略
└── 06_comprehensive/          # 综合示例
```

## 快速开始

### 运行单个示例

```bash
# 进入示例目录
cd 01_event_recording

# 运行应用
go run main.go

# 运行测试
go test -v
```

### 运行所有示例测试

```bash
# 从项目根目录运行所有测试
go test ./examples/sandbox/... -v
```

## 示例说明

### 01_event_recording - 事件录制与回放

演示如何录制用户操作序列，然后在测试中回放。

**关键 API**:
- `sandbox.NewEventRecorder(maxSize)` - 创建录制器
- `recorder.Record(event)` - 录制事件
- `recorder.Events()` - 获取录制的事件
- `recorder.SaveToFile(filename)` - 保存到文件
- `recorder.LoadFromFile(filename)` - 从文件加载

**使用场景**: 自动化测试、bug 重现、操作演示

### 02_snapshot - 快照系统

演示如何保存和恢复应用状态。

**关键 API**:
- `sb.Snapshot(level, tags)` - 创建快照
- `sb.Restore(snapshot)` - 恢复快照
- `sb.ListSnapshots()` - 列出所有快照
- `sb.FindSnapshots(tag)` - 按标签查找
- `sb.DeleteSnapshot(id)` - 删除快照
- `sb.ClearSnapshots()` - 清空所有快照

**快照级别**:
- `SnapshotMinimal` - 仅渲染缓冲区
- `SnapshotStandard` - 缓冲区 + 事件历史
- `SnapshotFull` - 包括应用状态

**使用场景**: 状态调试、测试隔离、检查点恢复

### 03_test_helper - TestHelper 链式 API

演示如何使用流畅的链式 API 简化测试代码。

**关键 API**:
```go
helper.
    Type("text").
    Press(platform.KeyEnter).
    Wait(100 * time.Millisecond).
    AssertRender("expected").
    Result()
```

**可用方法**:
- `Type(text)` / `TypeFast(text)` - 输入文本
- `Press(key)` - 按键
- `Tab()` / `Enter()` - 快捷键
- `Click(x, y)` - 鼠标点击
- `Wait(d)` - 等待
- `Process()` - 处理所有操作
- `AssertRender(text)` - 断言

**使用场景**: 简化测试代码、提高可读性

### 04_queue_stats - 队列统计与监控

演示如何监控事件队列的状态。

**关键 API**:
- `sb.QueueStats()` - 获取队列统计
- `stats.Length` - 当前队列长度
- `stats.MaxSize` - 队列容量
- `stats.MemoryUsed` - 已用内存
- `stats.MemoryLimit` - 内存限制
- `stats.EvictCount` - 淘汰事件数

**使用场景**: 性能分析、容量规划、内存监控

### 05_injection_strategy - 事件注入策略

演示不同的事件注入策略控制。

**三种策略**:
- `InjectProhibited` - 禁止注入（生产环境）
- `InjectAllowed` - 允许注入（测试环境，默认）
- `InjectRecorded` - 仅录制不处理

**关键 API**:
- `sb.Injector().SetStrategy(strategy)` - 设置策略
- `sb.Injector().Strategy()` - 获取当前策略

**使用场景**: 环境隔离、安全控制、调试模式

### 06_comprehensive - 综合示例

演示多个功能的组合使用。

**组合场景**:
- 录制 + 快照 + 回放
- 多步骤快照策略
- 性能监控
- 错误恢复

**使用场景**: 复杂测试场景、端到端测试

## 依赖要求

所有示例都依赖于以下包：

```go
import (
    "github.com/wwsheng009/mint/ui"
    "github.com/wwsheng009/mint/sandbox"
    "github.com/wwsheng009/mint/sandbox/mock"
    "github.com/wwsheng009/mint/runtime/platform"
)
```

## 运行环境

- Go 1.21+
- 需要 `github.com/wwsheng009/mint` 框架

## 相关文档

- [SANDBOX_ADVANCED_FEATURES.md](../../docs/sandbox/SANDBOX_ADVANCED_FEATURES.md) - 高级功能完整参考
- [QUICK_START_GUIDE.md](../../docs/sandbox/QUICK_START_GUIDE.md) - 快速入门指南
- [API_REFERENCE.md](../../docs/sandbox/API_REFERENCE.md) - API 参考

## 贡献

欢迎提交问题和改进建议！
