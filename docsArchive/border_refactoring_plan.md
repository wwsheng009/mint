# Border 组件重构计划

## 问题分析

### 当前设计缺陷

`BorderedNode` 作为独立组件存在，存在以下问题：

1. **语义不一致**
   - Border 应该是容器的**属性**（类似 CSS），而不是一个组件
   - 与 Box Model 的统一处理逻辑不符合

2. **逻辑分散**
   - Border 渲染逻辑在 `BorderedNode.RenderBorder()`
   - 而其他盒模型属性（Padding, Margin）由 `BoxModelProvider` 处理
   - 导致边界处理逻辑维护困难

3. **使用混淆**
   - 需要区分"带边框的容器组件" 和 "带边框属性的容器"
   - 用户可能不知道应该使用哪种方式

4. **扩展性差**
   - 未来可能需要支持圆角、阴影、渐变等更多视觉属性
   - 如果每个都是独立组件，会造成组件爆炸

## 目标设计

### 统一的容器属性系统

```go
// 所有容器都支持的样式属性
type BoxModelStyle struct {
    Border   Border
    Padding  Padding
    Margin   Margin
}

type Border struct {
    Width  int
    Style  BorderStyle
    Color  string
    Radius int  // 圆角半径
    // 未来可以扩展：
    // Gradient BorderGradient  // 渐变边框
    // Shadow   ShadowStyle       // 阴影
    // ...
}

// 任何容器都可以通过 BoxModelProvider 提供
func (c *ContainerNode) GetBoxModel() BoxModel {
    return BoxModel{
        Border:  c.style.Border,
        Padding: c.style.Padding,
        Margin:  c.style.Margin,
    }
}
```

### 使用方式对比

**当前（不推荐）：**
```go
// 创建一个带边框的容器（需要嵌套在 Borted 组件中）
Bordered().Label("Title").Child(
    VStack().
        Children(
            // 内容
        ).Build()
).Build()
```

**目标设计（推荐）：**
```go
// 容器直接设置 border 属性
VStack().
    Border(Border{
        Width: 1,
        Style: BorderRounded,
        Color: "blue",
    }).
    Label("标题").
    Children(
        // 内容
    ).Build()

// 或者通过 builder 链式方法
VStack().
    Border(1, BorderRounded, "blue").
    Label("标题").
    Children(/*...*/).Build()
```

## 迁移计划

### 阶段 1：准备阶段

1. **记录当前使用场景**
   ```bash
   # 搜索所有 Bordered() 的使用
   grep -r "Bordered()" --include="*.go" runtime/
   ```

2. **添加废弃警告**
   ```go
   // runtime/ui/layout.go
   // Deprecated: Use ContainerStyle.Border instead. This will be removed in v2.0.
   func Bordered() *BorderedBuilder {
       log.Println("WARNING: Bordered() is deprecated, use ContainerStyle.Border instead")
       // ...
   }
   ```

3. **添加迁移工具函数**
   ```go
   // 提供 BorderedNode 到带 border 属性的便捷转换
   func ConvertBorderedToStyledBorder(bn *BorderedNode) Border {
       return Border{
           Width:  1,
           Style:  bn.GetBorderStyle(),
           Color:  bn.GetBorderColor(),
           Label:  bn.GetBorderLabel(),
       }
   }
   ```

### 阶段 2：逐步迁移

1. **核心容器添加 Border 支持**
   - `VStack`, `HStack`, `Grid`
   - `Box`, `Modal`, `Panel`

   ```go
   type VStackBuilder struct {
       // 现有字段...
       border Border  // 新增
   }

   func (b *VStackBuilder) Border(border Border) *VStackBuilder {
       b.border = border
       return b
   }

   // 便捷方法
   func (b *VStackBuilder) Border(width int, style BorderStyle, color string) *VStackBuilder {
       b.border = Border{Width: width, Style: style, Color: color}
       return b
   }
   ```

2. **迁移示例代码**

   **迁移前：**
   ```go
   // 当前使用 Bordered 组件
   Bordered().Label("User Info").Child(
       VStack().
           Width(40).
           Children(
               Text("Name: Alice"),
               Text("Email: alice@example.com"),
           ).Build()
   ).Build()
   ```

   **迁移后：**
   ```go
   // 将 Border 作为属性
   VStack().
       Border(1, BorderRounded, "blue").
       Label("User Info").
       Width(40).  // 注意：Width 现在指内容宽度（不包含 border）
       Children(
           Text("Name: Alice"),
           Text("Email: alice@example.com"),
       ).Build()
   ```

3. **更新文档和示例**

### 阶段 3：废弃 BorderedNode

1. **v1.5 - 添加编译警告**
   ```go
   //go:build !ignore_bordered_deprecation

   // Deprecated: Use ContainerStyle.Border instead
   func Bordered() *BorderedBuilder { ... }
   ```

2. **v2.0 - 完全移除**
   - 删除 `BorderedNode` 和 `BorderedBuilder`
   - 删除相关测试代码
   - 清理文档和示例

## 具体实现步骤

### Step 1: 扩展 Border 结构体

```go
// runtime/layout/border.go
type Border struct {
    Width  int
    Style  BorderStyle
    Color  string

    // 扩展属性
    Label  string  // 顶部标签（兼容当前 BodedNode 的功能）
    Radius int     // 圆角半径
}
```

### Step 2: 容器实现 BoxModelProvider

确保所有容器都正确实现 `BoxModelProvider`：

```go
// runtime/ui/layout.go
func (v *VStackNode) GetBoxModel() BoxModel {
    return BoxModel{
        Border:  v.style.Border,
        Padding: v.style.Padding,
        Margin:  v.style.Margin,
    }
}
```

### Step 3: 渲染层支持 Border

渲染引擎需要从 `BoxModelProvider` 读取 border 信息：

```go
// runtime/render/renderer.go
func (r *Renderer) renderNode(node Node, layout *LayoutBox) {
    // 从 BoxModelProvider 获取 border
    if provider, ok := node.(BoxModelProvider); ok {
        boxModel := provider.GetBoxModel()
        if boxModel.Border.HasBorder() {
            r.renderBorder(boxModel.Border, layout)
        }
    }
    // ... 渲染内容
}
```

### Step 4: 更新 BorderedNode (过渡期)

过渡期让 `BorderedNode` 实现兼容转换：

```go
func (bn *BorderedNode) GetBoxModel() BoxModel {
    return BoxModel{
        Border: Border{
            Width: 1,
            Style: bn.GetBorderStyle(),
            Color: bn.GetBorderColor(),
            Label: bn.GetBorderLabel(),
        },
    }
}
```

## 兼容性考虑

### 1. 尺寸语义差异

**问题：**
- `BorderedNode` 输出的尺寸**包含** border
- 新设计下，容器尺寸**包含** padding + border（通过 BoxModel）
- 需要确保 `Width/Height` 的语义一致性

**解决方案：**
- 记录 `Width/Height` 指的是**内容尺寸**（不包含 padding 和 border）
- `Measure()` 返回的总尺寸 = 内容 + padding + border
- 保持与 Box Model 规范一致

### 2. 嵌套场景

**当前：**
```go
Bordered().Child(
    VStack().Children(/*...*/)
)
```

**迁移后：**
```go
VStack().
    Border(...).
    Children(/*...*/)
```

两者在布局结果上应该等效。

## 测试策略

### 1. 单元测试

```go
func TestContainerBorder(t *testing.T) {
    // 测试容器正确应用 border 属性
    container := VStack().
        Border(1, BorderRounded, "blue").
        Width(40).
        Build()

    boxModel := container.GetBoxModel()
    assert.True(t, boxModel.Border.HasBorder())
    assert.Equal(t, 1, boxModel.Border.Width)
    assert.Equal(t, BorderRounded, boxModel.Border.Style)
}
```

### 2. 集成测试

```go
func TestMigrationBorderedToBorder(t *testing.T) {
    // 验证迁移前后的布局结果一致
    oldStyle := Bordered().Label("Test").Child(
        VStack().Width(40).Children(Text("Content")).Build()
    ).Build()

    newStyle := VStack().
        Border(1, BorderSingle, "blue").
        Label("Test").
        Width(40).
        Children(Text("Content")).
        Build()

    // 两者应该产生相同的布局结果
    assert.EqualLayouts(t, oldStyle, newStyle)
}
```

## 时间表

| 阶段 | 版本 | 目标 |
|------|------|------|
| 准备 | v1.4 | 添加废弃警告，记录使用场景 |
| 迁移 | v1.5 | 核心容器支持 Border 属性 |
| 废弃 | v2.0 | 完全移除 BorderedNode |

## 相关文档

- [Box Model 设计](./box_model_design.md)
- [布局引擎架构](./layout_architecture.md)
- [CSS Box Model 参考](https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_Box_Model)

---

**最后更新**: 2026-03-02
**状态**: 规划中
**负责人**: @Qwen Code
