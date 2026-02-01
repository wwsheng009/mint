// Package reconciler 提供声明式 UI 的协调器系统。
//
// 协调器负责:
//   - 管理 VNode 树的创建和更新
//   - 执行 Diff 算法对比新旧树
//   - 管理 Fiber 节点的工作循环
//   - 调度组件的挂载、更新和卸载
//
// 核心类型:
//   - Reconciler: 协调器主结构
//   - Fiber: 工作单元节点
//   - WorkLoop: 工作循环
//
// ⚠️ 此包是内部实现，不对外暴露。
package reconciler
