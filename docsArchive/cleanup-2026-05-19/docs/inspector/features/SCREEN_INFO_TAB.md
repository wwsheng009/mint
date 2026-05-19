# Inspector Screen Info Tab

## 概述

Inspector 新增了 "Screen" (屏幕信息) 标签页，用于实时显示屏幕大小、鼠标位置、按钮检测等信息。

## 功能特性

### 1. 屏幕大小显示
- 显示当前终端的列数和行数
- 实时更新，响应终端 resize 事件

### 2. Overlay 位置信息
- 显示 Inspector overlay 的当前位置 (x, y)
- 显示 overlay 的大小
- 显示 overlay 的边界范围

### 3. 鼠标信息
- 实时显示鼠标位置坐标
- 显示鼠标事件类型（Press/Release/Move/Wheel）
- 显示鼠标按钮状态
- 指示鼠标是否在 Inspector overlay 内

### 4. 元素检测
- 显示当前鼠标悬停的元素
- 显示当前选中的元素（通过 Enter 键选择）
- 显示元素的路径信息

### 5. 按钮检测
- 自动检测鼠标下的元素是否为按钮
- 检测元素是否有 onClick 事件处理器
- 显示元素类型信息

### 6. Inspector 状态
- 显示 Inspector 的启用/禁用状态
- 显示 Inspector 的可见/隐藏状态
- 显示当前活动的标签页

## 使用方法

### 切换到 Screen 标签页

1. 按 `F12` 或 `Ctrl+D` 打开 Inspector
2. 按数字键 `7` 切换到 Screen 标签页
3. 或使用 `Tab` 键循环切换标签页

### 移动 Inspector 面板

- `Alt+H` 或 `Alt+←`: 向左移动
- `Alt+L` 或 `Alt+→`: 向右移动
- `Alt+K` 或 `Alt+↑`: 向上移动
- `Alt+J` 或 `Alt+↓`: 向下移动

### 屏幕大小自动更新

当终端窗口大小改变时，Inspector 会自动捕获 resize 事件并更新屏幕大小信息。

## API

### 设置屏幕大小

```go
inspector.SetScreenSize(width, height int)
```

### 获取屏幕大小

```go
width, height := inspector.GetScreenSize()
```

## 实现细节

### 数据流

```
Terminal Resize Event
    ↓
Engine.resizeEvent()
    ↓
App.Resize(width, height)
    ↓
Inspector.SetScreenSize(width, height)
    ↓
Screen Info Tab 显示更新
```

### 相关文件

- `internal/inspector/standalone_inspector.go`: Inspector 核心实现
- `framework/app.go`: App Resize 事件处理，更新 Inspector 屏幕大小

## 示例

运行示例程序：

```bash
cd examples/ui_demos/demo1_full_featured
go run main.go
```

1. 启动后按 `F12` 打开 Inspector
2. 按 `7` 切换到 Screen 标签页
3. 移动鼠标查看实时位置更新
4. 调整终端窗口大小，观察屏幕大小自动更新
5. 使用 `Alt+H/J/K/L` 移动 Inspector 面板
6. 观察鼠标位置和元素检测信息

## 输出示例

```
📺 Screen Information

─ Screen Size ─
Screen Size: 80 cols × 24 rows

─ Overlay Position ─
Overlay Position: (0, 0)
Overlay Size: 80 cols × 25 rows
Overlay Bounds: (0,0) to (80,25)

─ Mouse Information ─
Mouse Position: (42, 12)
Mouse Event: Move (None)
Mouse in Overlay: Yes

─ Element Detection ─
Hovered: button (0/1/2)
Selected: None

─ Button Detection ─
✓ Button detected: button

─ Inspector State ─
Inspector: Enabled | Visible | Tab: Screen

─ Instructions ─
Alt+J/K/L/H: Move overlay
Move mouse to see real-time updates
```
