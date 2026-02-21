# Mint 布局系统测试方案

## 目录
- [1. 测试目标](#1-测试目标)
- [2. 测试策略](#2-测试策略)
- [3. 单元测试](#3-单元测试)
- [4. 集成测试](#4-集成测试)
- [5. 边界测试](#5-边界测试)
- [6. 性能测试](#6-性能测试)
- [7. 测试工具](#7-测试工具)
- [8. 测试执行计划](#8-测试执行计划)

---

## 1. 测试目标

### 1.1 核心目标
1. **验证约束传播正确性**：确保约束在组件间传递时保持一致性
2. **验证维度计算正确性**：确保外部/内部维度转换准确
3. **验证边界行为**：测试极端情况和边界条件
4. **验证性能指标**：确保优化后性能达标
5. **保证向后兼容**：确保 API 改进不破坏现有代码

### 1.2 测试覆盖率目标

| 测试类型 | 当前覆盖率 | 目标覆盖率 | 提升 |
|---------|-----------|-----------|------|
| 单元测试 | 65% | 85% | +20% |
| 集成测试 | 40% | 70% | +30% |
| 边界测试 | 30% | 80% | +50% |

### 1.3 质量门禁

- 所有单元测试通过：✅
- 所有集成测试通过：✅
- 代码覆盖率 >= 80%：✅
- 性能回归不超过 5%：✅
- 没有 P0/P1 bug：✅

---

## 2. 测试策略

### 2.1 测试金字塔

```
        /\
       /  \        E2E Tests (5%)
      /----\       - 用户场景测试
     /      \      - 跨组件集成
    /--------\     Integration Tests (20%)
   /          \    - 组件间交互
  /            \   - 布局场景
 /--------------\ Unit Tests (75%)
/                \ - 单个组件测试
                 - 约束传播测试
                 - 维度计算测试
```

### 2.2 测试分层

#### 2.2.1 单元测试（75%）
- 测试单个组件的 Measure/Paint 行为
- 测试约束传播算法
- 测试维度计算函数

#### 2.2.2 集成测试（20%）
- 测试组件组合（Panel = Border + VStack）
- 测试真实布局场景（Panel in HStack）
- 测试约束传播链（Root → HStack → Panel → Text）

#### 2.2.3 E2E 测试（5%）
- 测试用户场景（创建复杂 UI）
- 测试跨平台行为（不同终端尺寸）

### 2.3 测试类型分类

| 测试类型 | 优先级 | 测试内容 | 工具 |
|---------|--------|---------|------|
| 功能测试 | P0 | 约束传播、维度计算 | Go testing |
| 回归测试 | P0 | 已知 bug 不复现 | Go testing |
| 边界测试 | P1 | 极端情况、错误输入 | Go testing |
| 性能测试 | P1 | Measure 性能、缓存命中率 | Benchmark |
| 兼容性测试 | P2 | API 向后兼容 | Go testing |

---

## 3. 单元测试

### 3.1 约束传播测试

#### 测试套件：Border 约束传播

```go
// 文件: ui/components/border/constraints_test.go

package border_test

import (
    "testing"

    "github.com/wwsheng009/mint/runtime/layout"
    "github.com/wwsheng009/mint/ui/components/border"
    newtext "github.com/wwsheng009/mint/ui/components/text"
)

func TestBorder_ConstraintPropagation(t *testing.T) {
    tests := []struct {
        name                    string
        borderConfig            borderConfig
        parentConstraints       layout.Constraints
        expectedChildConstraints layout.Constraints
        expectedSize            layout.Size
    }{
        {
            name: "Explicit width 20, auto height",
            borderConfig: borderConfig{
                width:  20,
                height: 0,  // auto
                style:  layout.BorderRounded,
            },
            parentConstraints: layout.Constraints{
                MinWidth:  0,
                MaxWidth:  50,
                MinHeight: 0,
                MaxHeight: 100,
            },
            // 内部宽度 = 20 - 2 = 18
            expectedChildConstraints: layout.Constraints{
                MinWidth:  18,
                MaxWidth:  18,
                MinHeight: 0,
                MaxHeight: 100,
            },
            // 外部尺寸 = 内部 + 边框
            // 假设子元素返回 18x4
            expectedSize: layout.Size{Width: 20, Height: 6},  // 18+2, 4+2
        },
        {
            name: "Auto width, explicit height 6",
            borderConfig: borderConfig{
                width:  0,  // auto
                height: 6,
                style:  layout.BorderRounded,
            },
            parentConstraints: layout.Constraints{
                MinWidth:  0,
                MaxWidth:  50,
                MinHeight: 0,
                MaxHeight: 100,
            },
            // 内部高度 = 6 - 2 = 4
            expectedChildConstraints: layout.Constraints{
                MinWidth:  0,
                MaxWidth:  48,  // 50 - 2
                MinHeight: 4,
                MaxHeight: 4,
            },
            // 外部尺寸
            expectedSize: layout.Size{Width: 50, Height: 6},  // 使用父宽度
        },
        {
            name: "Completely auto",
            borderConfig: borderConfig{
                width:  0,  // auto
                height: 0,  // auto
                style:  layout.BorderRounded,
            },
            parentConstraints: layout.Constraints{
                MinWidth:  0,
                MaxWidth:  50,
                MinHeight: 0,
                MaxHeight: 100,
            },
            // 全部使用父约束（减去边框）
            expectedChildConstraints: layout.Constraints{
                MinWidth:  0,
                MaxWidth:  48,  // 50 - 2
                MinHeight: 0,
                MaxHeight: 98,  // 100 - 2
            },
            expectedSize: layout.Size{Width: 20, Height: 4},  // 子元素决定的尺寸
        },
        {
            name: "Fixed width and height",
            borderConfig: borderConfig{
                width:  20,
                height: 6,
                style:  layout.BorderRounded,
            },
            parentConstraints: layout.Constraints{
                MinWidth:  0,
                MaxWidth:  50,
                MinHeight: 0,
                MaxHeight: 100,
            },
            // 内部尺寸
            expectedChildConstraints: layout.Constraints{
                MinWidth:  18,
                MaxWidth:  18,
                MinHeight: 4,
                MaxHeight: 4,
            },
            expectedSize: layout.Size{Width: 20, Height: 6},
        },
        {
            name: "Double border style",
            borderConfig: borderConfig{
                width:  20,
                height: 0,
                style:  layout.BorderDouble,
            },
            parentConstraints: layout.Constraints{
                MinWidth:  0,
                MaxWidth:  50,
                MinHeight: 0,
                MaxHeight: 100,
            },
            // Double border 仍然是 1 字符宽
            expectedChildConstraints: layout.Constraints{
                MinWidth:  18,  // 20 - 2
                MaxWidth:  18,
                MinHeight: 0,
                MaxHeight: 100,
            },
            expectedSize: layout.Size{Width: 20, Height: 6},
        },
        {
            name: "No border style",
            borderConfig: borderConfig{
                width:  20,
                height: 0,
                style:  layout.BorderNone,
            },
            parentConstraints: layout.Constraints{
                MinWidth:  0,
                MaxWidth:  50,
                MinHeight: 0,
                MaxHeight: 100,
            },
            // 无边框，直接传递
            expectedChildConstraints: layout.Constraints{
                MinWidth:  20,
                MaxWidth:  20,
                MinHeight: 0,
                MaxHeight: 100,
            },
            expectedSize: layout.Size{Width: 20, Height: 6},
        },
        {
            name: "Explicit dimension larger than parent constraint",
            borderConfig: borderConfig{
                width:  60,  // > parent MaxWidth (50)
                height: 0,
                style:  layout.BorderRounded,
            },
            parentConstraints: layout.Constraints{
                MinWidth:  0,
                MaxWidth:  50,
                MinHeight: 0,
                MaxHeight: 100,
            },
            // 父约束应该被尊重
            expectedChildConstraints: layout.Constraints{
                MinWidth:  18,  // 20 - 2 (使用父约束的 MaxWidth)
                MaxWidth:  18,
                MinHeight: 0,
                MaxHeight: 100,
            },
            expectedSize: layout.Size{Width: 20, Height: 6},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Arrange
            inst := createBorderInstance(tt.borderConfig)
            textNode := newtext.New("Test content that is long enough to wrap")
            inst.SetChild(textNode.CreateInstance())

            // Act
            size := inst.Measure(tt.parentConstraints)

            // Assert
            // 验证尺寸
            if size != tt.expectedSize {
                t.Errorf("Size mismatch:\n  Expected: %+v\n  Got:      %+v",
                    tt.expectedSize, size)
            }

            // 验证传递给子元素的约束
            //（需要 mock 或追踪子元素的 ReceiveConstraints）
            childConstraints := inst.getChildConstraints()  // 需要实现此方法
            if childConstraints != tt.expectedChildConstraints {
                t.Errorf("Child constraints mismatch:\n  Expected: %+v\n  Got:      %+v",
                    tt.expectedChildConstraints, childConstraints)
            }
        })
    }
}

func TestBorder_CalculateInnerDimensions(t *testing.T) {
    tests := []struct {
        name           string
        outerDimensions Dimensions
        borderStyle    layout.BorderStyle
        expectedInner  Dimensions
    }{
        {
            name: "Rounded border 20x10",
            outerDimensions: Dimensions{Width: 20, Height: 10},
            borderStyle:    layout.BorderRounded,
            expectedInner:  Dimensions{Width: 18, Height: 8},
        },
        {
            name: "Double border 30x15",
            outerDimensions: Dimensions{Width: 30, Height: 15},
            borderStyle:    layout.BorderDouble,
            expectedInner:  Dimensions{Width: 28, Height: 13},
        },
        {
            name: "No border 25x8",
            outerDimensions: Dimensions{Width: 25, Height: 8},
            borderStyle:    layout.BorderNone,
            expectedInner:  Dimensions{Width: 25, Height: 8},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            padding := 2 * border.GetBorderWidth(tt.borderStyle)
            innerWidth := tt.outerDimensions.Width - padding
            innerHeight := tt.outerDimensions.Height - padding

            if innerWidth != tt.expectedInner.Width {
                t.Errorf("Inner width mismatch: expected %d, got %d",
                    tt.expectedInner.Width, innerWidth)
            }
            if innerHeight != tt.expectedInner.Height {
                t.Errorf("Inner height mismatch: expected %d, got %d",
                    tt.expectedInner.Height, innerHeight)
            }
        })
    }
}

type borderConfig struct {
    width  int
    height int
    style  layout.BorderStyle
}

type Dimensions struct {
    Width  int
    Height int
}

func createBorderInstance(config borderConfig) *border.Instance {
    inst := border.NewInstance()
    inst.SetWidth(config.width)
    inst.SetHeight(config.height)
    inst.SetBorderStyle(config.style)
    return inst
}
```

### 3.2 维度计算测试

```go
// 文件: ui/components/border/dimensions_test.go

package border_test

import (
    "testing"

    "github.com/wwsheng009/mint/runtime/layout"
    "github.com/wwsheng009/mint/ui/components/border"
)

func TestBorder_DimensionTransformation(t *testing.T) {
    tests := []struct {
        name              string
        outerWidth        int
        outerHeight       int
        borderStyle       layout.BorderStyle
        childSize         layout.Size
        expectedOuterSize layout.Size
    }{
        {
            name:        "Rounded border, child 18x4",
            outerWidth:  20,
            outerHeight: 0,  // auto
            borderStyle: layout.BorderRounded,
            childSize:   layout.Size{Width: 18, Height: 4},
            expectedOuterSize: layout.Size{Width: 20, Height: 6},  // 18+2, 4+2
        },
        {
            name:        "Double border, child 28x6",
            outerWidth:  30,
            outerHeight: 0,
            borderStyle: layout.BorderDouble,
            childSize:   layout.Size{Width: 28, Height: 6},
            expectedOuterSize: layout.Size{Width: 30, Height: 8},
        },
        {
            name:        "No border, child 25x5",
            outerWidth:  25,
            outerHeight: 0,
            borderStyle: layout.BorderNone,
            childSize:   layout.Size{Width: 25, Height: 5},
            expectedOuterSize: layout.Size{Width: 25, Height: 5},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            inst := border.NewInstance()
            inst.SetWidth(tt.outerWidth)
            inst.SetBorderStyle(tt.borderStyle)

            // 模拟子元素返回的尺寸
            outerSize := inst.computeOuterSize(tt.childSize)

            if outerSize != tt.expectedOuterSize {
                t.Errorf("Outer size mismatch:\n  Expected: %+v\n  Got:      %+v",
                    tt.expectedOuterSize, outerSize)
            }
        })
    }
}

func TestBorder_GetBorderWidth(t *testing.T) {
    tests := []struct {
        name     string
        style    layout.BorderStyle
        expected int
    }{
        {"BorderNone", layout.BorderNone, 0},
        {"BorderSingle", layout.BorderSingle, 1},
        {"BorderDouble", layout.BorderDouble, 1},
        {"BorderRounded", layout.BorderRounded, 1},
        {"BorderDashed", layout.BorderDashed, 1},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            width := border.GetBorderWidth(tt.style)
            if width != tt.expected {
                t.Errorf("GetBorderWidth(%v) = %d, expected %d",
                    tt.style, width, tt.expected)
            }
        })
    }
}
```

### 3.3 Text.Wrap 测试

```go
// 文件: ui/components/text/wrap_test.go

package text_test

import (
    "testing"

    "github.com/wwsheng009/mint/runtime/layout"
    "github.com/wwsheng009/mint/runtime/paint"
    "github.com/wwsheng009/mint/ui/components/text"
)

func TestText_WrapMeasure(t *testing.T) {
    tests := []struct {
        name              string
        content           string
        constraints       layout.Constraints
        expectedSize      layout.Size
        expectedLineCount int
    }{
        {
            name:        "Short text, no wrap needed",
            content:     "Hello",
            constraints: layout.Constraints{MaxWidth: 10, MaxHeight: 100},
            expectedSize: layout.Size{Width: 5, Height: 1},
            expectedLineCount: 1,
        },
        {
            name:        "Long text, wraps to 3 lines",
            content:     "This is very long text",
            constraints: layout.Constraints{MaxWidth: 10, MaxHeight: 100},
            expectedSize: layout.Size{Width: 10, Height: 3},
            expectedLineCount: 3,
        },
        {
            name:        "MaxWidth 18, wraps to 4 lines",
            content:     "This is a very long content that should wrap",
            constraints: layout.Constraints{MaxWidth: 18, MaxHeight: 100},
            expectedSize: layout.Size{Width: 18, Height: 4},
            expectedLineCount: 4,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            inst := text.NewInstance()
            inst.SetContent(tt.content)
            inst.SetWrap(true)

            size := inst.Measure(tt.constraints)

            if size.Width != tt.expectedSize.Width {
                t.Errorf("Width mismatch: expected %d, got %d",
                    tt.expectedSize.Width, size.Width)
            }
            if size.Height != tt.expectedSize.Height {
                t.Errorf("Height mismatch: expected %d, got %d",
                    tt.expectedSize.Height, size.Height)
            }

            // 验证行数
            lineCount := inst.getLineCount()  // 需要实现此方法
            if lineCount != tt.expectedLineCount {
                t.Errorf("Line count mismatch: expected %d, got %d",
                    tt.expectedLineCount, lineCount)
            }
        })
    }
}

func TestText_WrapPaintWithHeightConstraint(t *testing.T) {
    tests := []struct {
        name         string
        content      string
        wrap         bool
        bounds       paint.Rect  // [x, y, width, height]
        expectCrop   bool
        expectLines  int
    }{
        {
            name:        "4 lines vs 2 rows bounds, content cropped",
            content:     "This is a very long content that should wrap",
            wrap:        true,
            bounds:      paint.Rect{0, 0, 18, 2},  // 只有 2 行高度
            expectCrop:  true,
            expectLines: 2,  // 只渲染 2 行
        },
        {
            name:        "4 lines vs 4 rows bounds, all rendered",
            content:     "This is a very long content that should wrap",
            wrap:        true,
            bounds:      paint.Rect{0, 0, 18, 4},  // 恰好 4 行高度
            expectCrop:  false,
            expectLines: 4,  // 渲染全部 4 行
        },
        {
            name:        "No wrap, short content",
            content:     "Hello",
            wrap:        false,
            bounds:      paint.Rect{0, 0, 10, 5},
            expectCrop:  false,
            expectLines: 1,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            inst := text.NewInstance()
            inst.SetContent(tt.content)
            inst.SetWrap(tt.wrap)

            buf := paint.NewBuffer(20, 10)
            ctx := text.PaintContext{
                Bounds: tt.bounds,
            }

            // Track painted lines
            var paintedLines int
            inst.Paint(ctx, buf, func(line string) {  // 需要 mock Paint
                paintedLines++
            })

            if paintedLines != tt.expectLines {
                t.Errorf("Painted lines mismatch: expected %d, got %d",
                    tt.expectLines, paintedLines)
            }
        })
    }
}

func TestText_ValidatePaintSize(t *testing.T) {
    tests := []struct {
        name           string
        measureSize    layout.Size
        paintBounds    paint.Rect
        allowCrop      bool
        expectError    bool
        errorMessage   string
    }{
        {
            name:        "Measure height equals paint height, OK",
            measureSize: layout.Size{Width: 18, Height: 4},
            paintBounds: paint.Rect{0, 0, 18, 4},
            allowCrop:   false,
            expectError: false,
        },
        {
            name:        "Measure height exceeds paint height allowCrop=true, OK",
            measureSize: layout.Size{Width: 18, Height: 4},
            paintBounds: paint.Rect{0, 0, 18, 2},
            allowCrop:   true,
            expectError: false,
        },
        {
            name:        "Measure height exceeds paint height allowCrop=false, Error",
            measureSize: layout.Size{Width: 18, Height: 4},
            paintBounds: paint.Rect{0, 0, 18, 2},
            allowCrop:   false,
            expectError: true,
            errorMessage: "Text height (4) exceeds paint bounds (2)",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            inst := text.NewInstance()
            inst.SetAllowCrop(tt.allowCrop)

            err := inst.ValidatePaintSize(tt.measureSize, tt.paintBounds)

            if tt.expectError {
                if err == nil {
                    t.Error("Expected error but got nil")
                }
                if !contains(err.Error(), tt.errorMessage) {
                    t.Errorf("Error message mismatch: expected to contain %q, got %q",
                        tt.errorMessage, err.Error())
                }
            } else {
                if err != nil {
                    t.Errorf("Unexpected error: %v", err)
                }
            }
        })
    }
}

func contains(s, substr string) bool {
    return len(s) >= len(substr) && s[:len(substr)] == substr ||
        len(s) > len(substr) && contains(s[1:], substr)
}
```

---

## 4. 集成测试

### 4.1 Panel 组合测试

```go
// 文件: ui/components/panel/integration_test.go

package panel_test

import (
    "testing"

    "github.com/wwsheng009/mint/runtime/layout"
    "github.com/wwsheng009/mint/runtime/paint"
    "github.com/wwsheng009/mint/ui/components/panel"
    newstack "github.com/wwsheng009/mint/ui/components/stack"
    newtext "github.com/wwsheng009/mint/ui/components/text"
)

func TestPanel_DimensionTransformation(t *testing.T) {
    tests := []struct {
        name             string
        panelConfig      panelConfig
        expectedOuter    layout.Size
        expectedInner    layout.Size
    }{
        {
            name: "Panel 20x6 with Rounded border",
            panelConfig: panelConfig{
                outerWidth:  20,
                outerHeight: 6,
                borderStyle: layout.BorderRounded,
                content:     newtext.New("Test"),
            },
            expectedOuter: layout.Size{Width: 20, Height: 6},
            expectedInner: layout.Size{Width: 18, Height: 4},
        },
        {
            name: "Panel with auto height",
            panelConfig: panelConfig{
                outerWidth:  20,
                outerHeight: 0,  // auto
                borderStyle: layout.BorderRounded,
                content:     newtext.New("Test").SetWrap(true),
            },
            expectedOuter: layout.Size{Width: 20, Height: 2},  // 最小高度
            expectedInner: layout.Size{Width: 18, Height: 0},  // 由内容决定
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            vnode := createPanelVNode(tt.panelConfig)
            inst := vnode.CreateInstance()

            size := inst.Measure(layout.Constraints{
                MaxWidth:  50,
                MaxHeight: 100,
            })

            if size != tt.expectedOuter {
                t.Errorf("Outer size mismatch:\n  Expected: %+v\n  Got:      %+v",
                    tt.expectedOuter, size)
            }

            // 获取内部维度
            innerWidth, innerHeight := vnode.GetInnerDimensions()
            if innerWidth != tt.expectedInner.Width {
                t.Errorf("Inner width mismatch: expected %d, got %d",
                    tt.expectedInner.Width, innerWidth)
            }
        })
    }
}

func TestPanel_InHStack(t *testing.T) {
    tests := []struct {
        name         string
        panels       []panelConfig
        hStackWidth  int
        expectedSize layout.Size
    }{
        {
            name: "Two panels: 20x3 and auto",
            panels: []panelConfig{
                {outerWidth: 20, outerHeight: 3, content: newtext.New("Fixed")},
                {outerWidth: 20, outerHeight: 0, content: newtext.New("Auto")},
            },
            hStackWidth:  50,
            expectedSize: layout.Size{Width: 50, Height: 3},  // max(3, 2)
        },
        {
            name: "Three panels: 20, auto, 15",
            panels: []panelConfig{
                {outerWidth: 20, outerHeight: 3, content: newtext.New("Fixed1")},
                {outerWidth: 0, outerHeight: 0, content: newtext.New("Auto")},
                {outerWidth: 15, outerHeight: 4, content: newtext.New("Fixed2")},
            },
            hStackWidth:  50,
            expectedSize: layout.Size{Width: 50, Height: 4},  // max(3, ?, 4)
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // 构建 HStack
            var children []rtui.VNode
            for _, pc := range tt.panels {
                children = append(children, createPanelVNode(pc))
            }

            hstack := newstack.New(newstack.Row).
                SetWidth(tt.hStackWidth).
                SetGap(0).
                SetChildrenList(children)

            inst := hstack.CreateInstance()
            size := inst.Measure(layout.Constraints{
                MaxWidth:  100,
                MaxHeight: 100,
            })

            if size != tt.expectedSize {
                t.Errorf("HStack size mismatch:\n  Expected: %+v\n  Got:      %+v",
                    tt.expectedSize, size)
            }
        })
    }
}

func TestPanel_WithWrappedText(t *testing.T) {
    longText := "This is a very long text that should wrap to multiple lines"
    vnode := panel.NewBuilder().
        SetContentWidth(18).
        WithAutoHeight().
        Content(newtext.New(longText).SetWrap(true)).
        Build()

    inst := vnode.CreateInstance()
    size := inst.Measure(layout.Constraints{
        MaxWidth:  50,
        MaxHeight: 100,
    })

    // 长文本在 18 宽度下应该换行成多行
    // 假设换成 4 行，则高度 = 4 + 2 = 6
    if size.Height < 2 {
        t.Errorf("Expected height >= 2, got %d", size.Height)
    }

    // 验证约束传播
    buf := paint.NewBuffer(20, 10)
    ctx := panel.PaintContext{
        Bounds: paint.Rect{0, 0, 20, size.Height},
    }

    inst.Paint(ctx, buf)

    // 验证内容没有溢出
    for y := 0; y < size.Height; y++ {
        for x := 0; x < 20; x++ {
            // 检查是否在边框内
            cell := buf.GetContent(x, y)
            // 验证...
        }
    }
}

type panelConfig struct {
    outerWidth  int
    outerHeight int
    borderStyle layout.BorderStyle
    content     rtui.VNode
}

func createPanelVNode(config panelConfig) rtui.VNode {
    builder := panel.NewBuilder().
        SetWidth(config.outerWidth).
        SetHeight(config.outerHeight)

    if config.borderStyle != layout.BorderNone {
        builder.SetBorderStyle(config.borderStyle)
    }

    if config.content != nil {
        builder.Content(config.content)
    }

    return builder.Build()
}
```

### 4.2 约束传播链测试

```go
// 文件: ui/layout/constraint_chain_test.go

package layout_test

import (
    "testing"
    "strings"

    "github.com/wwsheng009/mint/ui/layout/constraints"
    "github.com/wwsheng009/mint/ui/components/border"
    newstack "github.com/wwsheng009/mint/ui/components/stack"
    newtext "github.com/wwsheng009/mint/ui/components/text"
)

func TestConstraintPropagation_Chain(t *testing.T) {
    // 启用约束追踪
    constraints.Enable()
    defer constraints.Disable()

    // 构建复杂布局
    hstack := createComplexLayout()

    // Measure
    size := hstack.Measure(layout.Constraints{
        MinWidth:  0,
        MaxWidth:  60,
        MinHeight: 0,
        MaxHeight: 100,
    })

    // 验证尺寸
    if size.Width != 60 {
        t.Errorf("Expected width 60, got %d", size.Width)
    }

    // 验证约束传播链
    trace := constraints.Dump()
    verifyConstraintPropagation(t, trace)
}

func createComplexLayout() rtui.VNode {
    panel1 := panel.NewBuilder().
        SetWidth(15).
        SetHeight(5).
        Title("Fixed").
        Content(newtext.New("Content")).
        Build()

    panel2 := panel.NewBuilder().
        SetWidth(20).
        Title("Auto").
        Content(newtext.New("Auto content that wraps").SetWrap(true)).
        Build()

    panel3 := panel.NewBuilder().
        SetWidth(15).
        Title("Footer").
        Content(newtext.New("End")).
        Build()

    return newstack.New(newstack.Row).
        SetWidth(60).
        SetGap(0).
        SetChildrenList([]rtui.VNode{panel1, panel2, panel3})
}

func verifyConstraintPropagation(t *testing.T, trace string) {
    // 验证约束传递的关键步骤
    requiredSteps := []struct {
        from        string
        to          string
        expectedKey string
    }{
        {"hstack", "border", "panel1"},
        {"border", "vstack", "panel1"},
        {"vstack", "text", "panel1"},
        {"hstack", "border", "panel2"},
        {"border", "vstack", "panel2"},
        {"vstack", "text", "panel2 (wrap)"},
    }

    for _, step := range requiredSteps {
        if !strings.Contains(trace, step.from) {
            t.Errorf("Constraint trace missing step from %q", step.from)
        }
    }
}
```

---

## 5. 边界测试

### 5.1 极端尺寸测试

```go
// 文件: ui/layout/boundary_test.go

package layout_test

import (
    "testing"

    "github.com/wwsheng009/mint/runtime/layout"
    "github.com/wwsheng009/mint/ui/components/panel"
    newtext "github.com/wwsheng009/mint/ui/components/text"
)

func TestBoundary_ZeroDimensions(t *testing.T) {
    // 测试零尺寸
    panelNode := panel.New().
        SetWidth(0).
        SetHeight(0).
        SetContent(newtext.New(""))

    inst := panelNode.CreateInstance()
    size := inst.Measure(layout.Constraints{
        MaxWidth:  50,
        MaxHeight: 50,
    })

    // 应该使用父约束
    if size.Width <= 0 || size.Height <= 0 {
        t.Errorf("Expected positive size, got %+v", size)
    }
}

func TestBoundary_VeryLargeDimensions(t *testing.T) {
    // 测试超大尺寸（可能导致溢出）
    largeWidth := 1000000
    panelNode := panel.New().
        SetWidth(largeWidth).
        SetHeight(10).
        SetContent(newtext.New(""))

    inst := panelNode.CreateInstance()
    size := inst.Measure(layout.Constraints{
        MaxWidth:  largeWidth,
        MaxHeight: 1000000,
    })

    // 应该不会溢出
    if size.Width > largeWidth || size.Height > 1000000 {
        t.Errorf("Size too large: %+v", size)
    }
}

func TestBoundary_NegativeDimensions(t *testing.T) {
    // 测试负尺寸
    panelNode := panel.New().
        SetWidth(-10).
        SetHeight(-5).
        SetContent(newtext.New(""))

    inst := panelNode.CreateInstance()
    size := inst.Measure(layout.Constraints{})

    // 应该被正确处理（归零）
    if size.Width < 0 || size.Height < 0 {
        t.Errorf("Negative dimensions not handled correctly: %+v", size)
    }
}

func TestBoundary_MinGreaterThanMax(t *testing.T) {
    // 测试 Min > Max 的约束
    panelNode := panel.New().
        SetWidth(20).
        SetHeight(10).
        SetContent(newtext.New(""))

    constraints := layout.Constraints{
        MinWidth:  30,  // > MaxWidth
        MaxWidth:  20,
    }

    inst := panelNode.CreateInstance()
    size := inst.Measure(constraints)

    // 应该使用显式维度，不崩溃
    if size.Width <= 0 {
        t.Errorf("Invalid size constraint handling: %+v", size)
    }
}
```

### 5.2 空内容和极长内容测试

```go
func TestBoundary_EmptyContent(t *testing.T) {
    vnode := panel.New().
        SetWidth(20).
        SetHeight(0).  // auto
        Content(newtext.New(""))

    inst := vnode.CreateInstance()
    size := inst.Measure(layout.Constraints{
        MaxWidth:  50,
        MaxHeight: 50,
    })

    // 空内容应该有小高度（最小 1 行）
    if size.Height < 2 {  // 最小：1 行内容 + 边框
        t.Errorf("Expected minimum height >= 2, got %d", size.Height)
    }
}

func TestBoundary_VeryLongContent(t *testing.T) {
    // 生成超长文本（1000 字符）
    longText := strings.Repeat("This is a very long text. ", 100)

    vnode := panel.New().
        SetContentWidth(18).
        WithAutoHeight().
        Content(newtext.New(longText).SetWrap(true))

    inst := vnode.CreateInstance()
    size := inst.Measure(layout.Constraints{
        MaxWidth:  100,
        MaxHeight: 1000,
    })

    // 高度应该合理（不会溢出）
    if size.Height > 1000 {
        t.Errorf("Height too large: %d", size.Height)
    }

    // 验证渲染不溢出
    buf := paint.NewBuffer(20, size.Height)
    ctx := panel.PaintContext{
        Bounds: paint.Rect{0, 0, 20, size.Height},
    }

    // 应该不会 panic
    inst.Paint(ctx, buf)
}

func TestBoundary_SpecialCharacters(t *testing.T) {
    specialContent := "日本語한글العربية🎉🔥"

    vnode := panel.New().
        SetContentWidth(18).
        WithAutoHeight().
        Content(newtext.New(specialContent).SetWrap(true))

    inst := vnode.CreateInstance()
    size := inst.Measure(layout.Constraints{
        MaxWidth:  100,
        MaxHeight: 100,
    })

    // 应该正常处理 Unicode 字符
    if size.Height <= 0 {
        t.Errorf("Failed to handle special characters: height=%d", size.Height)
    }
}
```

### 5.3 嵌套深度测试

```go
func TestBoundary_DeepNesting(t *testing.T) {
    // 构建深层嵌套（10 层）
    var vnode rtui.VNode = newtext.New("Content")

    for i := 0; i < 10; i++ {
        vnode = panel.New().
            SetWidth(20).
            WithAutoHeight().
            Content(vnode)
    }

    inst := vnode.CreateInstance()

    // 应该不会栈溢出
    size := inst.Measure(layout.Constraints{
        MaxWidth:  100,
        MaxHeight: 100,
    })

    if size.Height <= 0 {
        t.Errorf("Deep nesting failed: height=%d", size.Height)
    }
}

func TestBoundary_WideNesting(t *testing.T) {
    // 构建宽嵌套（100 个 Panel）
    var children []rtui.VNode
    for i := 0; i < 100; i++ {
        children = append(children, panel.New().
            SetWidth(10).
            SetHeight(3).
            Content(newtext.New(strconv.Itoa(i))))
    }

    vnode := newstack.New(newstack.Row).
        SetWidth(1000).
        SetGap(0).
        SetChildrenList(children)

    inst := vnode.CreateInstance()

    // 应该正常处理
    size := inst.Measure(layout.Constraints{
        MaxWidth:  2000,
        MaxHeight: 100,
    })

    if size.Width < 1000 {
        t.Errorf("Wide nesting failed: width=%d", size.Width)
    }
}
```

---

## 6. 性能测试

### 6.1 Measure 性能测试

```go
// 文件: ui/layout/performance_test.go

package layout_test

import (
    "testing"

    "github.com/wwsheng009/mint/runtime/layout"
    "github.com/wwsheng009/mint/ui/components/panel"
    newstack "github.com/wwsheng009/mint/ui/components/stack"
    newtext "github.com/wwsheng009/mint/ui/components/text"
)

func BenchmarkPanel_SimpleMeasure(b *testing.B) {
    vnode := panel.New().
        SetWidth(20).
        SetHeight(6).
        Content(newtext.New("Test content"))

    inst := vnode.CreateInstance()
    constraints := layout.Constraints{
        MaxWidth:  50,
        MaxHeight: 100,
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        inst.Measure(constraints)
    }
}

func BenchmarkPanel_NestedMeasure(b *testing.B) {
    vnode := createNestedPanel(5)  // 5 层嵌套
    inst := vnode.CreateInstance()
    constraints := layout.Constraints{
        MaxWidth:  100,
        MaxHeight: 100,
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        inst.Measure(constraints)
    }
}

func BenchmarkHStack_Measure(b *testing.B) {
    vnode := createHStack(10)  // 10 个子元素
    inst := vnode.CreateInstance()
    constraints := layout.Constraints{
        MaxWidth:  200,
        MaxHeight: 100,
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        inst.Measure(constraints)
    }
}

func BenchmarkText_WrapMeasure(b *testing.B) {
    inst := text.NewInstance()
    inst.SetContent("This is a very long text that should wrap to multiple lines")
    inst.SetWrap(true)

    constraints := layout.Constraints{
        MaxWidth:  18,
        MaxHeight: 100,
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        inst.Measure(constraints)
    }
}

// Measure 缓存性能测试
func BenchmarkMeasureCache_WithCache(b *testing.B) {
    cache := NewMeasureCache()
    vnode := panel.New().
        SetWidth(20).
        SetHeight(6).
        Content(newtext.New("Content"))

    inst := vnode.CreateInstance()
    constraints := layout.Constraints{
        MaxWidth:  50,
        MaxHeight: 100,
    }

    // 第一次：无缓存
    inst.Measure(constraints)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        // 后续：使用缓存
        if size, ok := cache.Get(vnode, constraints); ok {
            // 使用缓存
            _ = size
        } else {
            size := inst.Measure(constraints)
            cache.Put(vnode, constraints, size)
        }
    }
}

func BenchmarkMeasureCache_WithoutCache(b *testing.B) {
    vnode := panel.New().
        SetWidth(20).
        SetHeight(6).
        Content(newtext.New("Content"))

    inst := vnode.CreateInstance()
    constraints := layout.Constraints{
        MaxWidth:  50,
        MaxHeight: 100,
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        inst.Measure(constraints)
    }
}

// 辅助函数
func createNestedPanel(depth int) rtui.VNode {
    vnode := panel.New().
        SetWidth(20).
        WithAutoHeight().
        Content(newtext.New("Nested"))

    for i := 0; i < depth-1; i++ {
        vnode = panel.New().
            SetWidth(20).
            WithAutoHeight().
            Content(vnode)
    }

    return vnode
}

func createHStack(count int) rtui.VNode {
    var children []rtui.VNode
    for i := 0; i < count; i++ {
        children = append(children, panel.New().
            SetWidth(10).
            SetHeight(3).
            Content(newtext.New(strconv.Itoa(i))))
    }

    return newstack.New(newstack.Row).
        SetWidth(count * 10).
        SetGap(0).
        SetChildrenList(children)
}
```

### 6.2 性能基准测试

```go
func TestPerformance_Baseline(t *testing.T) {
    // 建立性能基线
    results := runPerformanceTests()

    // 与历史基线比较
    baseline := loadBaselineMetrics()

    if results.MeasureTime > baseline.MeasureTime*1.05 {
        t.Errorf("Measure time degraded: %v vs baseline %v",
            results.MeasureTime, baseline.MeasureTime)
    }

    if results.PaintTime > baseline.PaintTime*1.05 {
        t.Errorf("Paint time degraded: %v vs baseline %v",
            results.PaintTime, baseline.PaintTime)
    }
}

func runPerformanceTests() PerformanceMetrics {
    return PerformanceMetrics{
        MeasureTime: benchmarkMeasure(),
        PaintTime:   benchmarkPaint(),
        MemoryUsage: benchmarkMemory(),
    }
}

type PerformanceMetrics struct {
    MeasureTime time.Duration
    PaintTime   time.Duration
    MemoryUsage int64
}
```

---

## 7. 测试工具

### 7.1 约束追踪比较工具

```go
// 文件: ui/layout/testing/constraint_comparator.go

package testing

import (
    "fmt"
    "testing"

    "github.com/wwsheng009/mint/runtime/layout"
    "github.com/wwsheng009/mint/ui/layout/constraints"
)

type ConstraintComparator struct {
    expected []constraints.Entry
    actual   []constraints.Entry
}

func NewConstraintComparator() *ConstraintComparator {
    return &ConstraintComparator{}
}

func (cc *ConstraintComparator) CaptureExpected() {
    // 捕获预期的约束传播
    constraints.Clear()
    // 执行布局...
    cc.expected = constraints.Entries()
}

func (cc *ConstraintComparator) CaptureActual() {
    // 捕获实际的约束传播
    constraints.Clear()
    // 执行布局...
    cc.actual = constraints.Entries()
}

func (cc *ConstraintComparator) Compare(t *testing.T) {
    if len(cc.expected) != len(cc.actual) {
        t.Errorf("Constraint trace length mismatch: expected %d, got %d",
            len(cc.expected), len(cc.actual))
        return
    }

    for i, exp := range cc.expected {
        act := cc.actual[i]
        cc.compareEntry(t, i, exp, act)
    }
}

func (cc *ConstraintComparator) compareEntry(t *testing.T, index int, exp, act constraints.Entry) {
    if exp.From != act.From || exp.To != act.To {
        t.Errorf("Step %d: mismatched nodes\n  Expected: %s → %s\n  Got:      %s → %s",
            index, exp.From, exp.To, act.From, act.To)
    }

    if exp.Input != act.Input {
        t.Errorf("Step %d: mismatched input constraints\n  Expected: %s\n  Got:      %s",
            index, exp.Input, act.Input)
    }

    if exp.Output != act.Output {
        t.Errorf("Step %d: mismatched output constraints\n  Expected: %s\n  Got:      %s",
            index, exp.Output, act.Output)
    }
}

func (cc *ConstraintComparator) PrintDiff() string {
    var buf strings.Builder

    for i := 0; i < max(len(cc.expected), len(cc.actual)); i++ {
        if i < len(cc.expected) && i < len(cc.actual) {
            exp := cc.expected[i]
            act := cc.actual[i]

            if exp == act {
                buf.WriteString(fmt.Sprintf("%d: ✓ %s → %s\n", i, exp.From, exp.To))
            } else {
                buf.WriteString(fmt.Sprintf("%d: ✗ DIFF\n", i))
                buf.WriteString(fmt.Sprintf("    Expected: %s → %s\n", exp.From, exp.To))
                buf.WriteString(fmt.Sprintf("    Got:      %s → %s\n", act.From, act.To))
            }
        } else if i < len(cc.expected) {
            buf.WriteString(fmt.Sprintf("%d: ✗ MISSING in actual\n", i))
        } else {
            buf.WriteString(fmt.Sprintf("%d: ✗ EXTRA in actual\n", i))
        }
    }

    return buf.String()
}
```

### 7.2 布局快照测试工具

```go
// 文件: ui/layout/testing/snapshot.go

package testing

import (
    "encoding/json"
    "testing"

    "github.com/wwsheng009/mint/runtime/layout"
    rtui "github.com/wwsheng009/mint/runtime/ui"
)

type LayoutSnapshot struct {
    Tree       NodeSnapshot      `json:"tree"`
    Constraints ConstraintMap    `json:"constraints"`
}

type NodeSnapshot struct {
    ID       string       `json:"id"`
    Tag      string       `json:"tag"`
    Bounds   layout.Rect  `json:"bounds"`
    Size     layout.Size  `json:"size"`
    Children []string     `json:"children,omitempty"`
}

type ConstraintMap map[string]layout.Constraints

func CaptureLayoutSnapshot(root rtui.VNode, parentConstraints layout.Constraints) (*LayoutSnapshot, error) {
    snapshot := &LayoutSnapshot{
        Constraints: make(ConstraintMap),
    }

    // 遍历布局树，捕获每个节点的状态
    err := captureNode(root, "", parentConstraints, snapshot)
    if err != nil {
        return nil, err
    }

    return snapshot, nil
}

func captureNode(node rtui.VNode, path string, constraints layout.Constraints, snapshot *LayoutSnapshot) error {
    nodeID := path + "/" + node.Tag()

    // 捕获节点信息
    snapshot.Tree = NodeSnapshot{
        ID:     nodeID,
        Tag:    node.Tag(),
        Bounds: layout.Rect{},  // 从 Layout 获取
        Size:   layout.Size{},  // 从 Measure 获取
    }

    // 记录约束
    snapshot.Constraints[nodeID] = constraints

    // 递归处理子节点
    children := node.Children()
    snapshot.Tree.Children = make([]string, len(children))

    for i, child := range children {
        childID := nodeID + "/" + strconv.Itoa(i)
        snapshot.Tree.Children[i] = childID

        childConstraints := computeChildConstraints(node, child, constraints)
        err := captureNode(child, childID, childConstraints, snapshot)
        if err != nil {
            return err
        }
    }

    return nil
}

func CompareSnapshots(t *testing.T, expectedPath, actualPath string, actual *LayoutSnapshot) {
    expected, err := LoadSnapshot(expectedPath)
    if err != nil {
        t.Fatalf("Failed to load expected snapshot: %v", err)
    }

    // 比较树结构
    if !compareNodes(t, expected.Tree, actual.Tree) {
        t.Errorf("Layout tree mismatch")
    }

    // 比较约束
    for id, expectedConstraints := range expected.Constraints {
        actualConstraints, ok := actual.Constraints[id]
        if !ok {
            t.Errorf("Missing constraints for node %q", id)
            continue
        }

        if expectedConstraints != actualConstraints {
            t.Errorf("Constraints mismatch for node %q:\n  Expected: %+v\n  Got:      %+v",
                id, expectedConstraints, actualConstraints)
        }
    }

    // 保存实际快照
    err = SaveSnapshot(actualPath, actual)
    if err != nil {
        t.Errorf("Failed to save actual snapshot: %v", err)
    }
}

func LoadSnapshot(path string) (*LayoutSnapshot, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    var snapshot LayoutSnapshot
    err = json.Unmarshal(data, &snapshot)
    if err != nil {
        return nil, err
    }

    return &snapshot, nil
}

func SaveSnapshot(path string, snapshot *LayoutSnapshot) error {
    data, err := json.MarshalIndent(snapshot, "", "  ")
    if err != nil {
        return err
    }

    return os.WriteFile(path, data, 0644)
}

func compareNodes(t *testing.T, expected, actual NodeSnapshot) bool {
    matched := true

    if expected.Tag != actual.Tag {
        t.Errorf("Node tag mismatch: expected %q, got %q", expected.Tag, actual.Tag)
        matched = false
    }

    if expected.Size != actual.Size {
        t.Errorf("Node size mismatch: expected %+v, got %+v", expected.Size, actual.Size)
        matched = false
    }

    return matched
}
```

---

## 8. 测试执行计划

### 8.1 持续集成

#### GitHub Actions 配置

```yaml
# .github/workflows/layout-tests.yml

name: Layout System Tests

on:
  push:
    branches: [main, feature/*]
    paths:
      - 'ui/components/panel/**'
      - 'ui/components/border/**'
      - 'ui/components/text/**'
      - 'ui/components/stack/**'
      - 'ui/layout/**'
  pull_request:
    paths:
      - 'ui/components/panel/**'
      - 'ui/components/border/**'
      - 'ui/components/text/**'
      - 'ui/components/stack/**'
      - 'ui/layout/**'

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run unit tests
        run: |
          go test -v -race -coverprofile=coverage.out \
            ./ui/components/panel/... \
            ./ui/components/border/... \
            ./ui/components/text/... \
            ./ui/components/stack/... \
            ./ui/layout/...

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out

  integration-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run integration tests
        run: go test -v ./ui/layout/integration/...

  performance-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run benchmarks
        run: go test -bench=. -benchmem ./ui/layout/performance/...

      - name: Compare with baseline
        run: go run ./tools/benchcmp/main.go baseline.json new.json
```

### 8.2 测试执行流程

```
开发阶段
  ↓
单元测试 (go test ./ui/components/...)
  ↓
集成测试 (go test ./ui/layout/integration/...)
  ↓
性能测试 (go test -bench=. ./ui/layout/performance/...)
  ↓
提交到 CI
  ↓
所有测试通过 → 合并
```

### 8.3 测试报告

#### 测试覆盖率报告

```bash
# 生成覆盖率报告
go test -coverprofile=coverage.out ./ui/components/...
go tool cover -html=coverage.out -o coverage.html
```

#### 性能报告

```bash
# 运行基准测试
go test -bench=. -benchmem -benchtime=10s ./ui/layout/performance/... > bench.txt

# 与基线比较
benchcmp old.txt new.txt > diff.txt
```

---

## 附录

### A. 测试检查清单

#### 单元测试检查清单
- [ ] Border 约束传播测试
- [ ] Border 维度计算测试
- [ ] Text.Wrap 测试
- [ ] Text 高度约束测试
- [ ] VStack/HStack 布局测试
- [ ] Panel 组合测试

#### 集成测试检查清单
- [ ] Panel in HStack 测试
- [ ] Panel in VStack 测试
- [ ] 多层嵌套布局测试
- [ ] 约束传播链测试

#### 边界测试检查清单
- [ ] 零尺寸测试
- [ ] 极大尺寸测试
- [ ] 负尺寸测试
- [ ] 空内容测试
- [ ] 极长内容测试
- [ ] 特殊字符测试

#### 性能测试检查清单
- [ ] Measure 性能测试
- [ ] Paint 性能测试
- [ ] 缓存效率测试
- [ ] 增量布局测试

### B. 测试数据模板

```go
// 测试数据收集模板
type TestResult struct {
    TestName      string
    Component     string
    Constraints   layout.Constraints
    ExpectedSize  layout.Size
    ActualSize    layout.Size
    Passed        bool
    Duration      time.Duration
    Notes         string
}
```

---

**文档版本**: 1.0
**最后更新**: 2026-02-21
**作者**: Qwen Code
