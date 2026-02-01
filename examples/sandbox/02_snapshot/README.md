# 快照系统示例

演示如何使用 MockSandbox 的快照系统保存和恢复应用状态。

## 功能说明

- **三种快照级别**:
  - `SnapshotMinimal`: 仅渲染缓冲区
  - `SnapshotStandard`: 缓冲区 + 事件历史
  - `SnapshotFull`: 包括应用状态

- **快照操作**:
  - 创建快照：`Snapshot(level, tags)`
  - 恢复快照：`Restore(snapshot)`
  - 列出快照：`ListSnapshots()`
  - 查找快照：`FindSnapshots(tag)`
  - 删除快照：`DeleteSnapshot(id)`

## 运行应用

```bash
go run main.go
```

## 运行测试

```bash
go test -v

# 运行特定测试
go test -v -run TestSnapshotLevels
go test -v -run TestSnapshotSaveAndRestore
go test -v -run TestMultipleSnapshots
go test -v -run TestSnapshotWithInput
go test -v -run TestSnapshotDelete
```

## API 参考

```go
// 创建快照
snapshot, err := sb.Snapshot(sandbox.SnapshotStandard, "tag1", "tag2")

// 恢复快照
err = sb.Restore(snapshot)

// 列出所有快照
snapshots := sb.ListSnapshots()

// 按标签查找
tagged := sb.FindSnapshots("tag1")

// 删除快照
sb.DeleteSnapshot(snapshot.ID)

// 清空所有快照
sb.ClearSnapshots()
```

## 相关文档

- [SANDBOX_ADVANCED_FEATURES.md](../../../docs/sandbox/SANDBOX_ADVANCED_FEATURES.md)
