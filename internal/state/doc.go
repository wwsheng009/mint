// Package state 提供状态管理系统。
//
// 状态管理负责:
//   - 组件实例 (ComponentInstance) 的生命周期
//   - Hooks 状态的存储和检索
//   - 交互状态 (焦点、悬停等)
//   - 焦点管理
//
// 核心类型:
//   - InstanceManager: 实例管理器
//   - ComponentInstance: 组件实例
//   - InteractionState: 交互状态
//   - FocusManager: 焦点管理器
//
// ⚠️ 此包是内部实现，不对外暴露。
package state
