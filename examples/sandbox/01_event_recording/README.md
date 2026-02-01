# 事件录制与回放示例

演示如何使用 MockSandbox 的事件录制和回放功能。

## 功能说明

- **事件录制**: 自动记录用户操作序列（键盘、鼠标、窗口调整等）
- **事件回放**: 在新实例中重现录制的操作
- **文件存储**: 支持将录制保存到文件并加载

## 运行应用

```bash
# 运行交互式应用
go run main.go
```

## 运行测试

```bash
# 运行所有演示测试
go test -v

# 运行特定测试
go test -v -run TestEventRecording
go test -v -run TestEventReplay
go test -v -run TestRecordingToFile
```

## 测试说明

### TestEventRecording
演示事件录制功能：
1. 创建 EventRecorder
2. 设置到 MockSandbox
3. 执行操作（自动录制）
4. 查看录制结果
5. 保存到文件

### TestEventReplay
演示事件回放功能：
1. 在第一个实例录制操作
2. 获取录制的事件
3. 在新实例回放这些事件
4. 验证结果一致性

### TestRecordingToFile
演示文件存储功能：
1. 录制操作序列
2. 保存到 JSON 文件
3. 从文件加载录制
4. 统计事件信息

## API 参考

```go
// 创建录制器
recorder := sandbox.NewEventRecorder(maxSize)

// 设置到沙箱
sb.SetRecorder(recorder)

// 获取录制的事件
events := recorder.Events()

// 保存/加载文件
recorder.SaveToFile("recording.json")
recorder.LoadFromFile("recording.json")

// 清空录制
recorder.Clear()
```

## 相关文档

- [SANDBOX_ADVANCED_FEATURES.md](../../../docs/sandbox/SANDBOX_ADVANCED_FEATURES.md)
