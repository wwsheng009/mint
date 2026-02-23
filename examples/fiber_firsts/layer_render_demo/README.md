# Layer Rendering Demo

多层级渲染演示程序，展示 Fiber-first 渲染机制。

## 功能

- 展示 5 个渲染层级的正确渲染顺序（Z-Order）
- 演示 Fiber-first 分层渲染机制
- 说明 PaintPaintablePlanes() 的工作流程
- 展示 Modal 背景灰化机制

## 编译运行

```bash
cd examples/fiber_firsts/layer_render_demo
go build -o main.exe
./main.exe
```

## 技术说明

### Fiber-first 渲染流程

1. **ComputeLayout()** - 计算各层布局树
   - 每个层独立计算布局
   - 支持不同约束条件

2. **PaintPaintablePlanes()** - 分层绘制
   - 创建独立的 PaintableBuffer
   - 按 renderOrder 顺序绘制
   - Layer 0 → Layer 4

3. **Buffer 合成**
   - 各层 buffer 按顺序叠加
   - 高层覆盖低层内容
   - 保持透明区域可见

4. **Modal 背景灰化**
   - 检测 Modal 区域（Layer 2）
   - 灰化下方区域
   - 保留原有内容，仅改变颜色

### 关键技术点

- **分层渲染**：每层独立 buffer，避免相互干扰
- **Z-Order 控制**：通过 renderOrder 数组定义渲染顺序
- **优化策略**：只渲染有变化的层（脏标记）
