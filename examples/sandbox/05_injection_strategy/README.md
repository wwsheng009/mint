# 事件注入策略示例

演示如何使用 MockSandbox 的事件注入策略控制事件行为。

## 功能说明

三种注入策略：

| 策略 | 描述 | 使用场景 |
|------|------|----------|
| `InjectProhibited` | 禁止注入，仅录制 | 生产环境 |
| `InjectAllowed` | 允许注入和录制 | 测试环境（默认） |
| `InjectRecorded` | 仅录制不处理 | 调试/分析 |

## 运行应用

```bash
go run main.go
```

## 运行测试

```bash
go test -v

# 运行特定测试
go test -v -run TestInjectAllowed
go test -v -run TestInjectProhibited
go test -v -run TestInjectRecorded
go test -v -run TestStrategySwitch
go test -v -run TestStrategyWithApp
go test -v -run TestInjectErrorHandling
go test -v -run TestRecordingWithDifferentStrategies
go test -v -run TestStrategyInProduction
```

## API 参考

```go
// 获取注入器
injector := sb.Injector()

// 设置策略
injector.SetStrategy(sandbox.InjectAllowed)
injector.SetStrategy(sandbox.InjectProhibited)
injector.SetStrategy(sandbox.InjectRecorded)

// 获取当前策略
strategy := injector.Strategy()

// 配置文件方式
config := &sandbox.Config{
    Event: sandbox.EventConfig{
        Strategy: sandbox.InjectProhibited,
    },
}
sb := mock.NewWithConfig(config)
```

## 使用场景

### 生产环境配置
```go
config := sandbox.RealConfig()  // 禁止注入
```

### 测试环境配置
```go
config := sandbox.MockConfig()  // 允许注入
```

### 调试/分析配置
```go
config := &sandbox.Config{
    Event: sandbox.EventConfig{
        Strategy: sandbox.InjectRecorded,  // 仅录制
    },
}
```

## 相关文档

- [SANDBOX_ADVANCED_FEATURES.md](../../../docs/sandbox/SANDBOX_ADVANCED_FEATURES.md)
