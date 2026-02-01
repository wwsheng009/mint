# 队列统计与监控示例

演示如何使用 MockSandbox 的队列统计功能监控事件队列。

## 功能说明

队列统计提供以下信息：

| 字段 | 描述 |
|------|------|
| `Length` | 当前队列长度 |
| `MaxSize` | 队列容量限制 |
| `MemoryUsed` | 已用内存（字节） |
| `MemoryLimit` | 内存限制 |
| `EvictCount` | 已淘汰事件数 |

## 运行应用

```bash
go run main.go
```

## 运行测试

```bash
go test -v

# 运行特定测试
go test -v -run TestQueueBasicStats
go test -v -run TestQueueAfterEvents
go test -v -run TestQueueMemoryMonitoring
go test -v -run TestQueueEviction
go test -v -run TestQueueDuringInteraction
go test -v -run TestQueueComparison
```

## API 参考

```go
// 获取队列统计
stats := sb.QueueStats()

// 访问统计信息
length := stats.Length
maxSize := stats.MaxSize
memoryUsed := stats.MemoryUsed
memoryLimit := stats.MemoryLimit
evictCount := stats.EvictCount

// 计算使用率
usage := float64(length) / float64(maxSize) * 100
memoryUsage := float64(memoryUsed) / float64(memoryLimit) * 100
```

## 使用场景

1. **性能监控**: 追踪事件处理速度
2. **内存分析**: 监控队列内存使用
3. **容量规划**: 根据统计调整队列配置
4. **调试**: 了解事件积压情况

## 相关文档

- [SANDBOX_ADVANCED_FEATURES.md](../../../docs/sandbox/SANDBOX_ADVANCED_FEATURES.md)
