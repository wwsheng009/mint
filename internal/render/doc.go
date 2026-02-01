// Package render 提供渲染引擎。
//
// 渲染引擎负责:
//   - VNode 到 RNode 的转换
//   - 布局计算 (Layout Engine)
//   - 绘制命令的收集 (DrawCmd)
//   - 栅格化 (Rasterizer)
//
// 核心类型:
//   - RNode: 真实渲染节点
//   - LayoutEngine: 布局引擎
//   - DrawCmd: 绘制命令接口
//   - Rasterizer: 栅格化器
//   - DeclarativeNode: 可独立渲染的声明式节点
//
// ⚠️ 此包是内部实现，不对外暴露。
package render
