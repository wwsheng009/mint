# TestHelper 链式 API 示例

演示如何使用 MockSandbox 的 TestHelper 链式 API 简化测试代码。

## 功能说明

TestHelper 提供流畅的链式 API，让测试代码更简洁易读：

```go
helper.
    Type("username").
    Press(platform.KeyTab).
    Type("password").
    Press(platform.KeyEnter).
    AssertRender("Welcome").
    Result()
```

## 可用方法

| 方法 | 描述 |
|------|------|
| `Type(text)` | 输入文本（带延迟） |
| `TypeFast(text)` | 快速输入（无延迟） |
| `Press(key)` | 按键 |
| `PressWithMod(key, mod)` | 带修饰符按键 |
| `Tab()` | 按 Tab 键 |
| `Enter()` | 按 Enter 键 |
| `Click(x, y)` | 鼠标点击 |
| `Wait(d)` | 等待指定时间 |
| `Process()` | 处理所有操作 |
| `AssertRender(text)` | 断言渲染包含文本 |
| `Result()` | 获取操作结果 |

## 运行应用

```bash
go run main.go
```

## 运行测试

```bash
go test -v

# 运行特定测试
go test -v -run TestHelperBasic
go test -v -run TestHelperFormSubmit
go test -v -run TestHelperWait
go test -v -run TestHelperClear
go test -v -run TestHelperComplexSequence
go test -v -run TestHelperKeyboardShortcuts
go test -v -run TestHelperTypeFast
```

## API 参考

```go
helper := sb.Helper()

// 简单输入
helper.Type("hello").Process()

// 带按键操作
helper.Type("username").
    Press(platform.KeyTab).
    Type("password").
    Press(platform.KeyEnter).
    Process()

// 带等待
helper.Type("test").
    Wait(100 * time.Millisecond).
    Press(platform.KeyEnter).
    Process()

// 带断言
helper.Type("search").
    Press(platform.KeyEnter).
    AssertRender("results").
    Result()

// 快速输入
helper.TypeFast("quicktyping").Process()

// 获取结果
result := helper.Type("test").Process().Result()
if !result.OK() {
    t.Error(result.Error())
}
```

## 相关文档

- [SANDBOX_ADVANCED_FEATURES.md](../../../docs/sandbox/SANDBOX_ADVANCED_FEATURES.md)
