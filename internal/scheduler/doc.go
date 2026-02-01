// Package scheduler 提供 UI 调度器。
//
// 调度器负责:
//   - 管理更新任务的优先级
//   - 调度任务的执行时间
//   - 支持任务中断和恢复
//   - 实现时间切片
//
// 优先级级别:
//   - Immediate: 立即执行 (如用户输入)
//   - UserBlocking: 阻塞用户操作
//   - Normal: 正常优先级
//   - Low: 低优先级 (如数据加载)
//   - Idle: 空闲时执行
//
// ⚠️ 此包是内部实现，不对外暴露。
package scheduler
