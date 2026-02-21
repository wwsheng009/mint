// Package types provides common type definitions shared across all runtime packages.
// This package has zero dependencies on other runtime packages to avoid import cycles.
package types

// =============================================================================
// Layer Type - 统一的渲染层级定义
// =============================================================================

// Layer 表示视觉渲染层级，用于叠加组件
// 更高层级渲染在更低层级之上
type Layer int

const (
	// LayerBase 基础层 - 普通 UI 内容的默认层级
	LayerBase Layer = iota

	// LayerOverlay 覆盖层 - 下拉菜单、弹出框等
	LayerOverlay

	// LayerModal 模态层 - 需要用户关注的模态对话框
	LayerModal

	// LayerTooltip 提示层 - 工具提示和提示信息
	LayerTooltip

	// LayerInspector 检查器层 - UI 检查器调试覆盖层
	LayerInspector
)

// String 返回层级的字符串表示
func (l Layer) String() string {
	switch l {
	case LayerBase:
		return "base"
	case LayerOverlay:
		return "overlay"
	case LayerModal:
		return "modal"
	case LayerTooltip:
		return "tooltip"
	case LayerInspector:
		return "inspector"
	default:
		return "unknown"
	}
}

// ZIndex 返回此层级的 z-index 值（越高越在上层）
func (l Layer) ZIndex() int {
	return int(l)
}

// IsValid 检查层级值是否有效
func (l Layer) IsValid() bool {
	return l >= LayerBase && l <= LayerInspector
}

// IsModal 检查是否是模态层
func (l Layer) IsModal() bool {
	return l == LayerModal
}

// IsOverlay 检查是否是覆盖层（非基础层）
func (l Layer) IsOverlay() bool {
	return l >= LayerOverlay
}
