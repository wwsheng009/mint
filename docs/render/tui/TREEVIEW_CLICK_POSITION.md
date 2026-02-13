# TreeView点击位置调试说明

## 问题描述

如果在Inspector中点击TreeView时，选中的条目位置不正确（例如点击上面的条目，下面的被选中），可以通过环境变量调整偏移量。

## 解决方法

### 方法1：使用环境变量调整偏移量（推荐）

运行demo时设置`TUI_TREEVIEW_OFFSET`环境变量：

```bash
# Bash/Linux/Mac
TUI_DEBUG_INSPECTOR=true TUI_TREEVIEW_OFFSET=0 go run ./examples/ui_demos/demo2_runtime_internals/inspector_overlay/main.go

# Windows PowerShell
$env:TUI_DEBUG_INSPECTOR="true"
$env:TUI_TREEVIEW_OFFSET="0"
go run ./examples/ui_demos/demo2_runtime_internals/inspector_overlay/main.go
```

### 方法2：查找正确的偏移量值

1. 启用verbose输出来查看当前的点击位置：
   ```bash
   TUI_DEBUG_INSPECTOR=true go run ./examples/ui_demos/demo2_runtime_internals/inspector_overlay/main.go
   ```

2. 按F12打开Inspector，点击Elements标签

3. 点击TreeView的不同条目，观察控制台输出：
   ```
   [Inspector] Set TreeView bounds: x=0, y=7, w=80, h=33
   [Inspector] TreeView click: localY=0, target=0
   ```

4. 根据实际情况调整`TUI_TREEVIEW_OFFSET`：
   - 如果点击line 0但选中了line 4：尝试`TUI_TREEVIEW_OFFSET=0`
   - 如果点击line 0但选中了line 2：尝试`TUI_TREEVIEW_OFFSET=2`
   - 如果点击line 4才选中line 0：尝试`TUI_TREEVIEW_OFFSET=6`

### 偏移量说明

偏移量表示TreeView在tab content内部的实际起始行数：

- `offset=0`: TreeView从tab content的第一行开始
- `offset=4`: TreeView从tab content的第5行开始（默认值）
- `offset=6`: TreeView从tab content的第7行开始

## 调试技巧

### 查看TreeView布局信息

运行时查看控制台输出，找到：
```
[Inspector] TreeView layout: contentY=3, offset=4, actualY=7, height=33
```

- `contentY`: tab content的起始Y坐标
- `offset`: 当前使用的偏移量
- `actualY`: TreeView的实际起始Y坐标
- `height`: TreeView的高度

### 测试点击位置

1. 记录你想点击的条目（比如第一行）
2. 在屏幕上点击该位置
3. 查看控制台输出的`focusIndex`值
4. 如果`focusIndex`与你点击的条目不符，调整偏移量

## 常见情况

### 情况1：点击上面的选中下面的
**症状**：点击tree line 0，实际选中tree line 4或更下面
**原因**：偏移量太大
**解决**：减小`TUI_TREEVIEW_OFFSET`值，尝试0或1

### 情况2：点击下面的选中上面的
**症状**：点击tree line 4，实际选中tree line 0或更上面
**原因**：偏移量太小
**解决**：增大`TUI_TREEVIEW_OFFSET`值，尝试6或8

### 情况3：点击总是在条目之间
**症状**：点击位置总是选中两个条目之间的空白位置
**原因**：偏移量是小数（布局对齐问题）
**解决**：微调`TUI_TREEVIEW_OFFSET`值，+1或-1

## 永久修复

找到正确的偏移量后，可以修改代码中的默认值：

编辑 `internal/inspector/standalone_inspector.go`，找到：
```go
offset := 4 // Default heuristic value
```

改为你的测试值：
```go
offset := 0 // 修正后的值
```

## 相关文件

- `internal/inspector/standalone_inspector.go`: handleOverlayClick函数
- `components/display/treeview.go`: TreeView.HandleEvent函数
