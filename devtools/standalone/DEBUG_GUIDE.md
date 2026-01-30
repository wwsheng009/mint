# 输入调试指南

## 问题：程序启动后按键没有反应

### 调试步骤

#### 第一步：测试原始输入读取

运行独立的输入测试程序，验证平台层输入是否正常：

```bash
cd E:\projects\yao\wwsheng009\mint
.\test_input.exe
```

这个程序会：
- 初始化 Logger 记录所有输入
- 直接使用 `platform.InputReader` 读取输入
- 打印每个输入事件的详细信息
- 按 ESC 退出

**预期输出**（按键测试）：
```
[Event #1]
  Type:        0
  Key:         a (97)
  Special:     0
  Modifiers:   0
  ...
------------------------------
```

如果这个程序有反应，说明平台层输入读取是正常的。

#### 第二步：分析日志文件

程序运行后，日志会保存到 `C:\Users\vince\.mint\devtools\logs\` 目录。

使用 `mint-debugger` 分析：

```bash
# 分析最新的日志
mint-debugger

# 或指定日志文件
mint-debugger -log "C:\Users\vince\.mint\devtools\logs\session_20260130_xxxxxx.log"
```

#### 第三步：对比 framework 和 engine 的差异

**framework/app.go 的工作流程：**
```
1. App.Init()
   ├─ platform.NewInputReader()
   ├─ frameworkevent.NewPump(reader)
   └─ pump.Start()

2. App.Run()
   ├─ for a.state == StateRunning
   │   ├─ case ev := <-eventChan:     // 从 pump.Events() 读取
   │   ├─ case <-ticker.C:             // 定时渲染
   │   └─ case <-a.quit:               // 退出
```

**engine/main.go 的工作流程：**
```
1. eng := engine.New()
   └─ inputEvents: make(chan RawInput, 50)

2. eng.Run()
   ├─ platform.NewInputReader()
   ├─ inputReader.Start(inputEvents)  // 发送到 inputEvents
   ├─ go convertInputLoop()            // 从 inputEvents 读取并转换
   └─ for {
       ├─ case ev := <-eventQueue:     // 从 eventQueue 读取
       ├─ case <-ticker.C:
       └─ case <-e.quit:
    }
```

### 可能的问题点

#### 1. `convertRawInput` 中的问题

检查 `engine.go` 中的 `convertRawInput` 方法是否正确处理了所有输入类型。

#### 2. 事件过滤问题

检查 `handleEvent` 中是否有事件被意外过滤掉。

#### 3. 布局框问题

检查 `SetLayoutBoxes` 是否被正确调用，布局框是否正确设置。

### 调试模式

设置环境变量启用调试输出：

```bash
# 启用渲染调试
set TUI_OUTPUT_MODE=debug

# 启用调试模式
set TUI_DEBUG=true

# 运行程序
.\test_engine.exe
```

### 日志事件对照表

| 事件 | 说明 | 期望看到 |
|------|------|----------|
| `session_start` | 程序启动 | ✅ 应该有 |
| `component_add` | 组件添加 | ✅ 应该有 (3个按钮) |
| `mouse` | 鼠标事件 | ⚠️ 需要鼠标操作 |
| `focus` | 焦点事件 | ⚠️ 需要鼠标点击 |
| `key` | 键盘事件 | ⚠️ 需要按键 |

如果缺少某些事件类型，说明对应的事件处理流程有问题。

### 已知问题和修复

#### 问题1：空 RawInput 被当作有效事件

**原因**：`InputKeyPress = 0`，所以空的 `RawInput{Timestamp: now}` 会被当作按键事件。

**修复**：在 `input_windows.go` 中，无效输入返回 `Type: -1`。

#### 问题2：按键释放事件被发送

**原因**：`KeyDown == 0` 时返回空输入。

**修复**：按键释放时返回 `Type: -1`。

### 下一步

如果测试输入程序正常工作，但 engine 程序不工作：

1. 检查 `engine.go` 中的 `convertRawInput` 方法
2. 检查事件是否正确发送到 `eventQueue`
3. 检查 `handleEvent` 是否正确处理事件
4. 添加更多日志到 engine 中的关键位置
