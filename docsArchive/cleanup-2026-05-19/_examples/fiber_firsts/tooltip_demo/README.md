# Tooltip & Toast Component Demo

Tooltip 和 Toast 组件演示程序。

## 功能

展示 Toast 组件的多种功能和样式：
- 默认 Toast（Info）
- 成功 Toast（Success）
- 警告 Toast（Warning）
- 错误 Toast（Error）
- 不同层级的 Toast（Modal/Tooltip/Inspector）

## 编译运行

```bash
cd examples/fiber_firsts/tooltip_demo
go build -o main.exe
./main.exe
```

## 说明

- 使用 `ui/components/tooltip` 包的 Toast 组件
- 展示语义化样式（Info/Success/Warning/Error）
- 演示 5 个渲染层级的 Toast 效果
- Toast 自动定位，避免视觉重叠
