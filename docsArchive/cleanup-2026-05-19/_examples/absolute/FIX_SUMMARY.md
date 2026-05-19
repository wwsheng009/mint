# Absolute 示例问题的修复汇总

## 修复的问题

### 1. ✓ Count 只能递增一次

**位置**：`examples/absolute/main.go:23-27`

**原因**：
- `ui.On` 使用全局 `registeredHandlers` 避免重复注册
- handler 只注册一次，闭包捕获第一次渲染时的 `count` 值（0）
- 每次点击执行 `setCount(0 + 1)`，count 总是变成 1

**修复**：
```go
// ❌ 修复前
ui.On(IncrementIntent{}, func() {
    setCount(count + 1)  // 闭包捕获旧值
})

// ✅ 修复后：使用函数式更新
ui.On(IncrementIntent{}, func() {
    setCount(func(c int) int {
        return c + 1
    })
})
```

### 2. ? Click count 无法显示

**状态**：已分析，需要进一步验证

**可能原因**：
- 布局计算异常（高度超出终端范围）
- 嵌套 VStack 高度计算错误
- AbsoluteBuilder 作为子元素占用 Flex 空间

**临时解决方案**：
1. 将 `WithHeight(15)` 增加到 `20`
2. 使用 `WithValue` 增加终端高度
3. 移除嵌套 VStack

## 文档

已创建以下文档：
- `ANALYSIS.md` - AbsoluteBuilder 空间占用分析
- `COUNT_FIX.md` - Count 递增问题详解
- `ZINDEX_EXPLAINED.md` - ZIndex 作用说明

## 编译验证

```bash
cd E:\projects\yao\wwsheng009\mint
go build -o mint.exe examples/absolute/main.go
```

编译成功 ✓

## 运行测试

```bash
./mint.exe
```

测试中（PID: 30152）

## 下一步

1. 测试 Count 是否能正常递增
2. 检查 "Click count" 是否能正常显示
3. 如果仍有问题，添加布局调试日志
