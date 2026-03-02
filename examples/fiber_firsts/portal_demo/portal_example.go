// Portal 实际使用示例 (Phase 3)
// 展示如何在应用中创建 PortalRoot 和 Portal 组件
package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// PortalUsageExample 展示完整的 Portal 使用流程
func PortalUsageExample() {
	/*
	=== PortalRoot 创建示例 ===

	// 1. 在应用顶层创建 PortalRoot
	// PortalRoot 是所有 Portal 的挂载目标
	portalRoot := rtui.VNode{
		Type: rtui.VNodeElement,
		Tag:  "div",  // 使用任意容器标签
		Props: rtui.Props{
			"portalRootId": "main-root",  // 标识该节点为 PortalRoot
			"width":       80,
			"height":      25,
		},
		Children: []rtui.VNode{
			// PortalRoot 可以包含默认内容，但在实际使用中通常是空的
		},
	}

	=== Portal 创建示例 ===

	// 2. 在组件树中创建 Portal 组件
	// Portal 的子组件将被渲染到 PortalRoot 中
	tooltip := rtui.VNode{
		Type: rtui.VNodeElement,
		Tag:  "div",
		Props: rtui.Props{
			"portalRoot": "main-root",  // 指定目标 PortalRoot
			"position":  "fixed",
			"anchor":    "bottom",
		},
		Children: []rtui.VNode{
			rtui.Text("This is a tooltip"),
		},
	}

	=== 完整应用示例 ===

	// 主应用组件
	renderApp := func(ctx *rtui.ComponentContext) rtui.VNode {
		return rtui.VNode{
			Type: rtui.VNodeElement,
			Tag:  "app",
			Children: []rtui.VNode{
				// PortalRoot (应用顶层)
				rtui.VNode{
					Type: rtui.VNodeElement,
					Tag:  "portal-root",
					Props: rtui.Props{
						"portalRootId": "main-root",
					},
				},

				// 主内容区域
				rtui.VNode{
					Type: rtui.VNodeElement,
					Tag:  "main-content",
					Children: []rtui.VNode{
						rtui.Text("Main Application Content"),
						// Tooltip Portal
						rtui.VNode{
							Type: rtui.VNodeElement,
							Tag:  "div",
							Props: rtui.Props{
								"portalRoot": "main-root",
							},
							Children: []rtui.VNode{
								rtui.Text("Tooltip"),
							},
						},
					},
				},
			},
		}
	}
	*/

	// 注意：实际渲染流程
	// 1. Render 阶段：VNode 转换为 Fiber
	// 2. Commit 阶段：Reconciler.linkPortalsToRoots() 建立链接
	// 3. Layout 阶段：使用 Fiber.PortalRoot 进行布局计算
	// 4. Paint 阶段：在 PortalRoot 的位置绘制 Portal 内容
}

// TestPortalRootAndPortal_Simple 简单的 PortalRoot + Portal 示例
func TestPortalRootAndPortal_Simple(t *testing.T) {
	// 创建 PortalRoot
	portalRoot := &rtui.Fiber{
		NodeID: 100,
		Key:    "main-portal-root",
		Type:   rtui.VNodeElement,
		Props: rtui.Props{
			"portalRootId": "main-root",
			"width":       80,
			"height":      25,
		},
	}

	// 创建 Portal
	portal := &rtui.Fiber{
		NodeID: 200,
		Key:    "tooltip-portal",
		Type:   rtui.VNodeElement,
		Props: rtui.Props{
			"portalRoot": "main-root",
			"position":  "fixed",
		},
		Child: &rtui.Fiber{
			NodeID: 201,
			Key:    "tooltip-text",
			Type:   rtui.VNodeText,
			Props: rtui.Props{
				"content": "Hello from Portal!",
			},
		},
	}

	// 在 commit 阶段，linkPortalsToRoots 会建立链接
	// portal.PortalRoot = portalRoot

	// 验证 PortalRoot 标识
	portalRootID, ok := portalRoot.Props["portalRootId"].(string)
	assert.True(t, ok)
	assert.Equal(t, "main-root", portalRootID)

	// 验证 Portal 目标
	portalRootRef, ok := portal.Props["portalRoot"].(string)
	assert.True(t, ok)
	assert.Equal(t, "main-root", portalRootRef)

	t.Log("PortalRoot and Portal created successfully")
}

// TestMultiplePortals 多 Portal 示例
func TestMultiplePortals(t *testing.T) {
	// 创建多个 PortalRoot
	// 这些 PortalRoot 通常放置在应用的顶层

	// Main PortalRoot (用于 Tooltip, Toast 等)
	mainRoot := &rtui.Fiber{
		NodeID: 10,
		Key:    "main-root",
		Type:   rtui.VNodeElement,
		Props: rtui.Props{
			"portalRootId": "main",
			"layer":       1,  // Overlay Layer
		},
	}

	// Modal PortalRoot (用于 Modal, Dialog 等)
	modalRoot := &rtui.Fiber{
		NodeID: 20,
		Key:    "modal-root",
		Type:   rtui.VNodeElement,
		Props: rtui.Props{
			"portalRootId": "modal",
			"layer":       2,  // Modal Layer
		},
	}

	// 创建多个 Portal 组件
	tooltip := &rtui.Fiber{
		NodeID: 100,
		Key:    "tooltip",
		Type:   rtui.VNodeElement,
		Props: rtui.Props{
			"portalRoot": "main",
		},
	}

	toast := &rtui.Fiber{
		NodeID: 101,
		Key:    "toast",
		Type:   rtui.VNodeElement,
		Props: rtui.Props{
			"portalRoot": "main",
		},
	}

	modal := &rtui.Fiber{
		NodeID: 200,
		Key:    "modal",
		Type:   rtui.VNodeElement,
		Props: rtui.Props{
			"portalRoot": "modal",
		},
	}

	// 构建树结构
	root := &rtui.Fiber{
		NodeID: 1,
		Key:    "root",
		Type:   rtui.VNodeElement,
	}

	root.Child = mainRoot
	mainRoot.Sibling = modalRoot

	mainRoot.Child = tooltip
	tooltip.Sibling = toast

	modalRoot.Child = modal

	// 模拟 PortalRoot 收集和链接
	// (实际由 Reconciler.linkPortalsToRoots 在 commit 阶段完成)
	portalRoots := map[string]*rtui.Fiber{
		"main":  mainRoot,
		"modal": modalRoot,
	}

	// 链接 Portal
	if target, exists := portalRoots[tooltip.Props["portalRoot"].(string)]; exists {
		tooltip.PortalRoot = target
	}

	if target, exists := portalRoots[toast.Props["portalRoot"].(string)]; exists {
		toast.PortalRoot = target
	}

	if target, exists := portalRoots[modal.Props["portalRoot"].(string)]; exists {
		modal.PortalRoot = target
	}

	// 验证链接
	assert.Equal(t, mainRoot.NodeID, tooltip.PortalRoot.NodeID)
	assert.Equal(t, mainRoot.NodeID, toast.PortalRoot.NodeID)
	assert.Equal(t, modalRoot.NodeID, modal.PortalRoot.NodeID)

	t.Log("Multiple portals successfully linked to their roots")
}

// TestPortalInRealScenario 真实场景示例
func TestPortalInRealScenario(t *testing.T) {
	/*
	真实场景：一个带有 Tooltip 的按钮组件

	=== 组件代码 ===

	type ButtonWithTooltip struct {
		ComponentBase
		Text        string
		TooltipText string
	}

	renderButtonWithTooltip := func(ctx *context) VNode {
		button := ButtonWithTooltip{...}

		return VNode{
			Tag: "button",
			Props: Props{
				"onEnter": func(e Event) {
					// Show tooltip
				},
				"onLeave": func(e Event) {
					// Hide tooltip
				},
			},
			Children: []VNode{
				Text(button.Text),
				// Tooltip Portal
				VNode{
					Tag: "div",
					Props: Props{
						"portalRoot": "main",
						"position":  "fixed",
						"anchor":    "bottom",
					},
					Children: []VNode{
						Text(button.TooltipText),
					},
				},
			},
		}
	}

	=== 应用顶层 ===

renderApp := func(ctx *context) VNode {
	return VNode{
		Tag: "app",
		Children: []VNode{
			// PortalRoot
			VNode{
				Tag: "portal-root",
				Props: Props{
					"portalRootId": "main",
				},
			},
			// 按钮
			ButtonWithTooltip{
				Text:        "Click me",
				TooltipText: "This is a tooltip!",
			},
		},
	}
}
	*/

	// 测试验证结构
	// 创建 Fiber 树表示
	app := &rtui.Fiber{
		NodeID: 1,
		Key:    "app",
		Type:   rtui.VNodeElement,
	}

	portalRoot := &rtui.Fiber{
		NodeID: 2,
		Key:    "portal-root",
		Type:   rtui.VNodeElement,
		Props: rtui.Props{
			"portalRootId": "main",
		},
	}

	button := &rtui.Fiber{
		NodeID: 3,
		Key:    "button",
		Type:   rtui.VNodeElement,
	}

	tooltip := &rtui.Fiber{
		NodeID: 4,
		Key:    "tooltip",
		Type:   rtui.VNodeElement,
		Props: rtui.Props{
			"portalRoot": "main",
		},
	}

	// 构建树
	app.Child = portalRoot
	portalRoot.Sibling = button
	button.Child = tooltip

	// 链接 (模拟 commit 阶段)
	portalRoots := map[string]*rtui.Fiber{"main": portalRoot}
	if target, exists := portalRoots[tooltip.Props["portalRoot"].(string)]; exists {
		tooltip.PortalRoot = target
	}

	// 验证
	assert.NotNil(t, tooltip.PortalRoot)
	assert.Equal(t, portalRoot.NodeID, tooltip.PortalRoot.NodeID)

	t.Log("Real scenario test passed")
}
