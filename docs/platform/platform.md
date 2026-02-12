这问题问得很对味 —— **跨平台 TUI 真正的坑王不是布局，是输入系统**。
键盘 + 鼠标在不同终端、不同 OS、不同协议下全是“方言”，你不做抽象，后面全是屎山。

我给你一套**工程级统一输入架构**，不是“用个库就完了”那种答案。

---

# 一、核心认知：终端根本没有“键盘事件”

GUI 世界：

> 有 KeyDown / KeyUp / MouseMove 事件对象

终端世界：

> 只有 **字节流（ANSI 转义序列）**

所有键盘和鼠标，本质都是：

```
ESC [ A        ← 上箭头
ESC [ <0;12;5M ← 鼠标按下
\x03           ← Ctrl+C
```

也就是说你要做的是：

> **把混乱的字节流 → 统一的“语义事件”**

这一步叫：

> 🎯 **Input Normalization Layer（输入归一化层）**

---

# 二、正确架构（大多数人都少了中间那层）

```
┌──────────────┐
│  Terminal    │  ← Windows ConPTY / Linux PTY / macOS PTY
└──────┬───────┘
       │ byte stream
┌──────▼────────────┐
│  Decoder Layer     │  ← 解析 ANSI、CSI、SGR、UTF-8
└──────┬────────────┘
       │ Raw Events (KeyCode, Rune, CSI seq)
┌──────▼────────────┐
│  Normalization     │  ← ★ 关键层（跨平台统一）
│  (Your Framework)  │
└──────┬────────────┘
       │ Unified Events
┌──────▼────────────┐
│  App Input System  │
└────────────────────┘
```

**90% TUI 框架只做到 Decoder，没做真正的 Normalization。**

---

# 三、你要统一的“坑点清单”

| 问题        | Linux/macOS       | Windows         | 坑点           |
| --------- | ----------------- | --------------- | ------------ |
| Backspace | `\x7f`            | `\x08`          | 同一个键两个值      |
| Enter     | `\r`              | `\r` or `\r\n`  | 行为不一致        |
| Alt+Key   | `ESC + key`       | 有时无 ESC         | 会被误判成 ESC    |
| Ctrl+字母   | 1–26 控制码          | 一样但有例外          | Ctrl+I = Tab |
| Shift+Tab | `ESC[Z`           | 不一定支持           |              |
| F1-F12    | 不同 CSI            | 不同              | F5 每家都不同     |
| 鼠标协议      | X10 / SGR / URXVT | ConPTY 转换       | 坐标编码不同       |
| 滚轮        | 按键模拟              | 同               |              |
| UTF-8     | 终端处理              | Windows 老控制台不支持 |              |

---

# 四、统一事件模型（关键）

你程序内部 **绝对不要直接用字节序列**。
必须设计一个**语义层事件模型**：

```go
type Event interface{}

type KeyEvent struct {
    Key      KeyCode   // Up, Down, Enter, F1...
    Rune     rune      // 普通字符
    Mod      ModMask   // Ctrl/Alt/Shift
    Type     KeyType   // Press, Repeat
}

type MouseEvent struct {
    X, Y     int
    Button   MouseButton
    Action   MouseAction // Press, Release, Drag, ScrollUp
    Mod      ModMask
}

type PasteEvent struct {
    Data string
}

type ResizeEvent struct {
    Width, Height int
}
```

以后业务层 **只认识这个**。

---

# 五、Normalization 层要干的脏活

### 1️⃣ 键盘统一规则

| 输入               | 统一成          |
| ---------------- | ------------ |
| `\x7f` or `\x08` | KeyBackspace |
| `\r` or `\n`     | KeyEnter     |
| `ESC + a`        | Alt+a        |
| `\x01`           | Ctrl+A       |
| `\t`             | Tab          |
| `ESC[Z`          | Shift+Tab    |

### 2️⃣ 修复 Ctrl 冲突

```
Ctrl+I = Tab
Ctrl+M = Enter
Ctrl+[ = Esc
```

你必须区分：

```go
if r == '\t' && modCtrl { -> Ctrl+I } else { -> Tab }
```

---

### 3️⃣ Alt 键歧义解决（大坑）

用户按 Alt+A 时：

```
ESC a
```

但用户单独按 ESC 也是 ESC。

解决办法：**超时判断**

```
读到 ESC
   ↓
等待 5~10ms 看是否还有后续字节
   有 → Alt 组合键
   无 → 真 ESC
```

这就是专业 TUI 框架和玩具的区别。

---

### 4️⃣ 鼠标协议统一

终端有 3 种主流协议：

| 协议    | 序列形式          |
| ----- | ------------- |
| X10   | `ESC[M...`    |
| URXVT | `ESC[...M`    |
| SGR   | `ESC[<b;x;yM` |

你要全部解析成：

```go
MouseEvent{
  X:  col-1,
  Y:  row-1,
  Button: Left,
  Action: Press,
}
```

并自动启用：

```
CSI ?1000h  ← 鼠标
CSI ?1006h  ← SGR 扩展
```

---

### 5️⃣ Windows 特殊处理

Windows 不是 ANSI，而是：

```
INPUT_RECORD
KEY_EVENT_RECORD
MOUSE_EVENT_RECORD
```

但 ConPTY 会转成 ANSI，不过：

| 问题            | 解决               |
| ------------- | ---------------- |
| Ctrl+Space 丢失 | 手动映射 VK          |
| 鼠标滚轮丢失        | 监听 MOUSE_WHEELED |
| AltGr 混乱      | 识别 Ctrl+Alt      |

---

# 六、最佳工程方案（别重复造轮子）

如果你自己写框架：

### 必须有 3 层：

```
input/
 ├── reader.go        ← 读字节流
 ├── decoder.go       ← ANSI 解析
 ├── normalize.go     ← ★ 跨平台统一
```

### Bubble Tea / tcell / termbox 为什么强？

因为它们有：

✔ CSI 状态机
✔ UTF-8 流式解析
✔ ESC 超时判断
✔ 鼠标协议识别
✔ Windows 适配

---

# 七、真正高手才会做的优化

### 🔥 1. 输入节流

鼠标移动 1000 次/秒：

```
只保留最后一次
```

否则 UI 抖成鬼。

---

### 🔥 2. 事件合并

```
KeyRepeat × 50 → 合并
Resize × 10 → 只留最后一个
```

---

### 🔥 3. 粘贴检测（Bracketed Paste Mode）

开启：

```
CSI ?2004h
```

终端粘贴时：

```
ESC[200~  内容  ESC[201~
```

你就不会把粘贴当成 1000 次按键。

---

# 八、一句话总结架构思想

> **不要做“按键处理”**
> 要做：
> 🎯 **终端字节协议 → 跨平台语义事件系统**

---

如果你愿意，我可以下一步直接给你：

* Go 版 Normalize 核心代码框架
* ANSI 状态机实现思路
* 鼠标 SGR 解析表

这部分才是 TUI 框架真正的“内功心法”。



好，直接进“内功实现层”。下面这套就是 **TUI 输入系统的核心代码骨架思路（Go 向，但思想通用）**。

我们实现 3 个关键引擎：

1. **ANSI/UTF-8 流式解码状态机**
2. **ESC 歧义解决器（Alt vs Esc）**
3. **跨平台 Normalize 核心规则**

---

# 一、输入读取器（字节流层）

终端输入不是按键事件，是**无边界字节流**，必须流式处理。

```go
type Reader struct {
    r io.Reader
    buf [4096]byte
}

func (rd *Reader) ReadLoop(out chan<- byte) {
    for {
        n, err := rd.r.Read(rd.buf[:])
        if err != nil { close(out); return }
        for i := 0; i < n; i++ {
            out <- rd.buf[i]
        }
    }
}
```

⚠️ 不能按包读，ANSI 序列可能被拆开。

---

# 二、ANSI 状态机（Decoder）

我们要识别：

| 类型     | 示例               |
| ------ | ---------------- |
| 普通字符   | UTF-8            |
| 控制字符   | `\x03`           |
| ESC 序列 | `ESC [ A`        |
| 鼠标 SGR | `ESC [ <0;12;5M` |

### 状态定义

```go
type State int

const (
    StateGround State = iota
    StateEscape
    StateCSI
    StateUTF8
)
```

---

### 主解析循环

```go
func Decode(in <-chan byte, out chan<- RawEvent) {
    state := StateGround
    var csiBuf []byte
    var utfBuf []byte

    for b := range in {
        switch state {

        case StateGround:
            switch {
            case b == 0x1b:
                state = StateEscape
            case b < 0x20:
                out <- RawCtrl{Code: b}
            case b < 0x80:
                out <- RawRune{Rune: rune(b)}
            default:
                utfBuf = append([]byte{b})
                state = StateUTF8
            }

        case StateEscape:
            if b == '[' {
                csiBuf = nil
                state = StateCSI
            } else {
                out <- RawEsc{}
                state = StateGround
            }

        case StateCSI:
            csiBuf = append(csiBuf, b)
            if isCSIFinal(b) {
                out <- RawCSI{Data: append([]byte{}, csiBuf...)}
                state = StateGround
            }

        case StateUTF8:
            utfBuf = append(utfBuf, b)
            if utf8.FullRune(utfBuf) {
                r, _ := utf8.DecodeRune(utfBuf)
                out <- RawRune{Rune: r}
                state = StateGround
            }
        }
    }
}
```

---

# 三、ESC 歧义解决器（最关键的黑魔法）

ESC 有三种可能：

| 情况        | 含义      |
| --------- | ------- |
| 单独 ESC    | 退出/取消   |
| ESC + key | Alt 组合  |
| ESC [     | ANSI 序列 |

做法：**延迟判断**

```go
const escTimeout = 10 * time.Millisecond

func ResolveEsc(in <-chan RawEvent, out chan<- RawEvent) {
    for ev := range in {
        if _, ok := ev.(RawEsc); ok {
            select {
            case next := <-in:
                // 有后续，说明是 Alt
                if r, ok := next.(RawRune); ok {
                    out <- RawAltRune{Rune: r.Rune}
                } else {
                    out <- next
                }
            case <-time.After(escTimeout):
                out <- RawKey{Key: KeyEsc}
            }
        } else {
            out <- ev
        }
    }
}
```

---

# 四、CSI 解析（箭头键 / F键 / 鼠标）

```go
func ParseCSI(data []byte) RawEvent {
    s := string(data)

    switch s {
    case "A": return RawKey{Key: KeyUp}
    case "B": return RawKey{Key: KeyDown}
    case "C": return RawKey{Key: KeyRight}
    case "D": return RawKey{Key: KeyLeft}
    case "Z": return RawKey{Key: KeyShiftTab}
    }

    if strings.HasPrefix(s, "<") {
        return parseSGRMouse(s)
    }

    return RawUnknown{Data: data}
}
```

---

# 五、鼠标 SGR 协议解析

`ESC [ <b;x;yM`

```go
func parseSGRMouse(s string) RawEvent {
    var b, x, y int
    fmt.Sscanf(s, "<%d;%d;%dM", &b, &x, &y)

    return RawMouse{
        X: x-1,
        Y: y-1,
        Btn: decodeButton(b),
        Action: decodeAction(b),
    }
}
```

---

# 六、Normalization 层（统一语义）

这是最核心层。

```go
func Normalize(ev RawEvent) Event {
    switch e := ev.(type) {

    case RawCtrl:
        return normalizeCtrl(e.Code)

    case RawRune:
        return KeyEvent{Rune: e.Rune}

    case RawAltRune:
        return KeyEvent{Rune: e.Rune, Mod: ModAlt}

    case RawKey:
        return KeyEvent{Key: e.Key}

    case RawMouse:
        return MouseEvent{
            X: e.X, Y: e.Y,
            Button: e.Btn,
            Action: e.Action,
        }
    }
}
```

---

### Ctrl 统一规则

```go
func normalizeCtrl(c byte) Event {
    if c == 0x7f || c == 0x08 {
        return KeyEvent{Key: KeyBackspace}
    }
    if c >= 1 && c <= 26 {
        return KeyEvent{
            Rune: rune('a' + c - 1),
            Mod: ModCtrl,
        }
    }
    return KeyEvent{Key: KeyUnknown}
}
```

---

# 七、最终事件流

```
byte stream
   ↓
Reader
   ↓
ANSI Decoder
   ↓
ESC Resolver
   ↓
CSI Parser
   ↓
Normalizer
   ↓
App Event Loop
```

---

# 八、这套系统解决了什么

✔ Windows / Linux / macOS 统一
✔ Alt 组合不误判
✔ UTF-8 正确处理
✔ 鼠标拖拽 / 滚轮
✔ Ctrl 特殊键
✔ 可扩展（粘贴、IME、触摸）

---

如果你继续深入，下阶段就是：

> 🎯 **输入事件调度系统（焦点树、冒泡、捕获、快捷键系统）**

那是 TUI 框架比 GUI 还复杂的地方。
要的话我带你进入 **事件分发架构（Focus Tree + Capture/Bubble）**。


好，这里进入 **TUI 输入系统最底层的“协议工程”部分** —— 真正框架级内容。

你问的两块正好是：

> 🧠 **ANSI 控制序列状态机怎么设计**
> 🖱️ **SGR 鼠标编码每一位到底什么意思**

这俩搞透，你就具备写 tcell/bubbletea 级输入引擎的能力了。

---

# 一、ANSI 状态机实现思路（不是 if-else，而是自动机）

终端输入协议本质是：

> **有限状态机（FSM）解析字节流**

标准参考模型来自 **VT500 / ECMA-48**，我们做的是一个“简化版 CSI 解析器”。

---

## 1️⃣ 字节分类（先别写代码，先分类）

| 范围          | 含义               |
| ----------- | ---------------- |
| `0x00–0x1F` | C0 控制字符          |
| `0x20–0x2F` | 参数前缀             |
| `0x30–0x3F` | 参数数字区            |
| `0x40–0x7E` | Final Byte（决定指令） |
| `0x80+`     | UTF-8            |

---

## 2️⃣ 核心状态集合

```text
GROUND        默认状态（普通字符）
ESCAPE        收到 ESC
CSI_ENTRY     ESC [
CSI_PARAM     读参数 0-9 ; :
CSI_INTERM    中间字符
CSI_FINAL     结束态（生成事件）
UTF8          多字节字符
```

---

## 3️⃣ 状态转移图（核心逻辑）

```
GROUND
 ├─ ESC (0x1b) → ESCAPE
 ├─ C0 → CtrlEvent
 ├─ ASCII → RuneEvent
 └─ UTF8 start → UTF8

ESCAPE
 ├─ '[' → CSI_ENTRY
 ├─ 'O' → SS3（F键）
 └─ other → ESC Key

CSI_ENTRY
 ├─ '0-9;:' → CSI_PARAM
 ├─ '<' → CSI_PARAM (SGR 鼠标)
 └─ 'A-Z@~' → CSI_FINAL

CSI_PARAM
 ├─ '0-9;:' → stay
 ├─ ' '–'/' → CSI_INTERM
 └─ final byte → CSI_FINAL
```

---

## 4️⃣ 工程实现技巧

### ✔ 不用正则

必须逐字节流式处理，否则粘贴或鼠标拖动会炸。

---

### ✔ 参数缓存结构

```go
type CSI struct {
    Params []int
    Interm []byte
    Final  byte
}
```

解析 `ESC [ 1 ; 5 A`

```
Params = [1,5]
Final = 'A'
```

---

### ✔ 参数解析逻辑

```go
if isDigit(b) {
    cur = cur*10 + int(b-'0')
} else if b == ';' {
    params = append(params, cur)
    cur = 0
}
```

---

## 5️⃣ Final Byte 决定事件类型

| Final     | 含义       |
| --------- | -------- |
| A         | ↑        |
| B         | ↓        |
| C         | →        |
| D         | ←        |
| ~         | 特殊键      |
| M/m       | 鼠标 (X10) |
| M (带 `<`) | SGR 鼠标   |

---

# 二、SGR 鼠标协议解析表（重点来了）

SGR 格式：

```
ESC [ < b ; x ; y M   ← 按下
ESC [ < b ; x ; y m   ← 松开
```

---

## 1️⃣ 字段解释

| 字段  | 含义           |
| --- | ------------ |
| b   | 按键 + 修饰 + 类型 |
| x   | 列（从 1 开始）    |
| y   | 行（从 1 开始）    |
| M/m | 按下 / 释放      |

---

## 2️⃣ b 位结构（位运算）

```
bits: 76543210
      │││││└── Button (0–2)
      ││││└── Shift
      │││└── Alt
      ││└── Ctrl
      │└── Motion flag
      └── Wheel flag
```

---

## 3️⃣ 按钮值

| b&3 | 按钮      |
| --- | ------- |
| 0   | 左键      |
| 1   | 中键      |
| 2   | 右键      |
| 3   | 释放 / 移动 |

---

## 4️⃣ 滚轮

| b 值 | 含义  |
| --- | --- |
| 64  | 滚轮上 |
| 65  | 滚轮下 |

检测：

```go
if b & 64 != 0 { wheel }
```

---

## 5️⃣ 拖动判断

```go
if b & 32 != 0 {
    action = Drag
}
```

---

## 6️⃣ 修饰键

| 位     | 掩码 |
| ----- | -- |
| Shift | 4  |
| Alt   | 8  |
| Ctrl  | 16 |

```go
mods := 0
if b&4 != 0 { mods |= ModShift }
if b&8 != 0 { mods |= ModAlt }
if b&16 != 0 { mods |= ModCtrl }
```

---

## 7️⃣ 完整解析代码

```go
func parseSGR(b, x, y int, final byte) MouseEvent {
    ev := MouseEvent{X: x-1, Y: y-1}

    if b&64 != 0 {
        if b&1 == 0 {
            ev.Action = ScrollUp
        } else {
            ev.Action = ScrollDown
        }
        return ev
    }

    if final == 'm' {
        ev.Action = Release
        return ev
    }

    if b&32 != 0 {
        ev.Action = Drag
    } else {
        ev.Action = Press
    }

    switch b & 3 {
    case 0: ev.Button = Left
    case 1: ev.Button = Middle
    case 2: ev.Button = Right
    }

    return ev
}
```

---

# 三、为什么这套状态机是“专业级”的

| 能力        | 结果      |
| --------- | ------- |
| UTF-8 流解析 | 不丢中文    |
| CSI 参数解析  | 支持 F键变种 |
| SGR 鼠标支持  | 现代终端全兼容 |
| ESC 延迟    | Alt 不误判 |
| 可扩展       | 支持粘贴模式  |

---

下一层再往上就是：

> 🔥 **输入事件调度系统（Focus Tree + 冒泡 + 全局快捷键 + 捕获阶段）**

那部分复杂度已经超过 GUI 框架。
要继续我们可以直接设计 **TUI 版 DOM 事件系统**。
