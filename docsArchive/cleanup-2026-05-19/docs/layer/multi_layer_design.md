# Multi-Layer Layout Design

## 架构概述

### 当前实现（需要修改）

```
Single Fiber Tree → Layout → Single LayoutResult
                          ↓
                    LayerManager (归一化 - 错误!)
                          ↓
                    PaintableBoxes (坐标已破坏)
```

**问题：**
1. LayerManager直接修改LayoutBox坐标，破坏原始位置
2. 所有layer共享一个布局树，无法实现独立的overlay行为
3. Modal等overlay元素的位置过于依赖base层布局

### 正确的设计

```
Multiple Fiber Trees (one per layer → Layout (independent engine instances or shared)
      ↓                      ↓                    ↓
   LayerBase          LayerModal           LayerTooltip
      ↓                      ↓                    ↓
LayoutResult 1      LayoutResult 2       LayoutResult 3
      ↓                      ↓                    ↓
   Converted             Converted            Converted
      ↓                      ↓                    ↓
PaintableBox 1      PaintableBox 2       PaintableBox 3
      ↓                      ↓                    ↓
   planes[0]            planes[2]            planes[3]
                                      ↓
                            PaintEngine (按Z-order渲染)
                                      ↓
                            Final Buffer
```

## 具体实现

### 1. Fiber树按Layer分离

```go
// 在渲染开始前，将Fiber树按Layer分组
func (n *DeclarativeNode) collectFibersByLayer(root *reconciler.Fiber) map[Layer]*reconciler.Fiber {
    layers := make(map[Layer]*reconciler.Fiber)
    
    var collector func(fiber *reconciler.Fiber, layer Layer)
    collector = func(fiber *reconciler.Fiber, layer Layer) {
        if fiber == nil {
            return
        }
        
        // 当前fiber的layer
        fiberLayer := layer
        
        // 如果fiber自己定义了layer，使用fiber的layer
        // 否则继承父层的layer
        if fiber.Tag == "modal" {
            fiberLayer = LayerModal
        } else if fiber.Tag == "tooltip" {
            fiberLayer = LayerTooltip
        } else if fiber.Tag == "inspector" {
            fiberLayer = LayerInspector
        }
        
        // 添加到对应的layer
        if fiberLayer == layer {
            // Base layer保持原树结构
            if layers[layer] == nil {
                layers[layer] = fiber
            }
        }
        
        // 递归处理子节点
        for child := fiber.Child; child != nil; child = child.Sibling {
            collector(child, fiberLayer)
        }
    }
    
    return layers
}
```

### 2. 每个Layer独立布局

```go
// 为每个layer独立执行布局
func (n *DeclarativeNode) layoutLayers(
    fibersByLayer map[Layer]*reconciler.Fiber,
    constraints layout.Constraints,
) map[Layer]*layout.LayoutResult {
    results := make(map[Layer]*layout.LayoutResult)
    
    for layer, fiberRoot := range fibersByLayer {
        if fiberRoot == nil {
            continue
        }
        
        // 每个layer使用相同的布局引擎，但独立计算
        // 也可以为不同layer使用不同的布局策略
        result := n.newLayoutEngine.LayoutFiber(fiberRoot, constraints)
        results[layer] = result
    }
    
    return results
}
```

### 3. 转换为PaintablePlanes

```go
// 将每个layer的LayoutResult转换为PaintableBox，并添加到planes
func (n *DeclarativeNode) convertToPaintablePlanes(
    layoutResults map[Layer]*layout.LayoutResult,
    fiberRoots map[Layer]*reconciler.Fiber,
) *paint.PaintablePlanes {
    planes := paint.NewPaintablePlanes()
    
    for layer, layoutResult := range layoutResults {
        if layoutResult == nil || layoutResult.Root == nil {
            continue
        }
        
        // 获取该layer的Fiber根节点
        fiberRoot := fiberRoots[layer]
        converter := render.NewFiberToPaintableConverter(fiberRoot)
        
        // 转换为PaintableBox
        paintableLayout := converter.ConvertToLayout(layoutResult.Root)
        
        // 添加到对应的plane
        var addBox func(box *paint.PaintableBox)
        addBox = func(box *paint.PaintableBox) {
            if box == nil {
                return
            }
            planes.AddToLayer(paint.RenderLayer(layer), box)
            for _, child := range box.Children {
                addBox(child)
            }
        }
        addBox(paintableLayout.Root)
    }
    
    return planes
}
```

### 4. 渲染流程

```go
// fiberFirstPaint 修改后的版本
func (n *DeclarativeNode) fiberFirstPaint(
    ctx component.PaintContext,
    buf *paint.Buffer,
) {
    // ... 获取fiberRoot ...
    
    // 1. 按layer分组Fibers
    fibersByLayer := n.collectFibersByLayer(fiberRoot)
    
    // 2. 每个layer独立布局
    viewportConstraints := layout.Constraints{
        MaxWidth:  buf.Width,
        MaxHeight: buf.Height,
    }
    layoutResults := n.layoutLayers(fibersByLayer, viewportConstraints)
    
    // 3. 转换为PaintablePlanes
    planes := n.convertToPaintablePlanes(layoutResults, fibersByLayer)
    
    // 4. 渲染（按Z-order）
    if err := n.paintEngine.PaintPaintablePlanes(planes, buf); err != nil {
        log.RenderLogger.Debug("[Paint FAILED: %v", err)
        n.legacyPaint(ctx, buf)
        return
    }
}
```

## 优势

1. **独立性**：每个layer可以有自己的布局策略
2. **灵活性**：Modal可以相对于viewport定位，不依赖base层
3. **正确性**：没有坐标归一化问题，原始位置被保留
4. **可扩展**：新的layer类型可以有不同的布局规则

## 向后兼容

对于现有代码，如果所有组件都在一个VNode树中，可以：
1. 保持当前的单树结构
2. 移除LayerManager的坐标修改
3. 直接使用原始坐标渲染
4. 各layer共享同一个布局结果

这样的简单方案也能工作，只是每个layer的位置会有一定的依赖关系。
