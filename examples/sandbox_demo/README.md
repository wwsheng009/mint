# Sandbox 使用示例

本目录演示如何使用 Sandbox 运行和测试 Mint TUI 应用。

## 目录

- `main.go` - 示例应用（计数器）
- `test_mock.go` - 使用 Mock 沙箱测试
- `test_real.go` - 使用真实沙箱运行
- `test_replay.go` - 事件回放示例
- `test_chain.go` - 链式 API 示例
- `test_snapshot.go` - 快照功能示例

## 快速开始

### 1. 运行示例应用

```bash
go run main.go
```

### 2. 使用 Mock 沙箱测试

```bash
go test -v test_mock.go main.go
```

### 3. 使用真实沙箱

```bash
go run test_real.go
```

## 应用结构

```go
// main.go
func Counter() ui.VNode {
    // ... 应用逻辑 ...
}

func main() {
    ui.Run(Counter)
}
```

## Sandbox 集成

### Mock 沙箱测试

```go
func TestCounterWithMock(t *testing.T) {
    testApp, err := ui.TestRun(Counter, ui.TestWithSize(40, 12))
    if err != nil {
        t.Fatal(err)
    }
    defer testApp.Close()

    helper := testApp.Helper()

    // 测试点击 "+" 按钮
    result := helper.
        Tab().                          // 导航到按钮
        Tab().
        Enter().                        // 点击
        Process().
        AssertRender("Count: 1").       // 验证计数增加
        Result()

    if !result.OK() {
        t.Error(result.Error())
    }
}
```

### 真实沙箱运行

```go
func main() {
    sb, err := real.New(80, 24)
    if err != nil {
        log.Fatal(err)
    }
    defer sb.Close()

    sb.Initialize(nil)
    sb.Run()

    // ... 应用运行，用户交互 ...
}
```

## 更多示例

详见各个测试文件，了解完整用法。
