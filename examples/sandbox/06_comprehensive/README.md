# 综合示例

演示如何组合使用 Sandbox 的多个高级功能。

## 功能组合

本示例展示了以下功能的组合使用：

| 功能 | 用途 |
|------|------|
| **事件录制** | 记录用户操作序列 |
| **快照系统** | 保存和恢复关键状态 |
| **TestHelper** | 简化操作序列编写 |
| **队列统计** | 监控性能指标 |

## 运行应用

```bash
go run main.go
```

## 运行测试

```bash
go test -v

# 运行特定测试
go test -v -run TestComprehensiveWorkflow
go test -v -run TestRecordingSnapshotReplay
go test -v -run TestMultiStepSnapshotStrategy
go test -v -run TestPerformanceMonitoring
go test -v -run TestErrorRecoveryWithSnapshot
```

## 测试场景

### 1. 综合工作流 (TestComprehensiveWorkflow)
演示完整的测试流程：
- 设置事件录制
- 监控队列状态
- 使用 TestHelper 执行操作
- 创建快照
- 验证结果

### 2. 录制+快照+回放 (TestRecordingSnapshotReplay)
演示三种功能协作：
- 录制操作序列
- 保存快照
- 在新实例回放
- 恢复快照对比

### 3. 多步骤快照策略 (TestMultiStepSnapshotStrategy)
演示状态管理：
- 在每个关键步骤创建快照
- 列出所有快照
- 恢复到任意历史状态

### 4. 性能监控 (TestPerformanceMonitoring)
演示性能分析：
- 大量事件注入
- 队列内存追踪
- 淘汰事件统计

### 5. 错误恢复 (TestErrorRecoveryWithSnapshot)
演示容错机制：
- 建立基准状态快照
- 模拟错误操作
- 恢复到基准状态

## 代码模式

### 模式 1: 带录制和监控的测试
```go
// 设置录制
recorder := sandbox.NewEventRecorder(1000)
sb.SetRecorder(recorder)

// 获取初始状态
initialStats := sb.QueueStats()

// 执行操作
helper.Type("test").Tab().Press(platform.KeyEnter).Process()

// 检查结果
finalStats := sb.QueueStats()
events := recorder.Events()
```

### 模式 2: 带状态恢复的测试
```go
// 建立基准状态
helper.Type("name").Tab().Process()
baseSnapshot, _ := sb.Snapshot(sandbox.SnapshotStandard, "baseline")

// 执行可能失败的操作
riskyOperation()

// 恢复到基准状态
sb.Restore(baseSnapshot)
```

### 模式 3: 录制和回放
```go
// 录制
recorder := sandbox.NewEventRecorder(1000)
sb.SetRecorder(recorder)
doOperation()
events := recorder.Events()

// 回放
sb2.InjectRaw(events...)
```

## 相关文档

- [SANDBOX_ADVANCED_FEATURES.md](../../../docs/sandbox/SANDBOX_ADVANCED_FEATURES.md)
- [01_event_recording](../01_event_recording/)
- [02_snapshot](../02_snapshot/)
- [03_test_helper](../03_test_helper/)
- [04_queue_stats](../04_queue_stats/)
