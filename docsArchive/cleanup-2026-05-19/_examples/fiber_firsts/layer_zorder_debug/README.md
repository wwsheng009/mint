# Layer Z-Order Debug Demo

Layer Z-Order 调试演示，用于详细输出层级信息和渲染过程。

## 功能

- 显示 5 个渲染层级的详细定义
- 展示 Fiber-first 渲染流程的关键步骤
- 提供调试检查点和问题排查指南
- 支持层级调试输出模式

## 编译运行

```bash
cd examples/fiber_firsts/layer_zorder_debug
go build -o main.exe
./main.exe
```

## 环境变量

```bash
MINT_DEBUG_LAYER=true  # 启用层级调试输出
```

## 调试检查点

1. **渲染顺序**：检查 PaintPaintablePlanes() 中的循环顺序
2. **Buffer 合成**：验证低层 buffer 先绘制，高层 buffer 后绘制
3. **Modal 背景灰化**：检查 modal bounds 计算、灰化区域、中文字符处理
4. **中文字符宽度**：确认 EastAsianWidth = true

## 常见问题排查

- Modal 不显示：检查 Modal 层 buffer 和 layout
- 背景不灰化：检查 modalBox 的 bounds 属性
- 中文字符显示异常：检查 EastAsianWidth 和 continuation 跳过逻辑
- 层级错乱：检查 renderOrder 数组和 GetLayer() 返回值
